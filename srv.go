// Copyright 2015 Rana Ian. All rights reserved.
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
	"sync/atomic"
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

	// Version determines whether the Srv.Version method is logged.
	//
	// The default is true.
	Version bool
}

// NewLogSrvCfg creates a LogSrvCfg with default values.
func NewLogSrvCfg() LogSrvCfg {
	c := LogSrvCfg{}
	c.Close = true
	c.OpenSes = true
	c.Version = true
	return c
}

// Srv represents an Oracle server.
type Srv struct {
	id     uint64
	cfg    SrvCfg
	mu     sync.Mutex
	env    *Env
	ocisrv *C.OCIServer
	isUTF8 int32

	openSess *sesList

	sysNamer
}

// Close disconnects from an Oracle server.
//
// Any open sessions associated with the server are closed.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) Close() (err error) {
	if srv == nil {
		return nil
	}
	srv.mu.Lock()
	if srv.env == nil {
		srv.mu.Unlock()
		return nil
	}
	srv.env.mu.Lock()
	srv.env.openSrvs.remove(srv)
	srv.env.mu.Unlock()
	defer srv.mu.Unlock()
	return srv.close()
}

// close disconnects from an Oracle server, without holding locks.
// Does not remove Srv from Ses.openSrvs
func (srv *Srv) close() (err error) {
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
		srv.openSess.clear()
		srv.env = nil
		srv.ocisrv = nil
		_drv.srvPool.Put(srv)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()

	srv.openSess.closeAll(errs) // close sessions

	// detach server
	r := C.OCIServerDetach(
		srv.ocisrv,     //OCIServer   *srvhp,
		srv.env.ocierr, //OCIError    *errhp,
		C.OCI_DEFAULT)  //ub4         mode );
	if r == C.OCI_ERROR {
		errs.PushBack(errE(srv.env.ociError()))
	}
	err = srv.env.freeOciHandle(unsafe.Pointer(srv.ocisrv), C.OCI_HTYPE_SERVER)
	if err != nil {
		return errE(err)
	}

	return nil
}

// OpenSes opens an Oracle session returning a *Ses and possible error.
func (srv *Srv) OpenSes(cfg *SesCfg) (ses *Ses, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	if srv == nil {
		return nil, er("srv may not be nil.")
	}
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
	// allocate service context handle
	ocisvcctx, err := srv.env.allocOciHandle(C.OCI_HTYPE_SVCCTX)
	if err != nil {
		return nil, errE(err)
	}
	// set server handle onto service context handle
	err = srv.env.setAttr(ocisvcctx, C.OCI_HTYPE_SVCCTX, unsafe.Pointer(srv.ocisrv), C.ub4(0), C.OCI_ATTR_SERVER)
	if err != nil {
		return nil, errE(err)
	}
	//srv.logF(true, "CRED_EXT? %t username=%q", credentialType == C.OCI_CRED_EXT, username)
	// set driver name on the session handle
	// driver name is specified to aid diagnostics; max 9 single-byte characters
	// driver name will be visible in V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO as CLIENT_DRIVER
	drvName := fmt.Sprintf("GO%s", Version)
	cDrvName := C.CString(drvName)
	defer C.free(unsafe.Pointer(cDrvName))
	if err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION,
		unsafe.Pointer(cDrvName), C.ub4(len(drvName)), C.OCI_ATTR_DRIVER_NAME,
	); err != nil {
		return nil, errE(err)
	}
	// http://docs.oracle.com/cd/B28359_01/appdev.111/b28395/oci07lob.htm#CHDDHFAB
	// Set LOB prefetch size to chunk size
	lobPrefetchSize := C.ub4(lobChunkSize)
	if err = srv.env.setAttr(ocises, C.OCI_HTYPE_SESSION,
		unsafe.Pointer(&lobPrefetchSize), C.ub4(0), C.OCI_ATTR_DEFAULT_LOBPREFETCH_SIZE,
	); err != nil {
		return nil, errE(err)
	}

	mode := C.ub4(C.OCI_DEFAULT)
	switch cfg.Mode {
	case SysDba:
		mode |= C.OCI_SYSDBA
	case SysOper:
		mode |= C.OCI_SYSOPER
	}
	// begin session
	r := C.OCISessionBegin(
		(*C.OCISvcCtx)(ocisvcctx), //OCISvcCtx     *svchp,
		srv.env.ocierr,            //OCIError      *errhp,
		(*C.OCISession)(ocises),   //OCISession    *usrhp,
		credentialType,            //ub4           credt,
		mode)                      //ub4           mode );
	if r == C.OCI_ERROR {
		return nil, errE(srv.env.ociError())
	}
	// set session handle on service context handle
	err = srv.env.setAttr(unsafe.Pointer(ocisvcctx), C.OCI_HTYPE_SVCCTX, ocises, C.ub4(0), C.OCI_ATTR_SESSION)
	if err != nil {
		return nil, errE(err)
	}
	// set stmt cache size to zero
	// https://docs.oracle.com/database/121/LNOCI/oci09adv.htm#LNOCI16655
	stmtCacheSize := C.ub4(0)
	err = srv.env.setAttr(unsafe.Pointer(ocisvcctx), C.OCI_HTYPE_SVCCTX, unsafe.Pointer(&stmtCacheSize), C.ub4(0), C.OCI_ATTR_STMTCACHESIZE)
	if err != nil {
		return nil, errE(err)
	}

	ses = _drv.sesPool.Get().(*Ses) // set *Ses
	ses.mu.Lock()
	ses.srv = srv
	ses.ocisvcctx = (*C.OCISvcCtx)(ocisvcctx)
	ses.ocises = (*C.OCISession)(ocises)
	if ses.id == 0 {
		ses.id = _drv.sesId.nextId()
	}
	ses.cfg = *cfg
	if ses.cfg.StmtCfg == nil && ses.srv.cfg.StmtCfg != nil {
		ses.cfg.StmtCfg = &(*ses.srv.cfg.StmtCfg) // copy by value so that user may change independently
	}
	srv.openSess.add(ses)
	ses.mu.Unlock()

	return ses, nil
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

// NumSes returns the number of open Oracle sessions.
func (srv *Srv) NumSes() int {
	if srv == nil {
		return 0
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.openSess.len()
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

// IsUTF8 returns whether the DB uses AL32UTF8 encoding.
func (srv *Srv) IsUTF8() bool {
	if srv == nil {
		return false
	}
	return atomic.LoadInt32(&srv.isUTF8) == 1
}

// IsOpen returns true when the server is open; otherwise, false.
//
// Calling Close will cause Srv.IsOpen to return false. Once closed, a server cannot
// be re-opened. Call Env.OpenSrv to open a new server.
func (srv *Srv) IsOpen() bool {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.checkClosed() == nil
}

// checkClosed returns an error if Srv is closed. No locking occurs.
func (srv *Srv) checkClosed() error {
	if srv == nil || srv.ocisrv == nil {
		return er("Srv is closed.")
	}
	return srv.env.checkClosed()
}

// sysName returns a string representing the Ses.
func (srv *Srv) sysName() string {
	if srv == nil {
		return "E_S_"
	}
	return srv.sysNamer.Name(func() string { return fmt.Sprintf("%sS%v", srv.env.sysName(), srv.id) })
}

// log writes a message with an Srv system name and caller info.
func (srv *Srv) log(enabled bool, v ...interface{}) {
	logL(srv.sysName(), enabled, v...)
}

// log writes a formatted message with an Srv system name and caller info.
func (srv *Srv) logF(enabled bool, format string, v ...interface{}) {
	logF(srv.sysName(), enabled, format, v...)
}
