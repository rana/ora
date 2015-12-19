// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"io"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultPoolSize      = 4
	DefaultEvictDuration = time.Minute
)

// NewSrvPool returns a connection pool, which evicts the idle sessions in every minute.
// The pool holds at most size idle Srv.
// If size is zero, DefaultPoolSize will be used.
func (env *Env) NewSrvPool(srvCfg *SrvCfg, size int) *SrvPool {
	p := &SrvPool{
		env:    env,
		srv:    newIdlePool(size),
		srvCfg: srvCfg,
	}
	p.poolEvictor = &poolEvictor{Evict: p.srv.Evict}
	p.SetEvictDuration(DefaultEvictDuration)
	return p
}

type SrvPool struct {
	env    *Env
	srvCfg *SrvCfg
	srv    *idlePool

	*poolEvictor
}

func (p *SrvPool) Close() error {
	return p.srv.Close()
}

// Get a connection.
func (p *SrvPool) Get() (*Srv, error) {
	for {
		x := p.srv.Get()
		if x == nil { // the pool is empty
			break
		}
		return x.(*Srv), nil
	}
	return p.env.OpenSrv(p.srvCfg)
}

// Put the connection back to the idle pool.
func (p *SrvPool) Put(srv *Srv) {
	if srv == nil || !srv.IsOpen() {
		return
	}
	p.srv.Put(srv)
}

// NewSesPool returns a session pool, which evicts the idle sessions in every minute.
// The pool holds at most size idle Ses.
// If size is zero, DefaultPoolSize will be used.
func (srv *Srv) NewSesPool(sesCfg *SesCfg, size int) *SesPool {
	p := &SesPool{
		env:    srv.env,
		srv:    srv,
		sesCfg: sesCfg,
		ses:    newIdlePool(size),
	}
	p.poolEvictor = &poolEvictor{Evict: p.ses.Evict}
	p.SetEvictDuration(DefaultEvictDuration)
	return p
}

type SesPool struct {
	env    *Env
	srv    *Srv
	sesCfg *SesCfg
	ses    *idlePool

	*poolEvictor
}

func (p *SesPool) Close() error {
	return p.ses.Close()
}

// Get a session from an idle Srv.
func (p *SesPool) Get() (*Ses, error) {
	for {
		x := p.ses.Get()
		if x == nil { // the pool is empty
			break
		}
		ses := x.(*Ses)
		if err := ses.Ping(); err == nil {
			return ses, nil
		}
		ses.Close()
	}
	return p.srv.OpenSes(p.sesCfg)
}

// Put the session back to the session pool.
func (p *SesPool) Put(ses *Ses) {
	if ses == nil || !ses.IsOpen() {
		return
	}
	p.ses.Put(ses)
}

type poolEvictor struct {
	Evict func(time.Duration)

	sync.Mutex
	evictDurSec uint32 // evict duration, in seconds
	tickerCh    chan *time.Ticker
}

// Set the eviction duration to the given.
// Also starts eviction if not yet started.
func (p *poolEvictor) SetEvictDuration(dur time.Duration) {
	p.Lock()
	if p.tickerCh == nil { // first initialize
		p.tickerCh = make(chan *time.Ticker)
		go func(tickerCh <-chan *time.Ticker) {
			ticker := <-tickerCh
			for {
				select {
				case <-ticker.C:
					dur := time.Second * time.Duration(atomic.LoadUint32(&p.evictDurSec))
					p.Lock()
					evict := p.Evict
					p.Unlock()
					evict(dur)
				case nxt := <-tickerCh:
					ticker.Stop()
					ticker = nxt
				}
			}
		}(p.tickerCh)
	}
	p.Unlock()
	atomic.StoreUint32(&p.evictDurSec, uint32(dur/time.Second))
	p.tickerCh <- time.NewTicker(dur)
}

// SplitDSN splits the user/password@dblink string to username, password and dblink,
// to be used as SesCfg.Username, SesCfg.Password, SrvCfg.Dblink.
func SplitDSN(dsn string) (username, password, sid string) {
	if strings.HasPrefix(dsn, "/@") { // shortcut
		return "", "", dsn[2:]
	}
	if i := strings.LastIndex(dsn, "@"); i >= 0 {
		sid, dsn = dsn[i+1:], dsn[:i]
	}
	if i := strings.IndexByte(dsn, '/'); i >= 0 {
		username, password = dsn[:i], dsn[i+1:]
	}
	return
}

// idlePool is a pool of io.Closers.
// Each element will be Closed on eviction.
//
// The backing store is a simple []io.Closer, which is treated as random store,
// to achive uniform reuse.
type idlePool struct {
	elems []io.Closer
	times []time.Time

	sync.Mutex
}

// NewidlePool returns an idlePool.
func newIdlePool(size int) *idlePool {
	return &idlePool{
		elems: make([]io.Closer, size),
		times: make([]time.Time, size),
	}
}

// Evict evicts idle items idle for more than the given duration.
func (p *idlePool) Evict(dur time.Duration) {
	p.Lock()
	defer p.Unlock()
	deadline := time.Now().Add(-dur)
	for i, t := range p.times {
		e := p.elems[i]
		if e == nil || t.After(deadline) {
			continue
		}
		e.Close()
		p.elems[i] = nil
	}
	return
}

// Get returns a closer or nil, if no pool found.
func (p *idlePool) Get() io.Closer {
	p.Lock()
	defer p.Unlock()
	for i, c := range p.elems {
		if c == nil {
			continue
		}
		p.elems[i] = nil
		return c
	}
	return nil
}

// Put a new element into the store. The slot is chosen randomly.
// If no empty slot is found, one (random) is Close()-d and this new
// element is put there.
// This way elements reused uniformly.
func (p *idlePool) Put(c io.Closer) {
	p.Lock()
	defer p.Unlock()
	n := len(p.elems)
	if n == 0 {
		c.Close()
		return
	}
	now := time.Now()
	i0 := 0
	if n != 1 {
		i0 = rand.Intn(n)
	}
	for i := 0; i < n; i++ {
		j := (i0 + i) % n
		if p.elems[j] == nil {
			p.elems[j] = c
			p.times[j] = now
			return
		}
	}
	if p.elems[i0] != nil {
		p.elems[i0].Close()
	}
	p.elems[i0] = c
	p.times[i0] = now
}

// Close all elements.
func (p *idlePool) Close() error {
	p.Lock()
	defer p.Unlock()
	var err error
	for i, c := range p.elems {
		p.elems[i] = nil
		if c == nil {
			continue
		}
		if closeErr := c.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}
