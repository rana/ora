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

// EnvCfg configures a new Env.
type EnvCfg struct {
	// StmtCfg configures new Stmts.
	StmtCfg *StmtCfg
}

// NewEnvCfg creates a EnvCfg with default values.
func NewEnvCfg() *EnvCfg {
	c := &EnvCfg{}
	c.StmtCfg = NewStmtCfg()
	return c
}

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
	id       uint64
	cfg      EnvCfg
	mu       sync.Mutex
	ocienv   *C.OCIEnv
	ocierr   *C.OCIError
	errBuf   [512]C.char
	ociHndMu sync.Mutex

	openSrvs *srvList
	openCons *conList

	sysNamer
}

// Close disconnects from servers and resets optional fields.
func (env *Env) Close() (err error) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.log(_drv.cfg.Log.Env.Close)
	err = env.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			errs.PushBack(errR(value))
		}
		_drv.openEnvs.remove(env)
		env.ocienv = nil
		env.ocierr = nil
		env.openSrvs.clear()
		env.openCons.clear()
		_drv.envPool.Put(env)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()
	env.openCons.closeAll(errs)
	env.openSrvs.closeAll(errs)

	// Free oci environment handle and all oci child handles
	env.freeOciHandle(unsafe.Pointer(env.ocierr), C.OCI_HTYPE_ERROR)
	env.freeOciHandle(unsafe.Pointer(env.ocienv), C.OCI_HTYPE_ENV)
	return nil
}

// OpenSrv connects to an Oracle server returning a *Srv and possible error.
func (env *Env) OpenSrv(cfg *SrvCfg) (srv *Srv, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	env.mu.Lock()
	defer env.mu.Unlock()
	env.log(_drv.cfg.Log.Env.OpenSrv)
	err = env.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	if cfg == nil {
		return nil, er("Parameter 'cfg' may not be nil.")
	}
	// allocate server handle
	ocisrv, err := env.allocOciHandle(C.OCI_HTYPE_SERVER)
	if err != nil {
		return nil, errE(err)
	}
	// attach to server
	cDblink := C.CString(cfg.Dblink)
	defer C.free(unsafe.Pointer(cDblink))
	r := C.OCIServerAttach(
		(*C.OCIServer)(ocisrv),                //OCIServer     *srvhp,
		env.ocierr,                            //OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(cDblink)), //const OraText *dblink,
		C.sb4(len(cfg.Dblink)),                //sb4           dblink_len,
		C.OCI_DEFAULT)                         //ub4           mode);
	if r == C.OCI_ERROR {
		return nil, errE(env.ociErrorNL())
	}

	srv = _drv.srvPool.Get().(*Srv) // set *Srv
	srv.mu.Lock()
	srv.env = env
	srv.ocisrv = (*C.OCIServer)(ocisrv)
	if srv.id == 0 {
		srv.id = _drv.srvId.nextId()
	}
	srv.cfg = *cfg
	if srv.cfg.StmtCfg == nil && srv.env.cfg.StmtCfg != nil {
		srv.cfg.StmtCfg = &(*srv.env.cfg.StmtCfg) // copy by value so that user may change independently
	}
	srv.mu.Unlock()
	env.openSrvs.add(srv)

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
	env.log(_drv.cfg.Log.Env.OpenCon)
	err = env.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	dsn = strings.TrimSpace(dsn)

	var srvCfg SrvCfg
	sesCfg := SesCfg{Mode: DSNMode(dsn)}
	sesCfg.Username, sesCfg.Password, srvCfg.Dblink = SplitDSN(dsn)
	srv, err := env.OpenSrv(&srvCfg)
	if err != nil {
		return nil, errE(err)
	}
	ses, err := srv.OpenSes(&sesCfg)
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
		ses.mu.Lock()
		var isUTF8 int32
		if cs == "AL32UTF8" {
			isUTF8 = 1
		}
		atomic.StoreInt32(&con.ses.srv.isUTF8, isUTF8)
		ses.mu.Unlock()
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
	env.openCons.add(con)

	return con, nil
}

// NumSrv returns the number of open Oracle servers.
func (env *Env) NumSrv() int {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.openSrvs.len()
}

// NumCon returns the number of open Oracle connections.
func (env *Env) NumCon() int {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.openCons.len()
}

// SetCfg applies the specified cfg to the Env.
//
// Open Srvs do not observe the specified cfg.
func (env *Env) SetCfg(cfg *EnvCfg) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.cfg = *cfg
}

// Cfg returns the Env's cfg.
func (env *Env) Cfg() *EnvCfg {
	env.mu.Lock()
	defer env.mu.Unlock()
	return &env.cfg
}

// IsOpen returns true when the environment is open; otherwise, false.
//
// Calling Close will cause IsOpen to return false. Once closed, the environment
// may be re-opened by calling Open.
func (env *Env) IsOpen() bool {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.ocienv != nil
}

// checkClosed returns an error if Env is closed. No locking occurs.
func (env *Env) checkClosed() error {
	if env == nil || env.ocienv == nil {
		return er("Env is closed.")
	}
	return nil
}

// sysName returns a string representing the Env.
func (env *Env) sysName() string {
	if env == nil {
		return "E_"
	}
	return env.sysNamer.Name(func() string { return fmt.Sprintf("E%v", env.id) })
}

// logL writes a message with an Env system name and caller info.
func logL(nm string, enabled bool, v ...interface{}) {
	Log := _drv.cfg.Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", nm, callInfo(2))
	} else {
		Log.Logger.Infof("%v %v %v", nm, callInfo(2), fmt.Sprint(v...))
	}
}

// logF writes a formatted message with an Env system name and caller info.
func logF(nm string, enabled bool, format string, v ...interface{}) {
	Log := _drv.cfg.Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", nm, callInfo(2))
	} else {
		Log.Logger.Infof("%v %v %v", nm, callInfo(2), fmt.Sprintf(format, v...))
	}
}

// logL writes a message with an Env system name and caller info.
func (env *Env) log(enabled bool, v ...interface{}) {
	logL(env.sysName(), enabled, v...)
}

// logF writes a formatted message with an Env system name and caller info.
func (env *Env) logF(enabled bool, format string, v ...interface{}) {
	logF(env.sysName(), enabled, format, v...)
}

// allocateOciHandle allocates an oci handle. No locking occurs.
func (env *Env) allocOciHandle(handleType C.ub4) (unsafe.Pointer, error) {
	env.ociHndMu.Lock()
	defer env.ociHndMu.Unlock()
	// OCIHandleAlloc returns: OCI_SUCCESS, OCI_INVALID_HANDLE
	var handle unsafe.Pointer
	r := C.OCIHandleAlloc(
		unsafe.Pointer(env.ocienv), //const void    *parenth,
		&handle,                    //void          **hndlpp,
		handleType,                 //ub4           type,
		C.size_t(0),                //size_t        xtramem_sz,
		nil)                        //void          **usrmempp
	if r == C.OCI_INVALID_HANDLE {
		return nil, er("Unable to allocate handle")
	}
	return handle, nil
}

// freeOciHandle deallocates an oci handle. No locking occurs.
func (env *Env) freeOciHandle(ociHandle unsafe.Pointer, handleType C.ub4) error {
	var err error
	defer func() {
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

// ociError gets an error returned by an Oracle server. Locks env.mu!
func (env *Env) ociError() error {
	env.mu.Lock()
	err := env.ociErrorNL()
	env.mu.Unlock()
	return err
}

// ociError gets an error returned by an Oracle server. No locking occurs.
func (env *Env) ociErrorNL() error {
	var errcode C.sb4
	C.OCIErrorGet(
		unsafe.Pointer(env.ocierr),
		1, nil,
		&errcode,
		(*C.OraText)(unsafe.Pointer(&env.errBuf[0])),
		C.ub4(len(env.errBuf)),
		C.OCI_HTYPE_ERROR)
	return er(&ORAError{
		code:    int(errcode),
		message: C.GoString(&env.errBuf[0]),
	})
}

type ORAError struct {
	code    int
	message string
}

func (e ORAError) Code() int {
	return e.code
}

func (e *ORAError) Error() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
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
