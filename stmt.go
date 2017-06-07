// +build go1.9

// Copyright 2017 Tamás Gulácsi
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package ora

/*
#cgo pkg-config: odpi

#include "dpiImpl.h"

const int sizeof_dpiData = sizeof(void);
*/
import "C"
import (
	"context"
	"database/sql/driver"
	"io"
	"log"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

var _ = driver.Stmt((*statement)(nil))
var _ = driver.StmtQueryContext((*statement)(nil))
var _ = driver.StmtExecContext((*statement)(nil))
var _ = driver.NamedValueChecker((*statement)(nil))

const sizeof_dpiData = C.sizeof_dpiData

type statement struct {
	*conn
	dpiStmt *C.dpiStmt
	query   string
	data    [][]*C.dpiData
	vars    []*C.dpiVar
}

// Close closes the statement.
//
// As of Go 1.1, a Stmt will not be closed if it's in use
// by any queries.
func (st *statement) Close() error {
	if C.dpiStmt_close(st.dpiStmt, nil, 0) == C.DPI_FAILURE {
		return st.getError()
	}
	return nil
}

// NumInput returns the number of placeholder parameters.
//
// If NumInput returns >= 0, the sql package will sanity check
// argument counts from callers and return errors to the caller
// before the statement's Exec or Query methods are called.
//
// NumInput may also return -1, if the driver doesn't know
// its number of placeholders. In that case, the sql package
// will not sanity check Exec or Query argument counts.
func (st *statement) NumInput() int {
	var colCount C.uint32_t
	if C.dpiStmt_execute(st.dpiStmt, C.DPI_MODE_EXEC_PARSE_ONLY, &colCount) == C.DPI_FAILURE {
		return -1
	}
	var cnt C.uint32_t
	if C.dpiStmt_getBindCount(st.dpiStmt, &cnt) == C.DPI_FAILURE {
		return -1
	}
	return int(cnt)
}

// Exec executes a query that doesn't return rows, such
// as an INSERT or UPDATE.
//
// Deprecated: Drivers should implement StmtExecContext instead (or additionally).
func (st *statement) Exec(args []driver.Value) (driver.Result, error) {
	nargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		nargs[i].Ordinal = i + 1
		nargs[i].Value = arg
	}
	return st.ExecContext(context.Background(), nargs)
}

// Query executes a query that may return rows, such as a
// SELECT.
//
// Deprecated: Drivers should implement StmtQueryContext instead (or additionally).
func (st *statement) Query(args []driver.Value) (driver.Rows, error) {
	nargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		nargs[i].Ordinal = i + 1
		nargs[i].Value = arg
	}
	return st.QueryContext(context.Background(), nargs)
}

// ExecContext executes a query that doesn't return rows, such as an INSERT or UPDATE.
//
// ExecContext must honor the context timeout and return when it is canceled.
func (st *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// bind variables
	if err := st.bindVars(args); err != nil {
		return nil, err
	}

	// execute
	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctx.Done():
			st.Break()
		case <-done:
			return
		}
	}()

	mode := C.dpiExecMode(C.DPI_MODE_EXEC_DEFAULT)
	if !st.inTransaction {
		mode |= C.DPI_MODE_EXEC_COMMIT_ON_SUCCESS
	}
	var colCount C.uint32_t
	res := C.dpiStmt_execute(st.dpiStmt, mode, &colCount)
	done <- struct{}{}
	if res == C.DPI_FAILURE {
		return nil, st.getError()
	}
	var count C.uint64_t
	if C.dpiStmt_getRowCount(st.dpiStmt, &count) == C.DPI_FAILURE {
		return nil, nil
	}
	return driver.RowsAffected(count), nil
}

// QueryContext executes a query that may return rows, such as a SELECT.
//
// QueryContext must honor the context timeout and return when it is canceled.
func (st *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// bind variables
	if err := st.bindVars(args); err != nil {
		return nil, err
	}

	// execute
	done := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctx.Done():
			st.Break()
		case <-done:
			return
		}
	}()
	var colCount C.uint32_t
	res := C.dpiStmt_execute(st.dpiStmt, C.DPI_MODE_EXEC_DEFAULT, &colCount)
	done <- struct{}{}
	if res == C.DPI_FAILURE {
		log.Printf("dpiStmt_execute: %+v", st.getError())
		return nil, st.getError()
	}
	return st.openRows(int(colCount))
}

// bindVars binds the given args into new variables.
//
// FIXME(tgulacsi): handle sql.Out params and arrays as ExecuteMany OR PL/SQL arrays.
func (st *statement) bindVars(args []driver.NamedValue) error {
	var named bool
	if cap(st.vars) < len(args) {
		st.vars = make([]*C.dpiVar, len(args))
	} else {
		st.vars = st.vars[:len(args)]
	}
	if cap(st.data) < len(args) {
		st.data = make([][]*C.dpiData, len(args))
	} else {
		st.data = st.data[:len(args)]
	}
	for i, a := range args {
		if !named {
			named = a.Name != ""
		}
		var set func(data *C.dpiData, v interface{}) error

		var typ C.dpiOracleTypeNum
		var natTyp C.dpiNativeTypeNum
		var bufSize C.uint32_t
		switch v := a.Value.(type) {
		case Lob:
			typ, natTyp = C.DPI_ORACLE_TYPE_BLOB, C.DPI_NATIVE_TYPE_LOB
			if v.IsClob {
				typ = C.DPI_ORACLE_TYPE_CLOB
			}
			set = func(data *C.dpiData, v interface{}) error {
				L := v.(Lob)
				var lob *C.dpiLob
				if C.dpiConn_newTempLob(st.dpiConn, typ, &lob) == C.DPI_FAILURE {
					return st.getError()
				}
				if C.dpiLob_openResource(lob) == C.DPI_FAILURE {
					return st.getError()
				}
				var offset C.uint64_t
				p := make([]byte, 1<<20)
				for {
					n, err := L.Read(p)
					if n > 0 {
						if C.dpiLob_writeBytes(lob, offset, (*C.char)(unsafe.Pointer(&p[0])), C.uint64_t(n)) == C.DPI_FAILURE {
							return st.getError()
						}
						offset += C.uint64_t(n)
					}
					if err != nil {
						if err == io.EOF {
							break
						}
						return err
					}
				}
				if C.dpiLob_closeResource(lob) == C.DPI_FAILURE {
					return st.getError()
				}
				C.dpiData_setLOB(data, lob)
				return nil
			}
		case int64:
			typ, natTyp = C.DPI_ORACLE_TYPE_NUMBER, C.DPI_NATIVE_TYPE_INT64
			set = func(data *C.dpiData, v interface{}) error {
				C.dpiData_setInt64(data, C.int64_t(v.(int64)))
				return nil
			}
		case uint64:
			typ, natTyp = C.DPI_ORACLE_TYPE_NUMBER, C.DPI_NATIVE_TYPE_UINT64
			set = func(data *C.dpiData, v interface{}) error {
				C.dpiData_setUint64(data, C.uint64_t(v.(uint64)))
				return nil
			}
		case float32:
			typ, natTyp = C.DPI_ORACLE_TYPE_NUMBER, C.DPI_NATIVE_TYPE_FLOAT
			set = func(data *C.dpiData, v interface{}) error {
				C.dpiData_setFloat(data, C.float(v.(float32)))
				return nil
			}
		case float64:
			typ, natTyp = C.DPI_ORACLE_TYPE_NUMBER, C.DPI_NATIVE_TYPE_DOUBLE
			set = func(data *C.dpiData, v interface{}) error {
				C.dpiData_setDouble(data, C.double(v.(float64)))
				return nil
			}
		case bool:
			typ, natTyp = C.DPI_ORACLE_TYPE_BOOLEAN, C.DPI_NATIVE_TYPE_BOOLEAN
			set = func(data *C.dpiData, v interface{}) error {
				b := C.int(0)
				if v.(bool) {
					b = 1
				}
				C.dpiData_setBool(data, b)
				return nil
			}
		case []byte:
			typ, natTyp = C.DPI_ORACLE_TYPE_RAW, C.DPI_NATIVE_TYPE_BYTES
			bufSize = C.uint32_t(len(v))
			set = func(data *C.dpiData, v interface{}) error {
				b := v.([]byte)
				C.dpiData_setBytes(data, (*C.char)(unsafe.Pointer(&b[0])), C.uint32_t(len(b)))
				return nil
			}
		case string:
			typ, natTyp = C.DPI_ORACLE_TYPE_VARCHAR, C.DPI_NATIVE_TYPE_BYTES
			bufSize = 4 * C.uint32_t(len(v))
			set = func(data *C.dpiData, v interface{}) error {
				b := []byte(v.(string))
				C.dpiData_setBytes(data, (*C.char)(unsafe.Pointer(&b[0])), C.uint32_t(len(b)))
				return nil
			}
		case time.Time:
			typ, natTyp = C.DPI_ORACLE_TYPE_TIMESTAMP_TZ, C.DPI_NATIVE_TYPE_TIMESTAMP
			set = func(data *C.dpiData, v interface{}) error {
				t := v.(time.Time)
				_, z := t.Zone()
				C.dpiData_setTimestamp(data,
					C.int16_t(t.Year()), C.uint8_t(t.Month()), C.uint8_t(t.Day()),
					C.uint8_t(t.Hour()), C.uint8_t(t.Minute()), C.uint8_t(t.Second()), C.uint32_t(t.Nanosecond()),
					C.int8_t(z/3600), C.int8_t((z%3600)/60),
				)
				return nil
			}
		default:
			return errors.Errorf("%d. arg: unknown type %T", i+1, a.Value)
		}
		var dataArr *C.dpiData
		if C.dpiConn_newVar(
			st.conn.dpiConn, typ, natTyp, 1,
			bufSize, 1, 0, nil,
			&st.vars[i], &dataArr,
		) == C.DPI_FAILURE {
			return st.getError()
		}
		st.data[i] = (*((*[maxArraySize]*C.dpiData)(unsafe.Pointer(&dataArr))))[:]

		if err := set(st.data[i][0], a.Value); err != nil {
			return err
		}
		log.Printf("set %d to %#v(%T): %#v", i, a.Value, a.Value, st.data[i][0])
	}

	return nil
}

// CheckNamedValue is called before passing arguments to the driver
// and is called in place of any ColumnConverter. CheckNamedValue must do type
// validation and conversion as appropriate for the driver.
//
// If CheckNamedValue returns ErrRemoveArgument, the NamedValue will not be included
// in the final query arguments.
// This may be used to pass special options to the query itself.
//
// If ErrSkip is returned the column converter error checking path is used
// for the argument.
// Drivers may wish to return ErrSkip after they have exhausted their own special cases.
func (st *statement) CheckNamedValue(nv *driver.NamedValue) error {
	if nv == nil {
		return nil
	}
	switch x := nv.Value.(type) {
	case int:
		nv.Value = int64(x)
	case uint:
		nv.Value = uint64(x)
	}
	return nil
}

func (st *statement) openRows(colCount int) (*rows, error) {
	C.dpiStmt_setFetchArraySize(st.dpiStmt, fetchRowCount)

	r := rows{
		statement: st,
		columns:   make([]Column, colCount),
		vars:      make([]*C.dpiVar, colCount),
		data:      make([][]*C.dpiData, colCount),
	}
	var info C.dpiQueryInfo
	for i := 0; i < colCount; i++ {
		if C.dpiStmt_getQueryInfo(st.dpiStmt, C.uint32_t(i+1), &info) == C.DPI_FAILURE {
			return nil, st.getError()
		}
		bufSize := maxArraySize * info.clientSizeInBytes
		//fmt.Println(typ, numTyp, info.precision, info.scale, info.clientSizeInBytes)
		switch info.defaultNativeTypeNum {
		case C.DPI_ORACLE_TYPE_NUMBER:
			info.defaultNativeTypeNum = C.DPI_NATIVE_TYPE_BYTES
		case C.DPI_ORACLE_TYPE_DATE:
			info.defaultNativeTypeNum = C.DPI_NATIVE_TYPE_TIMESTAMP
		}
		r.columns[i] = Column{
			Name:           C.GoStringN(info.name, C.int(info.nameLength)),
			Type:           info.oracleTypeNum,
			DefaultNumType: info.defaultNativeTypeNum,
			Size:           info.clientSizeInBytes,
			Precision:      info.precision,
			Scale:          info.scale,
			Nullable:       info.nullOk == 1,
			ObjectType:     info.objectType,
		}
		switch info.oracleTypeNum {
		case C.DPI_ORACLE_TYPE_VARCHAR, C.DPI_ORACLE_TYPE_NVARCHAR, C.DPI_ORACLE_TYPE_CHAR, C.DPI_ORACLE_TYPE_NCHAR:
			bufSize *= 4
		}
		var dataArr *C.dpiData
		if C.dpiConn_newVar(
			st.conn.dpiConn, info.oracleTypeNum, info.defaultNativeTypeNum, maxArraySize,
			bufSize, 1, 0, info.objectType,
			&r.vars[i], &dataArr,
		) == C.DPI_FAILURE {
			return nil, st.getError()
		}
		if C.dpiStmt_define(st.dpiStmt, C.uint32_t(i+1), r.vars[i]) == C.DPI_FAILURE {
			return nil, st.getError()
		}
		r.data[i] = (*((*[maxArraySize]*C.dpiData)(unsafe.Pointer(&dataArr))))[:]
	}
	if C.dpiStmt_addRef(st.dpiStmt) == C.DPI_FAILURE {
		return &r, st.getError()
	}
	return &r, nil
}

type Column struct {
	Name           string
	Type           C.dpiOracleTypeNum
	DefaultNumType C.dpiNativeTypeNum
	Size           C.uint32_t
	Precision      C.int16_t
	Scale          C.int8_t
	Nullable       bool
	ObjectType     *C.dpiObjectType
}
