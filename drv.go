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
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/golang/glog"
)

var _drv *Drv
var _locations map[string]*time.Location

// Drv is an Oracle database driver.
//
// Drv implements the driver.Driver interface.
type Drv struct {
	listPool sync.Pool
	envPool  sync.Pool
	conPool  sync.Pool
	srvPool  sync.Pool
	sesPool  sync.Pool
	stmtPool sync.Pool
	txPool   sync.Pool
	rsetPool sync.Pool

	bndPools []sync.Pool
	defPools []sync.Pool

	// TODO: make setter,/getter and cascade to env when set
	// TODO: decide where to load the config file? From env may be best
	stmtCfg StmtCfg

	envId uint64
	envs  *list.List

	// an environment for the database/sql package
	sqlEnv *Env
}

// GetDrv returns a the default driver.
func GetDrv() *Drv {
	// place init code in GetDrv to support testing; call order requires it
	if _drv == nil {
		_locations = make(map[string]*time.Location)
		_drv = &Drv{envs: list.New()}
		_drv.listPool.New = func() interface{} {
			return list.New()
		}
		_drv.envPool.New = func() interface{} {
			return &Env{srvs: list.New(), cons: list.New(), stmtCfg: NewStmtCfg()}
		}
		_drv.conPool.New = func() interface{} {
			return &Con{}
		}
		_drv.srvPool.New = func() interface{} {
			return &Srv{sess: list.New()}
		}
		_drv.sesPool.New = func() interface{} {
			return &Ses{stmts: list.New(), txs: list.New()}
		}
		_drv.stmtPool.New = func() interface{} {
			return &Stmt{rsets: list.New()}
		}
		_drv.txPool.New = func() interface{} {
			return &Tx{}
		}
		_drv.rsetPool.New = func() interface{} {
			return &Rset{}
		}

		// init bind pools
		_drv.bndPools = make([]sync.Pool, bndIdxNil+1)
		for n := range _drv.bndPools {
			switch n {
			case bndIdxInt64:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt64{}
				}
			case bndIdxInt32:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt32{}
				}
			case bndIdxInt16:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt16{}
				}
			case bndIdxInt8:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt8{}
				}
			case bndIdxUint64:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint64{}
				}
			case bndIdxUint32:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint32{}
				}
			case bndIdxUint16:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint16{}
				}
			case bndIdxUint8:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint8{}
				}
			case bndIdxFloat64:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat64{}
				}
			case bndIdxFloat32:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat32{}
				}

			case bndIdxInt64Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt64Ptr{}
				}
			case bndIdxInt32Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt32Ptr{}
				}
			case bndIdxInt16Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt16Ptr{}
				}
			case bndIdxInt8Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt8Ptr{}
				}
			case bndIdxUint64Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint64Ptr{}
				}
			case bndIdxUint32Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint32Ptr{}
				}
			case bndIdxUint16Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint16Ptr{}
				}
			case bndIdxUint8Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint8Ptr{}
				}
			case bndIdxFloat64Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat64Ptr{}
				}
			case bndIdxFloat32Ptr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat32Ptr{}
				}

			case bndIdxInt64Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt64Slice{}
				}
			case bndIdxInt32Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt32Slice{}
				}
			case bndIdxInt16Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt16Slice{}
				}
			case bndIdxInt8Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndInt8Slice{}
				}
			case bndIdxUint64Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint64Slice{}
				}
			case bndIdxUint32Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint32Slice{}
				}
			case bndIdxUint16Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint16Slice{}
				}
			case bndIdxUint8Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndUint8Slice{}
				}
			case bndIdxFloat64Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat64Slice{}
				}
			case bndIdxFloat32Slice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndFloat32Slice{}
				}

			case bndIdxTime:
				_drv.bndPools[n].New = func() interface{} {
					return &bndTime{}
				}
			case bndIdxTimePtr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndTimePtr{}
				}
			case bndIdxTimeSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndTimeSlice{}
				}

			case bndIdxString:
				_drv.bndPools[n].New = func() interface{} {
					return &bndString{}
				}
			case bndIdxStringPtr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndStringPtr{}
				}
			case bndIdxStringSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndStringSlice{}
				}

			case bndIdxBool:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBool{}
				}
			case bndIdxBoolPtr:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBoolPtr{}
				}
			case bndIdxBoolSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBoolSlice{}
				}

			case bndIdxBin:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBin{}
				}
			case bndIdxBinSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBinSlice{}
				}

			case bndIdxIntervalYM:
				_drv.bndPools[n].New = func() interface{} {
					return &bndIntervalYM{}
				}
			case bndIdxIntervalYMSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndIntervalYMSlice{}
				}
			case bndIdxIntervalDS:
				_drv.bndPools[n].New = func() interface{} {
					return &bndIntervalDS{}
				}
			case bndIdxIntervalDSSlice:
				_drv.bndPools[n].New = func() interface{} {
					return &bndIntervalDSSlice{}
				}

			case bndIdxRset:
				_drv.bndPools[n].New = func() interface{} {
					return &bndRset{}
				}
			case bndIdxBfile:
				_drv.bndPools[n].New = func() interface{} {
					return &bndBfile{}
				}
			case bndIdxNil:
				_drv.bndPools[n].New = func() interface{} {
					return &bndNil{}
				}
			}
		}

		// init def pools
		_drv.defPools = make([]sync.Pool, defIdxRowid+1)
		for n := range _drv.defPools {
			switch n {
			case defIdxInt64:
				_drv.defPools[n].New = func() interface{} {
					return &defInt64{}
				}
			case defIdxInt32:
				_drv.defPools[n].New = func() interface{} {
					return &defInt32{}
				}
			case defIdxInt16:
				_drv.defPools[n].New = func() interface{} {
					return &defInt16{}
				}
			case defIdxInt8:
				_drv.defPools[n].New = func() interface{} {
					return &defInt8{}
				}
			case defIdxUint64:
				_drv.defPools[n].New = func() interface{} {
					return &defUint64{}
				}
			case defIdxUint32:
				_drv.defPools[n].New = func() interface{} {
					return &defUint32{}
				}
			case defIdxUint16:
				_drv.defPools[n].New = func() interface{} {
					return &defUint16{}
				}
			case defIdxUint8:
				_drv.defPools[n].New = func() interface{} {
					return &defUint8{}
				}
			case defIdxFloat64:
				_drv.defPools[n].New = func() interface{} {
					return &defFloat64{}
				}
			case defIdxFloat32:
				_drv.defPools[n].New = func() interface{} {
					return &defFloat32{}
				}

			case defIdxTime:
				_drv.defPools[n].New = func() interface{} {
					return &defTime{}
				}
			case defIdxString:
				_drv.defPools[n].New = func() interface{} {
					return &defString{}
				}
			case defIdxBool:
				_drv.defPools[n].New = func() interface{} {
					return &defBool{}
				}

			case defIdxLob:
				_drv.defPools[n].New = func() interface{} {
					return &defLob{}
				}
			case defIdxRaw:
				_drv.defPools[n].New = func() interface{} {
					return &defRaw{}
				}
			case defIdxLongRaw:
				_drv.defPools[n].New = func() interface{} {
					return &defLongRaw{}
				}

			case defIdxBfile:
				_drv.defPools[n].New = func() interface{} {
					return &defBfile{}
				}
			case defIdxIntervalYM:
				_drv.defPools[n].New = func() interface{} {
					return &defIntervalYM{}
				}
			case defIdxIntervalDS:
				_drv.defPools[n].New = func() interface{} {
					return &defIntervalDS{}
				}
			case defIdxRowid:
				_drv.defPools[n].New = func() interface{} {
					return &defRowid{}
				}
			}
		}

		// database/sql/driver expects binaryFloat to return float64
		var err error
		_drv.sqlEnv, err = _drv.OpenEnv()
		if err != nil {
			glog.Errorln("GetDrv: ", err)
		}
		_drv.sqlEnv.isSqlPkg = true
		_drv.sqlEnv.stmtCfg.Rset.binaryFloat = F64
		sql.Register(Name, _drv)
	}
	return _drv
}

// NumEnv returns the number of open Oracle environments.
func (drv *Drv) NumEnv() int {
	return drv.envs.Len()
}

var (
	csIDAl32UTF8 C.ub2
	csMu         sync.Mutex
)

// OpenEnv opens an Oracle environment.
func (drv *Drv) OpenEnv() (*Env, error) {
	env := drv.envPool.Get().(*Env)
	if env.id == 0 {
		drv.envId++
		env.id = drv.envId
	}
	glog.Infof("OpenEnv %v", env.id)

	csMu.Lock()
	if csIDAl32UTF8 == 0 {
		// Get the code for AL32UTF8
		var ocienv *C.OCIEnv
		r := C.OCIEnvCreate(&ocienv, C.OCI_DEFAULT|C.OCI_THREADED,
			nil, nil, nil, nil, 0, nil)
		if r == C.OCI_ERROR {
			csMu.Unlock()
			return nil, errNewF("Unable to create environment handle (Return code = %d).", r)
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
		return nil, errNewF("Unable to create environment handle (Return code = %d).", r)
	}
	// alloc oci error handle
	ocierr, err := env.allocOciHandle(C.OCI_HTYPE_ERROR)
	if err != nil {
		return nil, err
	}

	// set env struct

	env.drv = drv
	env.ocierr = (*C.OCIError)(ocierr)
	env.elem = drv.envs.PushBack(env)

	return env, nil
}

// Open opens a connection to an Oracle server with the database/sql environment.
//
// This is meant to be called by the database/sql package only.
//
// As an alternative, create your own Env and call Env.OpenCon.
//
// Open is a member of the driver.Driver interface.
func (drv *Drv) Open(conStr string) (driver.Conn, error) {
	glog.Infoln("Open")
	con, err := drv.sqlEnv.OpenCon(conStr)
	if err != nil {
		return nil, err
	}

	return con, nil
}

// checkNumericColumn returns nil when the column type is numeric; otherwise, an error.
func checkNumericColumn(gct GoColumnType, columnName string) error {
	switch gct {
	case I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32:
		return nil
	}
	if columnName == "" {
		return errNewF("invalid go column type (%v) specified for numeric sql column. Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", gctName(gct))
	} else {
		return errNewF("invalid go column type (%v) specified for numeric sql column (%v). Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", gctName(gct), columnName)
	}
}

// checkTimeColumn returns nil when the column type is time; otherwise, an error.
func checkTimeColumn(gct GoColumnType) error {
	switch gct {
	case T, OraT:
		return nil
	}
	return errNewF("invalid go column type (%v) specified for time-based sql column. Expected go column type T or OraT.", gctName(gct))
}

// checkStringColumn returns nil when the column type is string; otherwise, an error.
func checkStringColumn(gct GoColumnType) error {
	switch gct {
	case S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified for string-based sql column. Expected go column type S or OraS.", gctName(gct))
}

// checkBoolOrStringColumn returns nil when the column type is bool; otherwise, an error.
func checkBoolOrStringColumn(gct GoColumnType) error {
	switch gct {
	case B, OraB, S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type B, OraB, S, or OraS.", gctName(gct))
}

// checkBitsOrU8Column returns nil when the column type is Bits or U8; otherwise, an error.
func checkBitsOrU8Column(gct GoColumnType) error {
	switch gct {
	case Bin, U8:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type Bits or U8.", gctName(gct))
}

// checkBitsColumn returns nil when the column type is Bits or OraBits; otherwise, an error.
func checkBitsColumn(gct GoColumnType) error {
	switch gct {
	case Bin, OraBin:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type Bits or OraBits.", gctName(gct))
}

func gctName(gct GoColumnType) string {
	switch gct {
	case D:
		return "D"
	case I64:
		return "I64"
	case I32:
		return "I32"
	case I16:
		return "I16"
	case I8:
		return "I8"
	case U64:
		return "U64"
	case U32:
		return "U32"
	case U16:
		return "U16"
	case U8:
		return "U8"
	case F64:
		return "F64"
	case F32:
		return "F32"
	case OraI64:
		return "OraI64"
	case OraI32:
		return "OraI32"
	case OraI16:
		return "OraI16"
	case OraI8:
		return "OraI8"
	case OraU64:
		return "OraU64"
	case OraU32:
		return "OraU32"
	case OraU16:
		return "OraU16"
	case OraU8:
		return "OraU8"
	case OraF64:
		return "OraF64"
	case OraF32:
		return "OraF32"
	case T:
		return "T"
	case OraT:
		return "OraT"
	case S:
		return "S"
	case OraS:
		return "OraS"
	case B:
		return "B"
	case OraB:
		return "OraB"
	case Bin:
		return "Bits"
	case OraBin:
		return "OraBits"
	}
	return ""
}

func stringTrimmed(buffer []byte, pad byte) string {
	// Find length of non-padded string value
	// String buffer returned from Oracle is padded with Space char (32)
	//fmt.Println("stringTrimmed: len(buffer): ", len(buffer))
	var n int
	for n = len(buffer) - 1; n > -1; n-- {
		if buffer[n] != pad {
			n++
			break
		}
	}
	if n > 0 {
		return string(buffer[:n])
	}
	return ""
}

func clear(buffer []byte, fill byte) {
	for n, _ := range buffer {
		buffer[n] = fill
	}
}

func errNew(str string) error {
	return errors.New("ora: " + str)
}

func errNewF(format string, a ...interface{}) error {
	return errNew(fmt.Sprintf(format, a...))
}

func errRecover(value interface{}) error {
	return errors.New(fmt.Sprintf("ora recovered: %v", value))
}

func recoverMsg(value interface{}) string {
	return fmt.Sprintf("recovered: %v", value)
}
