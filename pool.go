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

// NewPool returns an idle session pool,
// which evicts the idle sessions every minute,
// and automatically manages the required new connections (Srv).
//
// This is done by maintaining a 1-1 pairing between the Srv and its Ses.
//
// This pool does NOT limit the number of active connections, just helps
// reuse already established connections and sessions, lowering the resource
// usage on the server.
//
// If size <= 0, then DefaultPoolSize is used.
func (env *Env) NewPool(srvCfg *SrvCfg, sesCfg *SesCfg, size int) *Pool {
	if size <= 0 {
		size = DefaultPoolSize
	}
	p := &Pool{
		env:    env,
		srvCfg: srvCfg, sesCfg: sesCfg,
		srv: newIdlePool(size),
		ses: newIdlePool(size),
	}
	p.poolEvictor = &poolEvictor{
		Evict: func(d time.Duration) {
			p.ses.Evict(d)
			p.srv.Evict(d)
		}}
	p.SetEvictDuration(DefaultEvictDuration)
	return p
}

// NewPool returns a new session pool with default config.
func NewPool(dsn string, size int) (*Pool, error) {
	env, err := OpenEnv(NewEnvCfg())
	if err != nil {
		return nil, err
	}
	srvCfg := NewSrvCfg()
	sesCfg := NewSesCfg()
	sesCfg.Mode = DSNMode(dsn)
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	return env.NewPool(srvCfg, sesCfg, size), nil
}

type Pool struct {
	env    *Env
	srvCfg *SrvCfg
	sesCfg *SesCfg

	sync.Mutex
	srv, ses *idlePool

	*poolEvictor
}

// Close all idle sessions and connections.
func (p *Pool) Close() error {
	p.Lock()
	err := p.ses.Close()
	if err2 := p.srv.Close(); err2 != nil && err == nil {
		err = err2
	}
	p.Unlock()
	return err
}

func insteadSesClose(ses *Ses, pool *idlePool) func() error {
	return func() error {
		ses.insteadClose = nil
		pool.Put(ses)
		return nil
	}
}

// Get a session - either an idle session, or if such does not exist, then
// a new session on an idle connection; if such does not exist, then
// a new session on a new connection.
func (p *Pool) Get() (*Ses, error) {
	p.Lock()
	defer p.Unlock()

	Instead := func(ses *Ses) error {
		ses.insteadClose = nil // one-shot
		p.ses.Put(sesSrvPB{Ses: ses, p: p.srv})
		return nil
	}
	// try get session from the ses pool
	for {
		x := p.ses.Get()
		if x == nil { // the ses pool is empty
			break
		}
		ses := x.(sesSrvPB).Ses
		if err := ses.Ping(); err == nil {
			ses.insteadClose = Instead
			return ses, nil
		}
		ses.Close()
	}

	var srv *Srv
	// try to get srv from the srv pool
	for {
		x := p.srv.Get()
		if x == nil { // the srv pool is empty
			break
		}
		srv = x.(*Srv)
		p.sesCfg.StmtCfg = srv.env.cfg.StmtCfg
		if ses, err := srv.OpenSes(p.sesCfg); err == nil {
			ses.insteadClose = Instead
			return ses, nil
		}
		_ = srv.Close()
	}

	//fmt.Fprintf(os.Stderr, "POOL: create new srv!\n")
	srv, err := p.env.OpenSrv(p.srvCfg)
	if err != nil {
		return nil, err
	}
	p.sesCfg.StmtCfg = srv.env.cfg.StmtCfg
	ses, err := srv.OpenSes(p.sesCfg)
	if err != nil {
		return nil, err
	}
	ses.insteadClose = Instead
	return ses, nil
}

// Put the session back to the session pool.
// Ensure that on ses Close (eviction), srv is put back on the idle pool.
func (p *Pool) Put(ses *Ses) {
	if ses == nil || !ses.IsOpen() {
		return
	}
	//fmt.Fprintf(os.Stderr, "POOL: put back ses\n")
	p.ses.Put(sesSrvPB{Ses: ses, p: p.srv})
}

type sesSrvPB struct {
	*Ses
	p *idlePool
}

func (s sesSrvPB) Close() error {
	if s.Ses == nil {
		return nil
	}
	srv := s.Ses.srv
	//fmt.Fprintf(os.Stderr, "POOL: close ses\n")
	err := s.Ses.Close()
	if s.p != nil {
		//fmt.Fprintf(os.Stderr, "POOL: put back srv %v\n", srv)
		s.p.Put(srv)
	}
	return err
}

// NewSrvPool returns a connection pool, which evicts the idle connections in every minute.
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
		srv:    srv,
		sesCfg: sesCfg,
		ses:    newIdlePool(size),
	}
	p.poolEvictor = &poolEvictor{Evict: p.ses.Evict}
	p.SetEvictDuration(DefaultEvictDuration)
	return p
}

type SesPool struct {
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
	dsn = strings.TrimSpace(dsn)
	switch DSNMode(dsn) {
	case SysOper:
		dsn = dsn[:len(dsn)-11]
	case SysDba:
		dsn = dsn[:len(dsn)-10]
	}
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

// DSNMode returns the SessionMode (SysDefault/SysDba/SysOper).
func DSNMode(str string) SessionMode {
	if len(str) <= 11 {
		return SysDefault
	}
	end := strings.ToUpper(str[len(str)-11:])
	if strings.HasSuffix(end, " AS SYSDBA") {
		return SysDba
	} else if strings.HasSuffix(end, " AS SYSOPER") {
		return SysOper
	}
	return SysDefault
}

// NewEnvSrvSes is a comfort function which opens the environment,
// creates a connection (Srv) to the server,
// and opens a session (Ses), in one call.
//
// Ideal for simple use cases.
func NewEnvSrvSes(dsn string, envCfg *EnvCfg) (*Env, *Srv, *Ses, error) {
	env, err := OpenEnv(envCfg)
	if err != nil {
		return nil, nil, nil, err
	}
	srvCfg := NewSrvCfg()
	sesCfg := NewSesCfg()
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	srv, err := env.OpenSrv(srvCfg)
	if err != nil {
		env.Close()
		return nil, nil, nil, err
	}
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		srv.Close()
		env.Close()
		return nil, nil, nil, err
	}
	return env, srv, ses, nil
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
