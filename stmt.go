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
	"unsafe"
)

var _ = driver.Stmt((*statement)(nil))
var _ = driver.StmtQueryContext((*statement)(nil))
var _ = driver.StmtExecContext((*statement)(nil))

const sizeof_dpiData = C.sizeof_dpiData

type statement struct {
	*conn
	dpiStmt *C.dpiStmt
	query   string
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
		nargs[i].Ordinal = i
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
		nargs[i].Ordinal = i
		nargs[i].Value = arg
	}
	return st.QueryContext(context.Background(), nargs)
}

// ExecContext executes a query that doesn't return rows, such as an INSERT or UPDATE.
//
// ExecContext must honor the context timeout and return when it is canceled.
func (st *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	mode := C.dpiExecMode(C.DPI_MODE_EXEC_DEFAULT)
	if !st.inTransaction {
		mode |= C.DPI_MODE_EXEC_COMMIT_ON_SUCCESS
	}
	var colCount C.uint32_t
	if C.dpiStmt_execute(st.dpiStmt, mode, &colCount) == C.DPI_FAILURE {
		return nil, st.getError()
	}
	return nil, nil
}

// QueryContext executes a query that may return rows, such as a SELECT.
//
// QueryContext must honor the context timeout and return when it is canceled.
func (st *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var colCount C.uint32_t
	if C.dpiStmt_execute(st.dpiStmt, C.DPI_MODE_EXEC_DEFAULT, &colCount) == C.DPI_FAILURE {
		return nil, st.getError()
	}
	return st.openRows(int(colCount))
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
