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
	server     *Server
	element    *list.Element
	statements *list.List
	username   string
	ocises     *C.OCISession

	statementConfig StatementConfig
}

// Prepare readies a sql statement associated with a session and returns a *Statement.
func (session *Session) Prepare(sql string, goColumnTypes ...GoColumnType) (*Statement, error) {

	// Validate that the session is open
	err := session.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate a statement handle
	statementHandle, err := session.server.environment.allocateOciHandle(C.OCI_HTYPE_STMT)
	if err != nil {
		return nil, err
	}

	// Prepare sql text with statement handle
	sqlp := C.CString(sql)
	defer C.free(unsafe.Pointer(sqlp))
	r := C.OCIStmtPrepare(
		(*C.OCIStmt)(statementHandle),      // OCIStmt       *stmtp,
		session.server.environment.ocierr,  // OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(sqlp)), // const OraText *stmt,
		C.ub4(C.strlen(sqlp)),              // ub4           stmt_len,
		C.OCI_NTV_SYNTAX,                   // ub4           language,
		C.OCI_DEFAULT)                      // ub4           mode );
	if r == C.OCI_ERROR {
		return nil, session.server.environment.ociError()
	}

	// Get statement from pool
	statement := session.server.environment.statementPool.Get().(*Statement)
	statement.session = session
	statement.sql = sql
	statement.goColumnTypes = goColumnTypes
	statement.ocistmt = (*C.OCIStmt)(statementHandle)
	statement.Config = session.statementConfig

	// Determine statement type
	err = statement.attr(unsafe.Pointer(&statement.stmtType), 4, C.OCI_ATTR_STMT_TYPE)
	if err != nil {
		err2 := statement.Close()
		if err2 != nil {
			return nil, err2
		}
		return nil, err
	}

	// Add statement to session list; store element for later statement removal
	statement.element = session.statements.PushBack(statement)

	return statement, nil
}

// BeginTransaction starts a transaction and returns a *Transaction.
func (session *Session) BeginTransaction() (*Transaction, error) {
	// Validate that the session is open
	err := session.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// the number of seconds the transaction can be inactive
	// before it is automatically terminated by the system.
	var timeout C.uword = C.uword(60)
	r := C.OCITransStart(
		session.server.ocisvcctx,          //OCISvcCtx    *svchp,
		session.server.environment.ocierr, //OCIError     *errhp,
		timeout,         //uword        timeout,
		C.OCI_TRANS_NEW) //ub4          flags );
	if r == C.OCI_ERROR {
		return nil, session.server.environment.ociError()
	}
	return &Transaction{session: session}, nil
}

// CommitTransaction commits a transaction.
func (session *Session) CommitTransaction() error {
	// Validate that the session is open
	err := session.checkIsOpen()
	if err != nil {
		return err
	}

	r := C.OCITransCommit(
		session.server.ocisvcctx,          //OCISvcCtx    *svchp,
		session.server.environment.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)                     //ub4          flags );
	if r == C.OCI_ERROR {
		return session.server.environment.ociError()
	}
	return nil
}

// RollbackTransaction rolls back a transaction.
func (session *Session) RollbackTransaction() error {
	// Validate that the session is open
	err := session.checkIsOpen()
	if err != nil {
		return err
	}

	r := C.OCITransRollback(
		session.server.ocisvcctx,          //OCISvcCtx    *svchp,
		session.server.environment.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)                     //ub4          flags );
	if r == C.OCI_ERROR {
		return session.server.environment.ociError()
	}
	return nil
}

// checkIsOpen validates that a session is open.
//
// ErrClosedSession is returned if a session is closed.
func (session *Session) checkIsOpen() error {
	if !session.IsOpen() {
		return errNew("open session prior to method call")
	}
	return nil
}

// IsOpen returns true when a session is open; otherwise, false.
//
// Calling Close will cause Session.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Server.OpenSession to open a new session.
func (session *Session) IsOpen() bool {
	return session.ocises != nil
}

// Close ends a session on an Oracle server.
//
// Any open statements associated with the session are closed.
//
// Calling Close will cause Session.IsOpen to return false. Once closed, a session
// cannot be re-opened. Call Server.OpenSession to open a new session.
func (session *Session) Close() error {
	if session.IsOpen() {
		// Close statements
		for e := session.statements.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Statement).Close()
			if err != nil {
				return err
			}
		}

		// Close session
		r := C.OCISessionEnd(
			session.server.ocisvcctx,          //OCISvcCtx       *svchp,
			session.server.environment.ocierr, //OCIError        *errhp,
			session.ocises,                    //OCISession      *usrhp,
			C.OCI_DEFAULT)                     //ub4             mode );
		if r == C.OCI_ERROR {
			return session.server.environment.ociError()
		}
		// OCISessionEnd invalidates oci session handle; no need to free session.ocises

		// Remove session from server list
		if session.element != nil {
			session.server.sessions.Remove(session.element)
		}

		// Clear session fields
		server := session.server
		session.server = nil
		session.element = nil
		session.username = ""
		session.ocises = nil

		// Put session in pool
		server.environment.sessionPool.Put(session)
	}
	return nil
}

// Sets the StatementConfig on the Session and all open Session Statements.
func (session *Session) SetStatementConfig(c StatementConfig) {
	session.statementConfig = c
	for e := session.statements.Front(); e != nil; e = e.Next() {
		e.Value.(*Statement).Config = c
	}
}

// StatementConfig returns a *StatementConfig.
func (session *Session) StatementConfig() *StatementConfig {
	return &session.statementConfig
}
