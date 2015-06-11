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
	"database/sql"
	"sync"
	"time"
	"unsafe"
)

const (
	// The driver name registered with the database/sql package.
	Name string = "ora"

	// The driver version sent to an Oracle server and visible in
	// V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO.
	Version string = "v2.0.0"
)

var _drv *Drv
var _locations map[string]*time.Location

// init initializes the driver.
func init() {
	_drv = &Drv{envs: list.New()}
	_locations = make(map[string]*time.Location)
	_drv.LogOpenEnv = true
	_drv.LogOpen = true

	// init general pools
	_drv.listPool = newPool(func() interface{} { return list.New() })
	_drv.envPool = newPool(func() interface{} { return &Env{srvs: list.New(), cons: list.New(), stmtCfg: NewStmtCfg()} })
	_drv.conPool = newPool(func() interface{} { return &Con{} })
	_drv.srvPool = newPool(func() interface{} { return &Srv{sess: list.New()} })
	_drv.sesPool = newPool(func() interface{} { return &Ses{stmts: list.New(), txs: list.New()} })
	_drv.stmtPool = newPool(func() interface{} { return &Stmt{rsets: list.New()} })
	_drv.txPool = newPool(func() interface{} { return &Tx{} })
	_drv.rsetPool = newPool(func() interface{} { return &Rset{genByPool: true} })

	// init bind pools
	_drv.bndPools = make([]*pool, bndIdxNil+1)
	_drv.bndPools[bndIdxInt64] = newPool(func() interface{} { return &bndInt64{} })
	_drv.bndPools[bndIdxInt32] = newPool(func() interface{} { return &bndInt32{} })
	_drv.bndPools[bndIdxInt16] = newPool(func() interface{} { return &bndInt16{} })
	_drv.bndPools[bndIdxInt8] = newPool(func() interface{} { return &bndInt8{} })
	_drv.bndPools[bndIdxUint64] = newPool(func() interface{} { return &bndUint64{} })
	_drv.bndPools[bndIdxUint32] = newPool(func() interface{} { return &bndUint32{} })
	_drv.bndPools[bndIdxUint16] = newPool(func() interface{} { return &bndUint16{} })
	_drv.bndPools[bndIdxUint8] = newPool(func() interface{} { return &bndUint8{} })
	_drv.bndPools[bndIdxFloat64] = newPool(func() interface{} { return &bndFloat64{} })
	_drv.bndPools[bndIdxFloat32] = newPool(func() interface{} { return &bndFloat32{} })
	_drv.bndPools[bndIdxInt64Ptr] = newPool(func() interface{} { return &bndInt64Ptr{} })
	_drv.bndPools[bndIdxInt32Ptr] = newPool(func() interface{} { return &bndInt32Ptr{} })
	_drv.bndPools[bndIdxInt16Ptr] = newPool(func() interface{} { return &bndInt16Ptr{} })
	_drv.bndPools[bndIdxInt8Ptr] = newPool(func() interface{} { return &bndInt8Ptr{} })
	_drv.bndPools[bndIdxUint64Ptr] = newPool(func() interface{} { return &bndUint64Ptr{} })
	_drv.bndPools[bndIdxUint32Ptr] = newPool(func() interface{} { return &bndUint32Ptr{} })
	_drv.bndPools[bndIdxUint16Ptr] = newPool(func() interface{} { return &bndUint16Ptr{} })
	_drv.bndPools[bndIdxUint8Ptr] = newPool(func() interface{} { return &bndUint8Ptr{} })
	_drv.bndPools[bndIdxFloat64Ptr] = newPool(func() interface{} { return &bndFloat64Ptr{} })
	_drv.bndPools[bndIdxFloat32Ptr] = newPool(func() interface{} { return &bndFloat32Ptr{} })
	_drv.bndPools[bndIdxInt64Slice] = newPool(func() interface{} { return &bndInt64Slice{} })
	_drv.bndPools[bndIdxInt32Slice] = newPool(func() interface{} { return &bndInt32Slice{} })
	_drv.bndPools[bndIdxInt16Slice] = newPool(func() interface{} { return &bndInt16Slice{} })
	_drv.bndPools[bndIdxInt8Slice] = newPool(func() interface{} { return &bndInt8Slice{} })
	_drv.bndPools[bndIdxUint64Slice] = newPool(func() interface{} { return &bndUint64Slice{} })
	_drv.bndPools[bndIdxUint32Slice] = newPool(func() interface{} { return &bndUint32Slice{} })
	_drv.bndPools[bndIdxUint16Slice] = newPool(func() interface{} { return &bndUint16Slice{} })
	_drv.bndPools[bndIdxUint8Slice] = newPool(func() interface{} { return &bndUint8Slice{} })
	_drv.bndPools[bndIdxFloat64Slice] = newPool(func() interface{} { return &bndFloat64Slice{} })
	_drv.bndPools[bndIdxFloat32Slice] = newPool(func() interface{} { return &bndFloat32Slice{} })
	_drv.bndPools[bndIdxTime] = newPool(func() interface{} { return &bndTime{} })
	_drv.bndPools[bndIdxTimePtr] = newPool(func() interface{} { return &bndTimePtr{} })
	_drv.bndPools[bndIdxTimeSlice] = newPool(func() interface{} { return &bndTimeSlice{} })
	_drv.bndPools[bndIdxString] = newPool(func() interface{} { return &bndString{} })
	_drv.bndPools[bndIdxStringPtr] = newPool(func() interface{} { return &bndStringPtr{} })
	_drv.bndPools[bndIdxStringSlice] = newPool(func() interface{} { return &bndStringSlice{} })
	_drv.bndPools[bndIdxBool] = newPool(func() interface{} { return &bndBool{} })
	_drv.bndPools[bndIdxBoolPtr] = newPool(func() interface{} { return &bndBoolPtr{} })
	_drv.bndPools[bndIdxBoolSlice] = newPool(func() interface{} { return &bndBoolSlice{} })
	_drv.bndPools[bndIdxBin] = newPool(func() interface{} { return &bndBin{} })
	_drv.bndPools[bndIdxBinSlice] = newPool(func() interface{} { return &bndBinSlice{} })
	_drv.bndPools[bndIdxLob] = newPool(func() interface{} { return &bndLob{} })
	_drv.bndPools[bndIdxLobPtr] = newPool(func() interface{} { return &bndLobPtr{} })
	_drv.bndPools[bndIdxLobSlice] = newPool(func() interface{} { return &bndLobSlice{} })
	_drv.bndPools[bndIdxIntervalYM] = newPool(func() interface{} { return &bndIntervalYM{} })
	_drv.bndPools[bndIdxIntervalYMSlice] = newPool(func() interface{} { return &bndIntervalYMSlice{} })
	_drv.bndPools[bndIdxIntervalDS] = newPool(func() interface{} { return &bndIntervalDS{} })
	_drv.bndPools[bndIdxIntervalDSSlice] = newPool(func() interface{} { return &bndIntervalDSSlice{} })
	_drv.bndPools[bndIdxRset] = newPool(func() interface{} { return &bndRset{} })
	_drv.bndPools[bndIdxBfile] = newPool(func() interface{} { return &bndBfile{} })
	_drv.bndPools[bndIdxNil] = newPool(func() interface{} { return &bndNil{} })

	// init def pools
	_drv.defPools = make([]*pool, defIdxRowid+1)
	_drv.defPools[defIdxInt64] = newPool(func() interface{} { return &defInt64{} })
	_drv.defPools[defIdxInt32] = newPool(func() interface{} { return &defInt32{} })
	_drv.defPools[defIdxInt16] = newPool(func() interface{} { return &defInt16{} })
	_drv.defPools[defIdxInt8] = newPool(func() interface{} { return &defInt8{} })
	_drv.defPools[defIdxUint64] = newPool(func() interface{} { return &defUint64{} })
	_drv.defPools[defIdxUint32] = newPool(func() interface{} { return &defUint32{} })
	_drv.defPools[defIdxUint16] = newPool(func() interface{} { return &defUint16{} })
	_drv.defPools[defIdxUint8] = newPool(func() interface{} { return &defUint8{} })
	_drv.defPools[defIdxFloat64] = newPool(func() interface{} { return &defFloat64{} })
	_drv.defPools[defIdxFloat32] = newPool(func() interface{} { return &defFloat32{} })
	_drv.defPools[defIdxTime] = newPool(func() interface{} { return &defTime{} })
	_drv.defPools[defIdxString] = newPool(func() interface{} { return &defString{} })
	_drv.defPools[defIdxBool] = newPool(func() interface{} { return &defBool{} })
	_drv.defPools[defIdxLob] = newPool(func() interface{} { return &defLob{} })
	_drv.defPools[defIdxRaw] = newPool(func() interface{} { return &defRaw{} })
	_drv.defPools[defIdxLongRaw] = newPool(func() interface{} { return &defLongRaw{} })
	_drv.defPools[defIdxBfile] = newPool(func() interface{} { return &defBfile{} })
	_drv.defPools[defIdxIntervalYM] = newPool(func() interface{} { return &defIntervalYM{} })
	_drv.defPools[defIdxIntervalDS] = newPool(func() interface{} { return &defIntervalDS{} })
	_drv.defPools[defIdxRowid] = newPool(func() interface{} { return &defRowid{} })
}

// Register registers the ora database driver with the database/sql package.
//
// Call Register once before sql.Open when working with database/sql:
//
//	func init() {
//		ora.Register()
//	}
//
// Register is unnecessary if you work with the ora package directly.
func Register() {
	if _drv.sqlEnv == nil {
		var err error
		_drv.sqlEnv, err = OpenEnv()
		if err != nil {
			_drv.errE(err)
		}
		_drv.sqlEnv.isSqlPkg = true
		_drv.sqlEnv.stmtCfg.Rset.binaryFloat = F64 // database/sql/driver expects binaryFloat to return float64 (not float32)
		sql.Register(Name, _drv)
	}
}

// OpenEnv opens an Oracle environment.
func OpenEnv() (env *Env, err error) {
	_drv.log(_drv.LogOpenEnv)
	env = _drv.envPool.Get().(*Env)
	if env.id == 0 {
		env.id = _drv.envId.nextId()
	}
	var csIDAl32UTF8 C.ub2
	var csMu sync.Mutex
	csMu.Lock()
	if csIDAl32UTF8 == 0 {
		// Get the code for AL32UTF8
		var ocienv *C.OCIEnv
		r := C.OCIEnvCreate(&ocienv, C.OCI_DEFAULT|C.OCI_THREADED,
			nil, nil, nil, nil, 0, nil)
		if r == C.OCI_ERROR {
			csMu.Unlock()
			return nil, _drv.errF("Unable to create environment handle (Return code = %d).", r)
		}
		// http://docs.oracle.com/cd/B10501_01/server.920/a96529/ch8.htm#14284
		csName := []byte("AL32UTF8\x00")
		csIDAl32UTF8 = C.OCINlsCharSetNameToId(unsafe.Pointer(ocienv),
			(*C.oratext)(&csName[0]))
		C.OCIHandleFree(unsafe.Pointer(ocienv), C.OCI_HTYPE_ENV)
	}
	csMu.Unlock()

	// OCI_DEFAULT  - The default value, which is non-UTF-16 encoding.
	// OCI_THREADED - Uses threaded environment. Internal data structures not exposed to the user are protected from concurrent accesses by multiple threads.
	// OCI_OBJECT   - Uses object features such as OCINumber, OCINumberToInt, OCINumberFromInt. These are used in oracle-go type conversions.
	r := C.OCIEnvNlsCreate(
		&env.ocienv, //OCIEnv        **envhpp,
		C.OCI_DEFAULT|C.OCI_OBJECT|C.OCI_THREADED, //ub4           mode,
		nil,          //void          *ctxp,
		nil,          //void          *(*malocfp)
		nil,          //void          *(*ralocfp)
		nil,          //void          (*mfreefp)
		0,            //size_t        xtramemsz,
		nil,          //void          **usrmempp
		csIDAl32UTF8, //ub2           charset,
		csIDAl32UTF8) //ub2           ncharset );
	if r == C.OCI_ERROR {
		return nil, _drv.errF("Unable to create environment handle (Return code = %d).", r)
	}
	ocierr, err := env.allocOciHandle(C.OCI_HTYPE_ERROR) // alloc oci error handle
	if err != nil {
		return nil, _drv.errE(err)
	}
	env.drv = _drv // set env struct
	env.ocierr = (*C.OCIError)(ocierr)
	env.elem = _drv.envs.PushBack(env)
	env.LogClose = true
	env.LogOpenSrv = true
	return env, nil
}

// NumEnv returns the number of open Oracle environments.
func NumEnv() int {
	return _drv.envs.Len()
}
