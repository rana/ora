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
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// An Oracle environment.
type Environment struct {
	ocienv     *C.OCIEnv
	ocierr     *C.OCIError
	stmtConfig StatementConfig

	srvs     *list.List
	conPool  sync.Pool
	srvPool  sync.Pool
	sesPool  sync.Pool
	stmtPool sync.Pool

	int64BindPool   sync.Pool
	int32BindPool   sync.Pool
	int16BindPool   sync.Pool
	int8BindPool    sync.Pool
	uint64BindPool  sync.Pool
	uint32BindPool  sync.Pool
	uint16BindPool  sync.Pool
	uint8BindPool   sync.Pool
	float64BindPool sync.Pool
	float32BindPool sync.Pool

	int64PtrBindPool   sync.Pool
	int32PtrBindPool   sync.Pool
	int16PtrBindPool   sync.Pool
	int8PtrBindPool    sync.Pool
	uint64PtrBindPool  sync.Pool
	uint32PtrBindPool  sync.Pool
	uint16PtrBindPool  sync.Pool
	uint8PtrBindPool   sync.Pool
	float64PtrBindPool sync.Pool
	float32PtrBindPool sync.Pool

	int64SliceBindPool   sync.Pool
	int32SliceBindPool   sync.Pool
	int16SliceBindPool   sync.Pool
	int8SliceBindPool    sync.Pool
	uint64SliceBindPool  sync.Pool
	uint32SliceBindPool  sync.Pool
	uint16SliceBindPool  sync.Pool
	uint8SliceBindPool   sync.Pool
	float64SliceBindPool sync.Pool
	float32SliceBindPool sync.Pool

	int64DefinePool   sync.Pool
	int32DefinePool   sync.Pool
	int16DefinePool   sync.Pool
	int8DefinePool    sync.Pool
	uint64DefinePool  sync.Pool
	uint32DefinePool  sync.Pool
	uint16DefinePool  sync.Pool
	uint8DefinePool   sync.Pool
	float64DefinePool sync.Pool
	float32DefinePool sync.Pool

	oraInt64DefinePool   sync.Pool
	oraInt32DefinePool   sync.Pool
	oraInt16DefinePool   sync.Pool
	oraInt8DefinePool    sync.Pool
	oraUint64DefinePool  sync.Pool
	oraUint32DefinePool  sync.Pool
	oraUint16DefinePool  sync.Pool
	oraUint8DefinePool   sync.Pool
	oraFloat64DefinePool sync.Pool
	oraFloat32DefinePool sync.Pool

	timeBindPool      sync.Pool
	timePtrBindPool   sync.Pool
	timeSliceBindPool sync.Pool
	timeDefinePool    sync.Pool
	oraTimeDefinePool sync.Pool

	stringBindPool      sync.Pool
	stringPtrBindPool   sync.Pool
	stringSliceBindPool sync.Pool
	stringDefinePool    sync.Pool
	oraStringDefinePool sync.Pool

	bytesBindPool      sync.Pool
	bytesSliceBindPool sync.Pool
	lobDefinePool      sync.Pool
	longRawDefinePool  sync.Pool
	rawDefinePool      sync.Pool
	bfileBindPool      sync.Pool
	bfileDefinePool    sync.Pool
	nilBindPool        sync.Pool

	boolBindPool      sync.Pool
	boolPtrBindPool   sync.Pool
	boolSliceBindPool sync.Pool
	boolDefinePool    sync.Pool
	oraBoolDefinePool sync.Pool

	rowidDefinePool   sync.Pool
	resultSetBindPool sync.Pool

	oraIntervalYMBindPool      sync.Pool
	intervalYMDefinePool       sync.Pool
	oraIntervalYMSliceBindPool sync.Pool

	oraIntervalDSBindPool      sync.Pool
	intervalDSDefinePool       sync.Pool
	oraIntervalDSSliceBindPool sync.Pool
}

// NewEnv creates a new Oracle environment.
func NewEnv() (env *Environment) {
	env = &Environment{srvs: list.New()}
	env.stmtConfig = NewStmtConfig()

	env.conPool.New = func() interface{} {
		return &Connection{env: env}
	}
	env.srvPool.New = func() interface{} {
		return &Server{sess: list.New()}
	}
	env.sesPool.New = func() interface{} {
		return &Session{stmts: list.New()}
	}
	env.stmtPool.New = func() interface{} {
		return &Statement{env: env, rsts: list.New()}
	}

	env.int64BindPool.New = func() interface{} {
		return &int64Bind{env: env}
	}
	env.int32BindPool.New = func() interface{} {
		return &int32Bind{env: env}
	}
	env.int16BindPool.New = func() interface{} {
		return &int16Bind{env: env}
	}
	env.int8BindPool.New = func() interface{} {
		return &int8Bind{env: env}
	}
	env.uint64BindPool.New = func() interface{} {
		return &uint64Bind{env: env}
	}
	env.uint32BindPool.New = func() interface{} {
		return &uint32Bind{env: env}
	}
	env.uint16BindPool.New = func() interface{} {
		return &uint16Bind{env: env}
	}
	env.uint8BindPool.New = func() interface{} {
		return &uint8Bind{env: env}
	}
	env.float64BindPool.New = func() interface{} {
		return &float64Bind{env: env}
	}
	env.float32BindPool.New = func() interface{} {
		return &float32Bind{env: env}
	}

	env.int64PtrBindPool.New = func() interface{} {
		return &int64PtrBind{env: env}
	}
	env.int32PtrBindPool.New = func() interface{} {
		return &int32PtrBind{env: env}
	}
	env.int16PtrBindPool.New = func() interface{} {
		return &int16PtrBind{env: env}
	}
	env.int8PtrBindPool.New = func() interface{} {
		return &int8PtrBind{env: env}
	}
	env.uint64PtrBindPool.New = func() interface{} {
		return &uint64PtrBind{env: env}
	}
	env.uint32PtrBindPool.New = func() interface{} {
		return &uint32PtrBind{env: env}
	}
	env.uint16PtrBindPool.New = func() interface{} {
		return &uint16PtrBind{env: env}
	}
	env.uint8PtrBindPool.New = func() interface{} {
		return &uint8PtrBind{env: env}
	}
	env.float64PtrBindPool.New = func() interface{} {
		return &float64PtrBind{env: env}
	}
	env.float32PtrBindPool.New = func() interface{} {
		return &float32PtrBind{env: env}
	}

	env.int64SliceBindPool.New = func() interface{} {
		return &int64SliceBind{env: env}
	}
	env.int32SliceBindPool.New = func() interface{} {
		return &int32SliceBind{env: env}
	}
	env.int16SliceBindPool.New = func() interface{} {
		return &int16SliceBind{env: env}
	}
	env.int8SliceBindPool.New = func() interface{} {
		return &int8SliceBind{env: env}
	}
	env.uint64SliceBindPool.New = func() interface{} {
		return &uint64SliceBind{env: env}
	}
	env.uint32SliceBindPool.New = func() interface{} {
		return &uint32SliceBind{env: env}
	}
	env.uint16SliceBindPool.New = func() interface{} {
		return &uint16SliceBind{env: env}
	}
	env.uint8SliceBindPool.New = func() interface{} {
		return &uint8SliceBind{env: env}
	}
	env.float64SliceBindPool.New = func() interface{} {
		return &float64SliceBind{env: env}
	}
	env.float32SliceBindPool.New = func() interface{} {
		return &float32SliceBind{env: env}
	}

	env.int64DefinePool.New = func() interface{} {
		return &int64Define{env: env}
	}
	env.int32DefinePool.New = func() interface{} {
		return &int32Define{env: env}
	}
	env.int16DefinePool.New = func() interface{} {
		return &int16Define{env: env}
	}
	env.int8DefinePool.New = func() interface{} {
		return &int8Define{env: env}
	}
	env.uint64DefinePool.New = func() interface{} {
		return &uint64Define{env: env}
	}
	env.uint32DefinePool.New = func() interface{} {
		return &uint32Define{env: env}
	}
	env.uint16DefinePool.New = func() interface{} {
		return &uint16Define{env: env}
	}
	env.uint8DefinePool.New = func() interface{} {
		return &uint8Define{env: env}
	}
	env.float64DefinePool.New = func() interface{} {
		return &float64Define{env: env}
	}
	env.float32DefinePool.New = func() interface{} {
		return &float32Define{env: env}
	}

	env.oraInt64DefinePool.New = func() interface{} {
		return &oraInt64Define{env: env}
	}
	env.oraInt32DefinePool.New = func() interface{} {
		return &oraInt32Define{env: env}
	}
	env.oraInt16DefinePool.New = func() interface{} {
		return &oraInt16Define{env: env}
	}
	env.oraInt8DefinePool.New = func() interface{} {
		return &oraInt8Define{env: env}
	}
	env.oraUint64DefinePool.New = func() interface{} {
		return &oraUint64Define{env: env}
	}
	env.oraUint32DefinePool.New = func() interface{} {
		return &oraUint32Define{env: env}
	}
	env.oraUint16DefinePool.New = func() interface{} {
		return &oraUint16Define{env: env}
	}
	env.oraUint8DefinePool.New = func() interface{} {
		return &oraUint8Define{env: env}
	}
	env.oraFloat64DefinePool.New = func() interface{} {
		return &oraFloat64Define{env: env}
	}
	env.oraFloat32DefinePool.New = func() interface{} {
		return &oraFloat32Define{env: env}
	}

	env.timeBindPool.New = func() interface{} {
		return &timeBind{env: env}
	}
	env.timePtrBindPool.New = func() interface{} {
		return &timePtrBind{env: env}
	}
	env.timeSliceBindPool.New = func() interface{} {
		return &timeSliceBind{env: env}
	}
	env.timeDefinePool.New = func() interface{} {
		return &timeDefine{env: env}
	}
	env.oraTimeDefinePool.New = func() interface{} {
		return &oraTimeDefine{env: env}
	}
	env.stringBindPool.New = func() interface{} {
		return &stringBind{env: env}
	}
	env.stringPtrBindPool.New = func() interface{} {
		return &stringPtrBind{env: env}
	}
	env.stringSliceBindPool.New = func() interface{} {
		return &stringSliceBind{env: env}
	}
	env.stringDefinePool.New = func() interface{} {
		return &stringDefine{env: env}
	}
	env.oraStringDefinePool.New = func() interface{} {
		return &oraStringDefine{env: env}
	}
	env.bytesBindPool.New = func() interface{} {
		return &bytesBind{env: env}
	}
	env.bytesSliceBindPool.New = func() interface{} {
		return &bytesSliceBind{env: env}
	}
	env.lobDefinePool.New = func() interface{} {
		return &lobDefine{env: env}
	}
	env.longRawDefinePool.New = func() interface{} {
		return &longRawDefine{env: env}
	}
	env.rawDefinePool.New = func() interface{} {
		return &rawDefine{env: env}
	}
	env.bfileBindPool.New = func() interface{} {
		return &bfileBind{env: env}
	}
	env.bfileDefinePool.New = func() interface{} {
		return &bfileDefine{env: env}
	}
	env.nilBindPool.New = func() interface{} {
		return &nilBind{env: env}
	}

	env.boolBindPool.New = func() interface{} {
		return &boolBind{env: env}
	}
	env.boolPtrBindPool.New = func() interface{} {
		return &boolPtrBind{env: env}
	}
	env.boolSliceBindPool.New = func() interface{} {
		return &boolSliceBind{env: env}
	}
	env.boolDefinePool.New = func() interface{} {
		return &boolDefine{env: env}
	}
	env.oraBoolDefinePool.New = func() interface{} {
		return &oraBoolDefine{env: env}
	}

	env.rowidDefinePool.New = func() interface{} {
		return &rowidDefine{env: env}
	}
	env.resultSetBindPool.New = func() interface{} {
		return &resultSetBind{env: env}
	}

	env.oraIntervalYMBindPool.New = func() interface{} {
		return &oraIntervalYMBind{env: env}
	}
	env.intervalYMDefinePool.New = func() interface{} {
		return &intervalYMDefine{env: env}
	}
	env.oraIntervalYMSliceBindPool.New = func() interface{} {
		return &oraIntervalYMSliceBind{env: env}
	}

	env.oraIntervalDSBindPool.New = func() interface{} {
		return &oraIntervalDSBind{env: env}
	}
	env.intervalDSDefinePool.New = func() interface{} {
		return &intervalDSDefine{env: env}
	}
	env.oraIntervalDSSliceBindPool.New = func() interface{} {
		return &oraIntervalDSSliceBind{env: env}
	}

	return env
}

// Open starts an Oracle environment.
//
// Calling Open is required prior to any other Environment method call.
func (env *Environment) Open() error {
	if !env.IsOpen() {
		// OCI_DEFAULT  - The default value, which is non-UTF-16 encoding.
		// OCI_THREADED - Uses threaded environment. Internal data structures not exposed to the user are protected from concurrent accesses by multiple threads.
		// OCI_OBJECT   - Uses object features such as OCINumber, OCINumberToInt, OCINumberFromInt. These are used in oracle-go type conversions.
		// Returns: OCI_SUCCESS or OCI_ERROR
		r := C.OCIEnvNlsCreate(
			&env.ocienv, //OCIEnv        **envhpp,
			C.OCI_DEFAULT|C.OCI_OBJECT|C.OCI_THREADED, //ub4           mode,
			nil, //void          *ctxp,
			nil, //void          *(*malocfp)
			nil, //void          *(*ralocfp)
			nil, //void          (*mfreefp)
			0,   //size_t        xtramemsz,
			nil, //void          **usrmempp
			0,   //ub2           charset,
			0)   //ub2           ncharset );
		if r == C.OCI_ERROR {
			return errNewF("Unable to create environment handle (Return code = %d).", r)
		}

		handle, _ := env.allocateOciHandle(C.OCI_HTYPE_ERROR)
		env.ocierr = (*C.OCIError)(handle)
	}
	return nil
}

// checkIsOpen validates that the environment is open.
func (env *Environment) checkIsOpen() error {
	if !env.IsOpen() {
		return errNew("open Environment prior to method call")
	}
	return nil
}

// IsOpen returns true when the environment is open; otherwise, false.
//
// Calling Close will cause IsOpen to return false.
// Once closed, the environment may be re-opened by
// calling Open.
func (env *Environment) IsOpen() bool {
	return env.ocienv != nil
}

// Close disconnects from servers and resets optional fields.
func (env *Environment) Close() error {
	if env.IsOpen() {
		// Disconnect servers
		for e := env.srvs.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Server).Close()
			if err != nil {
				return err
			}
		}

		// Free oci environment handle and all oci child handles
		// The oci error handle is released as a child of the environment handle
		err := env.freeOciHandle(unsafe.Pointer(env.ocienv), C.OCI_HTYPE_ENV)
		if err != nil {
			return err
		}

		// Clear environment fields
		// environment.servers is cleared by previous calls to disconnect all servers
		env.ocienv = nil
		env.ocierr = nil
		env.stmtConfig.Reset()
	}
	return nil
}

// OpenConnection starts a connection to an Oracle server and returns a driver.Conn.
//
// The connection string has the form username/password@dbname e.g., scott/tiger@orcl
// dbname is a connection identifier such as a net service name,
// full connection identifier, or a simple connection identifier.
// The dbname may be defined in the client machine's tnsnames.ora file.
func (env *Environment) OpenConnection(connStr string) (driver.Conn, error) {
	// Validate that the environment is open
	err := env.checkIsOpen()
	if err != nil {
		return nil, err
	}

	var username string
	var password string
	var dbname string

	// Parse connection string
	connStr = strings.Trim(connStr, " ")
	connStr = strings.Replace(connStr, "/", " / ", 1)
	connStr = strings.Replace(connStr, "@", " @ ", 1)
	_, err = fmt.Sscanf(connStr, "%s / %s @ %s", &username, &password, &dbname)

	// Connect to server
	srv, err := env.OpenServer(dbname)
	if err != nil {
		return nil, err
	}

	// Open a session on the server
	ses, err := srv.OpenSession(username, password)
	if err != nil {
		return nil, err
	}

	// Get connection from pool
	con := env.conPool.Get().(*Connection)
	con.srv = srv
	con.ses = ses

	return con, nil
}

// OpenServer connects to an Oracle server.
func (env *Environment) OpenServer(dbname string) (*Server, error) {
	// Validate that the environment is open
	err := env.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate server handle
	//OCIHandleAlloc( (void  *) envhp, (void  **) &srvhp, (ub4)OCI_HTYPE_SERVER, 0, (void  **) 0);
	serverHandle, err := env.allocateOciHandle(C.OCI_HTYPE_SERVER)
	if err != nil {
		return nil, err
	}

	// Attach to server
	//OCIServerAttach(srvhp, errhp, (text *)"inst1_alias", strlen ("inst1_alias"), OCI_DEFAULT);
	dbnamep := C.CString(dbname)
	defer C.free(unsafe.Pointer(dbnamep))
	r := C.OCIServerAttach(
		(*C.OCIServer)(serverHandle),          //OCIServer     *srvhp,
		env.ocierr,                            //OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(dbnamep)), //const OraText *dbname,
		C.sb4(C.strlen(dbnamep)),              //sb4           dbname_len,
		C.OCI_DEFAULT)                         //ub4           mode);
	if r == C.OCI_ERROR {
		return nil, env.ociError()
	}

	// Allocate service context handle
	//OCIHandleAlloc( (void  *) envhp, (void  **) &svchp, (ub4)OCI_HTYPE_SVCCTX, 0, (void  **) 0);
	svcctxHandle, err := env.allocateOciHandle(C.OCI_HTYPE_SVCCTX)
	if err != nil {
		return nil, err
	}

	// Set server handle onto service context handle
	///* set attribute server context in the service context */
	//OCIAttrSet( (void  *) svchp, (ub4) OCI_HTYPE_SVCCTX, (void  *) srvhp, (ub4) 0, (ub4) OCI_ATTR_SERVER, (OCIError *) errhp);
	err = env.setOciAttribute(svcctxHandle, C.OCI_HTYPE_SVCCTX, serverHandle, C.ub4(0), C.OCI_ATTR_SERVER)
	if err != nil {
		return nil, err
	}

	// Get server from pool
	srv := env.srvPool.Get().(*Server)
	srv.env = env
	srv.dbname = dbname
	srv.ocisvr = (*C.OCIServer)(serverHandle)
	srv.ocisvcctx = (*C.OCISvcCtx)(svcctxHandle)
	srv.stmtConfig = env.stmtConfig

	// Add server to environment list; store element for later server removal
	srv.elem = env.srvs.PushBack(srv)

	return srv, nil
}

// allocateOciHandle allocates an oci handle.
func (env *Environment) allocateOciHandle(handleType C.ub4) (unsafe.Pointer, error) {
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
func (env *Environment) freeOciHandle(ociHandle unsafe.Pointer, handleType C.ub4) error {
	// OCIHandleFree returns: OCI_SUCCESS, OCI_INVALID_HANDLE, or OCI_ERROR
	r := C.OCIHandleFree(
		unsafe.Pointer(env.ocienv), //void      *hndlp,
		handleType)                 //ub4       type );
	if r == C.OCI_INVALID_HANDLE {
		return errNew("Unable to free handle")
	} else if r == C.OCI_ERROR {
		return env.ociError()
	}

	return nil
}

// setOciAttribute sets an attribute value on a handle or descriptor.
func (env *Environment) setOciAttribute(
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
func (env *Environment) ociError() error {
	var errcode C.sb4
	var errbuff [512]C.char
	C.OCIErrorGet(
		unsafe.Pointer(env.ocierr),
		1, nil,
		&errcode,
		(*C.OraText)(unsafe.Pointer(&errbuff[0])),
		512,
		C.OCI_HTYPE_ERROR)
	s := C.GoString(&errbuff[0])
	return errors.New(s)
}

// Sets the StatementConfig on the Environment and all open Environment Servers.
func (env *Environment) SetStatementConfig(c StatementConfig) {
	env.stmtConfig = c
	for e := env.srvs.Front(); e != nil; e = e.Next() {
		e.Value.(*Server).SetStatementConfig(c)
	}
}

// StatementConfig returns a *StatementConfig.
func (env *Environment) StatementConfig() *StatementConfig {
	return &env.stmtConfig
}
