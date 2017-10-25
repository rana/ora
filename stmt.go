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
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
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
	sync.RWMutex

	id uint64
	// protects that open/close should not happen at once
	cmu                 sync.Mutex
	cfg                 atomic.Value
	env                 atomic.Value // we need to cache the env here
	ses                 *Ses
	ocistmt             *C.OCIStmt
	stmtType            C.ub2
	sql                 string
	gcts                []GoColumnType
	bnds                []bnd
	hasPtrBind          bool
	stringPtrBufferSize int
	bindInfo

	openRsets *rsetList

	sysNamer
}

// Cfg returns the Stmt's StmtCfg, or it's Ses's, if not set.
// If the env is the PkgSqlEnv, that will override StmtCfg!
func (stmt *Stmt) Cfg() StmtCfg {
	cfg := stmt.SelfCfg()
	if !cfg.IsZero() {
		return cfg
	}
	if env := stmt.Env(); env.isPkgEnv {
		return env.Cfg()
	}
	return stmt.ses.Cfg().StmtCfg
}

// returns the Stmt's StmtCfg only
func (stmt *Stmt) SelfCfg() StmtCfg {
	if c := stmt.cfg.Load(); c != nil {
		return c.(StmtCfg)
	}
	return StmtCfg{}
}

func (stmt *Stmt) SetCfg(cfg StmtCfg) {
	stmt.cfg.Store(cfg)
}

func (stmt *Stmt) Env() *Env {
	e := stmt.env.Load()
	if e == nil {
		return nil
	}
	return e.(*Env)
}

// Close closes the SQL statement.
//
// Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Stmt.Prep to create a new statement.
func (stmt *Stmt) Close() (err error) {
	return stmt.closeWithRemove()
}
func (stmt *Stmt) closeWithRemove() error {
	if stmt == nil {
		return nil
	}
	stmt.RLock()
	ses := stmt.ses
	stmt.RUnlock()
	if ses == nil {
		return nil
	}
	if ses.openStmts != nil {
		ses.openStmts.remove(stmt)
	}
	return stmt.close()
}

// close closes the SQL statement, without locking stmt.
// does not remove Stmt from Ses.openStmts
func (stmt *Stmt) close() (err error) {
	//fmt.Println("close " + stmt.sysName())
	stmt.log(_drv.Cfg().Log.Stmt.Close)
	err = stmt.checkClosed()
	if err != nil {
		return errE(err)
	}

	stmt.cmu.Lock()
	defer stmt.cmu.Unlock()

	errs := _drv.listPool.Get().(*list.List)
	defer func() {
		if value := recover(); value != nil {
			err := errR(value)
			stmt.logF(true, "PANIC %v", err)
			errs.PushBack(err)
		}
		stmt.Lock()
		env := stmt.Env()
		ocistmt := stmt.ocistmt

		stmt.stringPtrBufferSize = 0
		stmt.env.Store((*Env)(nil))
		stmt.ses = nil
		stmt.ocistmt = nil
		stmt.stmtType = 0
		stmt.sql = ""
		stmt.gcts = nil
		stmt.bnds = nil
		stmt.hasPtrBind = false
		stmt.bindInfo = bindInfo{}
		stmt.openRsets.clear()
		_drv.stmtPool.Put(stmt)
		stmt.Unlock()
		stmt.SetCfg(StmtCfg{})

		if ocistmt != nil {
			// free ocistmt to release cursor on server
			// OCIStmtRelease must be called with OCIStmtPrepare2
			// See https://docs.oracle.com/database/121/LNOCI/oci09adv.htm#LNOCI16655
			r := C.OCIStmtRelease(
				ocistmt,       // OCIStmt        *stmthp
				env.ocierr,    // OCIError       *errhp,
				nil,           // const OraText  *key
				C.ub4(0),      // ub4 keylen
				C.OCI_DEFAULT, // ub4 mode
			)
			if r == C.OCI_ERROR {
				errs.PushBack(errE(env.ociError()))
				// Sometimes panics if free unconditionally - see #222.
				// https://github.com/rana/ora/issues/222
				C.OCIHandleFree(unsafe.Pointer(ocistmt), C.OCI_HTYPE_STMT)
			}
		}

		multiErr := newMultiErrL(errs)
		if multiErr != nil {
			err = errE(*multiErr)
		}
		errs.Init()
		_drv.listPool.Put(errs)
	}()
	// close binds
	stmt.Lock()
	for _, bind := range stmt.bnds {
		if bind == nil {
			continue
		}
		err = bind.close()
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	openRsets := stmt.openRsets
	stmt.Unlock()
	//fmt.Println("closeAll " + stmt.sysName())
	openRsets.closeAll(errs)

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

// Parse the statement, and return the syntax errors - WITHOUT executing it.
// Rejects ALTER statements, as they're executed anyway by Oracle...
func (stmt *Stmt) Parse() (err error) {
	if stmt == nil {
		return er("stmt may not be nil.")
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt.log(_drv.Cfg().Log.Stmt.Exe)
	err = stmt.checkClosed()
	if err != nil {
		return errE(err)
	}
	if stmt.stmtType == C.OCI_STMT_ALTER || stmt.stmtType == 0 {
		return er("parsing ALTER statement is perilous!")
	}
	// Execute statement on Oracle server
	stmt.RLock()
	env := stmt.Env()
	stmt.ses.RLock()
	r := C.OCIStmtExecute(
		stmt.ses.ocisvcctx, //OCISvcCtx           *svchp,
		stmt.ocistmt,       //OCIStmt             *stmtp,
		env.ocierr,         //OCIError            *errhp,
		C.ub4(1),           //ub4                 iters,
		C.ub4(0),           //ub4                 rowoff,
		nil,                //const OCISnapshot   *snap_in,
		nil,                //OCISnapshot         *snap_out,
		C.OCI_PARSE_ONLY)   //ub4                 mode );
	stmt.ses.RUnlock()
	stmt.RUnlock()
	if r == C.OCI_ERROR {
		return errE(env.ociError())
	}
	return nil
}

var spcRpl = strings.NewReplacer("\t", " ", "   ", " ", "  ", " ")

// exe executes a SQL statement on an Oracle server returning rowsAffected, lastInsertId and error.
func (stmt *Stmt) exe(params []interface{}, isAssocArray bool) (rowsAffected uint64, lastInsertId int64, err error) {
	return stmt.exeC(context.Background(), params, isAssocArray)
}
func (stmt *Stmt) exeC(ctx context.Context, params []interface{}, isAssocArray bool) (rowsAffected uint64, lastInsertId int64, err error) {
	if stmt == nil {
		return 0, 0, er("stmt may not be nil.")
	}
	if err = ctx.Err(); err != nil {
		return
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt.log(_drv.Cfg().Log.Stmt.Exe)
	err = stmt.checkClosed()
	if err != nil {
		return 0, 0, errE(err)
	}
	if cfg, ok := ctxStmtCfg(ctx); ok {
		stmt.SetCfg(cfg)
	}
	// for case of inserting and returning identity for database/sql package
	stmt.RLock()
	pkgEnvInsert := stmt.Env().isPkgEnv && stmt.stmtType == C.OCI_STMT_INSERT
	stmt.RUnlock()
	if pkgEnvInsert {
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
	if stmt.stmtType == C.OCI_STMT_SELECT {
		err = stmt.setPrefetchSize() // set prefetch size
		if err != nil {
			return 0, 0, errE(err)
		}
	}
	mode := C.ub4(C.OCI_DEFAULT) // determine auto-commit state; don't auto-comit if there's an explicit user transaction occuring
	var autoCommit bool
	if stmt.Cfg().IsAutoCommitting {
		stmt.RLock()
		n := stmt.ses.openTxs.len()
		stmt.RUnlock()
		if n == 0 {
			mode = C.OCI_COMMIT_ON_SUCCESS
			autoCommit = true
		}
	}
	stmt.logF(_drv.Cfg().Log.Stmt.Exe, "iterations=%d autoCommit=%t", iterations, autoCommit)
	// Execute statement on Oracle server
	stmt.RLock()
	env := stmt.Env()
	stmt.ses.RLock()
	r := C.OCIStmtExecute(
		stmt.ses.ocisvcctx, //OCISvcCtx           *svchp,
		stmt.ocistmt,       //OCIStmt             *stmtp,
		env.ocierr,         //OCIError            *errhp,
		C.ub4(iterations),  //ub4                 iters,
		C.ub4(0),           //ub4                 rowoff,
		nil,                //const OCISnapshot   *snap_in,
		nil,                //OCISnapshot         *snap_out,
		mode)               //ub4                 mode );
	stmt.ses.RUnlock()
	stmtType, hasPtrBind := stmt.stmtType, stmt.hasPtrBind
	stmt.RUnlock()
	stmt.logF(_drv.Cfg().Log.Stmt.Exe, "returned %d, hasPtrBind=%t", r, hasPtrBind)
	if r == C.OCI_ERROR {
		return 0, 0, errE(env.ociError())
	}
	// Get rowsAffected based on statement type
	switch stmtType {
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
			return 0, 0, errE(env.ociError())
		}
		//fmt.Printf("stmtType=%d\n", stmt.stmtType)
	}
	if hasPtrBind { // Set any bind pointers
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
	return stmt.qryC(context.Background(), params)
}
func (stmt *Stmt) qryC(ctx context.Context, params []interface{}) (rset *Rset, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt.log(_drv.Cfg().Log.Stmt.Qry)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cfg, ok := ctxStmtCfg(ctx); ok {
		stmt.SetCfg(cfg)
	}

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
	stmt.RLock()
	env := stmt.Env()
	stmt.ses.RLock()
	r := C.OCIStmtExecute(
		//stmt.ses.ocisvcctx,      //OCISvcCtx           *svchp,
		stmt.ses.ocisvcctx, //OCISvcCtx           *svchp,
		stmt.ocistmt,       //OCIStmt             *stmtp,
		env.ocierr,         //OCIError            *errhp,
		C.ub4(0),           //ub4                 iters,
		C.ub4(0),           //ub4                 rowoff,
		nil,                //const OCISnapshot   *snap_in,
		nil,                //OCISnapshot         *snap_out,
		C.OCI_DEFAULT)      //ub4                 mode );
	stmt.ses.RUnlock()
	hasPtrBind := stmt.hasPtrBind
	stmt.RUnlock()
	if r == C.OCI_ERROR {
		return nil, errE(env.ociError())
	}
	if hasPtrBind { // set any bind pointers
		err = stmt.setBindPtrs()
		if err != nil {
			return nil, errE(err)
		}
	}
	// create result set and open
	// FIXME(tgulacsi): reusing Rsets causes sporadic failures.
	//rset = _drv.rsetPool.Get().(*Rset)
	rset = &Rset{}
	//rset.Lock()
	rset.env = env
	if rset.id == 0 {
		rset.id = _drv.rsetId.nextId()
	}
	//rset.Unlock()
	err = rset.open(stmt, stmt.ocistmt)
	if err != nil {
		rset.close()
		return nil, errE(err)
	}
	stmt.RLock()
	stmt.openRsets.add(rset)
	stmt.RUnlock()

	return rset, nil
}

// setBindPtrs enables binds to set out pointers for some types such as time.Time, etc.
func (stmt *Stmt) setBindPtrs() (err error) {
	stmt.RLock()
	defer stmt.RUnlock()
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
	stmt.logF(_drv.Cfg().Log.Stmt.Bind, "Params %d", len(params))
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
	stmt.RLock()
	bnds := stmt.bnds
	stmt.RUnlock()
	if cap(bnds) < len(params) {
		bnds = make([]bnd, len(params))
	} else {
		bnds = bnds[:len(params)]
	}
	stmt.Lock()
	stmt.bnds = bnds
	defer stmt.Unlock()
	for n = range params {
		name, v := nameAndValue(params[n])
		pos := namedPos{Ordinal: n + 1, Name: name}
		//stmt.logF(_drv.Cfg().Log.Stmt.Bind, "params[%d]=(%v %T)", n, params[n], params[n])
		switch value := v.(type) {
		case int64:
			bnd := stmt.getBnd(bndIdxInt64).(*bndInt64)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case int32:
			bnd := stmt.getBnd(bndIdxInt32).(*bndInt32)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case int16:
			bnd := stmt.getBnd(bndIdxInt16).(*bndInt16)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case int8:
			bnd := stmt.getBnd(bndIdxInt8).(*bndInt8)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case uint64:
			bnd := stmt.getBnd(bndIdxUint64).(*bndUint64)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case uint32:
			bnd := stmt.getBnd(bndIdxUint32).(*bndUint32)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case uint16:
			bnd := stmt.getBnd(bndIdxUint16).(*bndUint16)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case uint8:
			bnd := stmt.getBnd(bndIdxUint8).(*bndUint8)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case float64:
			bnd := stmt.getBnd(bndIdxFloat64).(*bndFloat64)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case float32:
			bnd := stmt.getBnd(bndIdxFloat32).(*bndFloat32)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case Int64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt64).(*bndInt64)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Int64:
			bnd := stmt.getBnd(bndIdxInt64Ptr).(*bndInt64Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Int32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt32).(*bndInt32)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Int32:
			bnd := stmt.getBnd(bndIdxInt32Ptr).(*bndInt32Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}

			stmt.hasPtrBind = true
		case Int16:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt16).(*bndInt16)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Int16:
			bnd := stmt.getBnd(bndIdxInt16Ptr).(*bndInt16Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}

			stmt.hasPtrBind = true
		case Int8:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INT)
			} else {
				bnd := stmt.getBnd(bndIdxInt8).(*bndInt8)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Int8:
			bnd := stmt.getBnd(bndIdxInt8Ptr).(*bndInt8Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}

			stmt.hasPtrBind = true
		case Uint64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint64).(*bndUint64)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint32).(*bndUint32)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint16:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint16).(*bndUint16)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Uint8:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_UIN)
			} else {
				bnd := stmt.getBnd(bndIdxUint8).(*bndUint8)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case Float64:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BDOUBLE)
			} else {
				bnd := stmt.getBnd(bndIdxFloat64).(*bndFloat64)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Float64:
			bnd := stmt.getBnd(bndIdxFloat64Ptr).(*bndFloat64Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Float32:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BFLOAT)
			} else {
				bnd := stmt.getBnd(bndIdxFloat32).(*bndFloat32)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Float32:
			bnd := stmt.getBnd(bndIdxFloat32Ptr).(*bndFloat32Ptr)
			bnds[n] = bnd
			err = bnd.bind(&(value.Value), &(value.IsNull), pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Num:
			bnd := stmt.getBnd(bndIdxNumString).(*bndNumString)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case OraNum:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_VNU)
			} else {
				bnd := stmt.getBnd(bndIdxNumString).(*bndNumString)
				bnds[n] = bnd
				err = bnd.bind(Num(value.Value), pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case OCINum:
			bnd := stmt.getBnd(bndIdxOCINum).(*bndOCINum)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}

		case *int64:
			bnd := stmt.getBnd(bndIdxInt64Ptr).(*bndInt64Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int32:
			bnd := stmt.getBnd(bndIdxInt32Ptr).(*bndInt32Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int16:
			bnd := stmt.getBnd(bndIdxInt16Ptr).(*bndInt16Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *int8:
			bnd := stmt.getBnd(bndIdxInt8Ptr).(*bndInt8Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint64:
			bnd := stmt.getBnd(bndIdxUint64Ptr).(*bndUint64Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint32:
			bnd := stmt.getBnd(bndIdxUint32Ptr).(*bndUint32Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint16:
			bnd := stmt.getBnd(bndIdxUint16Ptr).(*bndUint16Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *uint8:
			bnd := stmt.getBnd(bndIdxUint8Ptr).(*bndUint8Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *float64:
			bnd := stmt.getBnd(bndIdxFloat64Ptr).(*bndFloat64Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *float32:
			bnd := stmt.getBnd(bndIdxFloat32Ptr).(*bndFloat32Ptr)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *Num:
			bnd := stmt.getBnd(bndIdxNumStringPtr).(*bndNumStringPtr)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *OraNum:
			bnd := stmt.getBnd(bndIdxNumStringPtr).(*bndNumStringPtr)
			bnds[n] = bnd
			err = bnd.bind((*Num)(&value.Value), pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *OCINum:
			bnd := stmt.getBnd(bndIdxOCINumPtr).(*bndOCINumPtr)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true

		case []int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []uint8: // the same as []byte !
			if stmt.Cfg().byteSlice == U8 {
				bnd := stmt.getBnd(bndIdxUint8Slice).(*bndUint8Slice)
				bnds[n] = bnd
				iterations, err = bnd.bind(&value, pos, stmt, isAssocArray)
				if err != nil {
					return iterations, err
				}
			} else {
				switch bnd := stmt.getBnd(bndIdxBin).(type) {
				case *bndBin:
					bnds[n] = bnd
					err = bnd.bind(value, pos, stmt)
					if err != nil {
						return iterations, err
					}
				case *bndLob:
					if value == nil {
						stmt.setNilBind(n, C.SQLT_BLOB)
					} else {
						bnds[n] = bnd
						err = bnd.bindReader(bytes.NewReader(value), pos, stmt.Cfg().lobBufferSize, C.SQLT_BLOB, stmt)
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
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case *[]uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case []float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			bnds[n] = bnd
			var err error
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			bnds[n] = bnd
			var err error
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}

		case []Num:
			bnd := stmt.getBnd(bndIdxNumStringSlice).(*bndNumStringSlice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, nil, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []OCINum:
			bnd := stmt.getBnd(bndIdxOCINumSlice).(*bndOCINumSlice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, nil, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case []Int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Int64:
			bnd := stmt.getBnd(bndIdxInt64Slice).(*bndInt64Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]Int32:
			bnd := stmt.getBnd(bndIdxInt32Slice).(*bndInt32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]Int16:
			bnd := stmt.getBnd(bndIdxInt16Slice).(*bndInt16Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case []Int8:
			bnd := stmt.getBnd(bndIdxInt8Slice).(*bndInt8Slice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint64:
			bnd := stmt.getBnd(bndIdxUint64Slice).(*bndUint64Slice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint32:
			bnd := stmt.getBnd(bndIdxUint32Slice).(*bndUint32Slice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint16:
			bnd := stmt.getBnd(bndIdxUint16Slice).(*bndUint16Slice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Uint8:
			bnd := stmt.getBnd(bndIdxUint8Slice).(*bndUint8Slice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
		case []Float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Float64:
			bnd := stmt.getBnd(bndIdxFloat64Slice).(*bndFloat64Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Float32:
			bnd := stmt.getBnd(bndIdxFloat32Slice).(*bndFloat32Slice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []OraNum:
			bnd := stmt.getBnd(bndIdxNumStringSlice).(*bndNumStringSlice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}

		case time.Time:
			bnd := stmt.getBnd(bndIdxTime).(*bndTime)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case *time.Time:
			bnd := stmt.getBnd(bndIdxTimePtr).(*bndTimePtr)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []time.Time:
			bnd := stmt.getBnd(bndIdxTimeSlice).(*bndTimeSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case Time:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_TIMESTAMP_TZ)
			} else {
				bnd := stmt.getBnd(bndIdxTime).(*bndTime)
				bnds[n] = bnd
				if err = bnd.bind(value.Value, pos, stmt); err != nil {
					return iterations, err
				}
			}
		case *Time:
			bnd := stmt.getBnd(bndIdxTimePtr).(*bndTimePtr)
			bnds[n] = bnd
			if err = bnd.bind(&value.Value, pos, stmt); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Time:
			bnd := stmt.getBnd(bndIdxTimeSlice).(*bndTimeSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case Date:
			if value.IsNull() {
				stmt.setNilBind(n, C.SQLT_DAT)
			} else {
				bnd := stmt.getBnd(bndIdxDate).(*bndDate)
				bnds[n] = bnd
				err = bnd.bind(value.Date, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Date:
			bnd := stmt.getBnd(bndIdxDatePtr).(*bndDatePtr)
			bnds[n] = bnd
			if err = bnd.bind(value, pos, stmt); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Date:
			bnd := stmt.getBnd(bndIdxDateSlice).(*bndDateSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case *[]Date:
			bnd := stmt.getBnd(bndIdxDateSlice).(*bndDateSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
		case string:
			bnd := stmt.getBnd(bndIdxString).(*bndString)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
		case *string:
			bnd := stmt.getBnd(bndIdxStringPtr).(*bndStringPtr)
			bnds[n] = bnd
			spbs := stmt.stringPtrBufferSize
			if spbs == 0 {
				spbs = stmt.Cfg().stringPtrBufferSize
			}
			err = bnd.bind(value, nil, pos, spbs, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case String:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				bnd := stmt.getBnd(bndIdxString).(*bndString)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *String:
			bnd := stmt.getBnd(bndIdxStringPtr).(*bndStringPtr)
			bnds[n] = bnd
			spbs := stmt.stringPtrBufferSize
			if spbs == 0 {
				spbs = stmt.Cfg().stringPtrBufferSize
			}
			str := &(value.Value)
			err = bnd.bind(str, &(value.IsNull), pos, spbs, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []string:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []String:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(&value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]string:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bind(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case *[]String:
			bnd := stmt.getBnd(bndIdxStringSlice).(*bndStringSlice)
			bnds[n] = bnd
			if iterations, err = bnd.bindOra(value, pos, stmt, isAssocArray); err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true

		case bool:
			bnd := stmt.getBnd(bndIdxBool).(*bndBool)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt.Cfg(), stmt)
			if err != nil {
				return iterations, err
			}
		case *bool:
			bnd := stmt.getBnd(bndIdxBoolPtr).(*bndBoolPtr)
			bnds[n] = bnd
			err = bnd.bind(value, pos, stmt.Cfg().TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Bool:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				bnd := stmt.getBnd(bndIdxBool).(*bndBool)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt.Cfg(), stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []bool:
			bnd := stmt.getBnd(bndIdxBoolSlice).(*bndBoolSlice)
			bnds[n] = bnd
			err = bnd.bind(value, nil, pos, stmt.Cfg().FalseRune, stmt.Cfg().TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			iterations = uint32(len(value))
			stmt.hasPtrBind = true
		case []Bool:
			bnd := stmt.getBnd(bndIdxBoolSlice).(*bndBoolSlice)
			bnds[n] = bnd
			err = bnd.bindOra(value, pos, stmt.Cfg().FalseRune, stmt.Cfg().TrueRune, stmt)
			if err != nil {
				return iterations, err
			}
			iterations = uint32(len(value))
			stmt.hasPtrBind = true

		case Raw:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_BIN)
			} else {
				bnd := stmt.getBnd(bndIdxBin).(*bndBin)
				bnds[n] = bnd
				err = bnd.bind(value.Value, pos, stmt)
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
				bnds[n] = bnd
				err = bnd.bindReader(value.Reader, pos, stmt.Cfg().lobBufferSize, sqlt, stmt)
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
				bnds[n] = bnd
				err = bnd.bindLob(value, pos, stmt.Cfg().lobBufferSize, sqlt, stmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			}
			stmt.hasPtrBind = true

		case [][]byte:
			bnd := stmt.getBnd(bndIdxBinSlice).(*bndBinSlice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, nil, pos, stmt.Cfg().lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Raw:
			bnd := stmt.getBnd(bndIdxBinSlice).(*bndBinSlice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(value, pos, stmt.Cfg().lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case []Lob:
			bnd := stmt.getBnd(bndIdxLobSlice).(*bndLobSlice)
			bnds[n] = bnd
			iterations, err = bnd.bindOra(value, pos, stmt.Cfg().lobBufferSize, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true

			// FIXME(tgulacsi): []*Lob ?

		case IntervalYM:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INTERVAL_YM)
			} else {
				bnd := stmt.getBnd(bndIdxIntervalYM).(*bndIntervalYM)
				bnds[n] = bnd
				err = bnd.bind(value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []IntervalYM:
			bnd := stmt.getBnd(bndIdxIntervalYMSlice).(*bndIntervalYMSlice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case IntervalDS:
			if value.IsNull {
				stmt.setNilBind(n, C.SQLT_INTERVAL_DS)
			} else {
				bnd := stmt.getBnd(bndIdxIntervalDS).(*bndIntervalDS)
				bnds[n] = bnd
				err = bnd.bind(value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case []IntervalDS:
			bnd := stmt.getBnd(bndIdxIntervalDSSlice).(*bndIntervalDSSlice)
			bnds[n] = bnd
			iterations, err = bnd.bind(value, pos, stmt, isAssocArray)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		case Bfile:
			if value.IsNull {
				err = stmt.setNilBind(n, C.SQLT_FILE)
			} else {
				bnd := stmt.getBnd(bndIdxBfile).(*bndBfile)
				bnds[n] = bnd
				err = bnd.bind(value, pos, stmt)
				if err != nil {
					return iterations, err
				}
			}
		case *Rset:
			bnd := stmt.getBnd(bndIdxRset).(*bndRset)
			bnds[n] = bnd
			value.env = stmt.Env()
			value.stmt = stmt
			err = bnd.bind(value, pos, stmt)
			if err != nil {
				return iterations, err
			}
			stmt.hasPtrBind = true
		default:
			if v == nil {
				err = stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				t := reflect.TypeOf(v)
				if t.Kind() == reflect.Slice &&
					t.Elem().Kind() == reflect.Interface {
					return iterations, errF("Invalid bind parameter. ([]interface{}) (%v).", v)
				}
				return iterations, errF("Invalid bind parameter (%v) (%T:%v).", t.Name(), v, v)
			}
		}
	}

	return iterations, err
}

// NumRset returns the number of open Oracle result sets.
func (stmt *Stmt) NumRset() int {
	stmt.RLock()
	defer stmt.RUnlock()
	return stmt.openRsets.len()
}

type bindInfo struct {
	BindNames, IndNames []string
	Duplicates          []bool
}

func (stmt *Stmt) getBindInfo() (bindNames, indNames []string, duplicates []bool, err error) {
	stmt.RLock()
	bi := stmt.bindInfo
	stmt.RUnlock()
	if bi.BindNames != nil {
		return bi.BindNames, bi.IndNames, bi.Duplicates, nil
	}

	const arrSize = 128

	cfg := _drv.Cfg()
	env := stmt.Env()
	startLoc := C.ub4(1)
	var found C.sb4
	var bndNms, indNms [arrSize]*C.OraText
	var bndNmLens, indNmLens, dups [arrSize]C.ub1
	var binds [arrSize]*C.OCIBind
	for {
		stmt.RLock()
		r := C.OCIStmtGetBindInfo(
			stmt.ocistmt,   // OCIStmt      *stmtp,
			env.ocierr,     // OCIError     *errhp,
			C.ub4(arrSize), // ub4          size,
			startLoc,       // ub4          startloc,
			&found,         // sb4          *found,
			&bndNms[0],     // OraText      *bvnp[],
			&bndNmLens[0],  // ub1          bvnl[],
			&indNms[0],     // OraText      *invp[],
			&indNmLens[0],  // ub1          inpl[],
			&dups[0],       // ub1          dupl[],
			&binds[0],      // OCIBind      *hndl[]
		)
		stmt.RUnlock()
		if r == C.OCI_ERROR {
			err = env.ociError()
			return
		}
		stmt.logF(cfg.Log.Stmt.Bind, "start=%d found=%d", startLoc, found)
		n := int(found)
		// The expression abs(found) gives the total number of bind variables
		// in the statement irrespective of the start position.a
		// Positive value if the number of bind variables returned is less than
		// the size provided, otherwise negative.
		if n < 0 {
			n = -n
		}
		n -= int(startLoc - 1)
		if n > arrSize {
			n = arrSize
		}
		for i := 0; i < n; i++ {
			bindNames = append(bindNames, C.GoStringN((*C.char)(unsafe.Pointer(bndNms[i])), C.int(bndNmLens[i])))
			indNames = append(indNames, C.GoStringN((*C.char)(unsafe.Pointer(indNms[i])), C.int(indNmLens[i])))
			duplicates = append(duplicates, dups[i] > 0)
		}
		if found >= 0 {
			stmt.Lock()
			stmt.bindInfo = bindInfo{BindNames: bindNames, IndNames: indNames, Duplicates: duplicates}
			stmt.Unlock()
			return
		}
		startLoc += C.ub4(n)
	}
}

// SetGcts sets a slice of GoColumnType used in a Stmt.Qry *ora.Rset.
//
// SetGcts is optional.
func (stmt *Stmt) SetGcts(gcts []GoColumnType) []GoColumnType {
	stmt.Lock()
	old := stmt.gcts
	stmt.gcts = gcts
	stmt.Unlock()
	return old
}

// Gcts returns a slice of GoColumnType specified by Ses.Prep or Stmt.SetGcts.
//
// Gcts is used by a Stmt.Qry *ora.Rset to determine which Go types are mapped
// to a sql select-list.
func (stmt *Stmt) Gcts() []GoColumnType {
	stmt.RLock()
	defer stmt.RUnlock()
	return stmt.gcts
}

// IsOpen returns true when a statement is open; otherwise, false.
//
// Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Stmt.Prep to create a new statement.
func (stmt *Stmt) IsOpen() bool {
	if stmt == nil {
		return false
	}
	stmt.RLock()
	defer stmt.RUnlock()
	return stmt.ocistmt != nil
}

// checkClosed returns an error if Stmt is closed. No locking occurs.
func (stmt *Stmt) checkClosed() error {
	if stmt == nil {
		return er("Stmt is closed.")
	}
	stmt.RLock()
	closed := stmt.ocistmt == nil
	stmt.RUnlock()
	if closed {
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
	if !_drv.Cfg().Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		_drv.Cfg().Log.Logger.Infof("%v %v", stmt.sysName(), callInfo(1))
	} else {
		_drv.Cfg().Log.Logger.Infof("%v %v %v", stmt.sysName(), callInfo(1), fmt.Sprint(v...))
	}
}

// log writes a formatted message with an Stmt system name and caller info.
func (stmt *Stmt) logF(enabled bool, format string, v ...interface{}) {
	Log := _drv.Cfg().Log
	if !Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		Log.Logger.Infof("%v %v", stmt.sysName(), callInfo(1))
	} else {
		Log.Logger.Infof("%v %v %v", stmt.sysName(), callInfo(1), fmt.Sprintf(format, v...))
	}
}

// set prefetch size. No locking occurs.
func (stmt *Stmt) setPrefetchSize() error {
	cfg := stmt.Cfg()
	if cfg.prefetchRowCount > 0 {
		//fmt.Println("stmt.setPrefetchSize: prefetchRowCount ", stmt.Cfg().prefetchRowCount)
		// set prefetch row count
		if err := stmt.setAttr(cfg.prefetchRowCount, C.OCI_ATTR_PREFETCH_ROWS); err != nil {
			return errE(err)
		}
	} else if cfg.prefetchMemorySize > 0 {
		//fmt.Println("stmt.setPrefetchSize: prefetchMemorySize ", stmt.Cfg().prefetchMemorySize)
		// Set prefetch memory size
		if err := stmt.setAttr(cfg.prefetchMemorySize, C.OCI_ATTR_PREFETCH_MEMORY); err != nil {
			return errE(err)
		}
	}
	return nil
}

// attr gets an attribute from the statement handle. No locking occurs.
func (stmt *Stmt) attr(attrSize C.ub4, attrType C.ub4) (unsafe.Pointer, error) {
	attrup := C.malloc(C.size_t(attrSize))
	stmt.RLock()
	env := stmt.Env()
	r := C.OCIAttrGet(
		unsafe.Pointer(stmt.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4         cfgtrghndltyp,
		attrup,                       //void           *attributep,
		&attrSize,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		env.ocierr,                   //OCIError       *errhp
	)
	stmt.RUnlock()
	if r == C.OCI_ERROR {
		C.free(unsafe.Pointer(attrup))
		return nil, env.ociError()
	}
	return attrup, nil
}

// setAttr sets an attribute on the statement handle. No locking occurs.
func (stmt *Stmt) setAttr(attrValue uint32, attrType C.ub4) error {
	stmt.RLock()
	env := stmt.Env()
	r := C.OCIAttrSet(
		unsafe.Pointer(stmt.ocistmt), //void        *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4         trghndltyp,
		unsafe.Pointer(&attrValue),   //void        *attributep,
		4,          //ub4         size,
		attrType,   //ub4         attrtype,
		env.ocierr) //OCIError    *errhp );
	stmt.RUnlock()
	if r == C.OCI_ERROR {
		return errE(env.ociError())
	}

	return nil
}

// setNilBind sets a nil bind. No locking occurs.
func (stmt *Stmt) setNilBind(index int, sqlt C.ub2) (err error) {
	bnd := _drv.bndPools[bndIdxNil].Get().(*bndNil)
	stmt.bnds[index] = bnd
	err = bnd.bind(namedPos{Ordinal: index + 1}, sqlt, stmt)
	return err
}
