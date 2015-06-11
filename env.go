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
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

// Env is an Oracle environment.
type Env struct {
	id     uint64
	drv    *Drv
	ocienv *C.OCIEnv
	ocierr *C.OCIError

	srvs     *list.List
	cons     *list.List
	elem     *list.Element
	stmtCfg  StmtCfg
	errBuf   [512]C.char
	isSqlPkg bool

	// LogClose determines whether the Env.Close method is logged.
	//
	// The default is true.
	LogClose bool

	// LogOpenSrv determines whether the Env.OpenSrv method is logged.
	//
	// The default is true.
	LogOpenSrv bool
}

// sysName returns a name which represents the current environment.
func (env *Env) sysName() string {
	return fmt.Sprintf("E%v", env.id)
}

// log writes a message with caller info.
func (env *Env) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			Log.Infof("%v %v", env.sysName(), callInfo(1))
		} else {
			Log.Infof("%v %v %v", env.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with caller info.
func (env *Env) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			Log.Infof("%v %v", env.sysName(), callInfo(1))
		} else {
			Log.Infof("%v %v %v", env.sysName(), callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}

// err creates an error with caller info.
func (env *Env) err(v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprint(v...)))
	Log.Errorln(err)
	return err
}

// errF creates a formatted error with caller info.
func (env *Env) errF(format string, v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprintf(format, v...)))
	Log.Errorln(err)
	return err
}

// errE wraps an error with caller info.
func (env *Env) errE(e error) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), e.Error()))
	Log.Errorln(err)
	return err
}

// NumSrv returns the number of open Oracle servers.
func (env *Env) NumSrv() int {
	return env.srvs.Len()
}

// NumCon returns the number of open Oracle connections.
func (env *Env) NumCon() int {
	return env.cons.Len()
}

// checkIsOpen validates that the environment is open.
func (env *Env) checkIsOpen() error {
	if !env.IsOpen() {
		return env.err("Env is closed.")
	}
	return nil
}

// IsOpen returns true when the environment is open; otherwise, false.
//
// Calling Close will cause IsOpen to return false.
// Once closed, the environment may be re-opened by
// calling Open.
func (env *Env) IsOpen() bool {
	return env.drv != nil
}

// Close disconnects from servers and resets optional fields.
func (env *Env) Close() (err error) {
	env.log(env.LogClose)
	if env.IsOpen() {
		errs := env.drv.listPool.Get().(*list.List)
		defer func() {
			if value := recover(); value != nil {
				Log.Errorln(recoverMsg(value))
				errs.PushBack(errRecover(value))
			}
			drv := env.drv
			drv.envs.Remove(env.elem)
			env.srvs.Init()
			env.drv = nil
			env.ocienv = nil
			env.ocierr = nil
			env.elem = nil
			drv.envPool.Put(env)
			multiErr := newMultiErrL(errs)
			if multiErr != nil {
				err = env.errE(*multiErr)
			}
			errs.Init()
			drv.listPool.Put(errs)
		}()
		for e := env.cons.Front(); e != nil; e = e.Next() { // close connections
			err0 := e.Value.(*Con).Close()
			errs.PushBack(err0)
		}
		for e := env.srvs.Front(); e != nil; e = e.Next() { // close servers
			err0 := e.Value.(*Srv).Close()
			errs.PushBack(err0)
		}
		// Free oci environment handle and all oci child handles
		// The oci error handle is released as a child of the environment handle
		err = env.freeOciHandle(unsafe.Pointer(env.ocienv), C.OCI_HTYPE_ENV)
		if err != nil {
			return env.errE(err)
		}
	}
	return nil
}

// OpenSrv connects to an Oracle server returning a *Srv and possible error.
func (env *Env) OpenSrv(dbname string) (*Srv, error) {
	env.logF(env.LogOpenSrv, "(dbname %v)", dbname)
	if err := env.checkIsOpen(); err != nil {
		return nil, env.errE(err)
	}
	// allocate server handle
	ocisrv, err := env.allocOciHandle(C.OCI_HTYPE_SERVER)
	if err != nil {
		return nil, env.errE(err)
	}
	// attach to server
	cDbname := C.CString(dbname)
	defer C.free(unsafe.Pointer(cDbname))
	r := C.OCIServerAttach(
		(*C.OCIServer)(ocisrv),                //OCIServer     *srvhp,
		env.ocierr,                            //OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(cDbname)), //const OraText *dbname,
		C.sb4(len(dbname)),                    //sb4           dbname_len,
		C.OCI_DEFAULT)                         //ub4           mode);
	if r == C.OCI_ERROR {
		return nil, env.ociError()
	}
	// allocate service context handle
	ocisvcctx, err := env.allocOciHandle(C.OCI_HTYPE_SVCCTX)
	if err != nil {
		return nil, env.errE(err)
	}
	// set server handle onto service context handle
	err = env.setAttr(ocisvcctx, C.OCI_HTYPE_SVCCTX, ocisrv, C.ub4(0), C.OCI_ATTR_SERVER)
	if err != nil {
		return nil, env.errE(err)
	}

	// set srv struct
	srv := env.drv.srvPool.Get().(*Srv)
	if srv.id == 0 {
		srv.id = _drv.srvId.nextId()
	}
	srv.env = env
	srv.ocisrv = (*C.OCIServer)(ocisrv)
	srv.ocisvcctx = (*C.OCISvcCtx)(ocisvcctx)
	srv.stmtCfg = env.stmtCfg
	srv.dbname = dbname
	srv.elem = env.srvs.PushBack(srv)
	return srv, nil
}

// OpenCon starts an Oracle session on a server returning a *Con and possible error.
//
// The connection string has the form username/password@dbname e.g., scott/tiger@orcl
// dbname is a connection identifier such as a net service name,
// full connection identifier, or a simple connection identifier.
// The dbname may be defined in the client machine's tnsnames.ora file.
func (env *Env) OpenCon(str string) (*Con, error) {
	if err := env.checkIsOpen(); err != nil {
		return nil, env.errE(err)
	}
	Log.Infof("E%v] OpenCon", env.id)
	// parse connection string
	var username string
	var password string
	var dbname string
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "/@") {
		dbname = str[2:]
	} else {
		str = strings.Replace(str, "/", " / ", 1)
		str = strings.Replace(str, "@", " @ ", 1)
		_, err := fmt.Sscanf(str, "%s / %s @ %s", &username, &password, &dbname)
		Log.Infof("E%v] OpenCon (dbname %v, username %v)", env.id, dbname, username)
		if err != nil {
			return nil, env.errE(err)
		}
	}
	// connect to server
	srv, err := env.OpenSrv(dbname)
	if err != nil {
		return nil, env.errE(err)
	}
	// open a session on the server
	ses, err := srv.OpenSes(username, password)
	if err != nil {
		return nil, env.errE(err)
	}
	// set con struct
	con := env.drv.conPool.Get().(*Con)
	if con.id == 0 {
		con.id = _drv.conId.nextId()
	}
	con.env = env
	con.srv = srv
	con.ses = ses
	con.elem = env.cons.PushBack(con)

	return con, nil
}

// allocateOciHandle allocates an oci handle.
func (env *Env) allocOciHandle(handleType C.ub4) (unsafe.Pointer, error) {
	// OCIHandleAlloc returns: OCI_SUCCESS, OCI_INVALID_HANDLE
	var handle unsafe.Pointer
	r := C.OCIHandleAlloc(
		unsafe.Pointer(env.ocienv), //const void    *parenth,
		&handle,                    //void          **hndlpp,
		handleType,                 //ub4           type,
		C.size_t(0),                //size_t        xtramem_sz,
		nil)                        //void          **usrmempp
	if r == C.OCI_INVALID_HANDLE {
		return nil, errNew("Unable to allocate handle")
	}
	return handle, nil
}

// freeOciHandle deallocates an oci handle.
func (env *Env) freeOciHandle(ociHandle unsafe.Pointer, handleType C.ub4) error {
	// OCIHandleFree returns: OCI_SUCCESS, OCI_INVALID_HANDLE, or OCI_ERROR
	r := C.OCIHandleFree(
		ociHandle,  //void      *hndlp,
		handleType) //ub4       type );
	if r == C.OCI_INVALID_HANDLE {
		return errNew("Unable to free handle")
	} else if r == C.OCI_ERROR {
		return env.ociError()
	}

	return nil
}

// setOciAttribute sets an attribute value on a handle or descriptor.
func (env *Env) setAttr(
	target unsafe.Pointer,
	targetType C.ub4,
	attribute unsafe.Pointer,
	attributeSize C.ub4,
	attributeType C.ub4) (err error) {

	r := C.OCIAttrSet(
		target,        //void        *trgthndlp,
		targetType,    //ub4         trghndltyp,
		attribute,     //void        *attributep,
		attributeSize, //ub4         size,
		attributeType, //ub4         attrtype,
		env.ocierr)    //OCIError    *errhp );
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

// getOciError gets an error returned by an Oracle server.
func (env *Env) ociError() error {
	var errcode C.sb4
	C.OCIErrorGet(
		unsafe.Pointer(env.ocierr),
		1, nil,
		&errcode,
		(*C.OraText)(unsafe.Pointer(&env.errBuf[0])),
		512,
		C.OCI_HTYPE_ERROR)
	return errNew(C.GoString(&env.errBuf[0]))
}

// Sets the StmtCfg on the Environment and all open Environment Servers.
func (env *Env) SetStmtCfg(c StmtCfg) {
	env.stmtCfg = c
	for e := env.srvs.Front(); e != nil; e = e.Next() {
		e.Value.(*Srv).SetStmtCfg(c)
	}
}

// StmtCfg returns a *StmtCfg.
func (env *Env) StmtCfg() *StmtCfg {
	return &env.stmtCfg
}
