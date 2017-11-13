//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"

	"gopkg.in/rana/ora.v4"
)

//// string or bool
//charB1     oracleColumnType = "char(1 byte) not null"
//charB1Null oracleColumnType = "char(1 byte) null"
//charC1     oracleColumnType = "char(1 char) not null"
//charC1Null oracleColumnType = "char(1 char) null"

func TestBindDefineBool(t *testing.T) {
	type testCase struct {
		gen func() interface{}
		ct  oracleColumnType
	}
	sc := ora.NewStmtCfg()
	testCases := make(map[string]testCase, 32)
	for _, ctName := range []string{"charB1", "charB1Null"} {
		ct := _T_colType[ctName]
		for _, typName := range []string{"bool", "OraBool", "boolSlice"} {
			for _, valName := range []string{"false", "true"} {
				testCases[fmt.Sprintf("%s_%s_%s", ctName, typName, valName)] = testCase{ct: ct, gen: _T_boolGen[typName+"_"+valName]}
			}
		}
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBindDefine(tc.gen(), tc.ct, t, sc)
		})
	}
}

func TestBindPtrBool(t *testing.T) {
	type testCase struct {
		gen func() interface{}
		ct  oracleColumnType
	}
	testCases := make(map[string]testCase, 16)
	for _, ctName := range []string{"charB1", "charB1Null", "charC1", "charC1Null"} {
		for _, valName := range []string{"false", "true"} {
			k := ctName + "_" + valName
			testCases[k] = testCase{
				gen: _T_boolGen[k],
				ct:  _T_colType[ctName],
			}
		}
	}

	for name, tc := range testCases {
		tc := tc
		if tc.gen == nil {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBindPtr(tc.gen(), tc.ct, t)
		})
	}
}

func TestMultiDefineBool(t *testing.T) {
	for _, ctName := range []string{
		"charB1", "charB1Null",
		"charC1", "charC1Null",
	} {
		gen := _T_boolGen[ctName+"_true"]
		if gen == nil {
			continue
		}
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(
				gen(),
				_T_colType[ctName],
				t,
			)
		})
	}
}

func TestWorkloadBool(t *testing.T) {
	for name, ct := range map[string]oracleColumnType{
		"charB1":     charB1,
		"charB1Null": charB1Null,
		"charC1":     charC1,
		"charC1Null": charC1Null,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//enableLogging(t)
			defer setC1Bool()()
			testWorkload(ct, t)
		})
	}
}

var _T_boolGen = map[string](func() interface{}){
	"bool_false":        func() interface{} { return gen_boolFalse() },
	"bool_true":         func() interface{} { return gen_boolTrue() },
	"OraBool_false":     func() interface{} { return gen_OraBoolFalse(false) },
	"OraBool_true":      func() interface{} { return gen_OraBoolTrue(false) },
	"boolSlice_false":   func() interface{} { return gen_boolSlice() },
	"boolSlice_true":    func() interface{} { return gen_boolSlice() },
	"OraBoolSlice_true": func() interface{} { return gen_OraBoolSlice(false) },
}

func setC1Bool() func() {
	cfg := ora.Cfg()
	old := cfg.Char1()
	cfg.Log.Logger.Infof("setting Char1 from %s to %s.", old, ora.OraB)
	ora.SetCfg(cfg.SetChar1(ora.OraB))
	return func() {
		cfg.Log.Logger.Infof("setting Char1 back from %s to %s.", ora.Cfg().Char1(), old)
		ora.SetCfg(cfg)
	}
}

// Issue89
func TestSelectChar(t *testing.T) {
	t.Parallel()
	tableName := tableName()
	if _, err := testDb.Exec("CREATE TABLE " + tableName + "(c1 CHAR(1), c2 CHAR(4))"); err != nil {
		t.Fatal(err)
	}
	testSes := getSes(t)
	defer testSes.Close()

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
