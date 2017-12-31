// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type PoolCfg struct {
	Type           PoolType
	Name           string
	Username       string
	Password       string
	Min, Max, Incr uint32
}

type PoolType uint8

const (
	NoPool  = PoolType(0)
	DRCPool = PoolType(1)
	SPool   = PoolType(2)
	CPool   = PoolType(3)
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
func (env *Env) NewPool(srvCfg SrvCfg, sesCfg SesCfg, size int) *Pool {
	if srvCfg.IsZero() {
		panic("srvCfg shall not be empty")
	}
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
	env, err := OpenEnv()
	if err != nil {
		return nil, err
	}
	srvCfg := SrvCfg{StmtCfg: NewStmtCfg(), Pool: DSNPool(dsn)}
	sesCfg := SesCfg{Mode: DSNMode(dsn)}
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	return env.NewPool(srvCfg, sesCfg, size), nil
}

type Pool struct {
	env    *Env
	srvCfg SrvCfg
	sesCfg SesCfg

	sync.Mutex
	srv, ses *idlePool

	*poolEvictor
}

// Close all idle sessions and connections.
func (p *Pool) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errR(r)
		}
	}()
	p.Lock()
	defer p.Unlock()
	for {
		x := p.ses.Get()
		if x == nil {
			break
		}
		ses := x.(sesSrvPB).Ses
		ses.insteadClose = nil // this is a must!
		ses.Close()
	}
	err = p.ses.Close() // close the pool
	if err2 := p.srv.Close(); err2 != nil && err == nil {
		err = err2
	}
	return err
}

// Get a session - either an idle session, or if such does not exist, then
// a new session on an idle connection; if such does not exist, then
// a new session on a new connection.
func (p *Pool) Get() (ses *Ses, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errR(r)
		}
	}()
	p.Lock()
	defer p.Unlock()

	// Instead of closing the session, put it back to the session pool.
	Instead := func(ses *Ses) error { p.Put(ses); return nil }
	// try get session from the ses pool
	for {
		x := p.ses.Get()
		if x == nil { // the ses pool is empty
			break
		}
		ses = x.(sesSrvPB).Ses
		if ses == nil || !ses.IsOpen() {
			continue
		}
		ses.insteadClose = Instead
		return ses, nil
	}

	var srv *Srv
	// try to get srv from the srv pool
	if p.sesCfg.IsZero() {
		p.sesCfg = NewSesCfg()
		p.sesCfg.StmtCfg = Cfg().StmtCfg
	}
	for {
		x := p.srv.Get()
		if x == nil { // the srv pool is empty
			break
		}
		srv = x.(*Srv)
		if srv == nil {
			continue
		}
		srv.RLock()
		ok := srv.env != nil
		srv.RUnlock()
		if !ok {
			continue
		}
		if ses, err = srv.OpenSes(p.sesCfg); err == nil {
			ses.insteadClose = Instead
			return ses, nil
		}
		_ = srv.Close()
	}

	//fmt.Fprintf(os.Stderr, "POOL: create new srv!\n")
	if srv, err = p.env.OpenSrv(p.srvCfg); err != nil {
		return nil, err
	}
	if ses, err = srv.OpenSes(p.sesCfg); err != nil {
		srv.Close()
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
	p.ses.Lock()
	ses.insteadClose = nil
	//fmt.Fprintf(os.Stderr, "POOL: put back ses\n")
	p.ses.putLocked(sesSrvPB{Ses: ses, p: p.srv})
	p.ses.Unlock()
}

type sesSrvPB struct {
	*Ses
	p *idlePool
}

// Close: after closing the session, put its srv into the pool,
// if it does not have more open sessions.
func (s sesSrvPB) Close() error {
	if s.Ses == nil {
		return nil
	}
	var srv *Srv
	s.Ses.Lock()
	if s.p != nil {
		srv = s.Ses.srv
	}
	s.Ses.insteadClose = nil
	s.Ses.Unlock()

	err := s.Ses.Close()
	if srv != nil { // there's only one ses per srv, so this should be safe
		s.p.Put(srv)
	}
	return err
}

// NewSrvPool returns a connection pool, which evicts the idle connections in every minute.
// The pool holds at most size idle Srv.
// If size is zero, DefaultPoolSize will be used.
func (env *Env) NewSrvPool(srvCfg SrvCfg, size int) *SrvPool {
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
	srvCfg SrvCfg
	srv    *idlePool

	*poolEvictor
}

func (p *SrvPool) Close() error {
	return p.srv.Close()
}

// Get a connection.
func (p *SrvPool) Get() (*Srv, error) {
	x := p.srv.Get()
	if x != nil {
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
func (srv *Srv) NewSesPool(sesCfg SesCfg, size int) *SesPool {
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
	sesCfg SesCfg
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
	ch := p.tickerCh
	fresh := ch == nil
	if fresh { // first initialize
		ch = make(chan *time.Ticker)
		p.tickerCh = ch
	}
	p.Unlock()

	if fresh {
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
		}(ch)
	}
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

// DSNPool returns the Pool config from dsn.
func DSNPool(str string) PoolCfg {
	if strings.HasSuffix(str, ":POOLED") || strings.Contains(str, "(SERVER=POOLED)") {
		pc := PoolCfg{Type: DRCPool, Min: 1, Max: 999, Incr: 1}
		pc.Username, pc.Password, _ = SplitDSN(str)
		return pc
	}
	return PoolCfg{}
}

// NewEnvSrvSes is a comfort function which opens the environment,
// creates a connection (Srv) to the server,
// and opens a session (Ses), in one call.
//
// Ideal for simple use cases.
func NewEnvSrvSes(dsn string) (*Env, *Srv, *Ses, error) {
	env, err := OpenEnv()
	if err != nil {
		return nil, nil, nil, err
	}
	srvCfg := SrvCfg{StmtCfg: env.Cfg(), Pool: DSNPool(dsn)}
	sesCfg := SesCfg{Mode: DSNMode(dsn)}
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	//fmt.Fprintf(os.Stderr, "dsn=% => srv=%#v ses=%#v", dsn, srvCfg, sesCfg)
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

const poolWaitGet = 10 * time.Millisecond
const poolWaitPut = 1 * time.Second

// idlePool is a pool of io.Closers.
// Each element will be Closed on eviction.
//
// The backing store is a simple []io.Closer, which is treated as random store,
// to achive uniform reuse.
type idlePool struct {
	sync.RWMutex
	elems atomic.Value
}

func (p *idlePool) Elems() chan io.Closer {
	i := p.elems.Load()
	if i == nil {
		return nil
	}
	return i.(chan io.Closer)
}
func (p *idlePool) SetElems(c chan io.Closer) chan io.Closer {
	b := p.Elems()
	p.elems.Store(c)
	return b
}

// NewidlePool returns an idlePool.
func newIdlePool(size int) *idlePool {
	var p idlePool
	p.SetElems(make(chan io.Closer, size))
	return &p
}

// Evict halves the idle items
func (p *idlePool) Evict(dur time.Duration) {
	p.RLock()
	defer p.RUnlock()
	elems := p.Elems()
	n := len(elems)/2 + 1
	for i := 0; i < n; i++ {
		select {
		case elem, ok := <-elems:
			if !ok {
				return
			}
			if elem != nil {
				elem.Close()
			}
		default:
			return
		}
	}
}

// Get returns a closer or nil, if no pool found.
func (p *idlePool) Get() io.Closer {
	p.RLock()
	defer p.RUnlock()
	for {
		elems := p.Elems()
		select {
		case elem := <-elems:
			if elem != nil {
				return elem
			}
		case <-time.After(poolWaitGet):
			return nil
		}
	}
}

// Put a new element into the store. The slot is chosen randomly.
// If no empty slot is found, one (random) is Close()-d and this new
// element is put there.
// This way elements reused uniformly.
func (p *idlePool) Put(c io.Closer) {
	p.RLock()
	p.putLocked(c)
	p.RUnlock()
}

func (p *idlePool) putLocked(c io.Closer) {
	select {
	case p.Elems() <- c:
		return
	default:
		go func() {
			select {
			case p.Elems() <- c:
				return
			case <-time.After(poolWaitPut):
				c.Close()
			}
		}()
	}
}

// Close all elements.
func (p *idlePool) Close() error {
	p.Lock()
	defer p.Unlock()
	elems := p.SetElems(nil)
	if elems == nil {
		return nil
	}
	close(elems)
	var err error
	for elem := range elems {
		if elem == nil {
			continue
		}
		if closeErr := elem.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}
