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
	"fmt"
	"unsafe"
)

// An Oracle server associated with an environment.
type Server struct {
	ocisvcctx  *C.OCISvcCtx
	ocisrv     *C.OCIServer
	stmtConfig StatementConfig

	env    *Environment
	elem   *list.Element
	sess   *list.List
	dbname string
}

// OpenSession opens a session on an Oracle server and returns a *Session.
func (srv *Server) OpenSession(username string, password string) (*Session, error) {
	// Validate that the server is open
	err := srv.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate session handle
	ocises, err := srv.env.allocateOciHandle(C.OCI_HTYPE_SESSION)
	if err != nil {
		return nil, err
	}

	// Set username on session handle
	usernamep := C.CString(username)
	defer C.free(unsafe.Pointer(usernamep))
	err = srv.env.setOciAttribute(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(usernamep), C.ub4(C.strlen(usernamep)), C.OCI_ATTR_USERNAME)
	if err != nil {
		return nil, err
	}

	// Set password on session handle
	passwordp := C.CString(password)
	defer C.free(unsafe.Pointer(passwordp))
	err = srv.env.setOciAttribute(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(passwordp), C.ub4(C.strlen(passwordp)), C.OCI_ATTR_PASSWORD)
	if err != nil {
		return nil, err
	}
	// Set driver name on the session handle
	// Driver name is specified to aid diagnostics
	// Driver name will be visible in V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO
	gop := C.CString(fmt.Sprintf("GO %v", DriverVersion))
	defer C.free(unsafe.Pointer(gop))
	err = srv.env.setOciAttribute(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(gop), C.ub4(C.strlen(gop)), C.OCI_ATTR_DRIVER_NAME)
	if err != nil {
		return nil, err
	}

	// Begin session
	r := C.OCISessionBegin(
		srv.ocisvcctx,           //OCISvcCtx     *svchp,
		srv.env.ocierr,          //OCIError      *errhp,
		(*C.OCISession)(ocises), //OCISession    *usrhp,
		C.OCI_CRED_RDBMS,        //ub4           credt,
		C.OCI_DEFAULT)           //ub4           mode );
	if r == C.OCI_ERROR {
		return nil, srv.env.ociError()
	}
	// Set session handle on service context handle
	err = srv.env.setOciAttribute(unsafe.Pointer(srv.ocisvcctx), C.OCI_HTYPE_SVCCTX, ocises, C.ub4(0), C.OCI_ATTR_SESSION)
	if err != nil {
		return nil, err
	}

	// Get session from pool
	ses := srv.env.sesPool.Get().(*Session)
	ses.ocises = (*C.OCISession)(ocises)
	ses.srv = srv
	ses.username = username
	ses.stmtConfig = srv.stmtConfig

	// Add session to server list; store element for later session removal
	ses.elem = srv.sess.PushBack(ses)

	return ses, nil
}

// Ping return nil when an Oracle server is contacted; otherwise, an error.
//
// Ping requires the server have at least one open session.
func (srv *Server) Ping() error {
	err := srv.checkIsOpen()
	if err != nil {
		return err
	}
	r := C.OCIPing(
		srv.ocisvcctx,  //OCISvcCtx     *svchp,
		srv.env.ocierr, //OCIError      *errhp,
		C.OCI_DEFAULT)  //ub4           mode );
	if r == C.OCI_ERROR {
		return srv.env.ociError()
	}
	return nil
}

// Version returns the Oracle database server version.
//
// Version requires the server have at least one open session.
func (srv *Server) Version() (string, error) {
	err := srv.checkIsOpen()
	if err != nil {
		return "", err
	}
	var buffer [512]C.char
	r := C.OCIServerVersion(
		unsafe.Pointer(srv.ocisrv),               //void         *hndlp,
		srv.env.ocierr,                           //OCIError     *errhp,
		(*C.OraText)(unsafe.Pointer(&buffer[0])), //OraText      *bufp,
		C.ub4(len(buffer)),                       //ub4          bufsz
		C.OCI_HTYPE_SERVER)                       //ub1          hndltype );
	if r == C.OCI_ERROR {
		return "", srv.env.ociError()
	}
	return C.GoString(&buffer[0]), nil
}

// checkIsOpen validates that the server is open.
//
// ErrClosedServer is returned if the server is closed.
func (srv *Server) checkIsOpen() error {
	if !srv.IsOpen() {
		return errNew("open Server prior to method call")
	}
	return nil
}

// IsOpen returns true when the server is open; otherwise, false.
//
// Calling Close will cause Server.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Environment.OpenServer to open a new server.
func (srv *Server) IsOpen() bool {
	return srv.ocisrv != nil
}

// Close disconnects from an Oracle server.
//
// Any open sessions associated with the server are closed.
//
// Calling Close will cause Server.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Environment.OpenServer to open a new server.
func (srv *Server) Close() error {
	if srv.IsOpen() {

		// Close sessions
		for e := srv.sess.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Session).Close()
			if err != nil {
				return err
			}
		}
		// Detach server
		r := C.OCIServerDetach(
			srv.ocisrv,     //OCIServer   *srvhp,
			srv.env.ocierr, //OCIError    *errhp,
			C.OCI_DEFAULT)  //ub4         mode );
		if r == C.OCI_ERROR {
			return srv.env.ociError()
		}
		// OCIServerDetach invalidates oci server handle; no need to free server.ocisvr
		// OCIServerDetach invalidates oci service context handle; no need to free server.ocisvcctx

		// Remove server from environment list
		if srv.elem != nil {
			srv.env.srvs.Remove(srv.elem)
		}

		// Clear server fields
		// srv.sess is cleared by previous calls to logoff all sessions
		env := srv.env
		srv.env = nil
		srv.elem = nil
		srv.dbname = ""
		srv.ocisrv = nil
		srv.ocisvcctx = nil

		// Put server in pool
		env.srvPool.Put(srv)
	}

	return nil
}

// Sets the StatementConfig on the Server and all open Server Sessions.
func (srv *Server) SetStatementConfig(c StatementConfig) {
	srv.stmtConfig = c
	for e := srv.sess.Front(); e != nil; e = e.Next() {
		e.Value.(*Session).SetStatementConfig(c)
	}
}

// StatementConfig returns a *StatementConfig.
func (srv *Server) StatementConfig() *StatementConfig {
	return &srv.stmtConfig
}
