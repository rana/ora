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
	"github.com/golang/glog"
	"strings"
	"unsafe"
)

// Ses is an Oracle session associated with a server.
type Ses struct {
	id     uint64
	srv    *Srv
	ocises *C.OCISession

	txId     uint64
	stmtId   uint64
	txs      *list.List
	stmts    *list.List
	elem     *list.Element
	stmtCfg  StmtCfg
	username string
}

// NumStmt returns the number of open Oracle statements.
func (ses *Ses) NumStmt() int {
	return ses.stmts.Len()
}

// NumTx returns the number of open Oracle transactions.
func (ses *Ses) NumTx() int {
	return ses.txs.Len()
}

// checkIsOpen validates that the session is open.
func (ses *Ses) checkIsOpen() error {
	if !ses.IsOpen() {
		return errNewF("Ses is closed (id %v)", ses.id)
	}
	return nil
}

// IsOpen returns true when a session is open; otherwise, false.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) IsOpen() bool {
	return ses.srv != nil
}

// Close ends a session on an Oracle server.
//
// Any open statements associated with the session are closed.
//
// Calling Close will cause Ses.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Srv.OpenSes to open a new session.
func (ses *Ses) Close() (err error) {
	if err := ses.checkIsOpen(); err != nil {
		return err
	}
	glog.Infof("E%vS%vS%v] Close", ses.srv.env.id, ses.srv.id, ses.id)
	errs := ses.srv.env.drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			glog.Errorln(recoverMsg(value))
			errs.PushBack(errRecover(value))
		}

		srv := ses.srv
		srv.sess.Remove(ses.elem)
		ses.txs.Init()
		ses.stmts.Init()
		ses.srv = nil
		ses.ocises = nil
		ses.elem = nil
		ses.username = ""
		srv.env.drv.sesPool.Put(ses)

		m := newMultiErrL(errs)
		if m != nil {
			err = *m
		}
		errs.Init()
		srv.env.drv.listPool.Put(errs)
	}()

	// close transactions
	// this does not rollback or commit any transactions
	// any open transactions will be timedout by the server
	for e := ses.txs.Front(); e != nil; e = e.Next() {
		e.Value.(*Tx).close()
	}
	// close statements
	for e := ses.stmts.Front(); e != nil; e = e.Next() {
		err0 := e.Value.(*Stmt).Close()
		errs.PushBack(err0)
	}
	// close session
	// OCISessionEnd invalidates oci session handle; no need to free session.ocises
	r := C.OCISessionEnd(
		ses.srv.ocisvcctx,  //OCISvcCtx       *svchp,
		ses.srv.env.ocierr, //OCIError        *errhp,
		ses.ocises,         //OCISession      *usrhp,
		C.OCI_DEFAULT)      //ub4             mode );
	if r == C.OCI_ERROR {
		errs.PushBack(ses.srv.env.ociError())
	}

	return err
}

// PrepAndExec prepares and executes a SQL statement returning the number of rows
// affected and a possible error.
func (ses *Ses) PrepAndExec(sql string, params ...interface{}) (uint64, error) {
	stmt, err := ses.Prep(sql)
	defer stmt.Close()
	if err != nil {
		return 0, err
	}
	return stmt.Exe(params...)
}

// PrepAndQry prepares a SQL statement and queries an Oracle server returning
// an *Rset and a possible error.
//
// If an error occurs during Prep or Query a nil *Rset will be returned.
//
// The *Stmt internal to this method is automatically closed when the *Rset
// retrieves all rows or returns an error.
func (ses *Ses) PrepAndQry(sql string, params ...interface{}) (*Rset, error) {
	stmt, err := ses.Prep(sql)
	if err != nil {
		defer stmt.Close()
		return nil, err
	}
	rset, err := stmt.Qry(params...)
	if err != nil {
		defer stmt.Close()
		return nil, err
	}
	rset.autoClose = true
	return rset, nil
}

// Prep prepares a sql statement returning a *Stmt and possible error.
func (ses *Ses) Prep(sql string, gcts ...GoColumnType) (*Stmt, error) {
	if err := ses.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vS%vS%v] Prep: %v", ses.srv.env.id, ses.srv.id, ses.id, sql)
	// allocate statement handle
	ocistmt, err := ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
	if err != nil {
		return nil, err
	}
	// prepare sql text with statement handle
	cSql := C.CString(sql)
	defer C.free(unsafe.Pointer(cSql))
	r := C.OCIStmtPrepare(
		(*C.OCIStmt)(ocistmt),              // OCIStmt       *stmtp,
		ses.srv.env.ocierr,                 // OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(cSql)), // const OraText *stmt,
		C.ub4(len(sql)),                    // ub4           stmt_len,
		C.OCI_NTV_SYNTAX,                   // ub4           language,
		C.OCI_DEFAULT)                      // ub4           mode );
	if r == C.OCI_ERROR {
		return nil, ses.srv.env.ociError()
	}

	// set stmt struct
	stmt := ses.srv.env.drv.stmtPool.Get().(*Stmt)
	if stmt.id == 0 {
		ses.stmtId++
		stmt.id = ses.stmtId
	}
	stmt.ses = ses
	stmt.ocistmt = (*C.OCIStmt)(ocistmt)
	// determine statement type
	var stmtType C.ub4
	err = stmt.attr(unsafe.Pointer(&stmtType), 4, C.OCI_ATTR_STMT_TYPE)
	if err != nil {
		err2 := stmt.Close()
		if err2 != nil {
			return nil, err2
		}
		return nil, err
	}
	stmt.gcts = gcts
	stmt.sql = sql
	stmt.stmtType = stmtType
	stmt.Cfg = ses.stmtCfg
	stmt.elem = ses.stmts.PushBack(stmt)

	return stmt, nil
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
func (ses *Ses) Sel(sqlFrom string, columnPairs ...interface{}) (*Rset, error) {
	if len(columnPairs) == 0 {
		return nil, errNew("no column name-type pairs specified")
	}
	if len(columnPairs)%2 != 0 {
		return nil, errNew("variadic parameter 'columnPairs' received an odd number of elements. Parameter 'columnPairs' expects an even number of elements")
	}
	// build select statement, gcts
	gcts := make([]GoColumnType, len(columnPairs)/2)
	buf := new(bytes.Buffer)
	buf.WriteString("SELECT ")
	for n := 0; n < len(columnPairs); n += 2 {
		columnName, ok := columnPairs[n].(string)
		if !ok {
			return nil, errNewF("variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
		}
		gct, ok := columnPairs[n+1].(GoColumnType)
		if !ok {
			return nil, errNewF("variadic parameter 'columnPairs' expected an element at index %v to be of type ora.GoColumnType", n+1)
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
		return nil, err
	}
	// qry
	rset, err := stmt.Qry() // TODO: add params for query?
	if err != nil {
		defer stmt.Close()
		return nil, err
	}
	rset.autoClose = true
	return rset, nil
}

// Ins composes, prepares and executes a sql INSERT statement returning a
// possible error.
//
// Ins offers convenience when specifying a long list of sql columns.
//
// Specify hasReturning to true to generate a RETURNING clause with the last
// column name-value pair. When specifying hasReturning as true the last value
// is expected to be a pointer capable of receiving the Oracle column value.
func (ses *Ses) Ins(tbl string, hasReturning bool, columnPairs ...interface{}) (err error) {
	if tbl == "" {
		return errNew("tbl is empty")
	}
	if len(columnPairs) == 0 {
		return errNew("no column name-value pairs specified")
	}
	if len(columnPairs)%2 != 0 {
		return errNew("variadic parameter 'columnPairs' received an odd number of elements. Parameter 'columnPairs' expects an even number of elements")
	}
	if hasReturning && len(columnPairs) == 1 {
		return errNew("len columnPairs must be greater than 1 when hasReturning is true.")
	}
	// build INSERT statement and params slice
	params := make([]interface{}, len(columnPairs)/2)
	buf := new(bytes.Buffer)
	buf.WriteString("INSERT INTO ")
	buf.WriteString(tbl)
	buf.WriteString(" (")
	if hasReturning {
		// returning clause
		lastColName := ""
		for p := 0; p < len(params); p++ {
			n := p * 2
			columnName, ok := columnPairs[n].(string)
			if !ok {
				return errNewF("variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
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
	} else {
		// no returning clause
		for p := 0; p < len(params); p++ {
			n := p * 2
			columnName, ok := columnPairs[n].(string)
			if !ok {
				return errNewF("variadic parameter 'columnPairs' expected an element at index %v to be of type string", n)
			}
			buf.WriteString(columnName)
			if p != len(params)-1 {
				buf.WriteString(", ")
			}
			params[p] = columnPairs[n+1]
		}
		buf.WriteString(") VALUES (")
		for n := 1; n <= len(params); n++ {
			buf.WriteString(fmt.Sprintf(":%v", n))
			if n != len(params) {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(")")
	}
	// prep
	stmt, err := ses.Prep(buf.String())
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exe(params...)
	return err
}

// StartTx starts an Oracle transaction returning a *Tx and possible error.
func (ses *Ses) StartTx() (*Tx, error) {
	if err := ses.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vS%vS%v] StartTx", ses.srv.env.id, ses.srv.id, ses.id)
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
		return nil, ses.srv.env.ociError()
	}

	// set tx struct
	tx := ses.srv.env.drv.txPool.Get().(*Tx)
	if tx.id == 0 {
		ses.txId++
		tx.id = ses.txId
	}
	tx.ses = ses
	tx.elem = ses.txs.PushFront(tx)

	return tx, nil
}

// Sets the StmtCfg on the Session and all open Session Statements.
func (ses *Ses) SetStmtCfg(c StmtCfg) {
	ses.stmtCfg = c
	for e := ses.stmts.Front(); e != nil; e = e.Next() {
		e.Value.(*Stmt).Cfg = c
	}
}

// StmtCfg returns a *StmtCfg.
func (ses *Ses) StmtCfg() *StmtCfg {
	return &ses.stmtCfg
}
