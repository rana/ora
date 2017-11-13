//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	"gopkg.in/rana/ora.v4"
)

func Test_cursor1_session(t *testing.T) {
	t.Parallel()
	testSes := getSes(t)
	defer testSes.Close()

	// create table
	tableName := tableName()
	createTblStmt, err := testSes.Prep(fmt.Sprintf("create table %v (c1 varchar2(48 char), c2 integer)", tableName))
	defer createTblStmt.Close()
	testErr(err, t)
	defer dropTable(tableName, testSes, t)
	_, err = createTblStmt.Exe()
	testErr(err, t)

	// insert records
	expectedStrs := []string{
		"Go is expressive, concise, clean, and efficient.",
		"Its concurrency mechanisms make it easy to",
		"Go compiles quickly to machine code yet has",
	}
	expectedInt64s := []int64{3, 7, 9}
	rowsAffected, err := testSes.PrepAndExe(
		fmt.Sprintf("insert into %v (c1, c2) values (:1, :2)", tableName),
		expectedStrs, expectedInt64s,
	)
	testErr(err, t)
	if rowsAffected != 3 {
		t.Fatalf("Expected 3 rows affected. (rowsAffected %v)", rowsAffected)
	}

	// create proc
	_, err = testSes.PrepAndExe(fmt.Sprintf("create or replace procedure proc1(p1 out sys_refcursor) as begin open p1 for select c1, c2 from %v order by c2; end proc1;", tableName))
	testErr(err, t)

	//enableLogging(t)
	// call proc
	stmt, err := testSes.Prep("call proc1(:1)")
	testErr(err, t)
	var rset ora.Rset
	_, err = stmt.Exe(&rset)
	testErr(err, t)

	if !rset.IsOpen() {
		t.Fatalf("rset %#v is closed!", rset)
	}
	for rset.Next() {
		if len(rset.Row) != 2 {
			t.Fatalf("select column count: expected(%v), actual(%v)", 2, len(rset.Row))
		}
		//fmt.Println("rset.Row ", rset.Row)
		compare(expectedStrs[0], rset.Row[0], ora.S, t)
		compare(expectedInt64s[0], rset.Row[1], ora.I64, t)
		expectedStrs = expectedStrs[1:]
		expectedInt64s = expectedInt64s[1:]
	}
	testErr(rset.Err(), t)
	if len(expectedStrs) > 0 {
		t.Errorf("didn't get wanted %v", expectedStrs)
	}
}

func Test_nested_rset(t *testing.T) {
	t.Parallel()
	testSes := getSes(t)
	defer testSes.Close()

	_, err := testSes.PrepAndExe(`CREATE OR REPLACE PROCEDURE proc2(p_cur OUT SYS_REFCURSOR) IS
BEGIN
  OPEN p_cur FOR
    SELECT CURSOR(SELECT A.* FROM user_objects A, (SELECT 1 FROM DUAL)) cur FROM DUAL;
END;`)
	if err != nil {
		t.Fatal(err)
	}
	//enableLogging(t)
	stmt, err := testSes.Prep("call proc2(:1)")
	testErr(err, t)
	//enableLogging(t)
	var rset ora.Rset
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatal(r)
			}
		}()
		_, err = stmt.Exe(&rset)
	}()
	if err != nil {
		errs, _ := GetCompileErrors(testSes, false)
		t.Errorf("errs: %#v", errs)
		t.Fatal(err)
	}

	for rset.Next() {
	}
}
