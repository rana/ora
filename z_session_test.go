// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"testing"
)

func TestSession_PrepCloseStmt(t *testing.T) {

	// setup
	env, err := GetDrv().OpenEnv()
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenSrv(testServerName)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSes(testUsername, testPassword)
	defer ses.Close()
	testErr(err, t)

	stmt, err := ses.Prep("select 'go' from dual")
	testErr(err, t)

	err = stmt.Close()
	testErr(err, t)
}

func TestSession_Tx_StartCommit(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	tx, err := testSes.StartTx()
	testErr(err, t)

	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	err = tx.Commit()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Query()
	testErr(err, t)

	for rset.Next() {

	}
	if 2 != rset.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 2, rset.Len())
	}
}

func TestSession_Tx_StartRollback(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	tx, err := testSes.StartTx()
	testErr(err, t)

	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	err = tx.Rollback()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Query()
	testErr(err, t)

	for rset.Next() {
	}
	if 0 != rset.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 0, rset.Len())
	}
}

func TestSession_PrepAndExec(t *testing.T) {
	rowsAffected, err := testSes.PrepAndExec(fmt.Sprintf("create table %v (c1 number)", tableName()))
	testErr(err, t)

	if rowsAffected != 0 {
		t.Fatalf("expected(%v), actual(%v)", 0, rowsAffected)
	}
}

func TestSession_PrepAndExec_Insert(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	rowsAffected, err := testSes.PrepAndExec(fmt.Sprintf("INSERT INTO %v (C1) VALUES (:C1)", tableName), values)
	testErr(err, t)

	if rowsAffected != 1000000 {
		t.Fatalf("expected(%v), actual(%v)", 1000000, rowsAffected)
	}
}

func TestSession_PrepAndQuery(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	// insert one row
	stmtIns, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmtIns.Exec()
	testErr(err, t)

	stmt, rset, err := testSes.PrepAndQuery(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)
	if stmt == nil {
		t.Fatalf("expected non-nil stmt")
	}
	defer stmt.Close()
	if rset == nil {
		t.Fatalf("expected non-nil rset")
	}

	row := rset.NextRow()
	if row[0] == 9 {
		t.Fatalf("expected(%v), actual(%v)", 9, row[0])
	}
}
