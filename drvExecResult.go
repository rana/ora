// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"math"
)

// DrvExecResult is an Oracle execution result.
//
// DrvExecResult implements the driver.Result interface.
type DrvExecResult struct {
	lastInsertId int64
	rowsAffected uint64
}

// LastInsertId returns the identity value from an insert statement.
//
// There are two setup steps required to reteive the LastInsertId.
// One, specify a 'returning into' clause in the SQL insert statement.
// And, two, specify a nil parameter to DB.Exec or DrvStmt.Exec.
//
// For example:
//
//	db, err := sql.Open("ora", "scott/tiger@orcl")
//
//	db.Exec("CREATE TABLE T1 (C1 NUMBER(19,0) GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1), C2 VARCHAR2(48 CHAR))")
//
//	result, err := db.Exec("INSERT INTO T1 (C2) VALUES ('GO') RETURNING C1 /*lastInsertId*/ INTO :C1", nil)
//
//	id, err := result.LastInsertId()
func (er *DrvExecResult) LastInsertId() (int64, error) {
	return er.lastInsertId, nil
}

// RowsAffected returns the number of rows affected by the exec statement.
func (er *DrvExecResult) RowsAffected() (int64, error) {
	var rowsAffected int64
	if er.rowsAffected > math.MaxInt64 {
		rowsAffected = math.MaxInt64
	} else {
		rowsAffected = int64(er.rowsAffected)
	}
	return rowsAffected, nil
}
