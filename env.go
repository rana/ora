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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// LogEnvCfg represents Env logging configuration values.
type LogEnvCfg struct {
	// Close determines whether the Env.Close method is logged.
	//
	// The default is true.
	Close bool

	// OpenSrv determines whether the Env.OpenSrv method is logged.
	//
	// The default is true.
	OpenSrv bool

	// OpenCon determines whether the Env.OpenCon method is logged.
	//
	// The default is true.
	OpenCon bool
}

// NewLogEnvCfg creates a LogEnvCfg with default values.
func NewLogEnvCfg() LogEnvCfg {
	c := LogEnvCfg{}
	c.Close = true
	c.OpenSrv = true
	c.OpenCon = true
	return c
}

// Env represents an Oracle environment.
type Env struct {
	sync.RWMutex

	id       uint64
	cmu      sync.Mutex
	cfg      atomic.Value
	ocienv   *C.OCIEnv
	ocierr   *C.OCIError
	errBuf   [512]C.char
	ociHndMu sync.Mutex
	isPkgEnv bool

	openSrvs *srvList
	openCons *conList

	sysNamer
}

func (env *Env) Cfg() StmtCfg {
	c := env.cfg.Load()
	if c == nil || c.(StmtCfg).IsZero() {
		return _drv.Cfg().StmtCfg
	}
	return c.(StmtCfg)
}
func (env *Env) SetCfg(cfg StmtCfg) {
	env.cfg.Store(cfg)
}

// Close disconnects from servers and resets optional fields.
func (env *Env) Close() (err error) {
	env.log(_drv.Cfg().Log.Env.Close)
	err = env.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)

	env.cmu.Lock()
	defer env.cmu.Unlock()

	env.RLock()
	openSrvs, openCons := env.openSrvs, env.openCons
	env.RUnlock()

	defer func() {
		if value := recover(); value != nil {
			errs.PushBack(errR(value))
		}
		_drv.openEnvs.remove(env)
		env.SetCfg(StmtCfg{})
		env.Lock()
		env.isPkgEnv = false
		env.ocienv = nil
		env.ocierr = nil
		env.Unlock()
		openSrvs.clear()
		openCons.clear()
		_drv.envPool.Put(env)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()
	openCons.closeAll(errs)
	openSrvs.closeAll(errs)

	// Free oci environment handle and all oci child handles
	env.RLock()
	env.freeOciHandle(unsafe.Pointer(env.ocierr), C.OCI_HTYPE_ERROR)
	env.freeOciHandle(unsafe.Pointer(env.ocienv), C.OCI_HTYPE_ENV)
	env.RUnlock()
	return nil
}

// OpenSrv connects to an Oracle server returning a *Srv and possible error.
func (env *Env) OpenSrv(cfg SrvCfg) (srv *Srv, err error) {
	if cfg.IsZero() {
		cfg.StmtCfg = env.Cfg()
		if cfg.IsZero() {
			panic("Parameter 'cfg' may not be empty.")
		}
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	env.log(_drv.Cfg().Log.Env.OpenSrv)
	err = env.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	// allocate server handle
	ocisrv, err := env.allocOciHandle(C.OCI_HTYPE_SERVER)
	if err != nil {
		return nil, errE(err)
	}
	// attach to server
	cDblink := C.CString(cfg.Dblink)
	defer func() { C.free(unsafe.Pointer(cDblink)) }()

	var (
		poolName           *C.OraText
		poolNameLen        C.ub4
		ocipool            unsafe.Pointer
		username, password *C.char
	)

	switch cfg.Pool.Type {
	case CPool, SPool, DRCPool:
		username, password = C.CString(cfg.Pool.Username), C.CString(cfg.Pool.Password)
		defer func() {
			C.free(unsafe.Pointer(username))
			C.free(unsafe.Pointer(password))
		}()
	}

	switch cfg.Pool.Type {
	case CPool:
		if ocipool, err = env.allocOciHandle(C.OCI_HTYPE_CPOOL); err != nil {
			env.freeOciHandle(unsafe.Pointer(ocisrv), C.OCI_HTYPE_SERVER)
			return nil, errE(err)
		}
		var pnl C.sb4
		env.RLock()
		r := C.OCIConnectionPoolCreate(
			env.ocienv,                               // OCIEnv           *envhp,
			env.ocierr,                               //           OCIError         *errhp,
			(*C.OCICPool)(ocipool),                   //                       OCISPool         *spoolhp,
			(**C.OraText)(unsafe.Pointer(&poolName)), //                        OraText          **poolName,
			&pnl, //                       ub4              *poolNameLen,
			(*C.OraText)(unsafe.Pointer(cDblink)),  //                        CONST OraText    *connStr,
			C.sb4(len(cfg.Dblink)),                 //                        ub4              connStrLen,
			C.ub4(cfg.Pool.Min),                    //                        ub4              sessMin,
			C.ub4(cfg.Pool.Max),                    //                        ub4              sessMax,
			C.ub4(cfg.Pool.Incr),                   //                        ub4              sessIncr,
			(*C.OraText)(unsafe.Pointer(username)), //     OraText          *userid,
			C.sb4(len(cfg.Pool.Username)),          //                        ub4              useridLen,
			(*C.OraText)(unsafe.Pointer(password)), // OraText          *password,
			C.sb4(len(cfg.Pool.Password)),          //            ub4              passwordLen,
			C.OCI_DEFAULT,                          //                        ub4              mode
		)
		env.RUnlock()
		if r == C.OCI_ERROR {
			err := env.ociError()
			env.log(_drv.Cfg().Log.Env.OpenSrv, fmt.Sprintf("ConnectionPoolCreate(u=%q p=%q link=%q): %+v", cfg.Pool.Username, cfg.Pool.Password, cfg.Dblink, err))
			env.freeOciHandle(unsafe.Pointer(ocisrv), C.OCI_HTYPE_SERVER)
			env.freeOciHandle(ocipool, C.OCI_HTYPE_CPOOL)
			return nil, errE(err)
		}
		poolNameLen = C.ub4(pnl)

	case SPool, DRCPool:
		ocipool, err := env.allocOciHandle(C.OCI_HTYPE_SPOOL)
		if err != nil {
			env.freeOciHandle(unsafe.Pointer(ocisrv), C.OCI_HTYPE_SERVER)
			return nil, errE(err)
		}

		env.RLock()
		r := C.OCISessionPoolCreate(
			env.ocienv,                               // OCIEnv           *envhp,
			env.ocierr,                               //           OCIError         *errhp,
			(*C.OCISPool)(ocipool),                   //                       OCISPool         *spoolhp,
			(**C.OraText)(unsafe.Pointer(&poolName)), //                        OraText          **poolName,
			&poolNameLen,                             //                       ub4              *poolNameLen,
			(*C.OraText)(unsafe.Pointer(cDblink)),    //                        CONST OraText    *connStr,
			C.ub4(len(cfg.Dblink)),                   //                        ub4              connStrLen,
			C.ub4(cfg.Pool.Min),                      //                        ub4              sessMin,
			C.ub4(cfg.Pool.Max),                      //                        ub4              sessMax,
			C.ub4(cfg.Pool.Incr),                     //                        ub4              sessIncr,
			(*C.OraText)(unsafe.Pointer(username)),   //     OraText          *userid,
			C.ub4(len(cfg.Pool.Username)),            //                        ub4              useridLen,
			(*C.OraText)(unsafe.Pointer(password)),   // OraText          *password,
			C.ub4(len(cfg.Pool.Password)),            //            ub4              passwordLen,
			C.OCI_DEFAULT,                            //                        ub4              mode
		)
		env.RUnlock()
		if r == C.OCI_ERROR {
			err := env.ociError()
			env.log(_drv.Cfg().Log.Env.OpenSrv, fmt.Sprintf("SessionPoolCreate(u=%q p=%q link=%q): %+v", cfg.Pool.Username, cfg.Pool.Password, cfg.Dblink, err))
			env.freeOciHandle(unsafe.Pointer(ocisrv), C.OCI_HTYPE_SERVER)
			env.freeOciHandle(ocipool, C.OCI_HTYPE_SPOOL)
			return nil, errE(err)
		}

	default:
		env.RLock()
		r := C.OCIServerAttach(
			(*C.OCIServer)(ocisrv),                //OCIServer     *srvhp,
			env.ocierr,                            //OCIError      *errhp,
			(*C.OraText)(unsafe.Pointer(cDblink)), //const OraText *dblink,
			C.sb4(len(cfg.Dblink)),                //sb4           dblink_len,
			C.OCI_DEFAULT)                         //ub4           mode);
		env.RUnlock()
		if r == C.OCI_ERROR {
			env.freeOciHandle(unsafe.Pointer(ocisrv), C.OCI_HTYPE_SERVER)
			return nil, errE(env.ociError())
		}
	}

	srv = _drv.srvPool.Get().(*Srv) // set *Srv
	srv.cmu.Lock()
	defer srv.cmu.Unlock()
	srv.Lock()
	srv.env = env
	srv.ocisrv = (*C.OCIServer)(ocisrv)
	if srv.id == 0 {
		srv.id = _drv.srvId.nextId()
	}
	srv.ocipool = ocipool
	srv.poolType = cfg.Pool.Type
	srv.ociPoolName, srv.ociPoolNameLen = poolName, poolNameLen
	srv.Unlock()
	srv.SetCfg(cfg)
	if cfg.StmtCfg.IsZero() && !srv.env.Cfg().IsZero() {
		cfg.StmtCfg = srv.env.Cfg()
		srv.SetCfg(cfg)
	}
	env.RLock()
	env.openSrvs.add(srv)
	env.RUnlock()

	return srv, nil
}

var (
	conCharset   = make(map[string]string, 2)
	conCharsetMu sync.Mutex
)

// OpenCon starts an Oracle session on a server returning a *Con and possible error.
//
// The connection string has the form username/password@dblink e.g., scott/tiger@orcl
// For connecting as SYSDBA or SYSOPER, append " AS SYSDBA" to the end of the connection string: "sys/sys as sysdba".
//
// dblink is a connection identifier such as a net service name,
// full connection identifier, or a simple connection identifier.
// The dblink may be defined in the client machine's tnsnames.ora file.
func (env *Env) OpenCon(dsn string) (con *Con, err error) {
	// do not lock; calls to env.OpenSrv will lock
	env.log(_drv.Cfg().Log.Env.OpenCon)
	err = env.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	dsn = strings.TrimSpace(dsn)

	srvCfg := SrvCfg{StmtCfg: env.Cfg(), Pool: DSNPool(dsn)}
	sesCfg := SesCfg{Mode: DSNMode(dsn)}
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	srv, err := env.OpenSrv(srvCfg)
	if err != nil {
		env.log(_drv.Cfg().Log.Env.OpenSrv, fmt.Sprintf("OpenSrv(%#v): %+v", srvCfg, err))
		return nil, errE(err)
	}
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		srv.Close()
		return nil, errE(err)
	}

	con = _drv.conPool.Get().(*Con) // set *Con
	con.env = env
	con.ses = ses
	if con.id == 0 {
		con.id = _drv.conId.nextId()
	}
	setUTF8 := func(cs string) {
		var isUTF8 int32
		if cs == "AL32UTF8" {
			isUTF8 = 1
		}
		atomic.StoreInt32(&ses.srv.isUTF8, isUTF8)
	}

	conCharsetMu.Lock()
	cs, ok := conCharset[srvCfg.Dblink]
	conCharsetMu.Unlock()

	if ok {
		setUTF8(cs)
		return con, nil
	}
	if rset, err := ses.PrepAndQry(
		`SELECT property_value FROM database_properties WHERE property_name = 'NLS_CHARACTERSET'`,
	); err != nil {
		//Log.Errorf("E%vS%vS%v] Determine database characterset: %v",
		//	env.id, con.id, ses.id, err)
	} else if rset != nil && rset.Next() && len(rset.Row) == 1 {
		//Log.Infof("E%vS%vS%v] Database characterset=%q",
		//	env.id, con.id, ses.id, rset.Row[0])
		if cs, ok := rset.Row[0].(string); ok {
			conCharsetMu.Lock()
			conCharset[srvCfg.Dblink] = cs
			conCharsetMu.Unlock()
			setUTF8(cs)
		}
	}
	env.RLock()
	env.openCons.add(con)
	env.RUnlock()

	return con, nil
}

// NumSrv returns the number of open Oracle servers.
func (env *Env) NumSrv() int {
	env.RLock()
	n := env.openSrvs.len()
	env.RUnlock()
	return n
}

// NumCon returns the number of open Oracle connections.
func (env *Env) NumCon() int {
	env.RLock()
	n := env.openCons.len()
	env.RUnlock()
	return n
}

// IsOpen returns true when the environment is open; otherwise, false.
//
// Calling Close will cause IsOpen to return false. Once closed, the environment
// may be re-opened by calling Open.
func (env *Env) IsOpen() bool {
	env.RLock()
	ok := env.ocienv != nil
	env.RUnlock()
	return ok
}

// checkClosed returns an error if Env is closed. No locking occurs.
func (env *Env) checkClosed() error {
	if env == nil {
		return er("Env is closed.")
	}
	env.RLock()
	closed := env.ocienv == nil
	env.RUnlock()
	if closed {
		return er("Env is closed.")
	}
	return nil
}

// sysName returns a string representing the Env.
func (env *Env) sysName() string {
	if env == nil {
		return "E_"
	}
	if env.isPkgEnv {
		return "EP"
	}
	return env.sysNamer.Name(func() string { return fmt.Sprintf("E%v", env.id) })
}

// log writes a message with an Env system name and caller info.
func (env *Env) log(enabled bool, v ...interface{}) {
	Log := _drv.Cfg().Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", env.sysName(), callInfo(1))
	} else {
		Log.Logger.Infof("%v %v %v", env.sysName(), callInfo(1), fmt.Sprint(v...))
	}
}

// log writes a formatted message with an Env system name and caller info.
func (env *Env) logF(enabled bool, format string, v ...interface{}) {
	Log := _drv.Cfg().Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", env.sysName(), callInfo(1))
	} else {
		Log.Logger.Infof("%v %v %v", env.sysName(), callInfo(1), fmt.Sprintf(format, v...))
	}
}

// allocateOciHandle allocates an oci handle. No locking occurs.
func (env *Env) allocOciHandle(handleType C.ub4) (unsafe.Pointer, error) {
	env.ociHndMu.Lock()
	defer env.ociHndMu.Unlock()
	// OCIHandleAlloc returns: OCI_SUCCESS, OCI_INVALID_HANDLE
	var handle unsafe.Pointer
	env.RLock()
	r := C.OCIHandleAlloc(
		unsafe.Pointer(env.ocienv), //const void    *parenth,
		&handle,                    //void          **hndlpp,
		handleType,                 //ub4           type,
		C.size_t(0),                //size_t        xtramem_sz,
		nil)                        //void          **usrmempp
	env.RUnlock()
	if r == C.OCI_INVALID_HANDLE {
		return nil, er("Unable to allocate handle")
	}
	return handle, nil
}

// freeOciHandle deallocates an oci handle. No locking occurs.
func (env *Env) freeOciHandle(ociHandle unsafe.Pointer, handleType C.ub4) error {
	var err error
	func() {
		if r := recover(); r != nil {
			err = errR(r)
		}
	}()
	env.ociHndMu.Lock()
	defer env.ociHndMu.Unlock()
	// OCIHandleFree returns: OCI_SUCCESS, OCI_INVALID_HANDLE, or OCI_ERROR
	r := C.OCIHandleFree(
		ociHandle,  //void      *hndlp,
		handleType) //ub4       type );
	if r == C.OCI_INVALID_HANDLE {
		err = er("Unable to free handle")
	} else if r == C.OCI_ERROR {
		err = errE(env.ociError())
	}
	return err
}

// setOciAttribute sets an attribute value on a handle or descriptor. No locking occurs.
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
		return errE(env.ociError())
	}
	return nil
}

// ociError gets an error returned by an Oracle server.
func (env *Env) ociError(prefix ...string) error {
	var errcode C.sb4
	env.RLock()
	C.OCIErrorGet(
		unsafe.Pointer(env.ocierr),
		1, nil,
		&errcode,
		(*C.OraText)(unsafe.Pointer(&env.errBuf[0])),
		C.ub4(len(env.errBuf)),
		C.OCI_HTYPE_ERROR)
	msg := C.GoString(&env.errBuf[0])
	env.RUnlock()
	return er(&ORAError{
		code:    int(errcode),
		prefix:  strings.Join(prefix, " "),
		message: msg,
	})
}

type ORAError struct {
	code            int
	prefix, message string
}

func (e ORAError) Code() int {
	return e.code
}

func (e *ORAError) Error() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		if e.prefix != "" {
			return e.prefix + ": " + e.message
		}
		return e.message
	}
	return fmt.Sprintf("ORA-%05d", e.code)
}

var b8Pool = sync.Pool{
	New: func() interface{} {
		p := unsafe.Pointer(C.malloc(8))
		runtime.SetFinalizer(&p, b8Free)
		return p
	},
}

func b8Free(pp *unsafe.Pointer) {
	if pp != nil && *pp != nil {
		C.free(*pp)
		*pp = nil
	}
}

func (env *Env) OCINumberFromInt(dest *C.OCINumber, value int64, byteLen int) error {
	val := (*C.sb8)(b8Pool.Get().(unsafe.Pointer))
	*val = C.sb8(value)
	r := C.OCINumberFromInt(
		env.ocierr,          //OCIError            *err,
		unsafe.Pointer(val), //const void          *inum,
		C.uword(byteLen),    //uword               inum_length,
		C.OCI_NUMBER_SIGNED, //uword               inum_s_flag,
		dest)                //OCINumber           *number );
	b8Pool.Put(unsafe.Pointer(val))
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

func (env *Env) OCINumberToInt(src *C.OCINumber, byteLen int) (int64, error) {
	val := b8Pool.Get().(unsafe.Pointer)
	r := C.OCINumberToInt(
		env.ocierr,          //OCIError              *err,
		src,                 //const OCINumber       *number,
		C.uword(byteLen),    //uword                 rsl_length,
		C.OCI_NUMBER_SIGNED, //uword                 rsl_flag,
		val)                 //void                  *rsl );
	if r == C.OCI_ERROR {
		return 0, env.ociError()
	}
	var ret int64
	switch byteLen {
	case 1:
		ret = int64(*((*C.sb1)(val)))
	case 2:
		ret = int64(*((*C.sb2)(val)))
	case 4:
		ret = int64(*((*C.sb4)(val)))
	default:
		ret = int64(*((*C.sb8)(val)))
	}
	b8Pool.Put(val)
	return ret, nil
}

func (env *Env) OCINumberFromUint(dest *C.OCINumber, value uint64, byteLen int) error {
	val := (*C.ub8)(b8Pool.Get().(unsafe.Pointer))
	*val = C.ub8(value)
	r := C.OCINumberFromInt(
		env.ocierr,            //OCIError            *err,
		unsafe.Pointer(val),   //const void          *inum,
		C.uword(byteLen),      //uword               inum_length,
		C.OCI_NUMBER_UNSIGNED, //uword               inum_s_flag,
		dest) //OCINumber           *number );
	b8Pool.Put(unsafe.Pointer(val))
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

func (env *Env) OCINumberToUint(src *C.OCINumber, byteLen int) (uint64, error) {
	val := b8Pool.Get().(unsafe.Pointer)
	r := C.OCINumberToInt(
		env.ocierr,            //OCIError              *err,
		src,                   //const OCINumber       *number,
		C.uword(byteLen),      //uword                 rsl_length,
		C.OCI_NUMBER_UNSIGNED, //uword                 rsl_flag,
		val) //void                  *rsl );
	if r == C.OCI_ERROR {
		return 0, env.ociError()
	}
	var ret uint64
	switch byteLen {
	case 1:
		ret = uint64(*(*C.ub1)(val))
	case 2:
		ret = uint64(*(*C.ub2)(val))
	case 4:
		ret = uint64(*(*C.ub4)(val))
	default:
		ret = uint64(*(*C.ub8)(val))
		b8Pool.Put(val)
	}
	return ret, nil
}

func (env *Env) OCINumberFromFloat(dest *C.OCINumber, value float64, byteLen int) error {
	val := (*C.double)(b8Pool.Get().(unsafe.Pointer))
	*val = C.double(value)
	r := C.OCINumberFromReal(
		env.ocierr,          //OCIError            *err,
		unsafe.Pointer(val), //const void          *inum,
		C.uword(byteLen),    //uword               inum_length,
		dest)                //OCINumber           *number );
	b8Pool.Put(unsafe.Pointer(val))
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

func (env *Env) OCINumberToFloat(src *C.OCINumber, byteLen int) (float64, error) {
	val := b8Pool.Get().(unsafe.Pointer)
	r := C.OCINumberToReal(
		env.ocierr,       //OCIError              *err,
		src,              //const OCINumber       *number,
		C.uword(byteLen), //uword                 rsl_length,
		val)              //void                  *rsl );
	if r == C.OCI_ERROR {
		return 0, env.ociError()
	}
	var ret float64
	if byteLen == 4 {
		ret = float64(*(*C.float)(val))
	} else {
		ret = float64(*(*C.double)(val))
	}
	b8Pool.Put(val)
	return ret, nil
}
