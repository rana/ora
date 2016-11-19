//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v3"
)

//// string or bool
//charB1     oracleColumnType = "char(1 byte) not null"
//charB1Null oracleColumnType = "char(1 byte) null"
//charC1     oracleColumnType = "char(1 char) not null"
//charC1Null oracleColumnType = "char(1 char) null"

//////////////////////////////////////////////////////////////////////////////////
//// charB1
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bool_charB1_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolFalse(), charB1, t, nil)
}

func TestBindDefine_bool_charB1_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolTrue(), charB1, t, nil)
}

func TestBindPtr_bool_charB1_false_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolFalse(), charB1, t)
}

func TestBindPtr_bool_charB1_true_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolTrue(), charB1, t)
}

func TestBindDefine_OraBool_charB1_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolFalse(false), charB1, t, nil)
}

func TestBindDefine_OraBool_charB1_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolTrue(false), charB1, t, nil)
}

func TestBindSlice_bool_charB1_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolSlice(), charB1, t, nil)
}

func TestBindSlice_OraBool_charB1_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolSlice(false), charB1, t, nil)
}

func TestMultiDefine_charB1_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_boolTrue(), charB1, t)
}

func TestWorkload_charB1_session(t *testing.T) {
	t.Parallel()
	//enableLogging(t)
	defer setC1Bool()()
	testWorkload(charB1, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// charB1Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bool_charB1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolFalse(), charB1Null, t, nil)
}

func TestBindDefine_bool_charB1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolTrue(), charB1Null, t, nil)
}

func TestBindPtr_bool_charB1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolFalse(), charB1Null, t)
}

func TestBindPtr_bool_charB1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolTrue(), charB1Null, t)
}

func TestBindDefine_OraBool_charB1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolFalse(true), charB1Null, t, nil)
}

func TestBindDefine_OraBool_charB1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolTrue(true), charB1Null, t, nil)
}

func TestBindSlice_bool_charB1Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolSlice(), charB1Null, t, nil)
}

func TestBindSlice_OraBool_charB1Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolSlice(true), charB1Null, t, nil)
}

func TestMultiDefine_charB1Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_boolTrue(), charB1Null, t)
}

func TestWorkload_charB1Null_session(t *testing.T) {
	t.Parallel()
	//enableLogging(t)
	defer setC1Bool()()
	testWorkload(charB1Null, t)
}

func TestBindDefine_charB1Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, charB1Null, t, nil)
}

//////////////////////////////////////////////////////////////////////////////////
//// charC1
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bool_charC1_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolFalse(), charC1, t, nil)
}

func TestBindDefine_bool_charC1_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolTrue(), charC1, t, nil)
}

func TestBindPtr_bool_charC1_false_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolFalse(), charC1, t)
}

func TestBindPtr_bool_charC1_true_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolTrue(), charC1, t)
}

func TestBindDefine_OraBool_charC1_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolFalse(false), charC1, t, nil)
}

func TestBindDefine_OraBool_charC1_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolTrue(false), charC1, t, nil)
}

func TestBindSlice_bool_charC1_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolSlice(), charC1, t, nil)
}

func TestBindSlice_OraBool_charC1_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolSlice(false), charC1, t, nil)
}

func TestMultiDefine_charC1_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_boolTrue(), charC1, t)
}

func TestWorkload_charC1_session(t *testing.T) {
	t.Parallel()
	//enableLogging(t)
	testWorkload(charC1, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// charC1Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bool_charC1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolFalse(), charC1Null, t, nil)
}

func TestBindDefine_bool_charC1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_boolTrue(), charC1Null, t, nil)
}

func TestBindPtr_bool_charC1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolFalse(), charC1Null, t)
}

func TestBindPtr_bool_charC1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_boolTrue(), charC1Null, t)
}

func TestBindDefine_OraBool_charC1Null_false_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolFalse(true), charC1Null, t, nil)
}

func TestBindDefine_OraBool_charC1Null_true_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolTrue(true), charC1Null, t, nil)
}

func TestBindSlice_bool_charC1Null_session(t *testing.T) {
	t.Parallel()
	//ora.Cfg().Log.Rset.Next = true
	//ora.Cfg().Log.Rset.BeginRow = true
	//enableLogging(t)
	testBindDefine(gen_boolSlice(), charC1Null, t, nil)
}

func TestBindSlice_OraBool_charC1Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraBoolSlice(true), charC1Null, t, nil)
}

func TestMultiDefine_charC1Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_boolTrue(), charC1Null, t)
}

func TestWorkload_charC1Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(charC1Null, t)
}

func TestBindDefine_charC1Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, charC1Null, t, nil)
}

func setC1Bool() func() {
	old := ora.Cfg().Env.StmtCfg.Rset.Char1()
	ora.Cfg().Log.Logger.Infof("setting Char1 from %s to %s.", old, ora.OraB)
	ora.Cfg().Env.StmtCfg.Rset.SetChar1(ora.OraB)
	return func() {
		ora.Cfg().Log.Logger.Infof("setting Char1 back from %s to %s.", ora.Cfg().Env.StmtCfg.Rset.Char1(), old)
		ora.Cfg().Env.StmtCfg.Rset.SetChar1(old)
	}
}

// Issue89
func TestSelectChar(t *testing.T) {
	t.Parallel()
	tableName := tableName()
	if _, err := testDb.Exec("CREATE TABLE " + tableName + "(c1 CHAR(1), c2 CHAR(4))"); err != nil {
		t.Fatal(err)
	}
	if _, err := testSes.PrepAndExe("INSERT INTO "+tableName+" VALUES (:1, :2)",
		"A", "ABCD"); err != nil {
		t.Fatal(err)
	}
	got := make([]interface{}, 0, 2)
	for tN, tC := range []struct {
		colDefs []ora.GoColumnType
		want    []interface{}
	}{
		{[]ora.GoColumnType{ora.B, ora.B}, []interface{}{false, false}},
		{[]ora.GoColumnType{ora.S, ora.S}, []interface{}{"A", "ABCD"}},
		{nil, []interface{}{"A", "ABCD"}},
	} {
		stmt, err := testSes.Prep("SELECT c1, c2 FROM "+tableName, tC.colDefs...)
		if err != nil {
			t.Fatal(err)
		}
		defer stmt.Close()
		rset, err := stmt.Qry()
		if err != nil {
			t.Fatal(err)
		}
		got = got[:0]
		rset.Next()
		got = append(got, rset.Row[0], rset.Row[1])
		t.Logf("%d. got %q, want %q.", tN, got, tC.want)
		if len(got) != len(tC.want) || got[0] != tC.want[0] || got[1] != tC.want[1] {
			t.Errorf("%d. got %q, want %q.", tN, got, tC.want)
		}
	}
}
