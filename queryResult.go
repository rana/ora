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
	resultSet *ResultSet
}

// Next populates the specified slice with the next row of data.
//
// Returns io.EOF when there are no more rows.
//
// Next is a member of the driver.Rows interface.
func (queryResult *QueryResult) Next(dest []driver.Value) (err error) {
	err = queryResult.resultSet.beginRow()
	defer queryResult.resultSet.endRow()
	if err != nil {
		return err
	}
	// Populate column values into destination slice
	for n, define := range queryResult.resultSet.defines {
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
func (queryResult *QueryResult) Columns() []string {
	return queryResult.resultSet.ColumnNames
}

// Close performs no operations.
//
// Close is a member of the driver.Rows interface.
func (queryResult *QueryResult) Close() error {
	return nil
}
