// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import (
	"bytes"
	"container/list"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// LogStmtCfg represents Stmt logging configuration values.
type LogStmtCfg struct {
	// Close determines whether the Stmt.Close method is logged.
	//
	// The default is true.
	Close bool

	// Exe determines whether the Stmt.Exe method is logged.
	//
	// The default is true.
	Exe bool

	// Qry determines whether the Stmt.Qry method is logged.
	//
	// The default is true.
	Qry bool

	// Bind determines whether the Stmt.bind method is logged.
	//
	// The default is true.
	Bind bool
}

// NewLogStmtCfg creates a LogStmtCfg with default values.
func NewLogStmtCfg() LogStmtCfg {
	c := LogStmtCfg{}
	c.Close = true
	c.Exe = true
	c.Qry = true
	c.Bind = true
	return c
}

// Stmt represents an Oracle statement.
type Stmt struct {
	id         uint64
	cfg        StmtCfg
	mu         sync.Mutex
	ses        *Ses
	ocistmt    *C.OCIStmt
	stmtType   C.ub2
	sql        string
	gcts       []GoColumnType
	bnds       []bnd
	hasPtrBind bool

	openRsets *rsetList

	sysNamer
}

// Close closes the SQL statement.
//
// Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Stmt.Prep to create a new statement.
func (stmt *Stmt) Close() (err error) {
	if stmt == nil {
		return nil
	}
	stmt.mu.Lock()
	if stmt.ses == nil {
		stmt.mu.Unlock()
		return nil
	}
	stmt.ses.mu.Lock()
	if stmt.ses.openStmts != nil {
		stmt.ses.openStmts.remove(stmt)
	}
	stmt.ses.mu.Unlock()
	defer stmt.mu.Unlock()
	return stmt.close()
}

// close closes the SQL statement, without locking stmt.
// does not remove Stmt from Ses.openStmts
func (stmt *Stmt) close() (err error) {
	stmt.log(_drv.cfg.Log.Stmt.Close)
	err = stmt.checkClosed()
	if err != nil {
		return errE(err)
	}
	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			err := errR(value)
			stmt.logF(true, "PANIC %v", err)
			errs.PushBack(err)
		}
		// free ocistmt to release cursor on server
		// OCIStmtRelease must be called with OCIStmtPrepare2
		// See https://docs.oracle.com/database/121/LNOCI/oci09adv.htm#LNOCI16655
		r := C.OCIStmtRelease(
			stmt.ocistmt,            // OCIStmt        *stmthp
			stmt.ses.srv.env.ocierr, // OCIError       *errhp,
			nil,           // const OraText  *key
			C.ub4(0),      // ub4 keylen
			C.OCI_DEFAULT, // ub4 mode
		)
		if r == C.OCI_ERROR {
			errs.PushBack(errE(stmt.ses.srv.env.ociError()))
		}

		stmt.ses = nil
		stmt.ocistmt = nil
		stmt.stmtType = 0
		stmt.sql = ""
		stmt.gcts = nil
		stmt.bnds = nil
		stmt.hasPtrBind = false
		stmt.openRsets.clear()
		_drv.stmtPool.Put(stmt)

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()
	// close binds
	for _, bind := range stmt.bnds {
		if bind == nil {
			continue
		}
		err = bind.close()
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	stmt.openRsets.closeAll(errs)

	return nil
}

// Exe executes a SQL statement on an Oracle server returning the number of
// rows affected and a possible error.
//
// Slice arguments should have the same length, as they'll be called in batch mode.
func (stmt *Stmt) Exe(params ...interface{}) (rowsAffected uint64, err error) {
	rowsAffected, _, err = stmt.exe(params, false)
	return rowsAffected, err
}

// ExeP executes an (PL/)SQL statement on an Oracle server returning the number of
// rows affected and a possible error.
//
// All arguments are sent as is (esp. slices).
func (stmt *Stmt) ExeP(params ...interface{}) (rowsAffected uint64, err error) {
	rowsAffected, _, err = stmt.exe(params, true)
	return rowsAffected, err
}

var spcRpl = strings.NewReplacer("\t", " ", "   ", " ", "  ", " ")

// exe executes a SQL statement on an Oracle server returning rowsAffected, lastInsertId and error.
func (stmt *Stmt) exe(params []interface{}, isAssocArray bool) (rowsAffected uint64, lastInsertId int64, err error) {
	if stmt == nil {
		return 0, 0, er("stmt may not be nil.")
	}
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt.log(_drv.cfg.Log.Stmt.Exe)
	err = stmt.checkClosed()
	if err != nil {
		return 0, 0, errE(err)
	}
	// for case of inserting and returning identity for database/sql package
	if _drv.sqlPkgEnv == stmt.ses.srv.env && stmt.stmtType == C.OCI_STMT_INSERT {
		lastIndex := strings.LastIndex(stmt.sql, ")")
		sqlEnd := spcRpl.Replace(stmt.sql[lastIndex+1 : len(stmt.sql)])
		sqlEnd = strings.ToUpper(sqlEnd)
		// add *int64 arg to capture identity
		if i := strings.LastIndex(sqlEnd, "RETURNING"); i >= 0 && strings.Contains(sqlEnd[i:], " /*LASTINSERTID*/ INTO ") {
			params[len(params)-1] = &lastInsertId
		}
	}
	iterations, err := stmt.bind(params, isAssocArray) // bind parameters
	if err != nil {
		return 0, 0, errE(err)
	}
	err = stmt.setPrefetchSize() // set prefetch size
	if err != nil {
		return 0, 0, errE(err)
	}
	var mode C.ub4 // determine auto-commit state; don't auto-comit if there's an explicit user transaction occuring
	if stmt.cfg.IsAutoCommitting && stmt.ses.openTxs.len() == 0 {
		mode = C.OCI_COMMIT_ON_SUCCESS
	} else {
		mode = C.OCI_DEFAULT
	}
	stmt.logF(_drv.cfg.Log.Stmt.Exe, "iterations=%d", iterations)
	// Execute statement on Oracle server
	r := C.OCIStmtExecute(
		stmt.ses.ocisvcctx,      //OCISvcCtx           *svchp,
		stmt.ocistmt,            //OCIStmt             *stmtp,
		stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(iterations),       //ub4                 iters,
		C.ub4(0),                //ub4                 rowoff,
		nil,                     //const OCISnapshot   *snap_in,
		nil,                     //OCISnapshot         *snap_out,
		mode)                    //ub4                 mode );
	stmt.logF(_drv.cfg.Log.Stmt.Exe, "returned %d", r)
	if r == C.OCI_ERROR {
		return 0, 0, errE(stmt.ses.srv.env.ociError())
	}
	// Get rowsAffected based on statement type
	switch stmt.stmtType {
	case C.OCI_STMT_SELECT, C.OCI_STMT_UPDATE, C.OCI_STMT_DELETE, C.OCI_STMT_INSERT:
		ra, err := stmt.attr(C.ROW_COUNT_LENGTH, C.OCI_ATTR_UB8_ROW_COUNT)
		if err != nil {
			return 0, 0, errE(err)
		}
		rowsAffected = uint64(*((*C.ROW_COUNT_TYPE)(ra)))
		C.free(ra)
		//case C.OCI_STMT_CREATE, C.OCI_STMT_DROP, C.OCI_STMT_ALTER, C.OCI_STMT_BEGIN:
	default:
		if r == C.OCI_NO_DATA {
			return 0, 0, errE(stmt.ses.srv.env.ociError())
		}
		//fmt.Printf("stmtType=%d\n", stmt.stmtType)
	}
	if stmt.hasPtrBind { // Set any bind pointers
		err = stmt.setBindPtrs()
		if err != nil {
			return rowsAffected, lastInsertId, errE(err)
		}
	}
	return rowsAffected, lastInsertId, nil
}

// Qry runs a SQL query on an Oracle server returning a *Rset and possible error.
func (stmt *Stmt) Qry(params ...interface{}) (*Rset, error) {
	return stmt.qry(params)
}

// qry runs a SQL query on an Oracle server returning a *Rset and possible error.
func (stmt *Stmt) qry(params []interface{}) (rset *Rset, err error) {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt.log(_drv.cfg.Log.Stmt.Qry)
	err = stmt.checkClosed()
	if err != nil {
		return nil, errE(err)
	}
	_, err = stmt.bind(params, false) // bind parameters
	if err != nil {
		return nil, errE(err)
	}
	err = stmt.setPrefetchSize() // set prefetch size
	if err != nil {
		return nil, errE(err)
	}
	// Query statement on Oracle server
	r := C.OCIStmtExecute(
		stmt.ses.ocisvcctx,      //OCISvcCtx           *svchp,
		stmt.ocistmt,            //OCIStmt             *stmtp,
		stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(0),                //ub4                 iters,
		C.ub4(0),                //ub4                 rowoff,
		nil,                     //const OCISnapshot   *snap_in,
		nil,                     //OCISnapshot         *snap_out,
		C.OCI_DEFAULT)           //ub4                 mode );
	if r == C.OCI_ERROR {
		return nil, errE(stmt.ses.srv.env.ociError())
	}
	if stmt.hasPtrBind { // set any bind pointers
		err = stmt.setBindPtrs()
		if err != nil {
			return nil, errE(err)
		}
	}
	// create result set and open
	rset = _drv.rsetPool.Get().(*Rset)
	if rset.id == 0 {
		rset.id = _drv.rsetId.nextId()
	}
	err = rset.open(stmt, stmt.ocistmt)
	if err != nil {
		rset.close()
		return nil, errE(err)
	}
	stmt.openRsets.add(rset)

	return rset, nil
}

// setBindPtrs enables binds to set out pointers for some types such as time.Time, etc.
func (stmt *Stmt) setBindPtrs() (err error) {
	for _, bind := range stmt.bnds {
		err = bind.setPtr()
		if err != nil {
			return errE(err)
		}
	}
	return nil
}

// gets a bind struct from a driver slice. No locking occurs.
func (stmt *Stmt) getBnd(idx int) interface{} {
	return _drv.bndPools[idx].Get()
}

// puts a bind struct in the driver slice. No locking occurs.
func (stmt *Stmt) putBnd(idx int, bnd bnd) {
	_drv.bndPools[idx].Put(bnd)
}

// bind associates Go variables to SQL string placeholders by the
// position of the variable and the position of the placeholder.
//
// The first placeholder starts at position 1.
//
// The placeholder represents an input bind when the value is a built-in value type
// or an array or slice of builtin value types. The placeholder represents an
// output bind when the value is a pointer to a built-in value type
// or an array or slice of pointers to builtin value types.
//
// No locking occurs.
func (stmt *Stmt) bind(params []interface{}, isAssocArray bool) (iterations uint32, err error) {
	stmt.logF(_drv.cfg.Log.Stmt.Bind, "Params %d", len(params))
	// Create binds for each parameter; bind position is 1-based
	if len(params) == 0 {
		return 1, nil
	}
	var n int
	defer func() {
		if err != nil {
			stmt.logF(true, "bind %d. (%T:%#v): %+v", n, params[n], params[n], err)
		}
	}()
	iterations = 1
	stmt.bnds = make([]bnd, len(params))
	for n = range params {
		//stmt.logF(_drv.cfg.Log.Stmt.Bind, "params[%d]=(%v %T)", n, params[n], params[n])
		switch value := params[n].(type) {
		case int64:
			bnd := stmt.getBnd(bndIdxInt64).(*bndInt64)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case int32:
			bnd := stmt.getBnd(bndIdxInt32).(*bndInt32)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case int16:
			bnd := stmt.getBnd(bndIdxInt16).(*bndInt16)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case int8:
			bnd := stmt.getBnd(bndIdxInt8).(*bndInt8)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case uint64:
			bnd := stmt.getBnd(bndIdxUint64).(*bndUint64)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case uint32:
			bnd := stmt.getBnd(bndIdxUint32).(*bndUint32)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case uint16:
			bnd := stmt.getBnd(bndIdxUint16).(*bndUint16)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case uint8:
			bnd := stmt.getBnd(bndIdxUint8).(*bndUint8)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case float64:
			bnd := stmt.getBnd(bndIdxFloat64).(*bndFloat64)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case float32:
			bnd := stmt.getBnd(bndIdxFloat32).(*bndFloat32)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case Int64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt64).(*bndInt64)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Int32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt32).(*bndInt32)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Int16:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt16).(*bndInt16)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Int8:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt8).(*bndInt8)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint64).(*bndUint64)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint32).(*bndUint32)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint16:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint16).(*bndUint16)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint8:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint8).(*bndUint8)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Float64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BDOUBLE)
			} else {
				bnd := stmt.getBnd(bndIdxFloat64).(*bndFloat64)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Float32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BFLOAT)
			} else {
				bnd := stmt.getBnd(bndIdxFloat32).(*bndFloat32)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Num:
			bnd := stmt.getBnd(bndIdxNumString).(*bndNumString)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case OraNum:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_VNU)
			} else {
				bnd := stmt.getBnd(bndIdxNumString).(*bndNumString)
				stmt.bnds[n] = bnd
				err = bnd.bind(Num(value.Value), n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *int64:
			bnd := stmt.getBnd(bndIdxInt64Ptr).(*bndInt64Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int32:
			bnd := stmt.getBnd(bndIdxInt32Ptr).(*bndInt32Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int16:
			bnd := stmt.getBnd(bndIdxInt16Ptr).(*bndInt16Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int8:
			bnd := stmt.getBnd(bndIdxInt8Ptr).(*bndInt8Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint64:
			bnd := stmt.getBnd(bndIdxUint64Ptr).(*bndUint64Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint32:
			bnd := stmt.getBnd(bndIdxUint32Ptr).(*bndUint32Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint16:
			bnd := stmt.getBnd(bndIdxUint16Ptr).(*bndUint16Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint8:
			bnd := stmt.getBnd(bndIdxUint8Ptr).(*bndUint8Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *float64:
			bnd := stmt.getBnd(bndIdxFloat64Ptr).(*bndFloat64Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *float32:
			bnd := stmt.getBnd(bndIdxFloat32Ptr).(*bndFloat32Ptr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true

		case []int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint8: // the same as []byte !
			if stmt.cfg.byteSlice == U8 {
				bnd := stmt.getBnd(bndIdxUint8Slice).(*bndUint8Slice)
				stmt.bnds[n] = bnd
				iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray)
				if err != nil {
					return iterations, err
				}
			} else {
				switch bnd := stmt.getBnd(bndIdxBin).(type) {
				case *bndBin:
					stmt.bnds[n] = bnd
					err = bnd.bind(value, n+1, stmt)
					if err != nil {
						return iterations, err
					}
				case *bndLob:
					if value == nil {
						stmt.setNilBind(n, C.SQLT_BLOB)
					} else {
						stmt.bnds[n] = bnd
						err = bnd.bindReader(bytes.NewReader(value), n+1, stmt.cfg.lobBufferSize, C.SQLT_BLOB, stmt)
						if err != nil {
							return iterations, err
						}
					}
				default:
					panic(fmt.Errorf("awaited *ora.bndBin, got %T", bnd))
				}
			}
		case *[]int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case []float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			stmt.bnds[n] = bnd
			var err error
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			stmt.bnds[n] = bnd
			var err error
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}

		case []Num:
			bnd := stmt.getBnd(bndIdxNumStringSlice).(*bndNumStringSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, nil, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case []Int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]Int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]Int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []Int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint8:
			bnd := stmt.getBnd(bndIdxUint8Slice).(*bndUint8Slice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []OraNum:
			bnd := stmt.getBnd(bndIdxNumStringSlice).(*bndNumStringSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case time.Time:
			bnd := stmt.getBnd(bndIdxTime).(*bndTime)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case *time.Time:
			bnd := stmt.getBnd(bndIdxTimePtr).(*bndTimePtr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []time.Time:
			bnd := stmt.getBnd(bndIdxTimeSlice).(*bndTimeSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case Time:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_TIMESTAMP_TZ)
			} else {
				bnd := stmt.getBnd(bndIdxTime).(*bndTime)
				stmt.bnds[n] = bnd
				if err = bnd.bind(value.Value, n+1, stmt); err != nil {
					return iterations, err
				}
			}
		case *Time:
			bnd := stmt.getBnd(bndIdxTimePtr).(*bndTimePtr)
			stmt.bnds[n] = bnd
			if err = bnd.bind(&value.Value, n+1, stmt); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Time:
			bnd := stmt.getBnd(bndIdxTimeSlice).(*bndTimeSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case Date:
			if value.IsNull() {
				stmt.setNilBind(n, C.SQLT_DAT)
			} else {
				bnd := stmt.getBnd(bndIdxDate).(*bndDate)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Date, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Date:
			bnd := stmt.getBnd(bndIdxDatePtr).(*bndDatePtr)
			stmt.bnds[n] = bnd
			if err = bnd.bind(value, n+1, stmt); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Date:
			bnd := stmt.getBnd(bndIdxDateSlice).(*bndDateSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Date:
			bnd := stmt.getBnd(bndIdxDateSlice).(*bndDateSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case string:
			bnd := stmt.getBnd(bndIdxString).(*bndString)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
		case *string:
			bnd := stmt.getBnd(bndIdxStringPtr).(*bndStringPtr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt.cfg.stringPtrBufferSize, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case String:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				bnd := stmt.getBnd(bndIdxString).(*bndString)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []string:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []String:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]string:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bind(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]String:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			stmt.bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, n+1, stmt, isAssocArray); err != nil {
				return iterations, err
			}

		case bool:
			bnd := stmt.getBnd(bndIdxBool).(*bndBool)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt.cfg, stmt)
			if err != nil {
				return iterations, err
			}
		case *bool:
			bnd := stmt.getBnd(bndIdxBoolPtr).(*bndBoolPtr)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt.cfg.TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Bool:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				bnd := stmt.getBnd(bndIdxBool).(*bndBool)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt.cfg, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []bool:
			bnd := stmt.getBnd(bndIdxBoolSlice).(*bndBoolSlice)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, nil, n+1, stmt.cfg.FalseRune, stmt.cfg.TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			iterations = uint32(len(value))
		case []Bool:
			bnd := stmt.getBnd(bndIdxBoolSlice).(*bndBoolSlice)
			stmt.bnds[n] = bnd
			err = bnd.bindOra(value, n+1, stmt.cfg.FalseRune, stmt.cfg.TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			iterations = uint32(len(value))
		case Raw:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BIN)
			} else {
				bnd := stmt.getBnd(bndIdxBin).(*bndBin)
				stmt.bnds[n] = bnd
				err = bnd.bind(value.Value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Lob:
			sqlt := C.ub2(C.SQLT_BLOB)
			if value.C {
				sqlt = C.SQLT_CLOB
			}
			if value.Reader == nil {
				stmt.setNilBind(n, sqlt)
			} else {
				bnd := stmt.getBnd(bndIdxLob).(*bndLob)
				stmt.bnds[n] = bnd
				err = bnd.bindReader(value.Reader, n+1, stmt.cfg.lobBufferSize, sqlt, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Lob:
			sqlt := C.ub2(C.SQLT_BLOB)
			if value != nil && value.C {
				sqlt = C.SQLT_CLOB
			}
			if value == nil {
				stmt.setNilBind(n, sqlt)
			} else {
				bnd := stmt.getBnd(bndIdxLobPtr).(*bndLobPtr)
				stmt.bnds[n] = bnd
				err = bnd.bindLob(value, n+1, stmt.cfg.lobBufferSize, sqlt, stmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			}

		case [][]byte:
			bnd := stmt.getBnd(bndIdxBinSlice).(*bndBinSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, nil, n+1, stmt.cfg.lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Raw:
			bnd := stmt.getBnd(bndIdxBinSlice).(*bndBinSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(value, n+1, stmt.cfg.lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Lob:
			bnd := stmt.getBnd(bndIdxLobSlice).(*bndLobSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bindOra(value, n+1, stmt.cfg.lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

			// FIXME(tgulacsi): []*Lob ?

		case IntervalYM:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INTERVAL_YM)
			} else {
				bnd := stmt.getBnd(bndIdxIntervalYM).(*bndIntervalYM)
				stmt.bnds[n] = bnd
				err = bnd.bind(value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []IntervalYM:
			bnd := stmt.getBnd(bndIdxIntervalYMSlice).(*bndIntervalYMSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case IntervalDS:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INTERVAL_DS)
			} else {
				bnd := stmt.getBnd(bndIdxIntervalDS).(*bndIntervalDS)
				stmt.bnds[n] = bnd
				err = bnd.bind(value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []IntervalDS:
			bnd := stmt.getBnd(bndIdxIntervalDSSlice).(*bndIntervalDSSlice)
			stmt.bnds[n] = bnd
			iterations, err = bnd.bind(value, n+1, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case Bfile:
			if value.IsNull {
				err = stmt.setNilBind(n, C.SQLT_FILE)
			} else {
				bnd := stmt.getBnd(bndIdxBfile).(*bndBfile)
				stmt.bnds[n] = bnd
				err = bnd.bind(value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Rset:
			bnd := stmt.getBnd(bndIdxRset).(*bndRset)
			stmt.bnds[n] = bnd
			err = bnd.bind(value, n+1, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		default:
			if params[n] == nil {
				err = stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				t := reflect.TypeOf(params[n])
				if t.Kind() == reflect.Slice {
					if t.Elem().Kind() == reflect.Interface {
						return iterations, errF("Invalid bind parameter. ([]interface{}) (%v).", params[n])
					}
				}
				return iterations, errF("Invalid bind parameter (%v) (%T:%v).", t.Name(), params[n], params[n])
			}
		}
	}

	return iterations, err
}

// NumRset returns the number of open Oracle result sets.
func (stmt *Stmt) NumRset() int {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	return stmt.openRsets.len()
}

// NumInput returns the number of placeholders in a sql statement.
func (stmt *Stmt) NumInput() int {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	bc, err := stmt.attr(4, C.OCI_ATTR_BIND_COUNT)
	if err != nil {
		return 0
	}
	bindCount := int(*((*C.ub4)(bc)))
	C.free(bc)
	return bindCount
}

// SetGcts sets a slice of GoColumnType used in a Stmt.Qry *ora.Rset.
//
// SetGcts is optional.
func (stmt *Stmt) SetGcts(gcts []GoColumnType) []GoColumnType {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	old := stmt.gcts
	stmt.gcts = gcts
	return old
}

// Gcts returns a slice of GoColumnType specified by Ses.Prep or Stmt.SetGcts.
//
// Gcts is used by a Stmt.Qry *ora.Rset to determine which Go types are mapped
// to a sql select-list.
func (stmt *Stmt) Gcts() []GoColumnType {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	return stmt.gcts
}

// SetCfg applies the specified cfg to the Stmt.
//
// Open Rsets do not observe the specified cfg.
func (stmt *Stmt) SetCfg(cfg *StmtCfg) {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	stmt.cfg = *cfg
}

// Cfg returns the Stmt's cfg.
func (stmt *Stmt) Cfg() *StmtCfg {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	return &stmt.cfg
}

// IsOpen returns true when a statement is open; otherwise, false.
//
// Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Stmt.Prep to create a new statement.
func (stmt *Stmt) IsOpen() bool {
	stmt.mu.Lock()
	defer stmt.mu.Unlock()
	return stmt.ocistmt != nil
}

// checkClosed returns an error if Stmt is closed. No locking occurs.
func (stmt *Stmt) checkClosed() error {
	if stmt.ocistmt == nil {
		return er("Stmt is closed.")
	}
	return nil
}

// sysName returns a string representing the Stmt.
func (stmt *Stmt) sysName() string {
	if stmt == nil {
		return "E_S_S_S_"
	}
	return stmt.sysNamer.Name(func() string { return fmt.Sprintf("%sS%v", stmt.ses.sysName(), stmt.id) })
}

// log writes a message with an Stmt system name and caller info.
func (stmt *Stmt) log(enabled bool, v ...interface{}) {
	logL(stmt.sysName(), enabled, v...)
}

// log writes a formatted message with an Stmt system name and caller info.
func (stmt *Stmt) logF(enabled bool, format string, v ...interface{}) {
	logF(stmt.sysName(), enabled, format, v...)
}

// set prefetch size. No locking occurs.
func (stmt *Stmt) setPrefetchSize() error {
	if stmt.cfg.prefetchRowCount > 0 {
		//fmt.Println("stmt.setPrefetchSize: prefetchRowCount ", stmt.cfg.prefetchRowCount)
		// set prefetch row count
		if err := stmt.setAttr(stmt.cfg.prefetchRowCount, C.OCI_ATTR_PREFETCH_ROWS); err != nil {
			return errE(err)
		}
	} else if stmt.cfg.prefetchMemorySize > 0 {
		//fmt.Println("stmt.setPrefetchSize: prefetchMemorySize ", stmt.cfg.prefetchMemorySize)
		// Set prefetch memory size
		if err := stmt.setAttr(stmt.cfg.prefetchMemorySize, C.OCI_ATTR_PREFETCH_MEMORY); err != nil {
			return errE(err)
		}
	}
	return nil
}

// attr gets an attribute from the statement handle. No locking occurs.
func (stmt *Stmt) attr(attrSize C.ub4, attrType C.ub4) (unsafe.Pointer, error) {
	attrup := C.malloc(C.size_t(attrSize))
	r := C.OCIAttrGet(
		unsafe.Pointer(stmt.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4         cfgtrghndltyp,
		attrup,                       //void           *attributep,
		&attrSize,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		stmt.ses.srv.env.ocierr,      //OCIError       *errhp
	)
	if r == C.OCI_ERROR {
		C.free(unsafe.Pointer(attrup))
		return nil, stmt.ses.srv.env.ociError()
	}
	return attrup, nil
}

// setAttr sets an attribute on the statement handle. No locking occurs.
func (stmt *Stmt) setAttr(attrValue uint32, attrType C.ub4) error {
	r := C.OCIAttrSet(
		unsafe.Pointer(stmt.ocistmt), //void        *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4         trghndltyp,
		unsafe.Pointer(&attrValue),   //void        *attributep,
		4,                       //ub4         size,
		attrType,                //ub4         attrtype,
		stmt.ses.srv.env.ocierr) //OCIError    *errhp );
	if r == C.OCI_ERROR {
		return errE(stmt.ses.srv.env.ociError())
	}

	return nil
}

// setNilBind sets a nil bind. No locking occurs.
func (stmt *Stmt) setNilBind(index int, sqlt C.ub2) (err error) {
	bnd := _drv.bndPools[bndIdxNil].Get().(*bndNil)
	stmt.bnds[index] = bnd
	err = bnd.bind(index+1, sqlt, stmt)
	return err
}
