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
	"fmt"
	"unsafe"
)

// Srv is an Oracle server associated with an environment.
type Srv struct {
	id        uint64
	env       *Env
	ocisvcctx *C.OCISvcCtx
	ocisrv    *C.OCIServer

	sesId   uint64
	sess    *list.List
	elem    *list.Element
	stmtCfg StmtCfg
	dbname  string
}

// NumSes returns the number of open Oracle sessions.
func (srv *Srv) NumSes() int {
	return srv.sess.Len()
}

// checkIsOpen validates that the server is open.
func (srv *Srv) checkIsOpen() error {
	if !srv.IsOpen() {
		return errNewF("Srv is closed (id %v)", srv.id)
	}
	return nil
}

// IsOpen returns true when the server is open; otherwise, false.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) IsOpen() bool {
	return srv.env != nil
}

// Close disconnects from an Oracle server.
//
// Any open sessions associated with the server are closed.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) Close() (err error) {
	if err := srv.checkIsOpen(); err != nil {
		return err
	}
	Log.Infof("E%vS%v] Close", srv.env.id, srv.id)
	errs := srv.env.drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			Log.Errorln(recoverMsg(value))
			errs.PushBack(errRecover(value))
		}

		env := srv.env
		env.srvs.Remove(srv.elem)
		srv.sess.Init()
		srv.env = nil
		srv.ocisrv = nil
		srv.ocisvcctx = nil
		srv.elem = nil
		srv.dbname = ""
		env.drv.srvPool.Put(srv)

		m := newMultiErrL(errs)
		if m != nil {
			err = *m
		}
		errs.Init()
		env.drv.listPool.Put(errs)
	}()

	// close sessions
	for e := srv.sess.Front(); e != nil; e = e.Next() {
		err0 := e.Value.(*Ses).Close()
		errs.PushBack(err0)
	}
	// detach server
	// OCIServerDetach invalidates oci server handle; no need to free server.ocisvr
	// OCIServerDetach invalidates oci service context handle; no need to free server.ocisvcctx
	r := C.OCIServerDetach(
		srv.ocisrv,     //OCIServer   *srvhp,
		srv.env.ocierr, //OCIError    *errhp,
		C.OCI_DEFAULT)  //ub4         mode );
	if r == C.OCI_ERROR {
		errs.PushBack(srv.env.ociError())
	}

	return err
}

// OpenSes opens an Oracle session returning a *Ses and possible error.
func (srv *Srv) OpenSes(username string, password string) (*Ses, error) {
	if err := srv.checkIsOpen(); err != nil {
		return nil, err
	}
	Log.Infof("E%vS%v] OpenSes (username %v)", srv.env.id, srv.id, username)
	// allocate session handle
	ocises, err := srv.env.allocOciHandle(C.OCI_HTYPE_SESSION)
	if err != nil {
		return nil, err
	}
	credentialType := C.ub4(C.OCI_CRED_EXT)
	if username != "" || password != "" {
		credentialType = C.OCI_CRED_RDBMS
		// set username on session handle
		cUsername := C.CString(username)
		defer C.free(unsafe.Pointer(cUsername))
		err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cUsername), C.ub4(len(username)), C.OCI_ATTR_USERNAME)
		if err != nil {
			return nil, err
		}
		// set password on session handle
		cPassword := C.CString(password)
		defer C.free(unsafe.Pointer(cPassword))
		err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cPassword), C.ub4(len(password)), C.OCI_ATTR_PASSWORD)
		if err != nil {
			return nil, err
		}
	}
	// set driver name on the session handle
	// driver name is specified to aid diagnostics; max 9 single-byte characters
	// driver name will be visible in V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO
	drvName := fmt.Sprintf("GO %v", Version)
	cDrvName := C.CString(drvName)
	defer C.free(unsafe.Pointer(cDrvName))
	err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cDrvName), C.ub4(len(drvName)), C.OCI_ATTR_DRIVER_NAME)
	if err != nil {
		return nil, err
	}
	Log.Infof("CRED_EXT? %t username=%q", credentialType == C.OCI_CRED_EXT, username)
	// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci07lob.htm#CHDDHFAB
	// Set LOB prefetch size to chunk size
	lobPrefetchSize := C.ub4(lobChunkSize)
	err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(&lobPrefetchSize), 0, C.OCI_ATTR_DEFAULT_LOBPREFETCH_SIZE)
	if err != nil {
		return nil, err
	}

	// begin session
	r := C.OCISessionBegin(
		srv.ocisvcctx,           //OCISvcCtx     *svchp,
		srv.env.ocierr,          //OCIError      *errhp,
		(*C.OCISession)(ocises), //OCISession    *usrhp,
		credentialType,          //ub4           credt,
		C.OCI_DEFAULT)           //ub4           mode );
	if r == C.OCI_ERROR {
		return nil, srv.env.ociError()
	}
	// set session handle on service context handle
	err = srv.env.setAttr(unsafe.Pointer(srv.ocisvcctx), C.OCI_HTYPE_SVCCTX, ocises, C.ub4(0), C.OCI_ATTR_SESSION)
	if err != nil {
		return nil, err
	}

	// set ses struct
	ses := srv.env.drv.sesPool.Get().(*Ses)
	if ses.id == 0 {
		srv.sesId++
		ses.id = srv.sesId
	}
	ses.srv = srv
	ses.ocises = (*C.OCISession)(ocises)
	ses.username = username
	ses.stmtCfg = srv.stmtCfg
	ses.elem = srv.sess.PushBack(ses)

	return ses, nil
}

// Ping return nil when an Oracle server is contacted; otherwise, an error.
//
// Ping requires the server have at least one open session.
func (srv *Srv) Ping() error {
	if err := srv.checkIsOpen(); err != nil {
		return err
	}
	Log.Infof("E%vS%v] Ping", srv.env.id, srv.id)
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
func (srv *Srv) Version() (string, error) {
	if err := srv.checkIsOpen(); err != nil {
		return "", err
	}
	Log.Infof("E%vS%v] Version", srv.env.id, srv.id)
	var buf [512]C.char
	r := C.OCIServerVersion(
		unsafe.Pointer(srv.ocisrv),            //void         *hndlp,
		srv.env.ocierr,                        //OCIError     *errhp,
		(*C.OraText)(unsafe.Pointer(&buf[0])), //OraText      *bufp,
		C.ub4(len(buf)),                       //ub4          bufsz
		C.OCI_HTYPE_SERVER)                    //ub1          hndltype );
	if r == C.OCI_ERROR {
		return "", srv.env.ociError()
	}
	return C.GoString(&buf[0]), nil
}

// Sets the StmtCfg on the Server and all open Server Sessions.
func (srv *Srv) SetStmtCfg(c StmtCfg) {
	srv.stmtCfg = c
	for e := srv.sess.Front(); e != nil; e = e.Next() {
		e.Value.(*Ses).SetStmtCfg(c)
	}
}

// StmtCfg returns a *StmtCfg.
func (srv *Srv) StmtCfg() *StmtCfg {
	return &srv.stmtCfg
}

// Break the currently running OCI function.
func (srv *Srv) Break() {
	Log.Infof("E%vS%v] Break", srv.env.id, srv.id)
	C.OCIBreak(unsafe.Pointer(srv.ocisvcctx), srv.env.ocierr)
}
