// Copyright 2016 Rana Ian, Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"container/list"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////
// envList
////////////////////////////////////////////////////////////////////////////////
type envList struct {
	items []*Env
	mu    sync.Mutex
}

func newEnvList() *envList {
	return &envList{items: make([]*Env, 0, 2)}
}

func (l *envList) add(e *Env) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, e) // append item
}

func (l *envList) remove(e *Env) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == e {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *envList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// srvList
////////////////////////////////////////////////////////////////////////////////
type srvList struct {
	items []*Srv
	mu    sync.Mutex
}

func newSrvList() *srvList {
	return &srvList{items: make([]*Srv, 0, 2)}
}

func (l *srvList) add(s *Srv) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, s) // append item
}

func (l *srvList) remove(s *Srv) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == s {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *srvList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range l.items {
		err := item.close() // close will not remove Srv from openSrvs
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Srvs from srvList
}

func (l *srvList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *srvList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// conList
////////////////////////////////////////////////////////////////////////////////
type conList struct {
	items []*Con
	mu    sync.Mutex
}

func newConList() *conList {
	return &conList{items: make([]*Con, 0, 8)}
}

func (l *conList) add(c *Con) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, c) // append item
}

func (l *conList) remove(c *Con) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == c {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *conList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range l.items {
		err := item.close() // close will not remove Con from openCons
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Cons from conList
}

func (l *conList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *conList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// sesList
////////////////////////////////////////////////////////////////////////////////
type sesList struct {
	items []*Ses
	mu    sync.Mutex
}

func newSesList() *sesList {
	return &sesList{items: make([]*Ses, 0, 8)}
}

func (l *sesList) add(s *Ses) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, s) // append item
}

func (l *sesList) remove(s *Ses) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == s {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *sesList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, ses := range l.items {
		err := ses.close() // close will not remove Ses from openSess
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Sess from sesList
}

func (l *sesList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *sesList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// txList
////////////////////////////////////////////////////////////////////////////////
type txList struct {
	items []*Tx
	mu    sync.Mutex
}

func newTxList() *txList {
	return &txList{items: make([]*Tx, 0, 2)}
}

func (l *txList) add(t *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, t) // append item
}

func (l *txList) remove(t *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == t {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *txList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range l.items {
		err := item.close() // close will not remove Tx from openTxs
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Txs from txList
}

func (l *txList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *txList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// stmtList
////////////////////////////////////////////////////////////////////////////////
type stmtList struct {
	items []*Stmt
	mu    sync.Mutex
}

func newStmtList() *stmtList {
	return &stmtList{items: make([]*Stmt, 0, 8)}
}

func (l *stmtList) add(s *Stmt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, s) // append item
}

func (l *stmtList) remove(s *Stmt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == s {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *stmtList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range l.items {
		err := item.close() // close will not remove Stmt from openStmts
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Stmts from stmtList
}

func (l *stmtList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *stmtList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// rsetList
////////////////////////////////////////////////////////////////////////////////
type rsetList struct {
	items []*Rset
	mu    sync.Mutex
}

func newRsetList() *rsetList {
	return &rsetList{items: make([]*Rset, 0, 8)}
}

func (l *rsetList) add(r *Rset) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, r) // append item
}

func (l *rsetList) remove(r *Rset) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n, item := range l.items {
		if item == r {
			l.items[n] = l.items[0]
			l.items = l.items[1:]
			break
		}
	}
}

func (l *rsetList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range l.items {
		err := item.close() // close will not remove Rset from openRsets
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Rsets from rsetList
}

func (l *rsetList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *rsetList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}
