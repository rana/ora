// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
)

// QueryResult contains methods to retrieve the results of a SQL select statement.
//
// Implements the driver.Rows interface.
type QueryResult struct {
	rst *ResultSet
}

// Next populates the specified slice with the next row of data.
//
// Returns io.EOF when there are no more rows.
//
// Next is a member of the driver.Rows interface.
func (qr *QueryResult) Next(dest []driver.Value) (err error) {
	err = qr.rst.beginRow()
	defer qr.rst.endRow()
	if err != nil {
		return err
	}
	// Populate column values into destination slice
	for n, define := range qr.rst.defines {
		value, err := define.value()
		if err != nil {
			return err
		}
		dest[n] = value
	}
	return nil
}

// Columns returns query column names.
//
// Columns is a member of the driver.Rows interface.
func (qr *QueryResult) Columns() []string {
	return qr.rst.ColumnNames
}

// Close performs no operations.
//
// Close is a member of the driver.Rows interface.
func (qr *QueryResult) Close() error {
	return nil
}
