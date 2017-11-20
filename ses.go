// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type SesCfg struct {
	Username string
	Password string
	Mode     SessionMode

	StmtCfg
}

func (c SesCfg) IsZero() bool { return false }
func NewSesCfg() SesCfg       { return SesCfg{} }

func (cfg SesCfg) SetStmtCfg(stmtCfg StmtCfg) SesCfg {
	cfg.StmtCfg = stmtCfg
	return cfg
}

func (c SesCfg) SetPrefetchRowCount(prefetchRowCount uint32) SesCfg {
	c.StmtCfg = c.StmtCfg.SetPrefetchRowCount(prefetchRowCount)
	return c
}
func (c SesCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) SesCfg {
	c.StmtCfg = c.StmtCfg.SetPrefetchMemorySize(prefetchMemorySize)
	return c
}
func (c SesCfg) SetLongBufferSize(size uint32) SesCfg {
	c.StmtCfg = c.StmtCfg.SetLongBufferSize(size)
	return c
}
func (c SesCfg) SetLongRawBufferSize(size uint32) SesCfg {
	c.StmtCfg = c.StmtCfg.SetLongRawBufferSize(size)
	return c
}
func (c SesCfg) SetLobBufferSize(size int) SesCfg {
	c.StmtCfg = c.StmtCfg.SetLobBufferSize(size)
	return c
}
func (c SesCfg) SetStringPtrBufferSize(size int) SesCfg {
	c.StmtCfg = c.StmtCfg.SetStringPtrBufferSize(size)
	return c
}
func (c SesCfg) SetByteSlice(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetByteSlice(gct)
	return c
}
func (c SesCfg) SetNumberInt(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetNumberInt(gct)
	return c
}
func (c SesCfg) SetNumberBigInt(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetNumberBigInt(gct)
	return c
}
func (c SesCfg) SetNumberFloat(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetNumberFloat(gct)
	return c
}
func (c SesCfg) SetNumberBigFloat(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetNumberBigFloat(gct)
	return c
}
func (c SesCfg) SetBinaryDouble(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetBinaryDouble(gct)
	return c
}
func (c SesCfg) SetBinaryFloat(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetBinaryFloat(gct)
	return c
}
func (c SesCfg) SetFloat(gct GoColumnType) SesCfg { c.StmtCfg = c.StmtCfg.SetFloat(gct); return c }
func (c SesCfg) SetDate(gct GoColumnType) SesCfg  { c.StmtCfg = c.StmtCfg.SetDate(gct); return c }
func (c SesCfg) SetTimestamp(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetTimestamp(gct)
	return c
}
func (c SesCfg) SetTimestampTz(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetTimestampTz(gct)
	return c
}
func (c SesCfg) SetTimestampLtz(gct GoColumnType) SesCfg {
	c.StmtCfg = c.StmtCfg.SetTimestampLtz(gct)
	return c
}
func (c SesCfg) SetChar1(gct GoColumnType) SesCfg   { c.StmtCfg = c.StmtCfg.SetChar1(gct); return c }
func (c SesCfg) SetChar(gct GoColumnType) SesCfg    { c.StmtCfg = c.StmtCfg.SetChar(gct); return c }
func (c SesCfg) SetVarchar(gct GoColumnType) SesCfg { c.StmtCfg = c.StmtCfg.SetVarchar(gct); return c }
func (c SesCfg) SetLong(gct GoColumnType) SesCfg    { c.StmtCfg = c.StmtCfg.SetLong(gct); return c }
func (c SesCfg) SetClob(gct GoColumnType) SesCfg    { c.StmtCfg = c.StmtCfg.SetClob(gct); return c }
func (c SesCfg) SetBlob(gct GoColumnType) SesCfg    { c.StmtCfg = c.StmtCfg.SetBlob(gct); return c }
func (c SesCfg) SetRaw(gct GoColumnType) SesCfg     { c.StmtCfg = c.StmtCfg.SetRaw(gct); return c }
func (c SesCfg) SetLongRaw(gct GoColumnType) SesCfg { c.StmtCfg = c.StmtCfg.SetLongRaw(gct); return c }

type SessionMode uint8

const (
	// SysDefault is the default, normal session mode.
	SysDefault = SessionMode(iota)
	// SysDba is for connecting as SYSDBA.
	SysDba
	// SysOper is for connectiong as SYSOPER.
	SysOper
)

// LogSesCfg represents Ses logging configuration values.
type LogSesCfg struct {
	// Close determines whether the Ses.Close method is logged.
	//
	// The default is true.
	Close bool

	// PrepAndExe determines whether the Ses.PrepAndExe method is logged.
	//
	// The default is true.
	PrepAndExe bool

	// PrepAndQry determines whether the Ses.PrepAndQry method is logged.
	//
	// The default is true.
	PrepAndQry bool

	// Prep determines whether the Ses.Prep method is logged.
	//
	// The default is true.
	Prep bool

	// Ins determines whether the Ses.Ins method is logged.
	//
	// The default is true.
	Ins bool

	// Upd determines whether the Ses.Upd method is logged.
	//
	// The default is true.
	Upd bool

	// Sel determines whether the Ses.Sel method is logged.
	//
	// The default is true.
	Sel bool

	// StartTx determines whether the Ses.StartTx method is logged.
	//
	// The default is true.
	StartTx bool

	// Ping determines whether the Ses.Ping method is logged.
	//
	// The default is true.
	Ping bool

	// Break determines whether the Ses.Break method is logged.
	//
	// The default is true.
	Break bool
}

// NewLogSesCfg creates a LogSesCfg with default values.
func NewLogSesCfg() LogSesCfg {
	c := LogSesCfg{}
	c.Close = true
	c.PrepAndExe = true
	c.PrepAndQry = true
	c.Prep = true
	c.Ins = true
	c.Upd = true
	c.Sel = true
	c.StartTx = true
	c.Ping = true
	c.Break = true
	return c
}

// Ses is an Oracle session associated with a server.
type Ses struct {
	sync.RWMutex

	cfg atomic.Value
	// protects that open/close should not happen at once
	cmu       sync.Mutex
	id        uint64
	env       atomic.Value // cached
	srv       *Srv
	ocisvcctx *C.OCISvcCtx
	ocises    *C.OCISession
	isLocked  bool

	openStmts *stmtList
	openTxs   *txList

	insteadClose func(ses *Ses) error
	timezone     *time.Location

	sysNamer
}

// Cfg returns the Ses's SesCfg, or it's Srv's, if not set.
// If the ses.srv.env is the PkgSqlEnv, that will override StmtCfg!
func (ses *Ses) Cfg() SesCfg {
	c := ses.cfg.Load()
	var cfg SesCfg
	//fmt.Fprintf(os.Stderr, "%s.Cfg=%#v\n", ses.sysName(), c)
	if c != nil {
		cfg = c.(SesCfg)
	}
	if env := ses.Env(); env.isPkgEnv {
		cfg.StmtCfg = env.Cfg()
	} else if cfg.StmtCfg.IsZero() {
		cfg.StmtCfg = ses.srv.Cfg().StmtCfg
	}
	return cfg
}
func (ses *Ses) SetCfg(cfg SesCfg) {
	ses.cfg.Store(cfg)
}

func (ses *Ses) Env() *Env {
	e := ses.env.Load()
	if e == nil {
		return nil
	}
	return e.(*Env)
}

// Close ends a session on an Oracle server.
//
// Any open statements associated with the session are closed.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) Close() (err error) {
	if ses == nil {
		return nil
	}
	return ses.closeWithRemove()
}

func (ses *Ses) closeWithRemove() error {
	if ses == nil {
		return nil
	}
	ses.RLock()
	srv := ses.srv
	insteadClose := ses.insteadClose
	ses.RUnlock()
	if srv == nil {
		return nil
	}
	if insteadClose != nil {
		return insteadClose(ses)
	}
	srv.openSess.remove(ses)
	return ses.close()
}

// close ends a session on an Oracle server, without holding the lock.
// does not remove Ses from Srv.openSess
func (ses *Ses) close() (err error) {
	ses.cmu.Lock()
	defer ses.cmu.Unlock()

	ses.log(_drv.Cfg().Log.Ses.Close)
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			errs.PushBack(errR(value))
		}

		ses.SetCfg(SesCfg{})
		ses.Lock()
		ses.env.Store((*Env)(nil))
		ses.srv = nil
		ses.ocisvcctx = nil
		ses.ocises = nil
		ses.openStmts.clear()
		ses.openTxs.clear()
		ses.Unlock()
		_drv.sesPool.Put(ses)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()

	// close transactions
	// close does not rollback or commit any transactions
	// Expect user to make explicit Commit or Rollback.
	// Any open transactions will be timedout by the server
	// if not explicitly committed or rolledback.
	ses.RLock()
	openTxs, openStmts := ses.openTxs, ses.openStmts
	env, srv := ses.Env(), ses.srv
	ocises, ocisvcctx := ses.ocises, ses.ocisvcctx
	ses.RUnlock()
	openTxs.closeAll(errs)
	openStmts.closeAll(errs) // close statements

	// close session
	var r C.sword
	if srv.poolType == NoPool {
		r = C.OCISessionEnd(
			ocisvcctx,     //OCISvcCtx       *svchp,
			env.ocierr,    //OCIError        *errhp,
			ocises,        //OCISession      *usrhp,
			C.OCI_DEFAULT) //ub4             mode );
	} else {
		r = C.OCISessionRelease(
			ocisvcctx,     //OCISvcCtx       *svchp,
			env.ocierr,    //OCIError        *errhp,
			nil,           //OraText         *tag,
			0,             //ub4             tag_len,
			C.OCI_DEFAULT, //ub4             mode );
		)
	}
	if r == C.OCI_ERROR {
		errs.PushBack(errE(env.ociError()))
	}

	env.RLock()
	err = env.freeOciHandle(unsafe.Pointer(ocises), C.OCI_HTYPE_SESSION)
	if err != nil {
		env.RUnlock()
		return errE(err)
	}
	err = env.freeOciHandle(unsafe.Pointer(ocisvcctx), C.OCI_HTYPE_SVCCTX)
	env.RUnlock()
	if err != nil {
		return errE(err)
	}

	return nil
}

// PrepAndExe prepares and executes a SQL statement returning the number of rows
// affected and a possible error, using Exe, calling in batch for arrays.
//
// WARNING: just as sql.QueryRow, the prepared statement is closed right after
// execution, with all its siblings (Lobs, Rsets...)!
//
// So if you want to retrieve and use such objects, you have to first Prep,
// then Exe separately (and close the Stmt returned by Prep after finishing with
// those objects).
func (ses *Ses) PrepAndExe(sql string, params ...interface{}) (rowsAffected uint64, err error) {
	return ses.prepAndExe(sql, false, params...)
}

// PrepAndExeP prepares and executes a SQL statement returning the number of rows
// affected and a possible error, using ExeP, so passing arrays as is.
func (ses *Ses) PrepAndExeP(sql string, params ...interface{}) (rowsAffected uint64, err error) {
	return ses.prepAndExe(sql, true, params...)
}

// prepAndExe prepares and executes a SQL statement returning the number of rows
// affected and a possible error.
func (ses *Ses) prepAndExe(sql string, isAssocArray bool, params ...interface{}) (rowsAffected uint64, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	ses.log(_drv.Cfg().Log.Ses.PrepAndExe)
	err = ses.checkClosed()
	if err != nil {
		return 0, errE(err)
	}
	stmt, err := ses.Prep(sql)
	defer func() {
		if stmt != nil {
			err0 := stmt.Close()
			if err == nil { // don't overwrite original err
				err = err0
			}
		}
	}()
	if err != nil {
		return 0, errE(err)
	}
	if isAssocArray {
		rowsAffected, err = stmt.ExeP(params...)
	} else {
		rowsAffected, err = stmt.Exe(params...)
	}
	if err != nil {
		return rowsAffected, errE(err)
	}
	return rowsAffected, nil
}

// PrepAndQry prepares a SQL statement and queries an Oracle server returning
// an *Rset and a possible error.
//
// If an error occurs during Prep or Qry a nil *Rset will be returned.
//
// The *Stmt internal to this method is automatically closed when the *Rset
// retrieves all rows or returns an error.
func (ses *Ses) PrepAndQry(sql string, params ...interface{}) (rset *Rset, err error) {
	ses.log(_drv.Cfg().Log.Ses.PrepAndQry)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	stmt, err := ses.Prep(sql)
	if err != nil {
		defer stmt.Close()
		return nil, errE(err)
	}
	rset, err = stmt.Qry(params...)
	if err != nil {
		defer stmt.Close()
		return nil, errE(err)
	}
	rset.autoClose = true
	return rset, nil
}

// Prep prepares a sql statement returning a *Stmt and possible error.
func (ses *Ses) Prep(sql string, gcts ...GoColumnType) (stmt *Stmt, err error) {
	if ses == nil {
		return nil, er("ses may not be nil.")
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	ses.log(_drv.Cfg().Log.Ses.Prep, sql)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	ocistmt := (*C.OCIStmt)(nil)
	cSql := C.CString(sql) // prepare sql text with statement handle
	ses.RLock()
	env := ses.Env()
	r := C.OCIStmtPrepare2(
		ses.ocisvcctx,                      // OCISvcCtx     *svchp,
		&ocistmt,                           // OCIStmt       *stmtp,
		env.ocierr,                         // OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(cSql)), // const OraText *stmt,
		C.ub4(len(sql)),                    // ub4           stmt_len,
		nil,                                // const OraText *key,
		C.ub4(0),                           // ub4           keylen,
		C.OCI_NTV_SYNTAX,                   // ub4           language,
		C.OCI_DEFAULT)                      // ub4           mode );
	ses.RUnlock()
	C.free(unsafe.Pointer(cSql))
	if r == C.OCI_ERROR {
		return nil, errE(env.ociError())
	}
	// set stmt struct
	stmt = _drv.stmtPool.Get().(*Stmt)
	stmtCfg := ses.Cfg().StmtCfg
	stmt.SetCfg(StmtCfg{}) // reset - always inherit from ses.Cfg().
	stmt.cmu.Lock()
	defer stmt.cmu.Unlock()
	stmt.Lock()
	stmt.ocistmt = (*C.OCIStmt)(ocistmt)
	ses.RLock()
	stmt.env.Store(env)
	stmt.ses = ses
	if ses.srv.IsUTF8() && stmtCfg.stringPtrBufferSize > 1000 {
		stmt.stringPtrBufferSize = 1000
	}
	ses.RUnlock()
	stmt.sql = sql
	stmt.gcts = gcts
	if stmt.id == 0 {
		stmt.id = _drv.stmtId.nextId()
	}
	stmt.Unlock()
	st, err := stmt.attr(2, C.OCI_ATTR_STMT_TYPE) // determine statement type
	if err != nil {
		return nil, errE(err)
	}
	stmt.Lock()
	stmt.stmtType = *((*C.ub2)(st))
	stmt.Unlock()

	C.free(unsafe.Pointer(st))
	ses.openStmts.add(stmt)

	//ses.logF(true, "\n ses.cfg=%#v\nstmt.cfg=%#v", ses.Cfg().StmtCfg, stmt.Cfg())

	return stmt, nil
}

// Ins composes, prepares and executes a sql INSERT statement returning a
// possible error.
//
// Ins offers convenience when specifying a long list of sql columns.
//
// Ins expects at least two column name-value pairs where the last pair will be
// a part of a sql RETURNING clause. The last column name is expected to be an
// identity column returning an Oracle-generated value. The last value specified
// to the variadic parameter 'columnPairs' is expected to be a pointer capable
// of receiving the identity value.
func (ses *Ses) Ins(tbl string, columnPairs ...interface{}) (err error) {
	ses.log(_drv.Cfg().Log.Ses.Ins)
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	if tbl == "" {
		return errF("tbl is empty.")
	}
	if len(columnPairs) < 2 {
		return errF("Parameter 'columnPairs' expects at least 2 column name-value pairs.")
	}
	if len(columnPairs)%2 != 0 {
		return errF("Variadic parameter 'columnPairs' received an odd number of elements. Parameter 'columnPairs' expects an even number of elements.")
	}
	// build INSERT statement, params slice
	params := make([]interface{}, len(columnPairs)/2)
	buf := new(bytes.Buffer)
	buf.WriteString("INSERT INTO ")
	buf.WriteString(tbl)
	buf.WriteString(" (")
	lastColName := ""
	for p := 0; p < len(params); p++ {
		n := p * 2
		columnName, ok := columnPairs[n].(string)
		if !ok {
			return errF("Variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
		}
		if p == len(params)-1 {
			lastColName = columnName
		} else {
			buf.WriteString(columnName)
			if p < len(params)-2 {
				buf.WriteString(", ")
			}
		}
		params[p] = columnPairs[n+1]
	}
	buf.WriteString(") VALUES (")
	for n := 1; n < len(params); n++ {
		buf.WriteString(fmt.Sprintf(":%v", n))
		if n < len(params)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(")")
	buf.WriteString(" RETURNING ")
	buf.WriteString(lastColName)
	buf.WriteString(" INTO :RET_VAL")
	stmt, err := ses.Prep(buf.String()) // prep
	if err != nil {
		return errE(err)
	}
	defer stmt.Close()
	_, err = stmt.Exe(params...) // exe
	if err != nil {
		return errE(err)
	}
	return nil
}

// Upd composes, prepares and executes a sql UPDATE statement returning a
// possible error.
//
// Upd offers convenience when specifying a long list of sql columns.
func (ses *Ses) Upd(tbl string, columnPairs ...interface{}) (err error) {
	ses.log(_drv.Cfg().Log.Ses.Upd)
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	if tbl == "" {
		return errF("tbl is empty.")
	}
	if len(columnPairs) < 2 {
		return errF("Parameter 'columnPairs' expects at least 2 column name-value pairs.")
	}
	if len(columnPairs)%2 != 0 {
		return errF("Variadic parameter 'columnPairs' received an odd number of elements. Parameter 'columnPairs' expects an even number of elements.")
	}
	// build UPDATE statement, params slice
	params := make([]interface{}, len(columnPairs)/2)
	buf := new(bytes.Buffer)
	buf.WriteString("UPDATE ")
	buf.WriteString(tbl)
	buf.WriteString(" SET ")
	lastColName := ""
	for p := 0; p < len(params); p++ {
		n := p * 2
		columnName, ok := columnPairs[n].(string)
		if !ok {
			return errF("Variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
		}
		if p == len(params)-1 {
			lastColName = columnName
		} else {
			buf.WriteString(columnName)
			buf.WriteString(fmt.Sprintf(" = :%v", p+1))
			if p < len(params)-2 {
				buf.WriteString(", ")
			}
		}
		params[p] = columnPairs[n+1]
	}
	buf.WriteString(" WHERE ")
	buf.WriteString(lastColName)
	buf.WriteString(" = :WHERE_VAL")
	stmt, err := ses.Prep(buf.String()) // prep
	defer func() {
		err = stmt.Close()
		if err != nil {
			err = errE(err)
		}
	}()
	if err != nil {
		return errE(err)
	}
	_, err = stmt.Exe(params...) // exe
	if err != nil {
		return errE(err)
	}
	return nil
}

// Sel composes, prepares and queries a sql SELECT statement returning an *ora.Rset
// and possible error.
//
// Sel offers convenience when specifying a long list of sql columns with
// non-default GoColumnTypes.
//
// Specify a sql FROM clause with one or more pairs of sql column
// name-GoColumnType pairs. The FROM clause may have additional SQL clauses
// such as WHERE, HAVING, etc.
func (ses *Ses) Sel(sqlFrom string, columnPairs ...interface{}) (rset *Rset, err error) {
	ses.log(_drv.Cfg().Log.Ses.Sel)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	if len(columnPairs) == 0 {
		return nil, errF("No column name-type pairs specified.")
	}
	if len(columnPairs)%2 != 0 {
		return nil, errF("Variadic parameter 'columnPairs' received an odd number of elements. Parameter 'columnPairs' expects an even number of elements.")
	}
	// build select statement, gcts
	gcts := make([]GoColumnType, len(columnPairs)/2)
	buf := new(bytes.Buffer)
	buf.WriteString("SELECT ")
	for n := 0; n < len(columnPairs); n += 2 {
		columnName, ok := columnPairs[n].(string)
		if !ok {
			return nil, errF("Variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
		}
		gct, ok := columnPairs[n+1].(GoColumnType)
		if !ok {
			return nil, errF("Variadic parameter 'columnPairs' expected an element at index %v to be of type ora.GoColumnType", n+1)
		}
		buf.WriteString(columnName)
		if n != len(columnPairs)-2 {
			buf.WriteRune(',')
		}
		buf.WriteRune(' ')
		gcts[n/2] = gct
	}
	// add FROM keyword?
	fromIndex := strings.Index(strings.ToUpper(sqlFrom), "FROM")
	if fromIndex < 0 {
		buf.WriteString("FROM ")
	}
	buf.WriteString(sqlFrom)
	// prep
	stmt, err := ses.Prep(buf.String(), gcts...)
	if err != nil {
		defer stmt.Close()
		return nil, errE(err)
	}
	// qry
	rset, err = stmt.Qry()
	if err != nil {
		defer stmt.Close()
		return nil, errE(err)
	}
	rset.autoClose = true
	return rset, nil
}

type TxOption func(*txOption)
type txOption struct {
	flags   uint32
	timeout time.Duration
}

func TxFlags(flags uint32) TxOption            { return func(o *txOption) { o.flags = flags } }
func TxTimeout(timeout time.Duration) TxOption { return func(o *txOption) { o.timeout = timeout } }

// StartTx starts an Oracle transaction returning a *Tx and possible error.
func (ses *Ses) StartTx(opts ...TxOption) (tx *Tx, err error) {
	ses.log(_drv.Cfg().Log.Ses.StartTx)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}

	var o txOption
	for _, opt := range opts {
		opt(&o)
	}
	// start transaction
	// the number of seconds the transaction can be inactive
	// before it is automatically terminated by the system.
	var timeout = C.uword(60)
	if o.timeout > 0 {
		timeout = C.uword(o.timeout / time.Second)
	}
	ses.RLock()
	env := ses.Env()
	r := C.OCITransStart(
		ses.ocisvcctx, //OCISvcCtx    *svchp,
		env.ocierr,    //OCIError     *errhp,
		timeout,       //uword        timeout,
		C.OCI_TRANS_NEW|C.ub4(o.flags)) //ub4          flags );
	ses.RUnlock()
	if r == C.OCI_ERROR {
		return nil, errE(env.ociError())
	}
	tx = _drv.txPool.Get().(*Tx) // set *Tx
	tx.cmu.Lock()
	tx.Lock()
	tx.ses = ses
	if tx.id == 0 {
		tx.id = _drv.txId.nextId()
	}
	tx.Unlock()
	tx.cmu.Unlock()
	ses.openTxs.add(tx)

	return tx, nil
}

// Ping returns nil when an Oracle server is contacted; otherwise, an error.
func (ses *Ses) Ping() (err error) {
	ses.log(_drv.Cfg().Log.Ses.Ping)
	defer func() {
		if r := recover(); r != nil {
			err = errR(r)
		}
	}()
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	ses.RLock()
	env := ses.Env()
	r := C.OCIPing(
		ses.ocisvcctx, //OCISvcCtx     *svchp,
		env.ocierr,    //OCIError      *errhp,
		C.OCI_DEFAULT) //ub4           mode );
	ses.RUnlock()
	if r == C.OCI_ERROR {
		err := errE(env.ociError())
		if cd, ok := err.(interface {
			Code() int
		}); ok {
			if cd.Code() == 1010 { // ORA-01010: invalid OCI operation for server < 10.2
				return nil
			}
		}
		return err
	}
	return nil
}

// Break stops the currently running OCI function.
func (ses *Ses) Break() (err error) {
	if ses == nil {
		return nil
	}
	ses.log(_drv.Cfg().Log.Ses.Break)
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	ses.Lock()
	defer ses.Unlock()
	env := ses.Env()
	if ses.ocisvcctx == nil || env == nil || env.ocierr == nil {
		return nil
	}
	if r := C.OCIBreak(unsafe.Pointer(ses.ocisvcctx), env.ocierr); r == C.OCI_ERROR {
		return errE(env.ociError())
	}
	if r := C.OCIReset(unsafe.Pointer(ses.ocisvcctx), env.ocierr); r == C.OCI_ERROR {
		return errE(env.ociError())
	}
	return nil
}

// NumStmt returns the number of open Oracle statements.
func (ses *Ses) NumStmt() int {
	ses.RLock()
	openStmts := ses.openStmts
	ses.RUnlock()
	return openStmts.len()
}

// NumTx returns the number of open Oracle transactions.
func (ses *Ses) NumTx() int {
	ses.RLock()
	openTxs := ses.openTxs
	ses.RUnlock()
	return openTxs.len()
}

// IsOpen returns true when a session is open; otherwise, false.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) IsOpen() bool {
	return ses.checkClosed() == nil
}

// checkClosed returns an error if Ses is closed. No locking occurs.
func (ses *Ses) checkClosed() error {
	if ses == nil {
		return er("Ses is closed.")
	}
	ses.RLock()
	closed := ses.ocises == nil
	srv := ses.srv
	ses.RUnlock()
	if closed {
		return er("Ses is closed.")
	}
	return srv.checkClosed()
}

// sysName returns a string representing the Ses.
func (ses *Ses) sysName() string {
	if ses == nil {
		return "E_S_S_"
	}
	return ses.sysNamer.Name(func() string { return fmt.Sprintf("%sS%v", ses.srv.sysName(), ses.id) })
}

// Timezone return the current session's timezone.
func (ses *Ses) Timezone() (*time.Location, error) {
	ses.RLock()
	tz := ses.timezone
	ses.RUnlock()
	if tz != nil {
		return tz, nil
	}
	rset, err := ses.PrepAndQry("select CAST(tz_offset(sessiontimezone) AS VARCHAR2(7)) from dual")
	if err != nil {
		return nil, err
	}
	defer func() {
		for rset.Next() {
		}
	}()
	hasRow := rset.Next()
	if !hasRow {
		return nil, errors.New("no time zone returned from database")
	}
	var value string
	switch x := rset.Row[0].(type) {
	case string:
		value = x
	case String:
		value = x.String()
	}
	value = strings.Trim(value, " \x00")
	if value == "" {
		return nil, fmt.Errorf("unable to retrieve database timezone (%#v)", rset.Row[0])
	}
	var sign int
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.Replace(value, "-", "", 1)
	} else {
		sign = 1
	}
	strs := strings.Split(value, ":")
	if strs == nil || len(strs) != 2 {
		return nil, errors.New("unable to parse database timezone offset")
	}
	hourOffset, err := strconv.ParseInt(strs[0], 10, 32)
	if err != nil {
		return nil, err
	}
	minStr := strs[1]
	minOffset, err := strconv.ParseInt(minStr, 10, 32)
	if err != nil {
		return nil, err
	}
	offset := sign * ((int(hourOffset) * 3600) + (int(minOffset) * 60))
	tz = time.FixedZone(value, offset)
	ses.Lock()
	ses.timezone = tz
	ses.Unlock()
	return tz, nil
}

// SetAction sets the MODULE and ACTION attribute of the session.
func (ses *Ses) SetAction(module, action string) error {
	if len(module) > 48 {
		module = module[:48]
	}
	cModule := C.CString(module)
	defer C.free(unsafe.Pointer(cModule))
	env := ses.Env()
	if err := env.setAttr(unsafe.Pointer(ses.ocises), C.OCI_HTYPE_SESSION,
		unsafe.Pointer(cModule), C.ub4(len(module)), C.OCI_ATTR_MODULE,
	); err != nil {
		return errE(err)
	}

	if len(action) > 32 {
		action = action[:32]
	}
	cAction := C.CString(action)
	defer C.free(unsafe.Pointer(cAction))
	if err := env.setAttr(unsafe.Pointer(ses.ocises), C.OCI_HTYPE_SESSION,
		unsafe.Pointer(cAction), C.ub4(len(action)), C.OCI_ATTR_ACTION,
	); err != nil {
		return errE(err)
	}
	return nil
}

// log writes a message with an Ses system name and caller info.
func (ses *Ses) log(enabled bool, v ...interface{}) {
	Log := _drv.Cfg().Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", ses.sysName(), callInfo(2))
	} else {
		Log.Logger.Infof("%v %v %v", ses.sysName(), callInfo(2), fmt.Sprint(v...))
	}
}

// log writes a formatted message with an Ses system name and caller info.
func (ses *Ses) logF(enabled bool, format string, v ...interface{}) {
	Log := _drv.Cfg().Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", ses.sysName(), callInfo(2))
	} else {
		Log.Logger.Infof("%v %v %v", ses.sysName(), callInfo(2), fmt.Sprintf(format, v...))
	}
}
