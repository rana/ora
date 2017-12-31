// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	ora "gopkg.in/rana/ora.v4"

	"github.com/pkg/errors"
)

func TestStmt_Exe_table_create_alter_drop(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	tableName := tableName()

	// create table
	stmt, err := testSes.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)

	// alter table
	stmt, err = testSes.Prep(fmt.Sprintf("alter table %v add c2 number", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)

	// drop table
	stmt, err = testSes.Prep(fmt.Sprintf("drop table %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)
}

func TestStmt_Exe_insert(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	tableName, err := createTable(1, numberP38S0, testSes)
	if err != nil {
		t.Fatal(err)
	}
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exe_update(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	tableName, err := createTable(1, numberP38S0, testSes)
	if err != nil {
		t.Fatal(err)
	}
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// update record
	stmt, err = testSes.Prep(fmt.Sprintf("update %v set c1 = 8 where c1 = 9", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exe_delete(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	tableName, err := createTable(1, numberP38S0, testSes)
	if err != nil {
		t.Fatal(err)
	}
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// delete record
	stmt, err = testSes.Prep(fmt.Sprintf("delete %v where c1 = 9", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}
}

func TestStmt_Exe_select(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	tableName, err := createTable(1, numberP38S0, testSes)
	if err != nil {
		t.Fatal(err)
	}
	defer dropTable(tableName, testSes, t)

	// insert record
	stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err := stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// insert record
	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	defer stmt.Close()
	testErr(err, t)
	rowsAffected, err = stmt.Exe()
	testErr(err, t)
	if 1 != rowsAffected {
		t.Fatalf("rows affected: expected(%v), actual(%v)", 1, rowsAffected)
	}

	// fetch records
	qry := fmt.Sprintf("select c1 from %v", tableName)

	oCfg := ora.Cfg()
	defer ora.SetCfg(oCfg)
	cfg := oCfg
	cfg.Log.Rset.Next = true
	ora.SetCfg(cfg)
	//enableLogging(t)

	stmt, err = testSes.Prep(qry)
	testErr(errors.Wrap(err, qry), t)
	defer stmt.Close()
	rset, err := stmt.Qry()
	testErr(errors.Wrap(err, qry), t)

	var length int
	for rset.Next() {
		length++
		switch rset.Len() - 1 {
		case 0:
			compare_int64(int64(9), rset.Row[0], t)
		case 1:
			compare_int64(int64(11), rset.Row[0], t)
		}
	}
	if 2 != length {
		t.Fatalf("rows affected: expected(%v), actual(%v) err=%+v\n%s", 2, rset.Len(), rset.Err(), qry)
	}
}

func Benchmark_SimpleInsert(b *testing.B) {
	testSes := getSes(b)
	defer testSes.Close()

	tableName := tableName()
	testSes.PrepAndExe("CREATE TABLE " + tableName + " (F_id NUMBER, F_text VARCHAR2(30))")
	defer testSes.PrepAndExe("DROP TABLE " + tableName)

	ids, names := mkBenchArrays()
	stmt, err := testSes.Prep("INSERT INTO " + tableName + " (F_id, F_text) VALUES (:1, :2)")
	testErr(err, b)

	b.SetBytes(int64(len(ids)) * 1024)
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for i := range ids {
			_, err = stmt.Exe(ids[i], names[i])
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_MultiInsert(b *testing.B) {
	testSes := getSes(b)
	defer testSes.Close()

	tableName := tableName()
	testSes.PrepAndExe("CREATE TABLE " + tableName + " (F_id NUMBER, F_text VARCHAR2(30))")
	defer testSes.PrepAndExe("DROP TABLE " + tableName)

	ids, names := mkBenchArrays()
	stmt, err := testSes.Prep("INSERT INTO " + tableName + " (F_id, F_text) VALUES (:1, :2)")
	testErr(err, b)

	b.SetBytes(int64(len(ids)) * 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = stmt.Exe(ids, names); err != nil {
			b.Fatal(err)
		}
	}
}

func mkBenchArrays() ([]int64, []string) {
	ids := make([]int64, 1000)
	names := make([]string, len(ids))
	for i := range ids {
		ids[i] = int64(i)
		names[i] = fmt.Sprintf("col%02d/%02d", i, len(ids))
	}
	return ids, names
}
