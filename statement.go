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
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// A sql statement associated with an Oracle session.
//
// Implements the driver.Stmt interface.
type Statement struct {
	ocistmt *C.OCIStmt
	Config  StatementConfig

	env           *Environment
	ses           *Session
	elem          *list.Element
	rsts          *list.List
	binds         []bind
	goColumnTypes []GoColumnType
	sql           string
	stmtType      C.ub4
	hasPtrBind    bool
}

// bindParams associates Go variables to SQL string placeholders by the
// of the position of the variable and the position of the placeholder.
//
// The first placeholder starts at position 1.
//
// The placeholder represents an input bind when the value is a built-in value type
// or an array or slice of builtin value types. The placeholder represents an
// output bind when the value is a pointer to a built-in value type
// or an array or slice of pointers to builtin value types.
func (stmt *Statement) bindParams(params []interface{}) (iterations uint32, err error) {
	//fmt.Printf("Statement.bindParams: len(params) (%v)\n", len(params))

	iterations = 1
	// Create binds for each arg; bind position is 1-based
	if params != nil && len(params) > 0 {
		stmt.binds = make([]bind, len(params))
		for n := 0; n < len(params); n++ {
			//fmt.Printf("Statement.bindParams: params[%v] (%v)\n", n, params[n])
			if value, ok := params[n].(int64); ok {
				int64Bind := stmt.ses.srv.env.int64BindPool.Get().(*int64Bind)
				stmt.binds[n] = int64Bind
				err = int64Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int32); ok {
				int32Bind := stmt.ses.srv.env.int32BindPool.Get().(*int32Bind)
				stmt.binds[n] = int32Bind
				err = int32Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int16); ok {
				int16Bind := stmt.ses.srv.env.int16BindPool.Get().(*int16Bind)
				stmt.binds[n] = int16Bind
				err = int16Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int8); ok {
				int8Bind := stmt.ses.srv.env.int8BindPool.Get().(*int8Bind)
				stmt.binds[n] = int8Bind
				err = int8Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint64); ok {
				uint64Bind := stmt.ses.srv.env.uint64BindPool.Get().(*uint64Bind)
				stmt.binds[n] = uint64Bind
				err = uint64Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint32); ok {
				uint32Bind := stmt.ses.srv.env.uint32BindPool.Get().(*uint32Bind)
				stmt.binds[n] = uint32Bind
				err = uint32Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint16); ok {
				uint16Bind := stmt.ses.srv.env.uint16BindPool.Get().(*uint16Bind)
				stmt.binds[n] = uint16Bind
				err = uint16Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint8); ok {
				uint8Bind := stmt.ses.srv.env.uint8BindPool.Get().(*uint8Bind)
				stmt.binds[n] = uint8Bind
				err = uint8Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(float64); ok {
				float64Bind := stmt.ses.srv.env.float64BindPool.Get().(*float64Bind)
				stmt.binds[n] = float64Bind
				err = float64Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(float32); ok {
				float32Bind := stmt.ses.srv.env.float32BindPool.Get().(*float32Bind)
				stmt.binds[n] = float32Bind
				err = float32Bind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(Int64); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INT)
				} else {
					int64Bind := stmt.ses.srv.env.int64BindPool.Get().(*int64Bind)
					stmt.binds[n] = int64Bind
					err = int64Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int32); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INT)
				} else {
					int32Bind := stmt.ses.srv.env.int32BindPool.Get().(*int32Bind)
					stmt.binds[n] = int32Bind
					err = int32Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int16); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INT)
				} else {
					int16Bind := stmt.ses.srv.env.int16BindPool.Get().(*int16Bind)
					stmt.binds[n] = int16Bind
					err = int16Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int8); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INT)
				} else {
					int8Bind := stmt.ses.srv.env.int8BindPool.Get().(*int8Bind)
					stmt.binds[n] = int8Bind
					err = int8Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint64); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_UIN)
				} else {
					uint64Bind := stmt.ses.srv.env.uint64BindPool.Get().(*uint64Bind)
					stmt.binds[n] = uint64Bind
					err = uint64Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint32); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_UIN)
				} else {
					uint32Bind := stmt.ses.srv.env.uint32BindPool.Get().(*uint32Bind)
					stmt.binds[n] = uint32Bind
					err = uint32Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint16); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_UIN)
				} else {
					uint16Bind := stmt.ses.srv.env.uint16BindPool.Get().(*uint16Bind)
					stmt.binds[n] = uint16Bind
					err = uint16Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint8); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_UIN)
				} else {
					uint8Bind := stmt.ses.srv.env.uint8BindPool.Get().(*uint8Bind)
					stmt.binds[n] = uint8Bind
					err = uint8Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Float64); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_BDOUBLE)
				} else {
					float64Bind := stmt.ses.srv.env.float64BindPool.Get().(*float64Bind)
					stmt.binds[n] = float64Bind
					err = float64Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Float32); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_BFLOAT)
				} else {
					float32Bind := stmt.ses.srv.env.float32BindPool.Get().(*float32Bind)
					stmt.binds[n] = float32Bind
					err = float32Bind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(*int64); ok {
				int64PtrBind := stmt.ses.srv.env.int64PtrBindPool.Get().(*int64PtrBind)
				stmt.binds[n] = int64PtrBind
				err = int64PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*int32); ok {
				int32PtrBind := stmt.ses.srv.env.int32PtrBindPool.Get().(*int32PtrBind)
				stmt.binds[n] = int32PtrBind
				err = int32PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*int16); ok {
				int16PtrBind := stmt.ses.srv.env.int16PtrBindPool.Get().(*int16PtrBind)
				stmt.binds[n] = int16PtrBind
				err = int16PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*int8); ok {
				int8PtrBind := stmt.ses.srv.env.int8PtrBindPool.Get().(*int8PtrBind)
				stmt.binds[n] = int8PtrBind
				err = int8PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*uint64); ok {
				uint64PtrBind := stmt.ses.srv.env.uint64PtrBindPool.Get().(*uint64PtrBind)
				stmt.binds[n] = uint64PtrBind
				err = uint64PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*uint32); ok {
				uint32PtrBind := stmt.ses.srv.env.uint32PtrBindPool.Get().(*uint32PtrBind)
				stmt.binds[n] = uint32PtrBind
				err = uint32PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*uint16); ok {
				uint16PtrBind := stmt.ses.srv.env.uint16PtrBindPool.Get().(*uint16PtrBind)
				stmt.binds[n] = uint16PtrBind
				err = uint16PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*uint8); ok {
				uint8PtrBind := stmt.ses.srv.env.uint8PtrBindPool.Get().(*uint8PtrBind)
				stmt.binds[n] = uint8PtrBind
				err = uint8PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*float64); ok {
				float64PtrBind := stmt.ses.srv.env.float64PtrBindPool.Get().(*float64PtrBind)
				stmt.binds[n] = float64PtrBind
				err = float64PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(*float32); ok {
				float32PtrBind := stmt.ses.srv.env.float32PtrBindPool.Get().(*float32PtrBind)
				stmt.binds[n] = float32PtrBind
				err = float32PtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].([]int64); ok {
				int64SliceBind := stmt.ses.srv.env.int64SliceBindPool.Get().(*int64SliceBind)
				stmt.binds[n] = int64SliceBind
				err = int64SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int32); ok {
				int32SliceBind := stmt.ses.srv.env.int32SliceBindPool.Get().(*int32SliceBind)
				stmt.binds[n] = int32SliceBind
				err = int32SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int16); ok {
				int16SliceBind := stmt.ses.srv.env.int16SliceBindPool.Get().(*int16SliceBind)
				stmt.binds[n] = int16SliceBind
				err = int16SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int8); ok {
				int8SliceBind := stmt.ses.srv.env.int8SliceBindPool.Get().(*int8SliceBind)
				stmt.binds[n] = int8SliceBind
				err = int8SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint64); ok {
				uint64SliceBind := stmt.ses.srv.env.uint64SliceBindPool.Get().(*uint64SliceBind)
				stmt.binds[n] = uint64SliceBind
				err = uint64SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint32); ok {
				uint32SliceBind := stmt.ses.srv.env.uint32SliceBindPool.Get().(*uint32SliceBind)
				stmt.binds[n] = uint32SliceBind
				err = uint32SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint16); ok {
				uint16SliceBind := stmt.ses.srv.env.uint16SliceBindPool.Get().(*uint16SliceBind)
				stmt.binds[n] = uint16SliceBind
				err = uint16SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint8); ok {
				if stmt.Config.byteSlice == U8 {
					uint8SliceBind := stmt.ses.srv.env.uint8SliceBindPool.Get().(*uint8SliceBind)
					stmt.binds[n] = uint8SliceBind
					err = uint8SliceBind.bind(value, nil, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
					iterations = uint32(len(value))
				} else {
					bytesBind := stmt.ses.srv.env.bytesBindPool.Get().(*bytesBind)
					stmt.binds[n] = bytesBind
					err = bytesBind.bind(value, n+1, stmt.Config.lobBufferSize, stmt.ses.srv.ocisvcctx, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]float64); ok {
				float64SliceBind := stmt.ses.srv.env.float64SliceBindPool.Get().(*float64SliceBind)
				stmt.binds[n] = float64SliceBind
				err = float64SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]float32); ok {
				float32SliceBind := stmt.ses.srv.env.float32SliceBindPool.Get().(*float32SliceBind)
				stmt.binds[n] = float32SliceBind
				err = float32SliceBind.bind(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))

			} else if value, ok := params[n].([]Int64); ok {
				int64SliceBind := stmt.ses.srv.env.int64SliceBindPool.Get().(*int64SliceBind)
				stmt.binds[n] = int64SliceBind
				err = int64SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int32); ok {
				int32SliceBind := stmt.ses.srv.env.int32SliceBindPool.Get().(*int32SliceBind)
				stmt.binds[n] = int32SliceBind
				err = int32SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int16); ok {
				int16SliceBind := stmt.ses.srv.env.int16SliceBindPool.Get().(*int16SliceBind)
				stmt.binds[n] = int16SliceBind
				err = int16SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int8); ok {
				int8SliceBind := stmt.ses.srv.env.int8SliceBindPool.Get().(*int8SliceBind)
				stmt.binds[n] = int8SliceBind
				err = int8SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint64); ok {
				uint64SliceBind := stmt.ses.srv.env.uint64SliceBindPool.Get().(*uint64SliceBind)
				stmt.binds[n] = uint64SliceBind
				err = uint64SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint32); ok {
				uint32SliceBind := stmt.ses.srv.env.uint32SliceBindPool.Get().(*uint32SliceBind)
				stmt.binds[n] = uint32SliceBind
				err = uint32SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint16); ok {
				uint16SliceBind := stmt.ses.srv.env.uint16SliceBindPool.Get().(*uint16SliceBind)
				stmt.binds[n] = uint16SliceBind
				err = uint16SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint8); ok {
				uint8SliceBind := stmt.ses.srv.env.uint8SliceBindPool.Get().(*uint8SliceBind)
				stmt.binds[n] = uint8SliceBind
				err = uint8SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Float64); ok {
				float64SliceBind := stmt.ses.srv.env.float64SliceBindPool.Get().(*float64SliceBind)
				stmt.binds[n] = float64SliceBind
				err = float64SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Float32); ok {
				float32SliceBind := stmt.ses.srv.env.float32SliceBindPool.Get().(*float32SliceBind)
				stmt.binds[n] = float32SliceBind
				err = float32SliceBind.bindOra(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(time.Time); ok {
				timeBind := stmt.ses.srv.env.timeBindPool.Get().(*timeBind)
				stmt.binds[n] = timeBind
				err = timeBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*time.Time); ok {
				timePtrBind := stmt.ses.srv.env.timePtrBindPool.Get().(*timePtrBind)
				stmt.binds[n] = timePtrBind
				err = timePtrBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(Time); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_TIMESTAMP_TZ)
				} else {
					timeBind := stmt.ses.srv.env.timeBindPool.Get().(*timeBind)
					stmt.binds[n] = timeBind
					err = timeBind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]time.Time); ok {
				timeSliceBind := stmt.ses.srv.env.timeSliceBindPool.Get().(*timeSliceBind)
				stmt.binds[n] = timeSliceBind
				err = timeSliceBind.bindTimeSlice(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Time); ok {
				timeSliceBind := stmt.ses.srv.env.timeSliceBindPool.Get().(*timeSliceBind)
				stmt.binds[n] = timeSliceBind
				err = timeSliceBind.bindOraTimeSlice(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(string); ok {
				stringBind := stmt.ses.srv.env.stringBindPool.Get().(*stringBind)
				stmt.binds[n] = stringBind
				err = stringBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*string); ok {
				stringPtrBind := stmt.ses.srv.env.stringPtrBindPool.Get().(*stringPtrBind)
				stmt.binds[n] = stringPtrBind
				err = stringPtrBind.bind(value, n+1, stmt.Config.stringPtrBufferSize, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(String); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_CHR)
				} else {
					stringBind := stmt.ses.srv.env.stringBindPool.Get().(*stringBind)
					stmt.binds[n] = stringBind
					err = stringBind.bind(value.Value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]string); ok {
				stringSliceBind := stmt.ses.srv.env.stringSliceBindPool.Get().(*stringSliceBind)
				stmt.binds[n] = stringSliceBind
				err = stringSliceBind.bindStringSlice(value, nil, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]String); ok {
				stringSliceBind := stmt.ses.srv.env.stringSliceBindPool.Get().(*stringSliceBind)
				stmt.binds[n] = stringSliceBind
				err = stringSliceBind.bindOraStringSlice(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(bool); ok {
				boolBind := stmt.ses.srv.env.boolBindPool.Get().(*boolBind)
				stmt.binds[n] = boolBind
				err = boolBind.bind(value, n+1, stmt.Config, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*bool); ok {
				boolPtrBind := stmt.ses.srv.env.boolPtrBindPool.Get().(*boolPtrBind)
				stmt.binds[n] = boolPtrBind
				err = boolPtrBind.bind(value, n+1, stmt.Config.TrueRune, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(Bool); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_CHR)
				} else {
					boolBind := stmt.ses.srv.env.boolBindPool.Get().(*boolBind)
					stmt.binds[n] = boolBind
					err = boolBind.bind(value.Value, n+1, stmt.Config, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]bool); ok {
				boolSliceBind := stmt.ses.srv.env.boolSliceBindPool.Get().(*boolSliceBind)
				stmt.binds[n] = boolSliceBind
				err = boolSliceBind.bindBoolSlice(value, nil, n+1, stmt.Config.FalseRune, stmt.Config.TrueRune, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Bool); ok {
				boolSliceBind := stmt.ses.srv.env.boolSliceBindPool.Get().(*boolSliceBind)
				stmt.binds[n] = boolSliceBind
				err = boolSliceBind.bindOraBoolSlice(value, n+1, stmt.Config.FalseRune, stmt.Config.TrueRune, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(Bytes); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_BLOB)
				} else {
					bytesBind := stmt.ses.srv.env.bytesBindPool.Get().(*bytesBind)
					stmt.binds[n] = bytesBind
					err = bytesBind.bind(value.Value, n+1, stmt.Config.lobBufferSize, stmt.ses.srv.ocisvcctx, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([][]byte); ok {
				bytesSliceBind := stmt.ses.srv.env.bytesSliceBindPool.Get().(*bytesSliceBind)
				stmt.binds[n] = bytesSliceBind
				err = bytesSliceBind.bindBytes(value, nil, n+1, stmt.Config.lobBufferSize, stmt.ses.srv.ocisvcctx, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Bytes); ok {
				bytesSliceBind := stmt.ses.srv.env.bytesSliceBindPool.Get().(*bytesSliceBind)
				stmt.binds[n] = bytesSliceBind
				err = bytesSliceBind.bindOraBytes(value, n+1, stmt.Config.lobBufferSize, stmt.ses.srv.ocisvcctx, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(*ResultSet); ok {
				resultSetBind := stmt.ses.srv.env.resultSetBindPool.Get().(*resultSetBind)
				stmt.binds[n] = resultSetBind
				err = resultSetBind.bind(value, n+1, stmt)
				if err != nil {
					return iterations, err
				}
				stmt.hasPtrBind = true
			} else if value, ok := params[n].(IntervalYM); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INTERVAL_YM)
				} else {
					oraIntervalYMBind := stmt.ses.srv.env.oraIntervalYMBindPool.Get().(*oraIntervalYMBind)
					stmt.binds[n] = oraIntervalYMBind
					err = oraIntervalYMBind.bind(value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(IntervalDS); ok {
				if value.IsNull {
					stmt.setNilBind(n, C.SQLT_INTERVAL_DS)
				} else {
					oraIntervalDSBind := stmt.ses.srv.env.oraIntervalDSBindPool.Get().(*oraIntervalDSBind)
					stmt.binds[n] = oraIntervalDSBind
					err = oraIntervalDSBind.bind(value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]IntervalYM); ok {
				oraIntervalYMSliceBind := stmt.ses.srv.env.oraIntervalYMSliceBindPool.Get().(*oraIntervalYMSliceBind)
				stmt.binds[n] = oraIntervalYMSliceBind
				err = oraIntervalYMSliceBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]IntervalDS); ok {
				oraIntervalDSSliceBind := stmt.ses.srv.env.oraIntervalDSSliceBindPool.Get().(*oraIntervalDSSliceBind)
				stmt.binds[n] = oraIntervalDSSliceBind
				err = oraIntervalDSSliceBind.bind(value, n+1, stmt.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(Bfile); ok {
				if value.IsNull {
					err = stmt.setNilBind(n, C.SQLT_FILE)
				} else {
					bfileBind := stmt.ses.srv.env.bfileBindPool.Get().(*bfileBind)
					stmt.binds[n] = bfileBind
					err = bfileBind.bind(value, n+1, stmt.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if params[n] == nil {
				err = stmt.setNilBind(n, C.SQLT_CHR)
			} else {
				return iterations, errNewF("Unsupported bind parameter (%v) (%v).", params[n], reflect.TypeOf(params[n]).Name())
			}
		}
	}

	return iterations, err
}

// setNilBind sets a nil bind.
func (stmt *Statement) setNilBind(index int, sqlt C.ub2) (err error) {
	nilBind := stmt.ses.srv.env.nilBindPool.Get().(*nilBind)
	stmt.binds[index] = nilBind
	err = nilBind.bind(index+1, sqlt, stmt.ocistmt)
	return err
}

// Execute runs a SQL statement on an Oracle server returning the number of
// rows affected and a possible error.
//
// Execute is meant to be called when working with the oracle package directly.
func (stmt *Statement) Execute(params ...interface{}) (rowsAffected uint64, err error) {
	rowsAffected, _, err = stmt.exec(false, params)
	return rowsAffected, err
}

// Exec runs a SQL statement on an Oracle server returning driver.Result and
// a possible error.
//
// Exec is meant to be called by the database/sql package.
//
// Exec is a member of the driver.Stmt interface.
func (stmt *Statement) Exec(values []driver.Value) (result driver.Result, err error) {
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	rowsAffected, lastInsertId, err := stmt.exec(true, params)
	if rowsAffected == 0 {
		result = driver.ResultNoRows
	} else {
		result = &ExecResult{rowsAffected: rowsAffected, lastInsertId: lastInsertId}
	}
	return result, err
}

// exec runs a SQL statement on an Oracle server returning rowsAffected, lastInsertId and error.
func (stmt *Statement) exec(tryAddBindForIdentity bool, params []interface{}) (rowsAffected uint64, lastInsertId int64, err error) {
	// Validate that the statement is open
	if err := stmt.checkIsOpen(); err != nil {
		return 0, 0, err
	}
	// For case of inserting and returning identity for database/sql package
	if tryAddBindForIdentity && stmt.stmtType == C.OCI_STMT_INSERT {
		lastIndex := strings.LastIndex(stmt.sql, ")")
		sqlEnd := stmt.sql[lastIndex+1 : len(stmt.sql)]
		sqlEnd = strings.ToUpper(sqlEnd)
		// add *int64 arg to capture identity
		if strings.Contains(sqlEnd, "RETURNING") {
			params[len(params)-1] = &lastInsertId
		}
	}
	// Bind parameters
	iterations, err := stmt.bindParams(params)
	if err != nil {
		return 0, 0, err
	}

	err = stmt.setPrefetchSize()
	if err != nil {
		return 0, 0, err
	}

	// Execute statement on Oracle server
	r := C.OCIStmtExecute(
		stmt.ses.srv.ocisvcctx,  //OCISvcCtx           *svchp,
		stmt.ocistmt,            //OCIStmt             *stmtp,
		stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(iterations),       //ub4                 iters,
		C.ub4(0),                //ub4                 rowoff,
		nil,                     //const OCISnapshot   *snap_in,
		nil,                     //OCISnapshot         *snap_out,
		C.OCI_DEFAULT)           //ub4                 mode );
	if r == C.OCI_ERROR {
		return 0, 0, stmt.ses.srv.env.ociError()
	}

	// Get row count based on statement type
	var rowCount C.ub8
	switch stmt.stmtType {
	case C.OCI_STMT_SELECT, C.OCI_STMT_UPDATE, C.OCI_STMT_DELETE, C.OCI_STMT_INSERT:
		err := stmt.attr(unsafe.Pointer(&rowCount), 8, C.OCI_ATTR_UB8_ROW_COUNT)
		if err != nil {
			return 0, 0, err
		}
		rowsAffected = uint64(rowCount)
	case C.OCI_STMT_CREATE, C.OCI_STMT_DROP, C.OCI_STMT_ALTER, C.OCI_STMT_BEGIN:
	}

	// Set any bind pointers
	if stmt.hasPtrBind {
		err = stmt.setBindPtrs()
		if err != nil {
			return rowsAffected, lastInsertId, err
		}
	}

	return rowsAffected, lastInsertId, nil
}

// Query runs a sql query on an Oracle server returning driver.Rows and a
// possible error.
//
// Query is meant to be called by the database/sql package.
//
// Query is a member of the driver.Stmt interface.
func (stmt *Statement) Query(values []driver.Value) (driver.Rows, error) {
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	rst, err := stmt.fetch(params)
	return &QueryResult{rst: rst}, err
}

// Fetch runs a SQL query on an Oracle server returning a *ResultSet and a possible
// error.
//
// Fetch is meant to be called when working with the oracle package directly.
func (stmt *Statement) Fetch(params ...interface{}) (*ResultSet, error) {
	return stmt.fetch(params)
}

// fetch runs a SQL query for Fetch and Query methods.
func (stmt *Statement) fetch(params []interface{}) (*ResultSet, error) {
	// Validate that the statement is open
	err := stmt.checkIsOpen()
	if err != nil {
		return nil, err
	}
	// Bind parameters
	_, err = stmt.bindParams(params)
	if err != nil {
		return nil, err
	}
	err = stmt.setPrefetchSize()
	if err != nil {
		return nil, err
	}
	// Run query
	r := C.OCIStmtExecute(
		stmt.ses.srv.ocisvcctx,  //OCISvcCtx           *svchp,
		stmt.ocistmt,            //OCIStmt             *stmtp,
		stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(0),                //ub4                 iters,
		C.ub4(0),                //ub4                 rowoff,
		nil,                     //const OCISnapshot   *snap_in,
		nil,                     //OCISnapshot         *snap_out,
		C.OCI_DEFAULT)           //ub4                 mode );
	if r == C.OCI_ERROR {
		return nil, stmt.ses.srv.env.ociError()
	}
	// Set any bind pointers
	if stmt.hasPtrBind {
		err = stmt.setBindPtrs()
		if err != nil {
			return nil, err
		}
	}
	// create and open result set
	rst := &ResultSet{}
	err = rst.open(stmt, stmt.ocistmt)
	if err != nil {
		rst.close()
		return nil, err
	}
	// store result set for later close call
	stmt.rsts.PushBack(rst)
	return rst, nil
}

func (stmt *Statement) setPrefetchSize() error {
	if stmt.Config.prefetchRowCount > 0 {
		//fmt.Println("statement.setPrefetchSize: prefetchRowCount ", statement.Config.prefetchRowCount)
		// Set prefetch row count
		if err := stmt.setAttr(unsafe.Pointer(&stmt.Config.prefetchRowCount), 4, C.OCI_ATTR_PREFETCH_ROWS); err != nil {
			return err
		}
	} else {
		//fmt.Println("statement.setPrefetchSize: prefetchMemorySize ", statement.Config.prefetchMemorySize)
		// Set prefetch memory size
		if err := stmt.setAttr(unsafe.Pointer(&stmt.Config.prefetchMemorySize), 4, C.OCI_ATTR_PREFETCH_MEMORY); err != nil {
			return err
		}
	}
	return nil
}

// NumInput returns the number of placeholders in a sql statement.
//
// NumInput is a member of the driver.Stmt interface.
func (stmt *Statement) NumInput() int {
	var bindCount uint32
	if err := stmt.attr(unsafe.Pointer(&bindCount), 4, C.OCI_ATTR_BIND_COUNT); err != nil {
		return 0
	}
	return int(bindCount)
}

// setBindPtrs enables binds to set out pointers for some types such as time.Time, etc.
func (stmt *Statement) setBindPtrs() (err error) {
	for _, bind := range stmt.binds {
		err = bind.setPtr()
		if err != nil {
			return err
		}
	}
	return nil
}

// checkIsOpen validates that a statement is open.
func (stmt *Statement) checkIsOpen() error {
	if !stmt.IsOpen() {
		return errNew("open statement prior to method call")
	}
	return nil
}

// IsOpen returns true when a statement is open; otherwise, false.
//
// Calling Close will cause Statement.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Statement.Prepare to create a new statement.
func (stmt *Statement) IsOpen() bool {
	return stmt.ocistmt != nil
}

// Close ends a sql statement.
//
// Calling Close will cause Statement.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Statement.Prepare to create a new statement.
//
// Close is a member of the driver.Stmt interface.
func (stmt *Statement) Close() error {
	if stmt.IsOpen() {
		// Close binds
		if len(stmt.binds) > 0 {
			for _, bind := range stmt.binds {
				//fmt.Printf("close bind %v\n", bind)
				if bind != nil {
					bind.close()
				}
			}
		}

		// Close result sets
		for e := stmt.rsts.Front(); e != nil; e = e.Next() {
			e.Value.(*ResultSet).close()
		}

		// Clear statement fields
		ses := stmt.ses
		stmt.ocistmt = nil
		stmt.ses = nil
		stmt.elem = nil
		stmt.binds = nil
		stmt.goColumnTypes = nil
		stmt.sql = ""
		stmt.stmtType = C.ub4(0)
		stmt.hasPtrBind = false

		// Put statement in pool
		ses.srv.env.stmtPool.Put(stmt)
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (stmt *Statement) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(stmt.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4            trghndltyp,
		attrup,                       //void           *attributep,
		&attrSize,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		stmt.ses.srv.env.ocierr)      //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return stmt.ses.srv.env.ociError()
	}
	return nil
}

// setAttr sets an attribute on the statement handle.
func (stmt *Statement) setAttr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrSet(
		unsafe.Pointer(stmt.ocistmt), //void        *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4         trghndltyp,
		attrup,                       //void        *attributep,
		attrSize,                     //ub4         size,
		attrType,                     //ub4         attrtype,
		stmt.ses.srv.env.ocierr)      //OCIError    *errhp );
	if r == C.OCI_ERROR {
		return stmt.ses.srv.env.ociError()
	}

	return nil
}
