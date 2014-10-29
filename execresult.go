// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"math"
)

// An execution result.
//
// Implements the driver.Result interface.
type ExecResult struct {
	lastInsertId int64
	rowsAffected uint64
}

// LastInsertId returns the identity value from an insert statement.
//
// There are two setup steps required to reteive the LastInsertId.
// One, specify a 'returning into' clause in the SQL insert statement.
// And, two, specify a nil parameter to DB.Exec or Stmt.Exec.
//
// For example:
//	db, err := sql.Open("oracle", "scott/tiger@orcl")
//
//	db.Exec("create table t1 (c1 number(19,0) generated always as identity (start with 1 increment by 1), c2 varchar2(48 char))")
//
//	result, err := db.Exec("insert into t1 (c2) values ('go') returning c1 into :c1", nil)
//
//	id, err := result.LastInsertId()
func (execResult *ExecResult) LastInsertId() (int64, error) {
	return execResult.lastInsertId, nil
}

// RowsAffected returns the number of rows affected by the exec statement.
func (execResult *ExecResult) RowsAffected() (int64, error) {
	var rowsAffected int64
	if execResult.rowsAffected > math.MaxInt64 {
		rowsAffected = math.MaxInt64
	} else {
		rowsAffected = int64(execResult.rowsAffected)
	}
	return rowsAffected, nil
}
