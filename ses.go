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
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

type SesCfg struct {
	Username string
	Password string
	StmtCfg  *StmtCfg
}

// NewSrvCfg creates a SrvCfg with default values.
func NewSesCfg() *SesCfg {
	c := &SesCfg{}
	c.StmtCfg = NewStmtCfg()
	return c
}

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
	return c
}

// Ses is an Oracle session associated with a server.
type Ses struct {
	id       uint64
	cfg      SesCfg
	mu       sync.Mutex
	srv      *Srv
	ocises   *C.OCISession
	isLocked bool

	openStmts *list.List
	openTxs   *list.List
	elem      *list.Element
}

// Close ends a session on an Oracle server.
//
// Any open statements associated with the session are closed.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) Close() (err error) {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	ses.log(_drv.cfg.Log.Ses.Close)
	err = ses.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			errs.PushBack(errR(value))
		}
		srv := ses.srv
		srv.openSess.Remove(ses.elem)
		ses.srv = nil
		ses.ocises = nil
		ses.elem = nil
		ses.openStmts.Init()
		ses.openTxs.Init()
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
	for e := ses.openTxs.Front(); e != nil; e = e.Next() {
		e.Value.(*Tx).close()
	}
	// close statements
	for e := ses.openStmts.Front(); e != nil; e = e.Next() {
		err = e.Value.(*Stmt).Close()
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	// close session
	// OCISessionEnd invalidates oci session handle; no need to free session.ocises
	r := C.OCISessionEnd(
		ses.srv.ocisvcctx,  //OCISvcCtx       *svchp,
		ses.srv.env.ocierr, //OCIError        *errhp,
		ses.ocises,         //OCISession      *usrhp,
		C.OCI_DEFAULT)      //ub4             mode );
	if r == C.OCI_ERROR {
		errs.PushBack(errE(ses.srv.env.ociError()))
	}
	return nil
}

// PrepAndExe prepares and executes a SQL statement returning the number of rows
// affected and a possible error.
func (ses *Ses) PrepAndExe(sql string, params ...interface{}) (rowsAffected uint64, err error) {
	ses.log(_drv.cfg.Log.Ses.PrepAndExe)
	err = ses.checkClosed()
	if err != nil {
		return 0, errE(err)
	}
	stmt, err := ses.Prep(sql)
	defer stmt.Close()
	if err != nil {
		return 0, errE(err)
	}
	rowsAffected, err = stmt.Exe(params...)
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
	ses.log(_drv.cfg.Log.Ses.PrepAndQry)
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
	ses.mu.Lock()
	defer ses.mu.Unlock()
	ses.log(_drv.cfg.Log.Ses.Prep, sql)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	// allocate statement handle
	upOciStmt, err := ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
	if err != nil {
		return nil, errE(err)
	}
	ocistmt := (*C.OCIStmt)(upOciStmt)
	cSql := C.CString(sql) // prepare sql text with statement handle
	defer C.free(unsafe.Pointer(cSql))
	r := C.OCIStmtPrepare2(
		ses.srv.ocisvcctx,                  // OCISvcCtx     *svchp,
		&ocistmt,                           // OCIStmt       *stmtp,
		ses.srv.env.ocierr,                 // OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(cSql)), // const OraText *stmt,
		C.ub4(len(sql)),                    // ub4           stmt_len,
		nil,                                // const OraText *key,
		C.ub4(0),                           // ub4           keylen,
		C.OCI_NTV_SYNTAX,                   // ub4           language,
		C.OCI_DEFAULT)                      // ub4           mode );
	if r == C.OCI_ERROR {
		return nil, errE(ses.srv.env.ociError())
	}
	// set stmt struct
	stmt = _drv.stmtPool.Get().(*Stmt)
	stmt.ses = ses
	stmt.ocistmt = (*C.OCIStmt)(ocistmt)
	stmtCfg := ses.cfg.StmtCfg
	if stmtCfg == nil {
		stmtCfg = NewStmtCfg()
	}
	stmt.cfg = *stmtCfg
	stmt.sql = sql
	stmt.gcts = gcts
	stmt.elem = ses.openStmts.PushBack(stmt)
	if stmt.id == 0 {
		stmt.id = _drv.stmtId.nextId()
	}
	err = stmt.attr(unsafe.Pointer(&stmt.stmtType), 4, C.OCI_ATTR_STMT_TYPE) // determine statement type
	if err != nil {
		return nil, errE(err)
	}
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
	ses.log(_drv.cfg.Log.Ses.Ins)
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
	defer stmt.Close()
	if err != nil {
		return errE(err)
	}
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
	ses.log(_drv.cfg.Log.Ses.Upd)
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
	ses.log(_drv.cfg.Log.Ses.Sel)
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

// StartTx starts an Oracle transaction returning a *Tx and possible error.
func (ses *Ses) StartTx() (tx *Tx, err error) {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	ses.log(_drv.cfg.Log.Ses.StartTx)
	err = ses.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	// start transaction
	// the number of seconds the transaction can be inactive
	// before it is automatically terminated by the system.
	// TODO: add timeout config value
	var timeout C.uword = C.uword(60)
	r := C.OCITransStart(
		ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		ses.srv.env.ocierr, //OCIError     *errhp,
		timeout,            //uword        timeout,
		C.OCI_TRANS_NEW)    //ub4          flags );
	if r == C.OCI_ERROR {
		return nil, errE(ses.srv.env.ociError())
	}
	tx = _drv.txPool.Get().(*Tx) // set *Tx
	tx.ses = ses
	tx.elem = ses.openTxs.PushFront(tx)
	if tx.id == 0 {
		tx.id = _drv.txId.nextId()
	}
	return tx, nil
}

// NumStmt returns the number of open Oracle statements.
func (ses *Ses) NumStmt() int {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	return ses.openStmts.Len()
}

// NumTx returns the number of open Oracle transactions.
func (ses *Ses) NumTx() int {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	return ses.openTxs.Len()
}

// SetCfg applies the specified cfg to the Ses.
//
// Open Stmts do not observe the specified cfg.
func (ses *Ses) SetCfg(cfg SesCfg) {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	ses.cfg = cfg
}

// Cfg returns the Ses's cfg.
func (ses *Ses) Cfg() *SesCfg {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	return &ses.cfg
}

// IsOpen returns true when a session is open; otherwise, false.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) IsOpen() bool {
	ses.mu.Lock()
	defer ses.mu.Unlock()
	return ses.ocises != nil
}

// checkClosed returns an error if Ses is closed. No locking occurs.
func (ses *Ses) checkClosed() error {
	if ses.ocises == nil {
		return er("Ses is closed.")
	}
	return nil
}

// sysName returns a string representing the Ses.
func (ses *Ses) sysName() string {
	if ses == nil {
		return "E_S_S_"
	}
	return ses.srv.sysName() + fmt.Sprintf("S%v", ses.id)
}

// log writes a message with an Ses system name and caller info.
func (ses *Ses) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", ses.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", ses.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with an Ses system name and caller info.
func (ses *Ses) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", ses.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", ses.sysName(), callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}
