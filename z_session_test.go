// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	"gopkg.in/rana/ora.v3"
)

func Test_open_cursors(t *testing.T) {
	// This needs "GRANT SELECT ANY DICTIONARY TO test"
	// or at least "GRANT SELECT ON v_$mystat TO test".
	// use 'opened cursors current' STATISTIC#=5 to determine open cursors
	// SELECT A.STATISTIC#, A.NAME, B.VALUE
	// FROM V$STATNAME A, V$MYSTAT B
	// WHERE A.STATISTIC# = B.STATISTIC#
	//enableLogging(t)
	env, err := ora.OpenEnv(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Close()

	rset, err := ses.PrepAndQry("SELECT VALUE FROM V$MYSTAT WHERE STATISTIC#=5")
	if err != nil {
		t.Fatal(err)
	}
	before := rset.NextRow()[0].(float64)
	rounds := 100
	for i := 0; i < rounds; i++ {
		func() {
			stmt, err := ses.Prep("SELECT 1 FROM user_objects WHERE ROWNUM < 100")
			if err != nil {
				t.Fatal(err)
			}
			defer stmt.Close()
			rset, err := stmt.Qry()
			if err != nil {
				t.Errorf("SELECT: %v", err)
				return
			}
			j := 0
			for rset.Next() {
				j++
			}
			//t.Logf("%d objects, error=%v", j, rset.Err)
		}()
	}
	rset, err = ses.PrepAndQry("SELECT VALUE FROM V$MYSTAT WHERE STATISTIC#=5")
	if err != nil {
		t.Fatal(err)
	}
	after := rset.NextRow()[0].(float64)
	if after-before >= float64(rounds) {
		t.Errorf("before=%d after=%d, awaited less than %d increment!", before, after, rounds)
		return
	}
	//t.Logf("before=%d after=%d", before, after)
}

func TestSession_PrepCloseStmt(t *testing.T) {

	// setup
	env, err := ora.OpenEnv(nil)
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSes(testSesCfg)
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
	_, err = stmt.Exe()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)

	err = tx.Commit()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Qry()
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
	_, err = stmt.Exe()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)

	err = tx.Rollback()
	testErr(err, t)

	stmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Qry()
	testErr(err, t)

	for rset.Next() {
	}
	if 0 != rset.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 0, rset.Len())
	}
}

func TestSession_PrepAndExe(t *testing.T) {
	rowsAffected, err := testSes.PrepAndExe(fmt.Sprintf("create table %v (c1 number)", tableName()))
	testErr(err, t)

	if rowsAffected != 0 {
		t.Fatalf("expected(%v), actual(%v)", 0, rowsAffected)
	}
}

func TestSession_PrepAndExe_Insert(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	rowsAffected, err := testSes.PrepAndExe(fmt.Sprintf("INSERT INTO %v (C1) VALUES (:C1)", tableName), values)
	testErr(err, t)

	if rowsAffected != 1000000 {
		t.Fatalf("expected(%v), actual(%v)", 1000000, rowsAffected)
	}
}

func TestSession_PrepAndQry(t *testing.T) {
	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	// insert one row
	stmtIns, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmtIns.Exe()
	testErr(err, t)

	rset, err := testSes.PrepAndQry(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)
	if rset == nil {
		t.Fatalf("expected non-nil rset")
	}

	row := rset.NextRow()
	if row[0] == 9 {
		t.Fatalf("expected(%v), actual(%v)", 9, row[0])
	}
}
