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
	"time"
	"unsafe"
)

const (
	// The driver name registered with the database/sql package.
	Name string = "ora"

	// The driver version sent to an Oracle server and visible in
	// V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO.
	Version string = "v3.0.0"
)

var _drv *Drv

// init initializes the driver.
func init() {
	_drv = &Drv{}
	_drv.locations = make(map[string]*time.Location)
	_drv.openEnvs = newEnvList()
	_drv.cfg = *NewDrvCfg()

	// init general pools
	_drv.listPool = newPool(func() interface{} { return list.New() })
	_drv.envPool = newPool(func() interface{} { return &Env{openSrvs: newSrvList(), openCons: newConList()} })
	_drv.conPool = newPool(func() interface{} { return &Con{} })
	_drv.srvPool = newPool(func() interface{} { return &Srv{openSess: newSesList()} })
	_drv.sesPool = newPool(func() interface{} { return &Ses{openStmts: newStmtList(), openTxs: newTxList()} })
	_drv.stmtPool = newPool(func() interface{} { return &Stmt{openRsets: newRsetList()} })
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

func init() {
	var err error
	_drv.sqlPkgEnv, err = OpenEnv(nil)
	if err != nil {
		errE(err)
	}
	// database/sql/driver expects binaryFloat to return float64 (not the Rset default of float32)
	_drv.sqlPkgEnv.cfg.StmtCfg.Rset.binaryFloat = F64
	sql.Register(Name, _drv)
}

// SetDrvCfg sets the used configuration options for the driver.
func SetDrvCfg(cfg *DrvCfg) {
	if cfg == nil {
		return
	}
	_drv.cfg = *cfg
	_drv.sqlPkgEnv.cfg = *cfg.Env
	_drv.sqlPkgEnv.cfg.StmtCfg.Rset.binaryFloat = F64
}

// Register used to register the ora database driver with the database/sql package,
// but this is automatic now - so this function is deprecated, has the same effect
// as SetDrvCfg.
func Register(cfg *DrvCfg) {
	SetDrvCfg(cfg)
}

// OpenEnv opens an Oracle environment.
//
// Optionally specify a cfg parameter. If cfg is nil, default cfg values are
// applied.
func OpenEnv(cfg *EnvCfg) (env *Env, err error) {
	_drv.mu.Lock()
	defer _drv.mu.Unlock()
	log(_drv.cfg.Log.OpenEnv)
	if cfg == nil { // ensure cfg
		tmp := *_drv.cfg.Env // copy by value to ensure independent cfgs
		cfg = &tmp
	}
	var csIDAl32UTF8 C.ub2
	if csIDAl32UTF8 == 0 { // Get the code for AL32UTF8
		var ocienv *C.OCIEnv
		r := C.OCIEnvCreate(&ocienv, C.OCI_DEFAULT|C.OCI_THREADED, nil, nil, nil, nil, 0, nil)
		if r == C.OCI_ERROR {
			return nil, errF("Unable to create environment handle (Return code = %d).", r)
		}
		csName := []byte("AL32UTF8\x00") // http://docs.oracle.com/cd/B10501_01/server.920/a96529/ch8.htm#14284
		csIDAl32UTF8 = C.OCINlsCharSetNameToId(unsafe.Pointer(ocienv), (*C.oratext)(&csName[0]))
		C.OCIHandleFree(unsafe.Pointer(ocienv), C.OCI_HTYPE_ENV)
	}
	// OCI_DEFAULT  - The default value, which is non-UTF-16 encoding.
	// OCI_THREADED - Uses threaded environment. Internal data structures not exposed to the user are protected from concurrent accesses by multiple threads.
	// OCI_OBJECT   - Uses object features such as OCINumber, OCINumberToInt, OCINumberFromInt. These are used in oracle-go type conversions.
	env = _drv.envPool.Get().(*Env) // set *Env
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
		return nil, errF("Unable to create environment handle (Return code = %d).", r)
	}
	ocierr, err := env.allocOciHandle(C.OCI_HTYPE_ERROR) // alloc oci error handle
	if err != nil {
		return nil, errE(err)
	}

	env.ocierr = (*C.OCIError)(ocierr)
	if env.id == 0 {
		env.id = _drv.envId.nextId()
	}
	env.cfg = *cfg
	_drv.openEnvs.add(env)

	return env, nil
}

// NumEnv returns the number of open Oracle environments.
func NumEnv() int {
	_drv.mu.Lock()
	defer _drv.mu.Unlock()
	return _drv.openEnvs.len()
}

// SetCfg applies the specified cfg to the ora database driver and any open Envs.
func SetCfg(cfg DrvCfg) {
	_drv.mu.Lock()
	defer _drv.mu.Unlock()
	_drv.cfg = cfg
	_drv.openEnvs.setAllCfg(cfg.Env)
}

// Cfg returns the ora database driver's cfg.
func Cfg() *DrvCfg {
	_drv.mu.Lock()
	defer _drv.mu.Unlock()
	return &_drv.cfg
}
