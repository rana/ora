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
	"container/list"
	"github.com/golang/glog"
	"unsafe"
)

// Ses is an Oracle session associated with a server.
type Ses struct {
	sesId  uint64
	srv    *Srv
	ocises *C.OCISession

	txId       uint64
	stmtId     uint64
	txs        *list.List
	stmts      *list.List
	elem       *list.Element
	stmtConfig StmtConfig
	username   string
}

// StmtCount returns the number of open Oracle statements.
func (ses *Ses) StmtCount() int {
	return ses.stmts.Len()
}

// TxCount returns the number of open Oracle transactions.
func (ses *Ses) TxCount() int {
	return ses.txs.Len()
}

// checkIsOpen validates that the session is open.
func (ses *Ses) checkIsOpen() error {
	if !ses.IsOpen() {
		return errNewF("Ses is closed (sesId %v)", ses.sesId)
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
	glog.Infof("E%vS%vS%v Close", ses.srv.env.envId, ses.srv.srvId, ses.sesId)
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
	return stmt.Exec(params...)
}

// PrepAndQuery prepares a SQL statement and queries an Oracle server returning
// a *Stmt, *Rset and a possible error.
//
// Call *Stmt.Close when down retieving data from *Rset.
//
// If an error occurs during Prep or Query a nil *Stmt and nil *Rset will be
// returned.
func (ses *Ses) PrepAndQuery(sql string, params ...interface{}) (*Stmt, *Rset, error) {
	stmt, err := ses.Prep(sql)
	if err != nil {
		defer stmt.Close()
		return nil, nil, err
	}
	rset, err := stmt.Query(params...)
	if err != nil {
		defer stmt.Close()
		return nil, nil, err
	}
	return stmt, rset, nil
}

// Prep prepares a sql statement returning a *Stmt and possible error.
func (ses *Ses) Prep(sql string, gcts ...GoColumnType) (*Stmt, error) {
	if err := ses.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vS%vS%v Prep", ses.srv.env.envId, ses.srv.srvId, ses.sesId)
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
	if stmt.stmtId == 0 {
		ses.stmtId++
		stmt.stmtId = ses.stmtId
	}
	glog.Infof("E%vS%vS%v Prep (stmtId %v)", ses.srv.env.envId, ses.srv.srvId, ses.sesId, stmt.stmtId)
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
	stmt.Config = ses.stmtConfig
	stmt.elem = ses.stmts.PushBack(stmt)

	return stmt, nil
}

// StartTx starts an Oracle transaction returning a *Tx and possible error.
func (ses *Ses) StartTx() (*Tx, error) {
	if err := ses.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vS%vS%v StartTx", ses.srv.env.envId, ses.srv.srvId, ses.sesId)
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
	if tx.txId == 0 {
		ses.txId++
		tx.txId = ses.txId
	}
	glog.Infof("E%vS%vS%v StartTx (txId %v)", ses.srv.env.envId, ses.srv.srvId, ses.sesId, tx.txId)
	tx.ses = ses
	tx.elem = ses.txs.PushFront(tx)

	return tx, nil
}

// Sets the StmtConfig on the Session and all open Session Statements.
func (ses *Ses) SetStmtConfig(c StmtConfig) {
	ses.stmtConfig = c
	for e := ses.stmts.Front(); e != nil; e = e.Next() {
		e.Value.(*Stmt).Config = c
	}
}

// StmtConfig returns a *StmtConfig.
func (ses *Ses) StmtConfig() *StmtConfig {
	return &ses.stmtConfig
}
