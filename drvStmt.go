// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import "database/sql/driver"

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
		return errNewF("DrvStmt is closed")
	}
	return nil
}

// Close closes the SQL statement.
//
// Close is a member of the driver.Stmt interface.
func (ds *DrvStmt) Close() error {
	if err := ds.checkIsOpen(); err != nil {
		return err
	}
	Log.Infof("E%vS%vS%vS%v] Close", ds.stmt.ses.srv.env.id, ds.stmt.ses.srv.id, ds.stmt.ses.id, ds.stmt.id)
	return ds.stmt.Close()
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
func (ds *DrvStmt) Exec(values []driver.Value) (result driver.Result, err error) {
	if err := ds.checkIsOpen(); err != nil {
		return nil, err
	}
	Log.Infof("E%vS%vS%vS%v] Exec", ds.stmt.ses.srv.env.id, ds.stmt.ses.srv.id, ds.stmt.ses.id, ds.stmt.id)
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	rowsAffected, lastInsertId, err := ds.stmt.exe(params)
	if rowsAffected == 0 {
		result = driver.ResultNoRows
	} else {
		result = &DrvExecResult{rowsAffected: rowsAffected, lastInsertId: lastInsertId}
	}
	return result, err
}

// Query runs a SQL query on an Oracle server. Query returns driver.Rows and a
// possible error.
//
// Query is a member of the driver.Stmt interface.
func (ds *DrvStmt) Query(values []driver.Value) (driver.Rows, error) {
	if err := ds.checkIsOpen(); err != nil {
		return nil, err
	}
	Log.Infof("E%vS%vS%vS%v] Query", ds.stmt.ses.srv.env.id, ds.stmt.ses.srv.id, ds.stmt.ses.id, ds.stmt.id)
	params := make([]interface{}, len(values))
	for n, _ := range values {
		params[n] = values[n]
	}
	rset, err := ds.stmt.qry(params)
	return &DrvQueryResult{rset: rset}, err
}
