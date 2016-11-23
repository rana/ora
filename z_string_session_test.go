//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/rana/ora.v4"
)

var _T_stringGen = map[string](func() interface{}){
	"string48":        func() interface{} { return gen_string48() },
	"OraString48":     func() interface{} { return gen_OraString48(false) },
	"OraString48Null": func() interface{} { return gen_OraString48(true) },
}

var _T_stringCols = []string{
	"charB48", "charB48Null",
	"charC48", "charC48Null",
	"nchar48", "nchar48Null",
	"varcharB48", "varcharB48Null",
	"varcharC48", "varcharC48Null",
	"varchar2B48", "varchar2B48Null",
	"varchar2C48", "varchar2C48Null",
	"nvarchar248", "nvarchar248Null",
}

func TestBindDefine_string(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, ctName := range _T_stringCols {
		for valName, gen := range _T_stringGen {
			t.Run(fmt.Sprintf("%s_%s", valName, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

func TestBindSlice_string(t *testing.T) {
	sc := ora.NewStmtCfg()
	for valName, gen := range map[string](func() interface{}){
		"stringSlice48":        func() interface{} { return gen_stringSlice48() },
		"OraStringSlice48":     func() interface{} { return gen_OraStringSlice48(false) },
		"OraStringSlice48Null": func() interface{} { return gen_OraStringSlice48(true) },
	} {
		for _, ctName := range _T_stringCols {
			t.Run(fmt.Sprintf("%s_%s", valName, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

func TestMultiDefine_string(t *testing.T) {
	for _, ctName := range _T_stringCols {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(gen_string48(), _T_colType[ctName], t)
		})
	}
}

func TestWorkload_charB48_session(t *testing.T) {
	for _, ctName := range _T_stringCols {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testWorkload(_T_colType[ctName], t)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// long
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_long(t *testing.T) {
	sc := ora.NewStmtCfg()
	for valName, gen := range map[string](func() interface{}){
		"string":             func() interface{} { return gen_string() },
		"stringSlice":        func() interface{} { return gen_stringSlice() },
		"OraString":          func() interface{} { return gen_OraString(false) },
		"OraStringSlice":     func() interface{} { return gen_OraString(false) },
		"OraStringNull":      func() interface{} { return gen_OraString(true) },
		"OraStringSliceNull": func() interface{} { return gen_OraString(true) },
	} {
		for _, ctName := range []string{
			"long", "longNull",
			"clob", "clobNull",
			"nclob", "nclobNull",
		} {
			if strings.HasSuffix(valName, "Null") && !strings.HasSuffix(ctName, "Null") {
				continue
			}
			t.Run(valName+"_"+ctName, func(t *testing.T) {
				if !strings.Contains(ctName, "lob") {
					t.Parallel()
				}
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

//func TestBindPtr_string_long_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), long, t)
//}

func TestMultiDefine_long_session(t *testing.T) {
	for _, ctName := range []string{
		"long", "longNull",
		"clob", "clobNull",
		"nclob", "nclobNull",
	} {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(gen_string(), _T_colType[ctName], t)
		})
	}
}

//func TestWorkload_long_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(long, t)
//}

//func TestBindPtr_string_longNull_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), longNull, t)
//}

//func TestWorkload_longNull_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(longNull, t)
//}
