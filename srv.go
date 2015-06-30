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
	"sync"
	"unsafe"
)

// SrvCfg configures a new Srv.
type SrvCfg struct {
	// Dblink specifies an Oracle database server. Dblink is a connect string
	// or a service point.
	Dblink string

	// StmtCfg configures new Stmts.
	StmtCfg *StmtCfg
}

// NewSrvCfg creates a SrvCfg with default values.
func NewSrvCfg() *SrvCfg {
	c := &SrvCfg{}
	c.StmtCfg = NewStmtCfg()
	return c
}

// LogSrvCfg represents Srv logging configuration values.
type LogSrvCfg struct {
	// Close determines whether the Srv.Close method is logged.
	//
	// The default is true.
	Close bool

	// OpenSes determines whether the Srv.OpenSes method is logged.
	//
	// The default is true.
	OpenSes bool

	// Ping determines whether the Srv.Ping method is logged.
	//
	// The default is true.
	Ping bool

	// Version determines whether the Srv.Version method is logged.
	//
	// The default is true.
	Version bool

	// Break determines whether the Srv.Break method is logged.
	//
	// The default is true.
	Break bool
}

// NewLogSrvCfg creates a LogSrvCfg with default values.
func NewLogSrvCfg() LogSrvCfg {
	c := LogSrvCfg{}
	c.Close = true
	c.OpenSes = true
	c.Ping = true
	c.Version = true
	c.Break = true
	return c
}

// Srv represents an Oracle server.
type Srv struct {
	id        uint64
	cfg       SrvCfg
	mu        sync.Mutex
	env       *Env
	ocisvcctx *C.OCISvcCtx
	ocisrv    *C.OCIServer
	dbIsUTF8  bool

	openSess *list.List
	elem     *list.Element
}

// Close disconnects from an Oracle server.
//
// Any open sessions associated with the server are closed.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) Close() (err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.log(_drv.cfg.Log.Srv.Close)
	err = srv.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			errs.PushBack(errR(value))
		}
		env := srv.env
		env.openSrvs.Remove(srv.elem)
		srv.openSess.Init()
		srv.env = nil
		srv.ocisrv = nil
		srv.ocisvcctx = nil
		srv.elem = nil
		_drv.srvPool.Put(srv)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()

	// close sessions
	for e := srv.openSess.Front(); e != nil; e = e.Next() {
		err = e.Value.(*Ses).Close()
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	// detach server
	// OCIServerDetach invalidates oci server handle; no need to free server.ocisvr
	// OCIServerDetach invalidates oci service context handle; no need to free server.ocisvcctx
	r := C.OCIServerDetach(
		srv.ocisrv,     //OCIServer   *srvhp,
		srv.env.ocierr, //OCIError    *errhp,
		C.OCI_DEFAULT)  //ub4         mode );
	if r == C.OCI_ERROR {
		errs.PushBack(errE(srv.env.ociError()))
	}
	return nil
}

// OpenSes opens an Oracle session returning a *Ses and possible error.
func (srv *Srv) OpenSes(cfg *SesCfg) (ses *Ses, err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.log(_drv.cfg.Log.Srv.OpenSes)
	err = srv.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	if cfg == nil {
		return nil, er("Parameter 'cfg' may not be nil.")
	}
	// allocate session handle
	ocises, err := srv.env.allocOciHandle(C.OCI_HTYPE_SESSION)
	if err != nil {
		return nil, errE(err)
	}
	credentialType := C.ub4(C.OCI_CRED_EXT)
	if cfg.Username != "" || cfg.Password != "" {
		credentialType = C.OCI_CRED_RDBMS
		// set username on session handle
		cUsername := C.CString(cfg.Username)
		defer C.free(unsafe.Pointer(cUsername))
		err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cUsername), C.ub4(len(cfg.Username)), C.OCI_ATTR_USERNAME)
		if err != nil {
			return nil, errE(err)
		}
		// set password on session handle
		cPassword := C.CString(cfg.Password)
		defer C.free(unsafe.Pointer(cPassword))
		err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cPassword), C.ub4(len(cfg.Password)), C.OCI_ATTR_PASSWORD)
		if err != nil {
			return nil, errE(err)
		}
	}
	//srv.logF(true, "CRED_EXT? %t username=%q", credentialType == C.OCI_CRED_EXT, username)
	// set driver name on the session handle
	// driver name is specified to aid diagnostics; max 9 single-byte characters
	// driver name will be visible in V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO
	drvName := fmt.Sprintf("GO %v", Version)
	cDrvName := C.CString(drvName)
	defer C.free(unsafe.Pointer(cDrvName))
	err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(cDrvName), C.ub4(len(drvName)), C.OCI_ATTR_DRIVER_NAME)
	if err != nil {
		return nil, errE(err)
	}
	// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci07lob.htm#CHDDHFAB
	// Set LOB prefetch size to chunk size
	lobPrefetchSize := C.ub4(lobChunkSize)
	err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION, unsafe.Pointer(&lobPrefetchSize), C.ub4(0), C.OCI_ATTR_DEFAULT_LOBPREFETCH_SIZE)
	if err != nil {
		return nil, errE(err)
	}
	// begin session
	r := C.OCISessionBegin(
		srv.ocisvcctx,           //OCISvcCtx     *svchp,
		srv.env.ocierr,          //OCIError      *errhp,
		(*C.OCISession)(ocises), //OCISession    *usrhp,
		credentialType,          //ub4           credt,
		C.OCI_DEFAULT)           //ub4           mode );
	if r == C.OCI_ERROR {
		return nil, errE(srv.env.ociError())
	}
	// set session handle on service context handle
	err = srv.env.setAttr(unsafe.Pointer(srv.ocisvcctx), C.OCI_HTYPE_SVCCTX, ocises, C.ub4(0), C.OCI_ATTR_SESSION)
	if err != nil {
		return nil, errE(err)
	}
	// set stmt cache size to zero
	// https://docs.oracle.com/database/121/LNOCI/oci09adv.htm#LNOCI16655
	stmtCacheSize := C.ub4(0)
	err = srv.env.setAttr(unsafe.Pointer(srv.ocisvcctx), C.OCI_HTYPE_SVCCTX, unsafe.Pointer(&stmtCacheSize), C.ub4(0), C.OCI_ATTR_STMTCACHESIZE)
	if err != nil {
		return nil, errE(err)
	}

	ses = _drv.sesPool.Get().(*Ses) // set *Ses
	ses.srv = srv
	ses.ocises = (*C.OCISession)(ocises)
	ses.elem = srv.openSess.PushBack(ses)
	if ses.id == 0 {
		ses.id = _drv.sesId.nextId()
	}
	ses.cfg = *cfg
	if ses.cfg.StmtCfg == nil && ses.srv.cfg.StmtCfg != nil {
		ses.cfg.StmtCfg = &(*ses.srv.cfg.StmtCfg) // copy by value so that user may change independently
	}

	return ses, nil
}

// Ping return nil when an Oracle server is contacted; otherwise, an error.
//
// Ping requires the server have at least one open session.
func (srv *Srv) Ping() (err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.log(_drv.cfg.Log.Srv.Ping)
	err = srv.checkClosed()
	if err != nil {
		return errE(err)
	}
	r := C.OCIPing(
		srv.ocisvcctx,  //OCISvcCtx     *svchp,
		srv.env.ocierr, //OCIError      *errhp,
		C.OCI_DEFAULT)  //ub4           mode );
	if r == C.OCI_ERROR {
		return errE(srv.env.ociError())
	}
	return nil
}

// Version returns the Oracle database server version.
//
// Version requires the server have at least one open session.
func (srv *Srv) Version() (ver string, err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.log(_drv.cfg.Log.Srv.Version)
	err = srv.checkClosed()
	if err != nil {
		return "", errE(err)
	}
	var buf [512]C.char
	r := C.OCIServerVersion(
		unsafe.Pointer(srv.ocisrv),            //void         *hndlp,
		srv.env.ocierr,                        //OCIError     *errhp,
		(*C.OraText)(unsafe.Pointer(&buf[0])), //OraText      *bufp,
		C.ub4(len(buf)),                       //ub4          bufsz
		C.OCI_HTYPE_SERVER)                    //ub1          hndltype );
	if r == C.OCI_ERROR {
		return "", errE(srv.env.ociError())
	}
	return C.GoString(&buf[0]), nil
}

// Break the currently running OCI function.
func (srv *Srv) Break() (err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.log(_drv.cfg.Log.Srv.Break)
	err = srv.checkClosed()
	if err != nil {
		return errE(err)
	}
	r := C.OCIBreak(unsafe.Pointer(srv.ocisvcctx), srv.env.ocierr)
	if r == C.OCI_ERROR {
		return errE(srv.env.ociError())
	}
	return nil
}

// NumSes returns the number of open Oracle sessions.
func (srv *Srv) NumSes() int {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.openSess.Len()
}

// SetCfg applies the specified cfg to the Srv.
//
// Open Sess do not observe the specified cfg.
func (srv *Srv) SetCfg(cfg SrvCfg) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.cfg = cfg
}

// Cfg returns the Srv's cfg.
func (srv *Srv) Cfg() *SrvCfg {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return &srv.cfg
}

// IsOpen returns true when the server is open; otherwise, false.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) IsOpen() bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.ocisrv != nil
}

// checkClosed returns an error if Srv is closed. No locking occurs.
func (srv *Srv) checkClosed() error {
	if srv.ocisrv == nil {
		return er("Srv is closed.")
	}
	return nil
}

// sysName returns a string representing the Ses.
func (srv *Srv) sysName() string {
	if srv == nil {
		return "E_S_"
	}
	return srv.env.sysName() + fmt.Sprintf("S%v", srv.id)
}

// log writes a message with an Srv system name and caller info.
func (srv *Srv) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", srv.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", srv.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with an Srv system name and caller info.
func (srv *Srv) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", srv.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", srv.sysName(), callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}
