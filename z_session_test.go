// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"gopkg.in/rana/ora.v4"
)

func TestIssue233(t *testing.T) {
	session, err := testSesPool.Get()
	if err != nil {
		t.Fatal(err)
	}

	session.PrepAndExe("DROP TABLE test_lob")
	for _, qry := range []string{
		"create table test_lob (c clob)",
		`begin
	  insert into test_lob values ('Hello');
	  insert into test_lob values ('world!');
	  commit;
	end;`,
		`CREATE OR REPLACE PROCEDURE sp_lob_test(o_cur_lob OUT SYS_REFCURSOR,
                                            o_cur_date OUT SYS_REFCURSOR) AS
BEGIN
  OPEN o_cur_lob FOR
    SELECT * FROM test_lob;

  OPEN o_cur_date FOR
    select sysdate from dual;
end sp_lob_test;`,
	} {
		if _, err = session.PrepAndExe(qry); err != nil {
			t.Fatal(errors.Wrap(err, qry))
		}
	}
	session.Close()

	qry := "call sp_lob_test(:o_cur_lob, :o_cur_date)"
	for i := 0; i < 1000000; i++ {
		session, err := testSesPool.Get()
		if err != nil {
			t.Fatal(err)
		}
		stmt, err := session.Prep(qry)
		if err != nil {
			t.Fatal(errors.Wrap(err, qry))

		}
		c1 := &ora.Rset{}
		c2 := &ora.Rset{}
		_, err = stmt.Exe(c1, c2)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("iteration #%d\n", i)
		fmt.Println(c1)
		c1.Exhaust()
		fmt.Println(c2)
		c2.Exhaust()

		stmt.Close()
		session.Close()
	}
}

func Test_open_cursors(t *testing.T) {
	t.Parallel()
	// This needs "GRANT SELECT ANY DICTIONARY TO test"
	// or at least "GRANT SELECT ON v_$mystat TO test".
	// use 'opened cursors current' STATISTIC#=5 to determine open cursors
	// SELECT A.STATISTIC#, A.NAME, B.VALUE
	// FROM V$STATNAME A, V$MYSTAT B
	// WHERE A.STATISTIC# = B.STATISTIC#
	//enableLogging(t)
	env, err := ora.OpenEnv()
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

	toNum := func(a interface{}) int {
		switch x := a.(type) {
		case int64:
			return int(x)
		case float64:
			return int(x)
		default:
			i, err := strconv.Atoi(fmt.Sprintf("%v", a))
			if err != nil {
				panic(err)
			}
			return i
		}
	}

	//qry := `SELECT VALUE FROM V$MYSTAT WHERE STATISTIC#=5`
	qry := `SELECT count(0) FROM v$open_cursor WHERE user_name = user AND cursor_type = 'OPEN'`
	//qry := `SELECT VALUE FROM v$sesstat WHERE statistic#=5 AND SID = sys_context('USERENV', 'SID')`
	countStmt, err := ses.Prep(qry)
	if err != nil {
		t.Fatal(errors.Wrap(err, qry))
	}
	defer countStmt.Close()
	count := func() int {
		rset, err := countStmt.Qry(qry)
		if err != nil {
			t.Skipf("%q: %v", qry, err)
		}
		return toNum(rset.NextRow()[0])
	}

	before := count()
	rounds := 2000
	if cgocheck() != 0 {
		rounds = 10
	}
	for i := 0; i < rounds; i++ {
		func() {
			stmt, err := ses.Prep("SELECT 1 FROM user_objects WHERE ROWNUM < 100")
			if err != nil {
				t.Fatal(err)
			}
			defer stmt.Close()
			//t.Logf("%d. in: %d", i, count())
			rset, err := stmt.Qry()
			if err != nil {
				t.Errorf("SELECT: %v", err)
				return
			}
			j := 0
			for rset.Next() {
				j++
			}
			t.Logf("%d objects, error=%v", j, rset.Err())
		}()
	}
	after := count()
	if after-before >= rounds {
		t.Errorf("before=%v after=%v, awaited less than %d increment!", before, after, rounds)
		return
	}
	//t.Logf("before=%d after=%d", before, after)
}

func TestSession_PrepCloseStmt(t *testing.T) {
	t.Parallel()

	// setup
	env, err := ora.OpenEnv()
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
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	tableName, err := createTable(1, numberP38S0, ses)
	testErr(err, t)
	defer dropTable(tableName, ses, t)

	defer ses.Close()
	tx, err := ses.StartTx()
	testErr(err, t)

	stmt, err := ses.Prep(fmt.Sprintf("insert into %v (c1) values (:1)", tableName))
	testErr(err, t)
	_, err = stmt.Exe(int64(9))
	testErr(err, t)
	_, err = stmt.Exe(int64(11))
	testErr(err, t)

	err = tx.Commit()
	testErr(err, t)

	stmt, err = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Qry()
	testErr(err, t)

	for rset.Next() {
		t.Logf("Row=%v", rset.Row)
	}
	if 2 != rset.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 2, rset.Len())
	}
}

func TestSession_Tx_StartRollback(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	tableName, err := createTable(1, numberP38S0, ses)
	testErr(err, t)
	defer dropTable(tableName, ses, t)

	cfg := ora.Cfg()
	cfg.Log.Tx.Commit, cfg.Log.Tx.Rollback = true, true
	ora.SetCfg(cfg)

	tx, err := ses.StartTx()
	testErr(err, t)

	enableLogging(t)
	stmt, err := ses.Prep(fmt.Sprintf("insert into %v (c1) values (:1)", tableName))
	testErr(err, t)
	_, err = stmt.Exe(int64(9))
	testErr(err, t)
	_, err = stmt.Exe(int64(11))
	testErr(err, t)

	err = tx.Rollback()
	testErr(err, t)

	stmt, err = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)

	rset, err := stmt.Qry()
	testErr(err, t)

	if 0 != rset.Len() {
		t.Fatalf("row count BEFORE execute: expected(%v), actual(%v)", 0, rset.Len())
	}
	for rset.Next() {
		t.Logf("Row=%v", rset.Row)
	}
	if 0 != rset.Len() {
		t.Fatalf("row count: expected(%v), actual(%v)", 0, rset.Len())
	}
}

func TestSession_PrepAndExe(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	rowsAffected, err := ses.PrepAndExe(fmt.Sprintf("create table %v (c1 number)", tableName()))
	testErr(err, t)

	if rowsAffected != 0 {
		t.Fatalf("expected(%v), actual(%v)", 0, rowsAffected)
	}
}

func TestSession_PrepAndExe_Insert(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	tableName, err := createTable(1, numberP38S0, ses)
	testErr(err, t)
	defer dropTable(tableName, ses, t)

	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}

	if cgc := cgocheck(); cgc > 0 && os.Getenv("NO_CGOCHECK_CHECK") != "1" {
		values = values[:2000]
		t.Logf("GODEBUG=%d so limiting slice to %d", cgc, len(values))
	}
	rowsAffected, err := ses.PrepAndExe(fmt.Sprintf("INSERT INTO %v (C1) VALUES (:C1)", tableName), values)
	testErr(err, t)

	if rowsAffected != uint64(len(values)) {
		t.Fatalf("expected(%v), actual(%v)", len(values), rowsAffected)
	}
}

func TestSession_PrepAndQry(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	tableName, err := createTable(1, numberP38S0, ses)
	testErr(err, t)
	defer dropTable(tableName, ses, t)

	// insert one row
	stmtIns, err := ses.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	testErr(err, t)
	_, err = stmtIns.Exe()
	testErr(err, t)

	rset, err := ses.PrepAndQry(fmt.Sprintf("select c1 from %v", tableName))
	testErr(err, t)
	if rset == nil {
		t.Fatalf("expected non-nil rset")
	}

	row := rset.NextRow()
	if row[0] == 9 {
		t.Fatalf("expected(%v), actual(%v)", 9, row[0])
	}
}

var _cgocheck int = 1

func cgocheck() int {
	return _cgocheck
}
func init() {
	gdbg := os.Getenv("GODEBUG")
	if gdbg != "" {
		for _, part := range strings.Split(gdbg, ",") {
			if strings.HasPrefix(part, "cgocheck=") {
				n, err := strconv.Atoi(part[9:])
				if err != nil {
					panic(err)
				}
				_cgocheck = n
				break
			}
		}
	}
}

func BenchmarkSession_PrepAndExe_Insert_WithCGOCheck(b *testing.B) {
	if cgocheck() == 0 {
		b.SkipNow()
	}
	benchmarkSession_PrepAndExe_Insert(b)
}
func BenchmarkSession_PrepAndExe_Insert_WithoutCGOCheck(b *testing.B) {
	if cgocheck() != 0 {
		b.SkipNow()
	}
	benchmarkSession_PrepAndExe_Insert(b)
}

func benchmarkSession_PrepAndExe_Insert(b *testing.B) {
	testSes := getSes(b)
	defer testSes.Close()

	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, b)
	defer dropTable(tableName, testSes, b)

	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	b.ResetTimer()
	const batchLen = 100
	for i := 0; i < b.N; i++ {
		rowsAffected, err := testSes.PrepAndExe(fmt.Sprintf("INSERT INTO %v (C1) VALUES (:C1)", tableName),
			values[i*batchLen:(i+1)*batchLen])
		if err != nil {
			b.Error(err)
			break
		}
		if rowsAffected != batchLen {
			b.Fatalf("expected(%v), actual(%v)", batchLen, rowsAffected)
		}
	}
}

func TestSessionCallPkg(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	if _, err := ses.PrepAndExe(`CREATE OR REPLACE PACKAGE mypkg AS
  FUNCTION myproc(user IN VARCHAR2, pass IN VARCHAR2) RETURN PLS_INTEGER;
END mypkg;`); err != nil {
		t.Fatal(err)
	}
	if _, err := ses.PrepAndExe(`CREATE OR REPLACE PACKAGE BODY mypkg AS
  FUNCTION myproc(user IN VARCHAR2, pass IN VARCHAR2) RETURN PLS_INTEGER IS
  BEGIN
    RETURN NVL(LENGTH(user), 0) + NVL(LENGTH(pass), 0);
  END myproc;
END mypkg;`); err != nil {
		t.Fatal(err)
	}
	rc := int64(-100)
	if _, err := ses.PrepAndExe("BEGIN :1 := MYPKG.MYPROC(:2, :3); END;", &rc, "a", "bc"); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d", rc)
	if rc != 3 {
		t.Errorf("got %d, awaited %d.", rc, 3)
	}
}

func TestIssue59(t *testing.T) {
	t.Parallel()
	ses, err := testSesPool.Get()
	testErr(err, t)
	defer ses.Close()

	if _, err := ses.PrepAndExe(`CREATE OR REPLACE
PROCEDURE test_59(theoutput OUT VARCHAR2, param1 IN VARCHAR2, param2 IN VARCHAR2, param3 IN VARCHAR2) IS
  TYPE vc_tab_typ IS TABLE OF VARCHAR2(32767) INDEX BY PLS_INTEGER;
  rows vc_tab_typ;
  res VARCHAR2(32767);
BEGIN
  SELECT ROWNUM||';'||A.object_name||';'||B.object_type||';'||param1||';'||param2||';'||param3
    BULK COLLECT INTO rows
    FROM user_objects B, user_objects A, (SELECT 1 FROM DUAL)
	WHERE ROWNUM < 1000;
  FOR i IN 1..rows.COUNT LOOP
    res := SUBSTR(res||CHR(10)||rows(i), 1, 32767);
    EXIT WHEN LENGTH(res) >= 32767;
  END LOOP;
  theoutput := SUBSTR(res, 1, 2000);
END test_59;`,
	); err != nil {
		t.Fatal(err)
	}
	ces, err := GetCompileErrors(ses, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ces) > 0 {
		for _, ce := range ces {
			t.Error(ce)
		}
	}

	res := strings.Repeat("\x00", 32768)
	if _, err := ses.PrepAndExe("CALL test_59(:1, :2, :3, :4)", &res, "a", "b", "c"); err != nil {
		t.Error(err)
	}
	t.Logf("res=%q", res)
}

// CompileError represents a compile-time error as in user_errors view.
type CompileError struct {
	Owner, Name, Type    string
	Line, Position, Code int64
	Text                 string
	Warning              bool
}

func (ce CompileError) Error() string {
	prefix := "ERROR "
	if ce.Warning {
		prefix = "WARN  "
	}
	return fmt.Sprintf("%s %s.%s %s %d:%d [%d] %s",
		prefix, ce.Owner, ce.Name, ce.Type, ce.Line, ce.Position, ce.Code, ce.Text)
}

// GetCompileErrors returns the slice of the errors in user_errors.
//
// If all is false, only errors are returned; otherwise, warnings, too.
func GetCompileErrors(ses *ora.Ses, all bool) ([]CompileError, error) {
	rows, err := ses.PrepAndQry(`
	SELECT USER owner, name, type, line, position, message_number, text, attribute
		FROM user_errors
		ORDER BY name, sequence`)
	if err != nil {
		return nil, err
	}
	var errors []CompileError
	var warn string
	for rows.Next() {
		var ce CompileError
		ce.Owner, ce.Name, ce.Type,
			ce.Line, ce.Position, ce.Code,
			ce.Text, warn =
			rows.Row[0].(string), rows.Row[1].(string), rows.Row[2].(string),
			int64(rows.Row[3].(float64)), int64(rows.Row[4].(float64)), int64(rows.Row[5].(float64)),
			rows.Row[6].(string), rows.Row[7].(string)
		ce.Warning = warn == "WARNING"
		if !ce.Warning || all {
			errors = append(errors, ce)
		}
	}
	return errors, rows.Err()
}
