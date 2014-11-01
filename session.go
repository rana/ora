// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"container/list"
	"unsafe"
)

// A session associated with an Oracle server.
type Session struct {
	ocises     *C.OCISession
	stmtConfig StatementConfig

	srv      *Server
	elem     *list.Element
	stmts    *list.List
	txs      *list.List
	username string
}

// Prepare readies a sql statement associated with a session and returns a *Statement.
func (ses *Session) Prepare(sql string, goColumnTypes ...GoColumnType) (*Statement, error) {

	// Validate that the session is open
	err := ses.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate a statement handle
	ocistmt, err := ses.srv.env.allocateOciHandle(C.OCI_HTYPE_STMT)
	if err != nil {
		return nil, err
	}

	// Prepare sql text with statement handle
	sqlp := C.CString(sql)
	defer C.free(unsafe.Pointer(sqlp))
	r := C.OCIStmtPrepare(
		(*C.OCIStmt)(ocistmt),              // OCIStmt       *stmtp,
		ses.srv.env.ocierr,                 // OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(sqlp)), // const OraText *stmt,
		C.ub4(C.strlen(sqlp)),              // ub4           stmt_len,
		C.OCI_NTV_SYNTAX,                   // ub4           language,
		C.OCI_DEFAULT)                      // ub4           mode );
	if r == C.OCI_ERROR {
		return nil, ses.srv.env.ociError()
	}

	// Get statement from pool
	stmt := ses.srv.env.stmtPool.Get().(*Statement)
	stmt.ses = ses
	stmt.sql = sql
	stmt.goColumnTypes = goColumnTypes
	stmt.ocistmt = (*C.OCIStmt)(ocistmt)
	stmt.Config = ses.stmtConfig

	// Determine statement type
	err = stmt.attr(unsafe.Pointer(&stmt.stmtType), 4, C.OCI_ATTR_STMT_TYPE)
	if err != nil {
		err2 := stmt.Close()
		if err2 != nil {
			return nil, err2
		}
		return nil, err
	}

	// Add statement to session list; store element for later statement removal
	stmt.elem = ses.stmts.PushBack(stmt)

	return stmt, nil
}

// BeginTransaction starts a transaction and returns a *Transaction.
func (ses *Session) BeginTransaction() (*Transaction, error) {
	// Validate that the session is open
	err := ses.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// the number of seconds the transaction can be inactive
	// before it is automatically terminated by the system.
	var timeout C.uword = C.uword(60)
	r := C.OCITransStart(
		ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		ses.srv.env.ocierr, //OCIError     *errhp,
		timeout,            //uword        timeout,
		C.OCI_TRANS_NEW)    //ub4          flags );
	if r == C.OCI_ERROR {
		return nil, ses.srv.env.ociError()
	}
	tx := &Transaction{ses: ses}
	// store tx for stmt to determine if can auto commit
	tx.elem = ses.txs.PushFront(tx)

	return tx, nil
}

// checkIsOpen validates that a session is open.
func (ses *Session) checkIsOpen() error {
	if !ses.IsOpen() {
		return errNew("open session prior to method call")
	}
	return nil
}

// IsOpen returns true when a session is open; otherwise, false.
//
// Calling Close will cause Session.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Server.OpenSession to open a new session.
func (ses *Session) IsOpen() bool {
	return ses.ocises != nil
}

// Close ends a session on an Oracle server.
//
// Any open statements associated with the session are closed.
//
// Calling Close will cause Session.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Server.OpenSession to open a new session.
func (ses *Session) Close() error {
	if ses.IsOpen() {
		// Close statements
		for e := ses.stmts.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Statement).Close()
			if err != nil {
				return err
			}
		}

		// remove transactions from list
		for e := ses.txs.Front(); e != nil; e = e.Next() {
			ses.txs.Remove(e)
		}

		// Close session
		r := C.OCISessionEnd(
			ses.srv.ocisvcctx,  //OCISvcCtx       *svchp,
			ses.srv.env.ocierr, //OCIError        *errhp,
			ses.ocises,         //OCISession      *usrhp,
			C.OCI_DEFAULT)      //ub4             mode );
		if r == C.OCI_ERROR {
			return ses.srv.env.ociError()
		}
		// OCISessionEnd invalidates oci session handle; no need to free session.ocises

		// Remove session from server list
		if ses.elem != nil {
			ses.srv.sess.Remove(ses.elem)
		}

		// Clear session fields
		srv := ses.srv
		ses.ocises = nil
		ses.srv = nil
		ses.elem = nil
		ses.username = ""

		// Put session in pool
		srv.env.sesPool.Put(ses)
	}
	return nil
}

// Sets the StatementConfig on the Session and all open Session Statements.
func (ses *Session) SetStatementConfig(c StatementConfig) {
	ses.stmtConfig = c
	for e := ses.stmts.Front(); e != nil; e = e.Next() {
		e.Value.(*Statement).Config = c
	}
}

// StatementConfig returns a *StatementConfig.
func (ses *Session) StatementConfig() *StatementConfig {
	return &ses.stmtConfig
}
