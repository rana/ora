// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"testing"
)

func TestStmt_Exec_table_create_alter_drop(t *testing.T) {
	tableName := tableName()

	// create table
	stmt, err := testSes.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	// alter table
	stmt, err = testSes.Prep(fmt.Sprintf("alter table %v add c2 number", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)

	// drop table
	stmt, err = testSes.Prep(fmt.Sprintf("drop table %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)
}

func TestStmt_Exec_insert(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exec_update(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// update record
	stmt, err = testSes.Prep(fmt.Sprintf("update %v set c1 = 8 where c1 = 9", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exec_delete(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// delete record
	stmt, err = testSes.Prep(fmt.Sprintf("delete %v where c1 = 9", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exec_select(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// insert record
	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exec()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// fetch records
	stmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	rset, err := stmt.Query()
	testErr(err, t)

	for rset.Next() {
		switch rset.Index {
		case 0:
			compare_int64(int64(9), rset.Row[0], t)
		case 1:
			compare_int64(int64(11), rset.Row[0], t)
		}
	}
	if 2 != rset.Len() {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 2, rset.Len())
	}
}
