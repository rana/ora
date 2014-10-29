// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"testing"
)

func TestSession_PrepareCloseStatement(t *testing.T) {

	// setup
	env := NewEnvironment()
	env.Open()
	defer env.Close()
	srv, err := env.OpenServer(testServerName)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSession(testUsername, testPassword)
	defer ses.Close()
	testErr(err, t)

	stmt, err := ses.Prepare("select 'go' from dual")
	testErr(err, t)

	err = stmt.Close()
	testErr(err, t)
}

func TestSession_Transaction_BeginCommit(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	tx, err := testSes.BeginTransaction()
	testErr(err, t)

	stmt, err := testSes.Prepare(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmt.Execute()
	testErr(err, t)

	stmt, err = testSes.Prepare(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Execute()
	testErr(err, t)

	err = tx.Commit()
	testErr(err, t)

	stmt, err = testSes.Prepare(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	resultSet, err := stmt.Fetch()
	testErr(err, t)

	for resultSet.Next() {

	}
	if 2 != resultSet.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 2, resultSet.Len())
	}
}

func TestSession_Transaction_BeginRollback(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	tx, err := testSes.BeginTransaction()
	testErr(err, t)

	stmt, err := testSes.Prepare(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmt.Execute()
	testErr(err, t)

	stmt, err = testSes.Prepare(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Execute()
	testErr(err, t)

	err = tx.Rollback()
	testErr(err, t)

	stmt, err = testSes.Prepare(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	resultSet, err := stmt.Fetch()
	testErr(err, t)

	for resultSet.Next() {
	}
	if 0 != resultSet.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 0, resultSet.Len())
	}
}
