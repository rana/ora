// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>

#cgo pkg-config: oci8
*/
import "C"
import (
	"container/list"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	// The driver name registered with the database/sql package.
	Name string = "ora"

	// The driver version sent to an Oracle server and visible in
	// V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO.
	Version string = "v4.1.15"
)

var _drv *Drv

// init initializes the driver.
func init() {
	_drv = &Drv{}
	_drv.locations = make(map[string]*time.Location)
	_drv.openEnvs = newEnvList()

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
	_drv.bndPools = make([]*sync.Pool, bndIdxNil+1)
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
	_drv.bndPools[bndIdxNumString] = newPool(func() interface{} { return &bndNumString{} })
	_drv.bndPools[bndIdxOCINum] = newPool(func() interface{} { return &bndOCINum{} })
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
	_drv.bndPools[bndIdxNumStringPtr] = newPool(func() interface{} { return &bndNumStringPtr{} })
	_drv.bndPools[bndIdxOCINumPtr] = newPool(func() interface{} { return &bndOCINumPtr{} })
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
	_drv.bndPools[bndIdxNumStringSlice] = newPool(func() interface{} { return &bndNumStringSlice{} })
	_drv.bndPools[bndIdxOCINumSlice] = newPool(func() interface{} { return &bndOCINumSlice{} })
	_drv.bndPools[bndIdxTime] = newPool(func() interface{} { return &bndTime{} })
	_drv.bndPools[bndIdxTimePtr] = newPool(func() interface{} { return &bndTimePtr{} })
	_drv.bndPools[bndIdxTimeSlice] = newPool(func() interface{} { return &bndTimeSlice{} })
	_drv.bndPools[bndIdxDate] = newPool(func() interface{} { return &bndDate{} })
	_drv.bndPools[bndIdxDatePtr] = newPool(func() interface{} { return &bndDatePtr{} })
	_drv.bndPools[bndIdxDateSlice] = newPool(func() interface{} { return &bndDateSlice{} })
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
	_drv.defPools = make([]*sync.Pool, defIdxRset+1)
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
	_drv.defPools[defIdxOCINum] = newPool(func() interface{} { return &defOCINum{} })
	_drv.defPools[defIdxTime] = newPool(func() interface{} { return &defTime{} })
	_drv.defPools[defIdxDate] = newPool(func() interface{} { return &defDate{} })
	_drv.defPools[defIdxString] = newPool(func() interface{} { return &defString{} })
	_drv.defPools[defIdxNumString] = newPool(func() interface{} { return &defNumString{} })
	_drv.defPools[defIdxOCINum] = newPool(func() interface{} { return &defOCINum{} })
	_drv.defPools[defIdxBool] = newPool(func() interface{} { return &defBool{} })
	_drv.defPools[defIdxLob] = newPool(func() interface{} { return &defLob{} })
	_drv.defPools[defIdxRaw] = newPool(func() interface{} { return &defRaw{} })
	_drv.defPools[defIdxLongRaw] = newPool(func() interface{} { return &defLongRaw{} })
	_drv.defPools[defIdxBfile] = newPool(func() interface{} { return &defBfile{} })
	_drv.defPools[defIdxIntervalYM] = newPool(func() interface{} { return &defIntervalYM{} })
	_drv.defPools[defIdxIntervalDS] = newPool(func() interface{} { return &defIntervalDS{} })
	_drv.defPools[defIdxRowid] = newPool(func() interface{} { return &defRowid{} })
	_drv.defPools[defIdxRset] = newPool(func() interface{} { return &defRset{} })

	var err error
	if _drv.sqlPkgEnv, err = OpenEnv(); err != nil {
		panic(fmt.Sprintf("OpenEnv: %v", err))
	}
	_drv.sqlPkgEnv.isPkgEnv = true
	// database/sql/driver expects binaryFloat to return float64 (not the Rset default of float32)
	cfg := _drv.sqlPkgEnv.Cfg()
	cfg.RsetCfg.binaryFloat = F64
	_drv.sqlPkgEnv.SetCfg(cfg)
	sql.Register(Name, _drv)
}

var csIDAl32UTF8 uint32

// OpenEnv opens an Oracle environment.
//
// Optionally specify a cfg parameter. If cfg is nil, default cfg values are
// applied.
func OpenEnv() (env *Env, err error) {
	cfg := _drv.Cfg()
	log(cfg.Log.OpenEnv)
	csid := C.ub2(atomic.LoadUint32(&csIDAl32UTF8))
	if csid == 0 { // Get the code for AL32UTF8
		var ocienv *C.OCIEnv
		r := C.OCIEnvCreate(&ocienv, C.OCI_DEFAULT|C.OCI_THREADED, nil, nil, nil, nil, 0, nil)
		if r == C.OCI_ERROR {
			return nil, errF("Unable to create environment handle (Return code = %d).", r)
		}
		csName := []byte("AL32UTF8\x00") // http://docs.oracle.com/cd/B10501_01/server.920/a96529/ch8.htm#14284
		csid = C.OCINlsCharSetNameToId(unsafe.Pointer(ocienv), (*C.oratext)(&csName[0]))
		C.OCIHandleFree(unsafe.Pointer(ocienv), C.OCI_HTYPE_ENV)
		atomic.StoreUint32(&csIDAl32UTF8, uint32(csid))
	}
	// OCI_DEFAULT  - The default value, which is non-UTF-16 encoding.
	// OCI_THREADED - Uses threaded environment. Internal data structures not exposed to the user are protected from concurrent accesses by multiple threads.
	// OCI_OBJECT   - Uses object features such as OCINumber, OCINumberToInt, OCINumberFromInt. These are used in oracle-go type conversions.
	_drv.RLock()
	env = _drv.envPool.Get().(*Env) // set *Env
	env.cmu.Lock()
	defer env.cmu.Unlock()
	r := C.OCIEnvNlsCreate(
		&env.ocienv, //OCIEnv        **envhpp,
		C.OCI_DEFAULT|C.OCI_OBJECT|C.OCI_THREADED, //ub4           mode,
		nil,  //void          *ctxp,
		nil,  //void          *(*malocfp)
		nil,  //void          *(*ralocfp)
		nil,  //void          (*mfreefp)
		0,    //size_t        xtramemsz,
		nil,  //void          **usrmempp
		csid, //ub2           charset,
		csid) //ub2           ncharset );
	_drv.RUnlock()
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
	env.SetCfg(cfg.StmtCfg)
	_drv.RLock()
	_drv.openEnvs.add(env)
	_drv.RUnlock()

	return env, nil
}

// NumEnv returns the number of open Oracle environments.
func NumEnv() int {
	_drv.RLock()
	defer _drv.RUnlock()
	return _drv.openEnvs.len()
}

// SetCfg applies the specified cfg to the ora database driver.
func SetCfg(cfg DrvCfg) {
	cfg.RsetCfg.binaryFloat = F64
	_drv.SetCfg(cfg)
	_drv.Lock()
	_drv.sqlPkgEnv.SetCfg(cfg.StmtCfg)
	_drv.Unlock()
}

// Cfg returns the ora database driver's cfg.
func Cfg() DrvCfg {
	return _drv.Cfg()
}

// Register used to register the ora database driver with the database/sql package,
// but this is automatic now - so this function is deprecated, has the same effect
// as SetCfg.
func Register(cfg DrvCfg) {
	SetCfg(cfg)
}

func newPool(f func() interface{}) *sync.Pool { return &sync.Pool{New: f} }
