// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
	"fmt"
)

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
	err = qr.rset.beginRow()
	defer qr.rset.endRow()
	if err != nil {
		return err
	}
	// Populate column values into destination slice
	for n, define := range qr.rset.defs {
		value, err := define.value()
		if err != nil {
			fmt.Printf("%d. %T (%#v): %v\n", n, define, define, err)
			return err
		}
		dest[n] = value
	}
	return nil
}

// Columns returns query column names.
//
// Columns is a member of the driver.Rows interface.
func (qr *DrvQueryResult) Columns() []string {
	return qr.rset.ColumnNames
}

// Close performs no operations.
//
// Close is a member of the driver.Rows interface.
func (qr *DrvQueryResult) Close() error {
	return nil
}
