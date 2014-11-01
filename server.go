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

// An Oracle server associated with an environment.
type Server struct {
	environment *Environment
	element     *list.Element
	sessions    *list.List
	dbname      string
	ocisvcctx   *C.OCISvcCtx
	ocisvr      *C.OCIServer

	statementConfig StatementConfig
}

// OpenSession opens a session on an Oracle server and returns a *Session.
func (server *Server) OpenSession(username string, password string) (*Session, error) {
	// Validate that the server is open
	err := server.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate session handle
	//OCIHandleAlloc((void  *)envhp, (void  **)&usrhp, (ub4)OCI_HTYPE_SESSION, (size_t) 0, (void  **) 0);
	sessionHandle, err := server.environment.allocateOciHandle(C.OCI_HTYPE_SESSION)
	if err != nil {
		return nil, err
	}

	// Set username on session handle
	//OCIAttrSet((void  *)usrhp, (ub4)OCI_HTYPE_SESSION, (void  *)"hr",(ub4)strlen("hr"), OCI_ATTR_USERNAME, errhp);
	usernamep := C.CString(username)
	defer C.free(unsafe.Pointer(usernamep))
	err = server.environment.setOciAttribute(sessionHandle, C.OCI_HTYPE_SESSION, unsafe.Pointer(usernamep), C.ub4(C.strlen(usernamep)), C.OCI_ATTR_USERNAME)
	if err != nil {
		return nil, err
	}

	// Set password on session handle
	//OCIAttrSet((void  *)usrhp, (ub4)OCI_HTYPE_SESSION, (void  *)"hr", (ub4)strlen("hr"), OCI_ATTR_PASSWORD, errhp);
	passwordp := C.CString(password)
	defer C.free(unsafe.Pointer(passwordp))
	err = server.environment.setOciAttribute(sessionHandle, C.OCI_HTYPE_SESSION, unsafe.Pointer(passwordp), C.ub4(C.strlen(passwordp)), C.OCI_ATTR_PASSWORD)
	if err != nil {
		return nil, err
	}
	// Set driver name on the session handle
	// Driver name is specified to aid diagnostics
	// Driver name will be visible in V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO
	// OCIAttrSet(authp, OCI_HTYPE_SESSION, client_driver, (ub4)(strlen(client_driver)), OCI_ATTR_DRIVER_NAME, errhp)
	gop := C.CString("GO")
	defer C.free(unsafe.Pointer(gop))
	err = server.environment.setOciAttribute(sessionHandle, C.OCI_HTYPE_SESSION, unsafe.Pointer(gop), C.ub4(C.strlen(gop)), C.OCI_ATTR_DRIVER_NAME)
	if err != nil {
		return nil, err
	}

	// Begin session
	//OCISessionBegin (svchp, errhp, usrhp, OCI_CRED_RDBMS, OCI_DEFAULT);
	r := C.OCISessionBegin(
		server.ocisvcctx,               //OCISvcCtx     *svchp,
		server.environment.ocierr,      //OCIError      *errhp,
		(*C.OCISession)(sessionHandle), //OCISession    *usrhp,
		C.OCI_CRED_RDBMS,               //ub4           credt,
		C.OCI_DEFAULT)                  //ub4           mode );
	if r == C.OCI_ERROR {
		return nil, server.environment.ociError()
	}
	// Set session handle on service context handle
	//OCIAttrSet((void  *)svchp, (ub4)OCI_HTYPE_SVCCTX, (void  *)usrhp,(ub4)0, OCI_ATTR_SESSION, errhp);
	err = server.environment.setOciAttribute(unsafe.Pointer(server.ocisvcctx), C.OCI_HTYPE_SVCCTX, sessionHandle, C.ub4(0), C.OCI_ATTR_SESSION)
	if err != nil {
		return nil, err
	}

	// Get session from pool
	session := server.environment.sessionPool.Get().(*Session)
	session.server = server
	session.username = username
	session.ocises = (*C.OCISession)(sessionHandle)
	session.statementConfig = server.statementConfig

	// Add session to server list; store element for later session removal
	session.element = server.sessions.PushBack(session)

	return session, nil
}

// Ping return nil when an Oracle server is contacted; otherwise, an error.
//
// Ping requires the server have at least one open session.
func (server *Server) Ping() error {
	err := server.checkIsOpen()
	if err != nil {
		return err
	}
	r := C.OCIPing(
		server.ocisvcctx,          //OCISvcCtx     *svchp,
		server.environment.ocierr, //OCIError      *errhp,
		C.OCI_DEFAULT)             //ub4           mode );
	if r == C.OCI_ERROR {
		return server.environment.ociError()
	}
	return nil
}

// Version returns the Oracle database server version.
//
// Version requires the server have at least one open session.
func (server *Server) Version() (string, error) {
	err := server.checkIsOpen()
	if err != nil {
		return "", err
	}
	var buffer [512]C.char
	r := C.OCIServerVersion(
		unsafe.Pointer(server.ocisvr),            //void         *hndlp,
		server.environment.ocierr,                //OCIError     *errhp,
		(*C.OraText)(unsafe.Pointer(&buffer[0])), //OraText      *bufp,
		C.ub4(len(buffer)),                       //ub4          bufsz
		C.OCI_HTYPE_SERVER)                       //ub1          hndltype );
	if r == C.OCI_ERROR {
		return "", server.environment.ociError()
	}
	return C.GoString(&buffer[0]), nil
}

// checkIsOpen validates that the server is open.
//
// ErrClosedServer is returned if the server is closed.
func (server *Server) checkIsOpen() error {
	if !server.IsOpen() {
		return errNew("open Server prior to method call")
	}
	return nil
}

// IsOpen returns true when the server is open; otherwise, false.
//
// Calling Close will cause Server.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Environment.OpenServer to open a new server.
func (server *Server) IsOpen() bool {
	return server.ocisvr != nil
}

// Close disconnects from an Oracle server.
//
// Any open sessions associated with the server are closed.
//
// Calling Close will cause Server.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Environment.OpenServer to open a new server.
func (server *Server) Close() error {
	if server.IsOpen() {

		// Close sessions
		for e := server.sessions.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Session).Close()
			if err != nil {
				return err
			}
		}
		// Detach server
		r := C.OCIServerDetach(
			server.ocisvr,             //OCIServer   *srvhp,
			server.environment.ocierr, //OCIError    *errhp,
			C.OCI_DEFAULT)             //ub4         mode );
		if r == C.OCI_ERROR {
			return server.environment.ociError()
		}
		// OCIServerDetach invalidates oci server handle; no need to free server.ocisvr
		// OCIServerDetach invalidates oci service context handle; no need to free server.ocisvcctx

		// Remove server from environment list
		if server.element != nil {
			server.environment.servers.Remove(server.element)
		}

		// Clear server fields
		// server.sessions is cleared by previous calls to logoff all sessions
		environment := server.environment
		server.environment = nil
		server.element = nil
		server.dbname = ""
		server.ocisvr = nil
		server.ocisvcctx = nil

		// Put server in pool
		environment.serverPool.Put(server)
	}

	return nil
}

// Sets the StatementConfig on the Server and all open Server Sessions.
func (server *Server) SetStatementConfig(c StatementConfig) {
	server.statementConfig = c
	for e := server.sessions.Front(); e != nil; e = e.Next() {
		e.Value.(*Session).SetStatementConfig(c)
	}
}

// StatementConfig returns a *StatementConfig.
func (server *Server) StatementConfig() *StatementConfig {
	return &server.statementConfig
}
