//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"reflect"
	"testing"
)

// test on heap table to retreive ROWID
func TestDefine_string_rowid_session(t *testing.T) {
	testRowid(false, t)
}

// test on indexed table to retrieve UROWID
func TestDefine_string_urowid_session(t *testing.T) {
	testRowid(true, t)
}

func testRowid(isUrowid bool, t *testing.T) {
	for n := 0; n < testIterations(); n++ {
		tableName := tableName()
		statement, err := testSes.Prepare(fmt.Sprintf("create table %v (c1 varchar2(48 byte))", tableName))
		defer statement.Close()
		testErr(err, t)
		_, err = statement.Execute()
		defer dropTable(tableName, testSes, t)
		testErr(err, t)
		// ROWID is returned from a table without an index
		// UROWID is returned from indexed tables
		if isUrowid {
			statement, err := testSes.Prepare(fmt.Sprintf("create unique index t1_pk on %v (c1)", tableName))
			defer statement.Close()
			testErr(err, t)
			_, err = statement.Execute()
			testErr(err, t)
		}

		// insert
		insertStmt, err := testSes.Prepare(fmt.Sprintf("insert into %v (c1) values ('go')", tableName))
		defer insertStmt.Close()
		testErr(err, t)
		rowsAffected, err := insertStmt.Execute()
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// select
		selectStmt, err := testSes.Prepare(fmt.Sprintf("select rowid from %v", tableName))
		defer selectStmt.Close()
		testErr(err, t)
		resultSet, err := selectStmt.Fetch()
		testErr(err, t)
		hasRow := resultSet.Next()
		testErr(resultSet.Err, t)
		if !hasRow {
			t.Fatalf("no row returned")
		} else if len(resultSet.Row) != 1 {
			t.Fatalf("select column count: expected(%v), actual(%v)", 1, len(resultSet.Row))
		} else {
			rowid, ok := resultSet.Row[0].(string)
			if !ok {
				t.Fatal("Expected string rowid. (%v, %v)", reflect.TypeOf(resultSet.Row[0]).Name(), resultSet.Row[0])
			}
			if rowid == "" {
				t.Fatalf("Expected non-empty rowid string. (%v)", rowid)
			}
			//fmt.Printf("rowid (%v)\n", rowid)

			updateStmt, err := testSes.Prepare(fmt.Sprintf("update %v set c1 = 'go go go' where rowid = :1", tableName))
			defer updateStmt.Close()
			testErr(err, t)
			rowsAffected, err = updateStmt.Execute(rowid)
			testErr(err, t)
			if rowsAffected != 1 {
				t.Fatalf("update rows affected: expected(%v), actual(%v)", 1, rowsAffected)
			}

			selectStmt2, err := testSes.Prepare(fmt.Sprintf("select c1 from %v", tableName))
			defer selectStmt2.Close()
			testErr(err, t)
			resultSet2, err := selectStmt2.Fetch()
			testErr(err, t)
			resultSet2.Next()
			testErr(resultSet2.Err, t)
			c1, ok := resultSet2.Row[0].(string)
			if !ok {
				t.Fatal("Expected string for c1 column. (%v, %v)", reflect.TypeOf(resultSet2.Row[0]).Name(), resultSet2.Row[0])
			}
			//fmt.Printf("c1 (%v)\n", c1)
			if c1 != "go go go" {
				t.Fatalf("Expected 'go go go' string. (%v)", c1)
			}
		}
	}
}
