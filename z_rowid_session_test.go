//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"
)

// test on heap table to retreive ROWID
func TestDefine_string_rowid_session(t *testing.T) {
	t.Parallel()
	testRowid(false, t)
}

// test on indexed table to retrieve UROWID
func TestDefine_string_urowid_session(t *testing.T) {
	t.Parallel()
	testRowid(true, t)
}

func testRowid(isUrowid bool, t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	for n := 0; n < testIterations(); n++ {
		tableName := tableName()
		stmt, err := testSes.Prep(fmt.Sprintf("create table %v (c1 varchar2(48 byte))", tableName))
		defer stmt.Close()
		testErr(err, t)
		_, err = stmt.Exe()
		defer dropTable(tableName, testSes, t)
		testErr(err, t)
		// ROWID is returned from a table without an index
		// UROWID is returned from indexed tables
		if isUrowid {
			stmt, err := testSes.Prep(fmt.Sprintf("create unique index t1_pk on %v (c1)", tableName))
			defer stmt.Close()
			testErr(err, t)
			_, err = stmt.Exe()
			testErr(err, t)
		}

		// insert
		insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values ('go')", tableName))
		defer insertStmt.Close()
		testErr(err, t)
		rowsAffected, err := insertStmt.Exe()
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// select
		selectStmt, err := testSes.Prep(fmt.Sprintf("select rowid from %v", tableName))
		defer selectStmt.Close()
		testErr(err, t)
		rset, err := selectStmt.Qry()
		testErr(err, t)
		hasRow := rset.Next()
		testErr(rset.Err(), t)
		if !hasRow {
			t.Fatalf("%d. no row returned", n)
		} else if len(rset.Row) != 1 {
			t.Fatalf("%d. select column count: expected(%v), actual(%v)", n, 1, len(rset.Row))
		}

		rowid, ok := rset.Row[0].(string)
		if !ok {
			t.Fatalf("%d. Expected string rowid. (%T, %v)", n, rset.Row[0], rset.Row[0])
		}
		if rowid == "" {
			t.Fatalf("%d. Expected non-empty rowid string. (%v)", n, rowid)
		}
		//fmt.Printf("rowid (%v)\n", rowid)

		want := "go go go"
		updateStmt, err := testSes.Prep(fmt.Sprintf("update %v set c1 = :1 where rowid = :2", tableName))
		defer updateStmt.Close()
		testErr(err, t)
		rowsAffected, err = updateStmt.Exe(want, rowid)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("%d. update rows affected: expected(%v), actual(%v)", n, 1, rowsAffected)
		}

		stmtSelect2, err := testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
		defer stmtSelect2.Close()
		testErr(err, t)
		rset2, err := stmtSelect2.Qry()
		testErr(err, t)
		rset2.Next()
		testErr(rset2.Err(), t)
		c1, ok := rset2.Row[0].(string)
		if !ok {
			t.Fatalf("%d. Expected string for c1 column. (%T, %v)", n, rset2.Row[0], rset2.Row[0])
		}
		//fmt.Printf("c1 (%v)\n", c1)
		if c1 != want {
			t.Fatalf("%d. Got %q, wanted %q.", n, c1, want)
		}
	}
}
