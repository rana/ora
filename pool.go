// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	DefaultPoolSize      = 4
	DefaultEvictDuration = time.Minute
)

// NewSrvPool returns a connection pool, which evicts the idle connections in every minute.
// The pool holds at most size idle Srv/Ses.
// If size is zero, DefaultPoolSize will be used.
func (env *Env) NewSrvPool(dsn string, size int) *SrvPool {
	srvCfg := NewSrvCfg()
	sesCfg := NewSesCfg()
	if strings.IndexByte(dsn, '@') == -1 {
		srvCfg.Dblink = dsn
	} else {
		sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	}
	p := &SrvPool{
		env:    env,
		srvCfg: srvCfg,
		sesCfg: sesCfg,
		evict:  DefaultEvictDuration,
		srv:    newIdlePool(size),
		ses:    newIdlePool(size),
	}
	p.SetEvictDuration(p.evict)
	return p
}

type SrvPool struct {
	env    *Env
	srvCfg *SrvCfg
	sesCfg *SesCfg

	mu       sync.Mutex // protects following fields
	evict    time.Duration
	tickerCh chan *time.Ticker
	srv      *idlePool
	ses      *idlePool
}

type pooledSes struct {
	*Ses
	sync.Mutex
}

func (ps *pooledSes) Close() error {
	if ps == nil {
		return nil
	}
	ps.Lock()
	defer ps.Unlock()
	if ps == nil || ps.Ses == nil {
		return nil
	}
	ps.Ses.mu.Lock()
	srv := ps.Ses.srv
	ps.Ses.mu.Unlock()
	err := ps.Ses.Close()
	ps.Ses = nil

	if srv == nil {
		return err
	}
	if srv.NumSes() == 0 {
		srv.Close()
	}
	return err
}

// Close closes all held Srvs.
func (p *SrvPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.ses.Close()
	if srvErr := p.srv.Close(); srvErr != nil && err == nil {
		return srvErr
	}
	return err
}

// Get returns a new Srv from the pool, or creates a new if no idle is in the pool.
func (p *SrvPool) Get() (*Srv, error) {
	if x := p.srv.Get(); x != nil {
		return x.(*Srv), nil
	}
	return p.env.OpenSrv(p.srvCfg)
}

// Put the unneeded srv back into the pool.
func (p *SrvPool) Put(srv *Srv) {
	if srv != nil && srv.IsOpen() {
		p.srv.Put(srv)
	}
}

// Get a session from an idle Srv.
func (p *SrvPool) GetSes() (*Ses, error) {
	for {
		x := p.ses.Get()
		if x == nil { // the pool is empty
			break
		}
		ps := x.(*pooledSes)
		ses := ps.Ses
		if err := ses.Ping(); err == nil {
			return ses, nil
		}
		ps.Close()
	}
	srv, err := p.Get()
	if err != nil {
		return nil, err
	}
	return srv.OpenSes(p.sesCfg)
}

// Put the session back to the session pool.
// Also puts back the srv into the pool if has no used sessions.
func (p *SrvPool) PutSes(ses *Ses) {
	if ses == nil || !ses.IsOpen() {
		return
	}
	ses.mu.Lock()
	srv := ses.srv
	if srv != nil {
		srv.mu.Lock()
		if srv.openSess.len() == 1 { // this is the only session
			p.srv.Put(srv)
			srv.mu.Unlock()
			ses.mu.Unlock()
			ses.Close()
			return
		}
		srv.mu.Unlock()
	}
	ses.mu.Unlock()
	p.ses.Put(&pooledSes{Ses: ses})
}

// Set the eviction duration to the given.
// Also starts eviction if not yet started.
func (p *SrvPool) SetEvictDuration(dur time.Duration) {
	p.mu.Lock()
	if p.tickerCh == nil { // first initialize
		p.tickerCh = make(chan *time.Ticker)
		go func(tickerCh <-chan *time.Ticker) {
			ticker := <-tickerCh
			for {
				select {
				case <-ticker.C:
					p.mu.Lock()
					dur := p.evict
					p.mu.Unlock()
					p.ses.Evict(dur)
					p.srv.Evict(dur)
				case nxt := <-tickerCh:
					ticker.Stop()
					ticker = nxt
				}
			}
		}(p.tickerCh)
	}
	p.evict = dur
	p.mu.Unlock()
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
	now := time.Now()
	n := len(p.elems)
	i0 := rand.Intn(n)
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
