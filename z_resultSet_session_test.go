//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	"gopkg.in/rana/ora.v3"
)

func Test_cursor1_session(t *testing.T) {

	// create table
	tableName := tableName()
	createTblStmt, err := testSes.Prep(fmt.Sprintf("create table %v (c1 varchar2(48 char), c2 integer)", tableName))
	defer createTblStmt.Close()
	testErr(err, t)
	defer dropTable(tableName, testSes, t)
	_, err = createTblStmt.Exe()
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
	insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1, c2) values (:1, :2)", tableName))
	testErr(err, t)
	rowsAffected, err := insertStmt.Exe(expectedStrs, expectedInt64s)
	testErr(err, t)
	if rowsAffected != 3 {
		t.Fatalf("Expected 3 rows affected. (rowsAffected %v)", rowsAffected)
	}

	// create proc
	createProcStmt, err := testSes.Prep(fmt.Sprintf("create or replace procedure proc1(p1 out sys_refcursor) as begin open p1 for select c1, c2 from %v order by c2; end proc1;", tableName))
	defer createProcStmt.Close()
	testErr(err, t)
	_, err = createProcStmt.Exe()
	testErr(err, t)

	// call proc
	callProcStmt, err := testSes.Prep("call proc1(:1)")
	defer callProcStmt.Close()
	testErr(err, t)
	rset := &ora.Rset{}
	_, err = callProcStmt.Exe(rset)
	testErr(err, t)
	if rset.IsOpen() {
		for rset.Next() {
			if len(rset.Row) != 2 {
				t.Fatalf("select column count: expected(%v), actual(%v)", 2, len(rset.Row))
			} else {
				//fmt.Println("rset.Row ", rset.Row)
				compare(expectedStrs[rset.Index], rset.Row[0], ora.S, t)
				compare(expectedInt64s[rset.Index], rset.Row[1], ora.I64, t)
			}
		}
		testErr(rset.Err, t)
	}
}
