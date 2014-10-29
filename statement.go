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
	environment   *Environment
	session       *Session
	element       *list.Element
	resultSets    *list.List
	binds         []bind
	goColumnTypes []GoColumnType
	sql           string
	stmtType      C.ub4
	ocistmt       *C.OCIStmt
	hasPtrBind    bool

	Config StatementConfig
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
func (statement *Statement) bindParams(params []interface{}) (iterations uint32, err error) {
	//fmt.Printf("Statement.bindParams: len(params) (%v)\n", len(params))

	iterations = 1
	// Create binds for each arg; bind position is 1-based
	if params != nil && len(params) > 0 {
		statement.binds = make([]bind, len(params))
		for n := 0; n < len(params); n++ {
			//fmt.Printf("Statement.bindParams: params[%v] (%v)\n", n, params[n])
			if value, ok := params[n].(int64); ok {
				int64Bind := statement.session.server.environment.int64BindPool.Get().(*int64Bind)
				statement.binds[n] = int64Bind
				err = int64Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int32); ok {
				int32Bind := statement.session.server.environment.int32BindPool.Get().(*int32Bind)
				statement.binds[n] = int32Bind
				err = int32Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int16); ok {
				int16Bind := statement.session.server.environment.int16BindPool.Get().(*int16Bind)
				statement.binds[n] = int16Bind
				err = int16Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(int8); ok {
				int8Bind := statement.session.server.environment.int8BindPool.Get().(*int8Bind)
				statement.binds[n] = int8Bind
				err = int8Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint64); ok {
				uint64Bind := statement.session.server.environment.uint64BindPool.Get().(*uint64Bind)
				statement.binds[n] = uint64Bind
				err = uint64Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint32); ok {
				uint32Bind := statement.session.server.environment.uint32BindPool.Get().(*uint32Bind)
				statement.binds[n] = uint32Bind
				err = uint32Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint16); ok {
				uint16Bind := statement.session.server.environment.uint16BindPool.Get().(*uint16Bind)
				statement.binds[n] = uint16Bind
				err = uint16Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(uint8); ok {
				uint8Bind := statement.session.server.environment.uint8BindPool.Get().(*uint8Bind)
				statement.binds[n] = uint8Bind
				err = uint8Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(float64); ok {
				float64Bind := statement.session.server.environment.float64BindPool.Get().(*float64Bind)
				statement.binds[n] = float64Bind
				err = float64Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(float32); ok {
				float32Bind := statement.session.server.environment.float32BindPool.Get().(*float32Bind)
				statement.binds[n] = float32Bind
				err = float32Bind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(Int64); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INT)
				} else {
					int64Bind := statement.session.server.environment.int64BindPool.Get().(*int64Bind)
					statement.binds[n] = int64Bind
					err = int64Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int32); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INT)
				} else {
					int32Bind := statement.session.server.environment.int32BindPool.Get().(*int32Bind)
					statement.binds[n] = int32Bind
					err = int32Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int16); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INT)
				} else {
					int16Bind := statement.session.server.environment.int16BindPool.Get().(*int16Bind)
					statement.binds[n] = int16Bind
					err = int16Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Int8); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INT)
				} else {
					int8Bind := statement.session.server.environment.int8BindPool.Get().(*int8Bind)
					statement.binds[n] = int8Bind
					err = int8Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint64); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_UIN)
				} else {
					uint64Bind := statement.session.server.environment.uint64BindPool.Get().(*uint64Bind)
					statement.binds[n] = uint64Bind
					err = uint64Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint32); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_UIN)
				} else {
					uint32Bind := statement.session.server.environment.uint32BindPool.Get().(*uint32Bind)
					statement.binds[n] = uint32Bind
					err = uint32Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint16); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_UIN)
				} else {
					uint16Bind := statement.session.server.environment.uint16BindPool.Get().(*uint16Bind)
					statement.binds[n] = uint16Bind
					err = uint16Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Uint8); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_UIN)
				} else {
					uint8Bind := statement.session.server.environment.uint8BindPool.Get().(*uint8Bind)
					statement.binds[n] = uint8Bind
					err = uint8Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Float64); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_BDOUBLE)
				} else {
					float64Bind := statement.session.server.environment.float64BindPool.Get().(*float64Bind)
					statement.binds[n] = float64Bind
					err = float64Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(Float32); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_BFLOAT)
				} else {
					float32Bind := statement.session.server.environment.float32BindPool.Get().(*float32Bind)
					statement.binds[n] = float32Bind
					err = float32Bind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(*int64); ok {
				int64PtrBind := statement.session.server.environment.int64PtrBindPool.Get().(*int64PtrBind)
				statement.binds[n] = int64PtrBind
				err = int64PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*int32); ok {
				int32PtrBind := statement.session.server.environment.int32PtrBindPool.Get().(*int32PtrBind)
				statement.binds[n] = int32PtrBind
				err = int32PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*int16); ok {
				int16PtrBind := statement.session.server.environment.int16PtrBindPool.Get().(*int16PtrBind)
				statement.binds[n] = int16PtrBind
				err = int16PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*int8); ok {
				int8PtrBind := statement.session.server.environment.int8PtrBindPool.Get().(*int8PtrBind)
				statement.binds[n] = int8PtrBind
				err = int8PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*uint64); ok {
				uint64PtrBind := statement.session.server.environment.uint64PtrBindPool.Get().(*uint64PtrBind)
				statement.binds[n] = uint64PtrBind
				err = uint64PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*uint32); ok {
				uint32PtrBind := statement.session.server.environment.uint32PtrBindPool.Get().(*uint32PtrBind)
				statement.binds[n] = uint32PtrBind
				err = uint32PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*uint16); ok {
				uint16PtrBind := statement.session.server.environment.uint16PtrBindPool.Get().(*uint16PtrBind)
				statement.binds[n] = uint16PtrBind
				err = uint16PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*uint8); ok {
				uint8PtrBind := statement.session.server.environment.uint8PtrBindPool.Get().(*uint8PtrBind)
				statement.binds[n] = uint8PtrBind
				err = uint8PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*float64); ok {
				float64PtrBind := statement.session.server.environment.float64PtrBindPool.Get().(*float64PtrBind)
				statement.binds[n] = float64PtrBind
				err = float64PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(*float32); ok {
				float32PtrBind := statement.session.server.environment.float32PtrBindPool.Get().(*float32PtrBind)
				statement.binds[n] = float32PtrBind
				err = float32PtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].([]int64); ok {
				int64SliceBind := statement.session.server.environment.int64SliceBindPool.Get().(*int64SliceBind)
				statement.binds[n] = int64SliceBind
				err = int64SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int32); ok {
				int32SliceBind := statement.session.server.environment.int32SliceBindPool.Get().(*int32SliceBind)
				statement.binds[n] = int32SliceBind
				err = int32SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int16); ok {
				int16SliceBind := statement.session.server.environment.int16SliceBindPool.Get().(*int16SliceBind)
				statement.binds[n] = int16SliceBind
				err = int16SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]int8); ok {
				int8SliceBind := statement.session.server.environment.int8SliceBindPool.Get().(*int8SliceBind)
				statement.binds[n] = int8SliceBind
				err = int8SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint64); ok {
				uint64SliceBind := statement.session.server.environment.uint64SliceBindPool.Get().(*uint64SliceBind)
				statement.binds[n] = uint64SliceBind
				err = uint64SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint32); ok {
				uint32SliceBind := statement.session.server.environment.uint32SliceBindPool.Get().(*uint32SliceBind)
				statement.binds[n] = uint32SliceBind
				err = uint32SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint16); ok {
				uint16SliceBind := statement.session.server.environment.uint16SliceBindPool.Get().(*uint16SliceBind)
				statement.binds[n] = uint16SliceBind
				err = uint16SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]uint8); ok {
				if statement.Config.byteSlice == U8 {
					uint8SliceBind := statement.session.server.environment.uint8SliceBindPool.Get().(*uint8SliceBind)
					statement.binds[n] = uint8SliceBind
					err = uint8SliceBind.bind(value, nil, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
					iterations = uint32(len(value))
				} else {
					bytesBind := statement.session.server.environment.bytesBindPool.Get().(*bytesBind)
					statement.binds[n] = bytesBind
					err = bytesBind.bind(value, n+1, statement.Config.lobBufferSize, statement.session.server.ocisvcctx, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]float64); ok {
				float64SliceBind := statement.session.server.environment.float64SliceBindPool.Get().(*float64SliceBind)
				statement.binds[n] = float64SliceBind
				err = float64SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]float32); ok {
				float32SliceBind := statement.session.server.environment.float32SliceBindPool.Get().(*float32SliceBind)
				statement.binds[n] = float32SliceBind
				err = float32SliceBind.bind(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))

			} else if value, ok := params[n].([]Int64); ok {
				int64SliceBind := statement.session.server.environment.int64SliceBindPool.Get().(*int64SliceBind)
				statement.binds[n] = int64SliceBind
				err = int64SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int32); ok {
				int32SliceBind := statement.session.server.environment.int32SliceBindPool.Get().(*int32SliceBind)
				statement.binds[n] = int32SliceBind
				err = int32SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int16); ok {
				int16SliceBind := statement.session.server.environment.int16SliceBindPool.Get().(*int16SliceBind)
				statement.binds[n] = int16SliceBind
				err = int16SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Int8); ok {
				int8SliceBind := statement.session.server.environment.int8SliceBindPool.Get().(*int8SliceBind)
				statement.binds[n] = int8SliceBind
				err = int8SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint64); ok {
				uint64SliceBind := statement.session.server.environment.uint64SliceBindPool.Get().(*uint64SliceBind)
				statement.binds[n] = uint64SliceBind
				err = uint64SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint32); ok {
				uint32SliceBind := statement.session.server.environment.uint32SliceBindPool.Get().(*uint32SliceBind)
				statement.binds[n] = uint32SliceBind
				err = uint32SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint16); ok {
				uint16SliceBind := statement.session.server.environment.uint16SliceBindPool.Get().(*uint16SliceBind)
				statement.binds[n] = uint16SliceBind
				err = uint16SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Uint8); ok {
				uint8SliceBind := statement.session.server.environment.uint8SliceBindPool.Get().(*uint8SliceBind)
				statement.binds[n] = uint8SliceBind
				err = uint8SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Float64); ok {
				float64SliceBind := statement.session.server.environment.float64SliceBindPool.Get().(*float64SliceBind)
				statement.binds[n] = float64SliceBind
				err = float64SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Float32); ok {
				float32SliceBind := statement.session.server.environment.float32SliceBindPool.Get().(*float32SliceBind)
				statement.binds[n] = float32SliceBind
				err = float32SliceBind.bindOra(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(time.Time); ok {
				timeBind := statement.session.server.environment.timeBindPool.Get().(*timeBind)
				statement.binds[n] = timeBind
				err = timeBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*time.Time); ok {
				timePtrBind := statement.session.server.environment.timePtrBindPool.Get().(*timePtrBind)
				statement.binds[n] = timePtrBind
				err = timePtrBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(Time); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_TIMESTAMP_TZ)
				} else {
					timeBind := statement.session.server.environment.timeBindPool.Get().(*timeBind)
					statement.binds[n] = timeBind
					err = timeBind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]time.Time); ok {
				timeSliceBind := statement.session.server.environment.timeSliceBindPool.Get().(*timeSliceBind)
				statement.binds[n] = timeSliceBind
				err = timeSliceBind.bindTimeSlice(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Time); ok {
				timeSliceBind := statement.session.server.environment.timeSliceBindPool.Get().(*timeSliceBind)
				statement.binds[n] = timeSliceBind
				err = timeSliceBind.bindOraTimeSlice(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(string); ok {
				stringBind := statement.session.server.environment.stringBindPool.Get().(*stringBind)
				statement.binds[n] = stringBind
				err = stringBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*string); ok {
				stringPtrBind := statement.session.server.environment.stringPtrBindPool.Get().(*stringPtrBind)
				statement.binds[n] = stringPtrBind
				err = stringPtrBind.bind(value, n+1, statement.Config.stringPtrBufferSize, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(String); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_CHR)
				} else {
					stringBind := statement.session.server.environment.stringBindPool.Get().(*stringBind)
					statement.binds[n] = stringBind
					err = stringBind.bind(value.Value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]string); ok {
				stringSliceBind := statement.session.server.environment.stringSliceBindPool.Get().(*stringSliceBind)
				statement.binds[n] = stringSliceBind
				err = stringSliceBind.bindStringSlice(value, nil, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]String); ok {
				stringSliceBind := statement.session.server.environment.stringSliceBindPool.Get().(*stringSliceBind)
				statement.binds[n] = stringSliceBind
				err = stringSliceBind.bindOraStringSlice(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(bool); ok {
				boolBind := statement.session.server.environment.boolBindPool.Get().(*boolBind)
				statement.binds[n] = boolBind
				err = boolBind.bind(value, n+1, statement.Config, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
			} else if value, ok := params[n].(*bool); ok {
				boolPtrBind := statement.session.server.environment.boolPtrBindPool.Get().(*boolPtrBind)
				statement.binds[n] = boolPtrBind
				err = boolPtrBind.bind(value, n+1, statement.Config.TrueRune, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(Bool); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_CHR)
				} else {
					boolBind := statement.session.server.environment.boolBindPool.Get().(*boolBind)
					statement.binds[n] = boolBind
					err = boolBind.bind(value.Value, n+1, statement.Config, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]bool); ok {
				boolSliceBind := statement.session.server.environment.boolSliceBindPool.Get().(*boolSliceBind)
				statement.binds[n] = boolSliceBind
				err = boolSliceBind.bindBoolSlice(value, nil, n+1, statement.Config.FalseRune, statement.Config.TrueRune, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Bool); ok {
				boolSliceBind := statement.session.server.environment.boolSliceBindPool.Get().(*boolSliceBind)
				statement.binds[n] = boolSliceBind
				err = boolSliceBind.bindOraBoolSlice(value, n+1, statement.Config.FalseRune, statement.Config.TrueRune, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(Bytes); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_BLOB)
				} else {
					bytesBind := statement.session.server.environment.bytesBindPool.Get().(*bytesBind)
					statement.binds[n] = bytesBind
					err = bytesBind.bind(value.Value, n+1, statement.Config.lobBufferSize, statement.session.server.ocisvcctx, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([][]byte); ok {
				bytesSliceBind := statement.session.server.environment.bytesSliceBindPool.Get().(*bytesSliceBind)
				statement.binds[n] = bytesSliceBind
				err = bytesSliceBind.bindBytes(value, nil, n+1, statement.Config.lobBufferSize, statement.session.server.ocisvcctx, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]Bytes); ok {
				bytesSliceBind := statement.session.server.environment.bytesSliceBindPool.Get().(*bytesSliceBind)
				statement.binds[n] = bytesSliceBind
				err = bytesSliceBind.bindOraBytes(value, n+1, statement.Config.lobBufferSize, statement.session.server.ocisvcctx, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(*ResultSet); ok {
				resultSetBind := statement.session.server.environment.resultSetBindPool.Get().(*resultSetBind)
				statement.binds[n] = resultSetBind
				err = resultSetBind.bind(value, n+1, statement)
				if err != nil {
					return iterations, err
				}
				statement.hasPtrBind = true
			} else if value, ok := params[n].(IntervalYM); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INTERVAL_YM)
				} else {
					oraIntervalYMBind := statement.session.server.environment.oraIntervalYMBindPool.Get().(*oraIntervalYMBind)
					statement.binds[n] = oraIntervalYMBind
					err = oraIntervalYMBind.bind(value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].(IntervalDS); ok {
				if value.IsNull {
					statement.setNilBind(n, C.SQLT_INTERVAL_DS)
				} else {
					oraIntervalDSBind := statement.session.server.environment.oraIntervalDSBindPool.Get().(*oraIntervalDSBind)
					statement.binds[n] = oraIntervalDSBind
					err = oraIntervalDSBind.bind(value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if value, ok := params[n].([]IntervalYM); ok {
				oraIntervalYMSliceBind := statement.session.server.environment.oraIntervalYMSliceBindPool.Get().(*oraIntervalYMSliceBind)
				statement.binds[n] = oraIntervalYMSliceBind
				err = oraIntervalYMSliceBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].([]IntervalDS); ok {
				oraIntervalDSSliceBind := statement.session.server.environment.oraIntervalDSSliceBindPool.Get().(*oraIntervalDSSliceBind)
				statement.binds[n] = oraIntervalDSSliceBind
				err = oraIntervalDSSliceBind.bind(value, n+1, statement.ocistmt)
				if err != nil {
					return iterations, err
				}
				iterations = uint32(len(value))
			} else if value, ok := params[n].(Bfile); ok {
				if value.IsNull {
					err = statement.setNilBind(n, C.SQLT_FILE)
				} else {
					bfileBind := statement.session.server.environment.bfileBindPool.Get().(*bfileBind)
					statement.binds[n] = bfileBind
					err = bfileBind.bind(value, n+1, statement.ocistmt)
					if err != nil {
						return iterations, err
					}
				}
			} else if params[n] == nil {
				err = statement.setNilBind(n, C.SQLT_CHR)
			} else {
				return iterations, errNewF("Unsupported bind parameter (%v) (%v).", params[n], reflect.TypeOf(params[n]).Name())
			}
		}
	}

	return iterations, err
}

// setNilBind sets a nil bind.
func (statement *Statement) setNilBind(index int, sqlt C.ub2) (err error) {
	nilBind := statement.session.server.environment.nilBindPool.Get().(*nilBind)
	statement.binds[index] = nilBind
	err = nilBind.bind(index+1, sqlt, statement.ocistmt)
	return err
}

// Execute runs a SQL statement on an Oracle server returning the number of
// rows affected and a possible error.
//
// Execute is meant to be called when working with the oracle package directly.
func (statement *Statement) Execute(params ...interface{}) (rowsAffected uint64, err error) {
	rowsAffected, _, err = statement.exec(false, params)
	return rowsAffected, err
}

// Exec runs a SQL statement on an Oracle server returning driver.Result and
// a possible error.
//
// Exec is meant to be called by the database/sql package.
//
// Exec is a member of the driver.Stmt interface.
func (statement *Statement) Exec(values []driver.Value) (result driver.Result, err error) {
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	rowsAffected, lastInsertId, err := statement.exec(true, params)
	if rowsAffected == 0 {
		result = driver.ResultNoRows
	} else {
		result = &ExecResult{rowsAffected: rowsAffected, lastInsertId: lastInsertId}
	}
	return result, err
}

// exec runs a SQL statement on an Oracle server returning rowsAffected, lastInsertId and error.
func (statement *Statement) exec(tryAddBindForIdentity bool, params []interface{}) (rowsAffected uint64, lastInsertId int64, err error) {
	// Validate that the statement is open
	if err := statement.checkIsOpen(); err != nil {
		return 0, 0, err
	}
	// For case of inserting and returning identity for database/sql package
	if tryAddBindForIdentity && statement.stmtType == C.OCI_STMT_INSERT {
		lastIndex := strings.LastIndex(statement.sql, ")")
		sqlEnd := statement.sql[lastIndex+1 : len(statement.sql)]
		sqlEnd = strings.ToUpper(sqlEnd)
		// add *int64 arg to capture identity
		if strings.Contains(sqlEnd, "RETURNING") {
			params[len(params)-1] = &lastInsertId
		}
	}
	// Bind parameters
	iterations, err := statement.bindParams(params)
	if err != nil {
		return 0, 0, err
	}

	err = statement.setPrefetchSize()
	if err != nil {
		return 0, 0, err
	}

	// Execute statement on Oracle server
	r := C.OCIStmtExecute(
		statement.session.server.ocisvcctx,          //OCISvcCtx           *svchp,
		statement.ocistmt,                           //OCIStmt             *stmtp,
		statement.session.server.environment.ocierr, //OCIError            *errhp,
		C.ub4(iterations),                           //ub4                 iters,
		C.ub4(0),                                    //ub4                 rowoff,
		nil,                                         //const OCISnapshot   *snap_in,
		nil,                                         //OCISnapshot         *snap_out,
		C.OCI_DEFAULT)                               //ub4                 mode );
	if r == C.OCI_ERROR {
		return 0, 0, statement.session.server.environment.ociError()
	}

	// Get row count based on statement type
	var rowCount C.ub8
	switch statement.stmtType {
	case C.OCI_STMT_SELECT, C.OCI_STMT_UPDATE, C.OCI_STMT_DELETE, C.OCI_STMT_INSERT:
		err := statement.attr(unsafe.Pointer(&rowCount), 8, C.OCI_ATTR_UB8_ROW_COUNT)
		if err != nil {
			return 0, 0, err
		}
		rowsAffected = uint64(rowCount)
	case C.OCI_STMT_CREATE, C.OCI_STMT_DROP, C.OCI_STMT_ALTER, C.OCI_STMT_BEGIN:
	}

	// Set any bind pointers
	if statement.hasPtrBind {
		err = statement.setBindPtrs()
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
func (statement *Statement) Query(values []driver.Value) (driver.Rows, error) {
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	resultSet, err := statement.fetch(params)
	return &QueryResult{resultSet: resultSet}, err
}

// Fetch runs a SQL query on an Oracle server returning a *ResultSet and a possible
// error.
//
// Fetch is meant to be called when working with the oracle package directly.
func (statement *Statement) Fetch(params ...interface{}) (*ResultSet, error) {
	return statement.fetch(params)
}

// fetch runs a SQL query for Fetch and Query methods.
func (statement *Statement) fetch(params []interface{}) (*ResultSet, error) {
	// Validate that the statement is open
	err := statement.checkIsOpen()
	if err != nil {
		return nil, err
	}
	// Bind parameters
	_, err = statement.bindParams(params)
	if err != nil {
		return nil, err
	}
	err = statement.setPrefetchSize()
	if err != nil {
		return nil, err
	}
	// Run query
	r := C.OCIStmtExecute(
		statement.session.server.ocisvcctx,          //OCISvcCtx           *svchp,
		statement.ocistmt,                           //OCIStmt             *stmtp,
		statement.session.server.environment.ocierr, //OCIError            *errhp,
		C.ub4(0),      //ub4                 iters,
		C.ub4(0),      //ub4                 rowoff,
		nil,           //const OCISnapshot   *snap_in,
		nil,           //OCISnapshot         *snap_out,
		C.OCI_DEFAULT) //ub4                 mode );
	if r == C.OCI_ERROR {
		return nil, statement.session.server.environment.ociError()
	}
	// Set any bind pointers
	if statement.hasPtrBind {
		err = statement.setBindPtrs()
		if err != nil {
			return nil, err
		}
	}
	// create and open result set
	resultSet := &ResultSet{}
	err = resultSet.open(statement, statement.ocistmt)
	if err != nil {
		resultSet.close()
		return nil, err
	}
	// store result set for later close call
	statement.resultSets.PushBack(resultSet)
	return resultSet, nil
}

func (statement *Statement) setPrefetchSize() error {
	if statement.Config.prefetchRowCount > 0 {
		//fmt.Println("statement.setPrefetchSize: prefetchRowCount ", statement.Config.prefetchRowCount)
		// Set prefetch row count
		if err := statement.setAttr(unsafe.Pointer(&statement.Config.prefetchRowCount), 4, C.OCI_ATTR_PREFETCH_ROWS); err != nil {
			return err
		}
	} else {
		//fmt.Println("statement.setPrefetchSize: prefetchMemorySize ", statement.Config.prefetchMemorySize)
		// Set prefetch memory size
		if err := statement.setAttr(unsafe.Pointer(&statement.Config.prefetchMemorySize), 4, C.OCI_ATTR_PREFETCH_MEMORY); err != nil {
			return err
		}
	}
	return nil
}

// NumInput returns the number of placeholders in a sql statement.
//
// NumInput is a member of the driver.Stmt interface.
func (statement *Statement) NumInput() int {
	var bindCount uint32
	if err := statement.attr(unsafe.Pointer(&bindCount), 4, C.OCI_ATTR_BIND_COUNT); err != nil {
		return 0
	}
	return int(bindCount)
}

// setBindPtrs enables binds to set out pointers for some types such as time.Time, etc.
func (statement *Statement) setBindPtrs() (err error) {
	for _, bind := range statement.binds {
		err = bind.setPtr()
		if err != nil {
			return err
		}
	}
	return nil
}

// checkIsOpen validates that a statement is open.
//
// ErrClosedStatement is returned if the statement is closed.
func (statement *Statement) checkIsOpen() error {
	if !statement.IsOpen() {
		return errNew("open Statement prior to method call")
	}
	return nil
}

// IsOpen returns true when a statement is open; otherwise, false.
//
// Calling Close will cause Statement.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Statement.Prepare to create a new statement.
func (statement *Statement) IsOpen() bool {
	return statement.ocistmt != nil
}

// Close ends a sql statement.
//
// Calling Close will cause Statement.IsOpen to return false. Once closed, a statement
// cannot be re-opened. Call Statement.Prepare to create a new statement.
//
// Close is a member of the driver.Stmt interface.
func (statement *Statement) Close() error {
	if statement.IsOpen() {
		// Close binds
		if len(statement.binds) > 0 {
			for _, bind := range statement.binds {
				//fmt.Printf("close bind %v\n", bind)
				if bind != nil {
					bind.close()
				}
			}
		}

		// Close result sets
		for e := statement.resultSets.Front(); e != nil; e = e.Next() {
			e.Value.(*ResultSet).close()
		}

		// Clear statement fields
		session := statement.session
		statement.session = nil
		statement.element = nil
		statement.binds = nil
		statement.goColumnTypes = nil
		statement.sql = ""
		statement.stmtType = C.ub4(0)
		statement.ocistmt = nil
		statement.hasPtrBind = false

		// Put statement in pool
		session.server.environment.statementPool.Put(statement)
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (statement *Statement) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(statement.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,                  //ub4            trghndltyp,
		attrup,                            //void           *attributep,
		&attrSize,                         //ub4            *sizep,
		attrType,                          //ub4            attrtype,
		statement.session.server.environment.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return statement.session.server.environment.ociError()
	}
	return nil
}

// setAttr sets an attribute on the statement handle.
func (statement *Statement) setAttr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrSet(
		unsafe.Pointer(statement.ocistmt), //void        *trgthndlp,
		C.OCI_HTYPE_STMT,                  //ub4         trghndltyp,
		attrup,                            //void        *attributep,
		attrSize,                          //ub4         size,
		attrType,                          //ub4         attrtype,
		statement.session.server.environment.ocierr) //OCIError    *errhp );
	if r == C.OCI_ERROR {
		return statement.session.server.environment.ociError()
	}

	return nil
}
