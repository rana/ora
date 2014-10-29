//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"testing"
)

func Test_cursor1_session(t *testing.T) {

	// create table
	tableName := tableName()
	createTblStmt, err := testSes.Prepare(fmt.Sprintf("create table %v (c1 varchar2(48 char), c2 number)", tableName))
	defer createTblStmt.Close()
	testErr(err, t)
	defer dropTable(tableName, testSes, t)
	_, err = createTblStmt.Execute()
	testErr(err, t)

	// insert records
	expectedStrs := make([]string, 3)
	expectedStrs[0] = "Go is expressive, concise, clean, and efficient."
	expectedStrs[1] = "Its concurrency mechanisms make it easy to"
	expectedStrs[2] = "Go compiles quickly to machine code yet has"
	expectedInt64s := make([]int64, 3)
	expectedInt64s[0] = 3
	expectedInt64s[1] = 7
	expectedInt64s[2] = 9
	insertStmt, err := testSes.Prepare(fmt.Sprintf("insert into %v (c1, c2) values (:1, :2)", tableName))
	testErr(err, t)
	rowsAffected, err := insertStmt.Execute(expectedStrs, expectedInt64s)
	testErr(err, t)
	if rowsAffected != 3 {
		t.Fatalf("Expected 3 rows affected. (rowsAffected %v)", rowsAffected)
	}

	// create proc
	createProcStmt, err := testSes.Prepare(fmt.Sprintf("create or replace procedure proc1(p1 out sys_refcursor) as begin open p1 for select c1, c2 from %v order by c2; end proc1;", tableName))
	defer createProcStmt.Close()
	testErr(err, t)
	_, err = createProcStmt.Execute()
	testErr(err, t)

	// call proc
	callProcStmt, err := testSes.Prepare("call proc1(:1)")
	defer callProcStmt.Close()
	testErr(err, t)
	resultSet := &ResultSet{}
	_, err = callProcStmt.Execute(resultSet)
	testErr(err, t)
	if resultSet.IsOpen() {
		for resultSet.Next() {
			if len(resultSet.Row) != 2 {
				t.Fatalf("select column count: expected(%v), actual(%v)", 2, len(resultSet.Row))
			} else {
				//fmt.Println("resultSet.Row ", resultSet.Row)
				compare(expectedStrs[resultSet.Index], resultSet.Row[0], S, t)
				compare(expectedInt64s[resultSet.Index], resultSet.Row[1], I64, t)
			}
		}
		testErr(resultSet.Err, t)
	}
}
