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
	ocienv          *C.OCIEnv
	ocierr          *C.OCIError
	statementConfig StatementConfig

	servers        *list.List
	connectionPool sync.Pool
	serverPool     sync.Pool
	sessionPool    sync.Pool
	statementPool  sync.Pool

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

// NewEnvironment creates a new Oracle environment.
func NewEnvironment() (environment *Environment) {
	environment = &Environment{servers: list.New()}
	environment.statementConfig = NewStatementConfig()

	environment.connectionPool.New = func() interface{} {
		return &Connection{environment: environment}
	}
	environment.serverPool.New = func() interface{} {
		return &Server{sessions: list.New()}
	}
	environment.sessionPool.New = func() interface{} {
		return &Session{statements: list.New()}
	}
	environment.statementPool.New = func() interface{} {
		return &Statement{environment: environment, resultSets: list.New()}
	}

	environment.int64BindPool.New = func() interface{} {
		return &int64Bind{environment: environment}
	}
	environment.int32BindPool.New = func() interface{} {
		return &int32Bind{environment: environment}
	}
	environment.int16BindPool.New = func() interface{} {
		return &int16Bind{environment: environment}
	}
	environment.int8BindPool.New = func() interface{} {
		return &int8Bind{environment: environment}
	}
	environment.uint64BindPool.New = func() interface{} {
		return &uint64Bind{environment: environment}
	}
	environment.uint32BindPool.New = func() interface{} {
		return &uint32Bind{environment: environment}
	}
	environment.uint16BindPool.New = func() interface{} {
		return &uint16Bind{environment: environment}
	}
	environment.uint8BindPool.New = func() interface{} {
		return &uint8Bind{environment: environment}
	}
	environment.float64BindPool.New = func() interface{} {
		return &float64Bind{environment: environment}
	}
	environment.float32BindPool.New = func() interface{} {
		return &float32Bind{environment: environment}
	}

	environment.int64PtrBindPool.New = func() interface{} {
		return &int64PtrBind{environment: environment}
	}
	environment.int32PtrBindPool.New = func() interface{} {
		return &int32PtrBind{environment: environment}
	}
	environment.int16PtrBindPool.New = func() interface{} {
		return &int16PtrBind{environment: environment}
	}
	environment.int8PtrBindPool.New = func() interface{} {
		return &int8PtrBind{environment: environment}
	}
	environment.uint64PtrBindPool.New = func() interface{} {
		return &uint64PtrBind{environment: environment}
	}
	environment.uint32PtrBindPool.New = func() interface{} {
		return &uint32PtrBind{environment: environment}
	}
	environment.uint16PtrBindPool.New = func() interface{} {
		return &uint16PtrBind{environment: environment}
	}
	environment.uint8PtrBindPool.New = func() interface{} {
		return &uint8PtrBind{environment: environment}
	}
	environment.float64PtrBindPool.New = func() interface{} {
		return &float64PtrBind{environment: environment}
	}
	environment.float32PtrBindPool.New = func() interface{} {
		return &float32PtrBind{environment: environment}
	}

	environment.int64SliceBindPool.New = func() interface{} {
		return &int64SliceBind{environment: environment}
	}
	environment.int32SliceBindPool.New = func() interface{} {
		return &int32SliceBind{environment: environment}
	}
	environment.int16SliceBindPool.New = func() interface{} {
		return &int16SliceBind{environment: environment}
	}
	environment.int8SliceBindPool.New = func() interface{} {
		return &int8SliceBind{environment: environment}
	}
	environment.uint64SliceBindPool.New = func() interface{} {
		return &uint64SliceBind{environment: environment}
	}
	environment.uint32SliceBindPool.New = func() interface{} {
		return &uint32SliceBind{environment: environment}
	}
	environment.uint16SliceBindPool.New = func() interface{} {
		return &uint16SliceBind{environment: environment}
	}
	environment.uint8SliceBindPool.New = func() interface{} {
		return &uint8SliceBind{environment: environment}
	}
	environment.float64SliceBindPool.New = func() interface{} {
		return &float64SliceBind{environment: environment}
	}
	environment.float32SliceBindPool.New = func() interface{} {
		return &float32SliceBind{environment: environment}
	}

	environment.int64DefinePool.New = func() interface{} {
		return &int64Define{environment: environment}
	}
	environment.int32DefinePool.New = func() interface{} {
		return &int32Define{environment: environment}
	}
	environment.int16DefinePool.New = func() interface{} {
		return &int16Define{environment: environment}
	}
	environment.int8DefinePool.New = func() interface{} {
		return &int8Define{environment: environment}
	}
	environment.uint64DefinePool.New = func() interface{} {
		return &uint64Define{environment: environment}
	}
	environment.uint32DefinePool.New = func() interface{} {
		return &uint32Define{environment: environment}
	}
	environment.uint16DefinePool.New = func() interface{} {
		return &uint16Define{environment: environment}
	}
	environment.uint8DefinePool.New = func() interface{} {
		return &uint8Define{environment: environment}
	}
	environment.float64DefinePool.New = func() interface{} {
		return &float64Define{environment: environment}
	}
	environment.float32DefinePool.New = func() interface{} {
		return &float32Define{environment: environment}
	}

	environment.oraInt64DefinePool.New = func() interface{} {
		return &oraInt64Define{environment: environment}
	}
	environment.oraInt32DefinePool.New = func() interface{} {
		return &oraInt32Define{environment: environment}
	}
	environment.oraInt16DefinePool.New = func() interface{} {
		return &oraInt16Define{environment: environment}
	}
	environment.oraInt8DefinePool.New = func() interface{} {
		return &oraInt8Define{environment: environment}
	}
	environment.oraUint64DefinePool.New = func() interface{} {
		return &oraUint64Define{environment: environment}
	}
	environment.oraUint32DefinePool.New = func() interface{} {
		return &oraUint32Define{environment: environment}
	}
	environment.oraUint16DefinePool.New = func() interface{} {
		return &oraUint16Define{environment: environment}
	}
	environment.oraUint8DefinePool.New = func() interface{} {
		return &oraUint8Define{environment: environment}
	}
	environment.oraFloat64DefinePool.New = func() interface{} {
		return &oraFloat64Define{environment: environment}
	}
	environment.oraFloat32DefinePool.New = func() interface{} {
		return &oraFloat32Define{environment: environment}
	}

	environment.timeBindPool.New = func() interface{} {
		return &timeBind{environment: environment}
	}
	environment.timePtrBindPool.New = func() interface{} {
		return &timePtrBind{environment: environment}
	}
	environment.timeSliceBindPool.New = func() interface{} {
		return &timeSliceBind{environment: environment}
	}
	environment.timeDefinePool.New = func() interface{} {
		return &timeDefine{environment: environment}
	}
	environment.oraTimeDefinePool.New = func() interface{} {
		return &oraTimeDefine{environment: environment}
	}
	environment.stringBindPool.New = func() interface{} {
		return &stringBind{environment: environment}
	}
	environment.stringPtrBindPool.New = func() interface{} {
		return &stringPtrBind{environment: environment}
	}
	environment.stringSliceBindPool.New = func() interface{} {
		return &stringSliceBind{environment: environment}
	}
	environment.stringDefinePool.New = func() interface{} {
		return &stringDefine{environment: environment}
	}
	environment.oraStringDefinePool.New = func() interface{} {
		return &oraStringDefine{environment: environment}
	}
	environment.bytesBindPool.New = func() interface{} {
		return &bytesBind{environment: environment}
	}
	environment.bytesSliceBindPool.New = func() interface{} {
		return &bytesSliceBind{environment: environment}
	}
	environment.lobDefinePool.New = func() interface{} {
		return &lobDefine{environment: environment}
	}
	environment.longRawDefinePool.New = func() interface{} {
		return &longRawDefine{environment: environment}
	}
	environment.rawDefinePool.New = func() interface{} {
		return &rawDefine{environment: environment}
	}
	environment.bfileBindPool.New = func() interface{} {
		return &bfileBind{environment: environment}
	}
	environment.bfileDefinePool.New = func() interface{} {
		return &bfileDefine{environment: environment}
	}
	environment.nilBindPool.New = func() interface{} {
		return &nilBind{environment: environment}
	}

	environment.boolBindPool.New = func() interface{} {
		return &boolBind{environment: environment}
	}
	environment.boolPtrBindPool.New = func() interface{} {
		return &boolPtrBind{environment: environment}
	}
	environment.boolSliceBindPool.New = func() interface{} {
		return &boolSliceBind{environment: environment}
	}
	environment.boolDefinePool.New = func() interface{} {
		return &boolDefine{environment: environment}
	}
	environment.oraBoolDefinePool.New = func() interface{} {
		return &oraBoolDefine{environment: environment}
	}

	environment.rowidDefinePool.New = func() interface{} {
		return &rowidDefine{environment: environment}
	}
	environment.resultSetBindPool.New = func() interface{} {
		return &resultSetBind{environment: environment}
	}

	environment.oraIntervalYMBindPool.New = func() interface{} {
		return &oraIntervalYMBind{environment: environment}
	}
	environment.intervalYMDefinePool.New = func() interface{} {
		return &intervalYMDefine{environment: environment}
	}
	environment.oraIntervalYMSliceBindPool.New = func() interface{} {
		return &oraIntervalYMSliceBind{environment: environment}
	}

	environment.oraIntervalDSBindPool.New = func() interface{} {
		return &oraIntervalDSBind{environment: environment}
	}
	environment.intervalDSDefinePool.New = func() interface{} {
		return &intervalDSDefine{environment: environment}
	}
	environment.oraIntervalDSSliceBindPool.New = func() interface{} {
		return &oraIntervalDSSliceBind{environment: environment}
	}

	return environment
}

// Open starts an Oracle environment.
//
// Calling Open is required prior to any other Environment method call.
func (environment *Environment) Open() error {
	if !environment.IsOpen() {
		// OCI_DEFAULT  - The default value, which is non-UTF-16 encoding.
		// OCI_THREADED - Uses threaded environment. Internal data structures not exposed to the user are protected from concurrent accesses by multiple threads.
		// OCI_OBJECT   - Uses object features such as OCINumber, OCINumberToInt, OCINumberFromInt. These are used in oracle-go type conversions.
		// Returns: OCI_SUCCESS or OCI_ERROR
		r := C.OCIEnvNlsCreate(
			&environment.ocienv,                       //OCIEnv        **envhpp,
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

		handle, _ := environment.allocateOciHandle(C.OCI_HTYPE_ERROR)
		environment.ocierr = (*C.OCIError)(handle)
	}
	return nil
}

// checkIsOpen validates that the environment is open.
//
// ErrClosedEnvironment is returned if the environment is closed.
func (environment *Environment) checkIsOpen() error {
	if !environment.IsOpen() {
		return errNew("open Environment prior to method call")
	}
	return nil
}

// IsOpen returns true when the environment is open; otherwise, false.
//
// Calling Close will cause IsOpen to return false.
// Once closed, the environment may be re-opened by
// calling Open.
func (environment *Environment) IsOpen() bool {
	return environment.ocienv != nil
}

// Close disconnects from servers and resets optional fields.
func (environment *Environment) Close() error {
	if environment.IsOpen() {
		// Disconnect servers
		for e := environment.servers.Front(); e != nil; e = e.Next() {
			err := e.Value.(*Server).Close()
			if err != nil {
				return err
			}
		}

		// Free oci environment handle and all oci child handles
		// The oci error handle is released as a child of the environment handle
		err := environment.freeOciHandle(unsafe.Pointer(environment.ocienv), C.OCI_HTYPE_ENV)
		if err != nil {
			return err
		}

		// Clear environment fields
		// environment.servers is cleared by previous calls to disconnect all servers
		environment.ocienv = nil
		environment.ocierr = nil
		environment.statementConfig.Reset()
	}
	return nil
}

// OpenConnection starts a connection to an Oracle server and returns a driver.Conn.
//
// The connection string has the form username/password@dbname.
// dbname is a connection identifier such as a net service name,
// full connection identifier, or a simple connection identifier.
func (environment *Environment) OpenConnection(connStr string) (driver.Conn, error) {
	// Validate that the environment is open
	err := environment.checkIsOpen()
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
	server, err := environment.OpenServer(dbname)
	if err != nil {
		return nil, err
	}

	// Open a session on the server
	session, err := server.OpenSession(username, password)
	if err != nil {
		return nil, err
	}

	// Get connection from pool
	connection := environment.connectionPool.Get().(*Connection)
	connection.server = server
	connection.session = session

	return connection, nil
}

// OpenServer connects to an Oracle server.
func (environment *Environment) OpenServer(dbname string) (*Server, error) {
	// Validate that the environment is open
	err := environment.checkIsOpen()
	if err != nil {
		return nil, err
	}

	// Allocate server handle
	//OCIHandleAlloc( (void  *) envhp, (void  **) &srvhp, (ub4)OCI_HTYPE_SERVER, 0, (void  **) 0);
	serverHandle, err := environment.allocateOciHandle(C.OCI_HTYPE_SERVER)
	if err != nil {
		return nil, err
	}

	// Attach to server
	//OCIServerAttach(srvhp, errhp, (text *)"inst1_alias", strlen ("inst1_alias"), OCI_DEFAULT);
	dbnamep := C.CString(dbname)
	defer C.free(unsafe.Pointer(dbnamep))
	r := C.OCIServerAttach(
		(*C.OCIServer)(serverHandle),          //OCIServer     *srvhp,
		environment.ocierr,                    //OCIError      *errhp,
		(*C.OraText)(unsafe.Pointer(dbnamep)), //const OraText *dbname,
		C.sb4(C.strlen(dbnamep)),              //sb4           dbname_len,
		C.OCI_DEFAULT)                         //ub4           mode);
	if r == C.OCI_ERROR {
		return nil, environment.ociError()
	}

	// Allocate service context handle
	//OCIHandleAlloc( (void  *) envhp, (void  **) &svchp, (ub4)OCI_HTYPE_SVCCTX, 0, (void  **) 0);
	svcctxHandle, err := environment.allocateOciHandle(C.OCI_HTYPE_SVCCTX)
	if err != nil {
		return nil, err
	}

	// Set server handle onto service context handle
	///* set attribute server context in the service context */
	//OCIAttrSet( (void  *) svchp, (ub4) OCI_HTYPE_SVCCTX, (void  *) srvhp, (ub4) 0, (ub4) OCI_ATTR_SERVER, (OCIError *) errhp);
	err = environment.setOciAttribute(svcctxHandle, C.OCI_HTYPE_SVCCTX, serverHandle, C.ub4(0), C.OCI_ATTR_SERVER)
	if err != nil {
		return nil, err
	}

	// Get server from pool
	server := environment.serverPool.Get().(*Server)
	server.environment = environment
	server.dbname = dbname
	server.ocisvr = (*C.OCIServer)(serverHandle)
	server.ocisvcctx = (*C.OCISvcCtx)(svcctxHandle)
	server.statementConfig = environment.statementConfig

	// Add server to environment list; store element for later server removal
	server.element = environment.servers.PushBack(server)

	return server, nil
}

// allocateOciHandle allocates an oci handle.
func (environment *Environment) allocateOciHandle(handleType C.ub4) (unsafe.Pointer, error) {
	// OCIHandleAlloc returns: OCI_SUCCESS, OCI_INVALID_HANDLE
	var handle unsafe.Pointer
	r := C.OCIHandleAlloc(
		unsafe.Pointer(environment.ocienv), //const void    *parenth,
		&handle,     //void          **hndlpp,
		handleType,  //ub4           type,
		C.size_t(0), //size_t        xtramem_sz,
		nil)         //void          **usrmempp
	if r == C.OCI_INVALID_HANDLE {
		return nil, errNew("Unable to allocate handle")
	}
	return handle, nil
}

// freeOciHandle deallocates an oci handle.
func (environment *Environment) freeOciHandle(ociHandle unsafe.Pointer, handleType C.ub4) error {
	// OCIHandleFree returns: OCI_SUCCESS, OCI_INVALID_HANDLE, or OCI_ERROR
	r := C.OCIHandleFree(
		unsafe.Pointer(environment.ocienv), //void      *hndlp,
		handleType)                         //ub4       type );
	if r == C.OCI_INVALID_HANDLE {
		return errNew("Unable to free handle")
	} else if r == C.OCI_ERROR {
		return environment.ociError()
	}

	return nil
}

// setOciAttribute sets an attribute value on a handle or descriptor.
func (environment *Environment) setOciAttribute(
	target unsafe.Pointer,
	targetType C.ub4,
	attribute unsafe.Pointer,
	attributeSize C.ub4,
	attributeType C.ub4) (err error) {

	r := C.OCIAttrSet(
		target,             //void        *trgthndlp,
		targetType,         //ub4         trghndltyp,
		attribute,          //void        *attributep,
		attributeSize,      //ub4         size,
		attributeType,      //ub4         attrtype,
		environment.ocierr) //OCIError    *errhp );
	if r == C.OCI_ERROR {
		return environment.ociError()
	}
	return nil
}

// getOciError gets an error returned by an Oracle server.
func (environment *Environment) ociError() error {
	var errcode C.sb4
	var errbuff [512]C.char
	C.OCIErrorGet(
		unsafe.Pointer(environment.ocierr),
		1, nil,
		&errcode,
		(*C.OraText)(unsafe.Pointer(&errbuff[0])),
		512,
		C.OCI_HTYPE_ERROR)
	s := C.GoString(&errbuff[0])
	return errors.New(s)
}

// Sets the StatementConfig on the Environment and all open Environment Servers.
func (environment *Environment) SetStatementConfig(c StatementConfig) {
	environment.statementConfig = c
	for e := environment.servers.Front(); e != nil; e = e.Next() {
		e.Value.(*Server).SetStatementConfig(c)
	}
}

// StatementConfig returns a *StatementConfig.
func (environment *Environment) StatementConfig() *StatementConfig {
	return &environment.statementConfig
}
