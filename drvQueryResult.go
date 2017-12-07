// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

/*
#include <oci.h>
*/
import "C"

// DrvQueryResult contains methods to retrieve the results of a SQL select statement.
//
// DrvQueryResult implements the driver.Rows interface.
type DrvQueryResult struct {
	rset *Rset
}

// Next populates the specified slice with the next row of data.
//
// Returns io.EOF when there are no more rows.
//
// Next is a member of the driver.Rows interface.
func (qr *DrvQueryResult) Next(dest []driver.Value) (err error) {
	if qr.rset == nil {
		return er("empty Rset")
	}
	err = qr.rset.beginRow()
	if err != nil {
		// FIXME(tgulacsi): this results in erroneous short iteration!
		qr.rset.closeWithRemove()
		// but without this close, memory consumption grows!
		qr.rset = nil
		return err
	}
	defer qr.rset.endRow()

	// Populate column values into destination slice
	qr.rset.RLock()
	defer qr.rset.RUnlock()
	if len(dest) < len(qr.rset.defs) {
		return fmt.Errorf("Short dest: got %d, wanted %d.", len(dest), len(qr.rset.defs))
	}
	offset := int(qr.rset.offset)
	for n, define := range qr.rset.defs {
		value, err := define.value(offset)
		if err != nil {
			fmt.Printf("%d. %T (%#v): %v\n", n, define, define, err)
			return err
		}
		dest[n] = value
	}
	return nil
}

// HasNextResultSet reports whether there is another result set after the current one.
func (qr *DrvQueryResult) HasNextResultSet() bool { return false }

// NextResultSet advances the driver to the next result set even
// if there are remaining rows in the current result set.
func (qr *DrvQueryResult) NextResultSet() error { return io.EOF }

// Columns returns query column names.
//
// Columns is a member of the driver.Rows interface.
func (qr *DrvQueryResult) Columns() []string {
	if qr.rset == nil {
		return nil
	}
	qr.rset.RLock()
	names := make([]string, len(qr.rset.Columns))
	for i, c := range qr.rset.Columns {
		names[i] = c.Name
	}
	qr.rset.RUnlock()
	return names
}

// ColumnTypeDatabaseTypeName returns the database system type name
// without the length, in uppercase.
func (qr *DrvQueryResult) ColumnTypeDatabaseTypeName(index int) string {
	if qr.rset == nil {
		return ""
	}
	// https://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16271
	qr.rset.RLock()
	x := qr.rset.Columns[index].Type
	qr.rset.RUnlock()
	switch x {
	case C.SQLT_CHR:
		return "VARCHAR2"
	case C.SQLT_NUM:
		return "NUMBER"
	case C.SQLT_INT:
		return "INTEGER"
	case C.SQLT_FLT:
		return "FLOAT"
	case C.SQLT_STR:
		return "STRING"
	case C.SQLT_VNU:
		return "VARNUM"
	case C.SQLT_LNG:
		return "LONG"
	case C.SQLT_VCS:
		return "VARCHAR"
	case C.SQLT_DAT:
		return "DATE"
	case C.SQLT_VBI:
		return "VARRAW"
	case C.SQLT_BFLOAT:
		return "NATIVE FLOAT"
	case C.SQLT_BDOUBLE:
		return "NATIVE DOUBLE"
	case C.SQLT_BIN:
		return "RAW"
	case C.SQLT_LBI:
		return "LONG RAW"
	case C.SQLT_UIN:
		return "UNSIGNED INT"
	case C.SQLT_LVC:
		return "LONG VARCHAR"
	case C.SQLT_LVB:
		return "LONG VARRAW"
	case C.SQLT_AFC:
		return "CHAR"
	case C.SQLT_AVC:
		return "CHARZ"
	case C.SQLT_RDD:
		return "ROWID"
	case C.SQLT_NTY:
		return "NAMED"
	case C.SQLT_REF:
		return "REF"
	case C.SQLT_CLOB:
		return "CLOB"
	case C.SQLT_BLOB:
		return "BLOB"
	case C.SQLT_FILE:
		return "FILE"
	case C.SQLT_VST:
		return "OCI STRING"
	case C.SQLT_ODT:
		return "OCI DATE"
	case C.SQLT_DATE:
		return "ANSI DATE"
	case C.SQLT_TIMESTAMP:
		return "TIMESTAMP"
	case C.SQLT_TIMESTAMP_TZ:
		return "TIMESTAMP WITH TIME ZONE"
	case C.SQLT_INTERVAL_YM:
		return "INTERVAL YEAR TO MONTH"
	case C.SQLT_INTERVAL_DS:
		return "INTERVAL DAY TO SECOND"
	case C.SQLT_TIMESTAMP_LTZ:
		return "TIMESTAMP WITH LOCAL TIME ZONE"
	default:
		return strconv.Itoa(int(x))
	}
}

// ColumnTypeLength returns the length of the column type
// if the column is a variable length type.
// If the column is not a variable length type ok should return false.
// If length is not limited other than system limits,
// it should return math.MaxInt64.
func (qr *DrvQueryResult) ColumnTypeLength(index int) (length int64, ok bool) {
	if qr.rset == nil {
		return 0, false
	}
	qr.rset.RLock()
	c := qr.rset.Columns[index]
	qr.rset.RUnlock()
	if c.Length == 0 {
		return 0, false
	}
	return int64(c.Length), true
}

// ColumnTypeNullable returns true if it is known the column may be null,
// or false if the column is known to be not nullable.
// If the column nullability is unknown, ok should be false.
func (qr *DrvQueryResult) ColumnTypeNullable(index int) (nullable, ok bool) {
	return true, true
}

// ColumnTypePrecisionScale return the precision and scale for decimal types.
// If not applicable, ok should be false.
func (qr *DrvQueryResult) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	if qr.rset == nil {
		return 0, 0, false
	}
	qr.rset.RLock()
	c := qr.rset.Columns[index]
	qr.rset.RUnlock()
	if c.Type == C.SQLT_NUM || c.Type == C.SQLT_INT {
		return int64(c.Precision), int64(c.Scale), true
	}
	return 0, 0, false
}

func (qr *DrvQueryResult) ColumnTypeScanType(index int) reflect.Type {
	if qr.rset == nil {
		return nil
	}
	qr.rset.RLock()
	x := qr.rset.Columns[index].Type
	qr.rset.RUnlock()
	switch x {
	case C.SQLT_CHR, C.SQLT_STR, C.SQLT_LNG, C.SQLT_VCS, C.SQLT_LVC, C.SQLT_AFC, C.SQLT_AVC, C.SQLT_CLOB, C.SQLT_VST:
		return reflect.TypeOf("")
	case C.SQLT_NUM, C.SQLT_FLT, C.SQLT_VNU, C.SQLT_BDOUBLE:
		return reflect.TypeOf(float64(0))
	case C.SQLT_INT:
		return reflect.TypeOf(int64(0))
	case C.SQLT_DAT, C.SQLT_ODT, C.SQLT_DATE, C.SQLT_TIMESTAMP, C.SQLT_TIMESTAMP_TZ, C.SQLT_TIMESTAMP_LTZ:
		return reflect.TypeOf(time.Time{})
	case C.SQLT_VBI, C.SQLT_BIN, C.SQLT_LBI, C.SQLT_LVB, C.SQLT_RDD, C.SQLT_BLOB, C.SQLT_FILE:
		return reflect.TypeOf([]byte{})
	case C.SQLT_BFLOAT:
		return reflect.TypeOf(float32(0))
	case C.SQLT_UIN:
		return reflect.TypeOf(uint64(0))

	case C.SQLT_NTY, C.SQLT_REF:
		return reflect.TypeOf([]byte{})
	case C.SQLT_INTERVAL_YM, C.SQLT_INTERVAL_DS:
		return reflect.TypeOf(time.Duration(0))

	default:
		var x interface{}
		return reflect.TypeOf(x)
	}

}

// Close performs no operations.
//
// Close is a member of the driver.Rows interface.
func (qr *DrvQueryResult) Close() error {
	if qr.rset == nil {
		return nil
	}
	return qr.rset.closeWithRemove()
}
