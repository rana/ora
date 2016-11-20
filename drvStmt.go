// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
	"fmt"
)

// DrvStmt is an Oracle statement associated with a session.
//
// DrvStmt wraps Stmt and is intended for use by the database/sql/driver package.
//
// DrvStmt implements the driver.Stmt interface.
type DrvStmt struct {
	stmt *Stmt
}

// checkIsOpen validates that the server is open.
func (ds *DrvStmt) checkIsOpen() error {
	if ds.stmt == nil {
		return er("DrvStmt is closed.")
	}
	return nil
}

// Close closes the SQL statement.
//
// Close is a member of the driver.Stmt interface.
func (ds *DrvStmt) Close() error {
	ds.log(true)
	if err := ds.checkIsOpen(); err != nil {
		return errE(err)
	}
	if err := ds.stmt.Close(); err != nil {
		return errE(err)
	}
	return nil
}

// NumInput returns the number of placeholders in a sql statement.
//
// NumInput is a member of the driver.Stmt interface.
func (ds *DrvStmt) NumInput() int {
	if ds.stmt == nil {
		return 0
	}
	return ds.stmt.NumInput()
}

// Exec executes an Oracle SQL statement on a server. Exec returns a driver.Result
// and a possible error.
//
// Exec is a member of the driver.Stmt interface.
func (ds *DrvStmt) Exec(values []driver.Value) (driver.Result, error) {
	ds.log(true)
	if err := ds.checkIsOpen(); err != nil {
		return nil, errE(err)
	}
	params := make([]interface{}, len(values))
	for n := range values {
		params[n] = values[n]
	}
	rowsAffected, lastInsertId, err := ds.stmt.exe(params, false)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	if rowsAffected == 0 {
		return driver.RowsAffected(0), nil
	}
	return &DrvExecResult{rowsAffected: rowsAffected, lastInsertId: lastInsertId}, nil
}

// Query runs a SQL query on an Oracle server. Query returns driver.Rows and a
// possible error.
//
// Query is a member of the driver.Stmt interface.
func (ds *DrvStmt) Query(values []driver.Value) (driver.Rows, error) {
	ds.log(true)
	if err := ds.checkIsOpen(); err != nil {
		return nil, errE(err)
	}
	params := make([]interface{}, len(values))
	for n := range values {
		params[n] = values[n]
	}
	rset, err := ds.stmt.qry(params)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return &DrvQueryResult{rset: rset}, nil
}

// sysName returns a string representing the DrvStmt.
func (ds *DrvStmt) sysName() string {
	if ds == nil {
		return "E_S_S_S_S_"
	}
	return ds.stmt.sysName() + fmt.Sprintf("S%v", ds.stmt.id)
}

// log writes a message with an DrvStmt system name and caller info.
func (ds *DrvStmt) log(enabled bool, v ...interface{}) {
	cfg := _drv.Cfg()
	if !cfg.Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		cfg.Log.Logger.Infof("%v %v", ds.sysName(), callInfo(1))
	} else {
		cfg.Log.Logger.Infof("%v %v %v", ds.sysName(), callInfo(1), fmt.Sprint(v...))
	}
}
