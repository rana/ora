//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/rana/ora.v4"
)

//// numeric
//numberP38S0Identity oracleColumnType = "number(38,0) generated always as identity (start with 1 increment by 1)"
//numberP38S0         oracleColumnType = "number(38,0) not null"
//numberP38S0Null     oracleColumnType = "number(38,0) null"
//numberP16S15        oracleColumnType = "number(16,15) not null"
//numberP16S15Null    oracleColumnType = "number(16,15) null"
//binaryDouble        oracleColumnType = "binary_double not null"
//binaryDoubleNull    oracleColumnType = "binary_double null"
//binaryFloat         oracleColumnType = "binary_float not null"
//binaryFloatNull     oracleColumnType = "binary_float null"
//floatP126           oracleColumnType = "float(126) not null"
//floatP126Null       oracleColumnType = "float(126) null"

var _T_numericGen = map[string](func() interface{}){
	"int64":        func() interface{} { return gen_int64() },
	"int32":        func() interface{} { return gen_int32() },
	"int16":        func() interface{} { return gen_int16() },
	"int8":         func() interface{} { return gen_int8() },
	"OraInt64":     func() interface{} { return gen_OraInt64(false) },
	"OraInt32":     func() interface{} { return gen_OraInt32(false) },
	"OraInt16":     func() interface{} { return gen_OraInt16(false) },
	"OraInt8":      func() interface{} { return gen_OraInt8(false) },
	"OraInt64Null": func() interface{} { return gen_OraInt64(true) },
	"OraInt32Null": func() interface{} { return gen_OraInt32(true) },
	"OraInt16Null": func() interface{} { return gen_OraInt16(true) },
	"OraInt8Null":  func() interface{} { return gen_OraInt8(true) },

	"uint64":        func() interface{} { return gen_uint64() },
	"uint32":        func() interface{} { return gen_uint32() },
	"uint16":        func() interface{} { return gen_uint16() },
	"uint8":         func() interface{} { return gen_uint8() },
	"OraUint64":     func() interface{} { return gen_OraUint64(false) },
	"OraUint32":     func() interface{} { return gen_OraUint32(false) },
	"OraUint16":     func() interface{} { return gen_OraUint16(false) },
	"OraUint8":      func() interface{} { return gen_OraUint8(false) },
	"OraUint64Null": func() interface{} { return gen_OraUint64(true) },
	"OraUint32Null": func() interface{} { return gen_OraUint32(true) },
	"OraUint16Null": func() interface{} { return gen_OraUint16(true) },
	"OraUint8Null":  func() interface{} { return gen_OraUint8(true) },

	"float64":             func() interface{} { return gen_float64() },
	"float32":             func() interface{} { return gen_float32() },
	"float64Trunc":        func() interface{} { return gen_float64Trunc() },
	"float32Trunc":        func() interface{} { return gen_float32Trunc() },
	"OraFloat64Trunc":     func() interface{} { return gen_OraFloat64Trunc(false) },
	"OraFloat32Trunc":     func() interface{} { return gen_OraFloat32Trunc(false) },
	"OraFloat64TruncNull": func() interface{} { return gen_OraFloat64Trunc(true) },
	"OraFloat32TruncNull": func() interface{} { return gen_OraFloat32Trunc(true) },

	"numString":      func() interface{} { return gen_NumString() },
	"numStringTrunc": func() interface{} { return gen_NumStringTrunc() },
}

var _T_numericCols = []string{
	"numberP38S0", "numberP38S0Null",
	"numberP16S15", "numberP16S15Null",
	"binaryDouble", "binaryDoubleNull",
	"binaryFloat", "binaryFloatNull",
	"floatP126", "floatP126Null",
}

func TestBindNumericString(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	tableName, err := createTable(1, numberP38S0, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	stmt, err := testSes.Prep("INSERT INTO " + tableName + " (c1) VALUES (:1)")
	testErr(err, t)
	defer stmt.Close()
	s := ""
	for i := 1; i < 38; i++ {
		s = s + strconv.Itoa(i%10)
		if _, err := stmt.Exe(s); err != nil {
			t.Errorf("Insert %q: %v", s, err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE numberP38S0
// BIND PTR numberP38S0
// BIND DEFINE VALUE numberP38S0Null
// BIND DEFINE VALUE numberP16S15
// BIND PTR numberP16S15
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_numeric(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, ctName := range _T_numericCols {
		ct := _T_colType[ctName]
		for valName, gen := range _T_numericGen {
			if gen == nil {
				continue
			}
			gen := gen
			if !strings.Contains(valName, "int") && !strings.Contains(valName, "Int") && !strings.HasSuffix(valName, "Trunc") {
				continue
			}
			if strings.HasSuffix(valName, "Null") && !strings.HasSuffix(ctName, "Null") {
				continue
			}
			//t.Logf("v=%T (%#v)", v, gen)
			t.Run(valName+"_"+ctName, func(t *testing.T) {
				t.Parallel()
				v := gen()
				testBindDefine(v, ct, t, sc)
			})
			vName := fmt.Sprintf("%T", gen())
			if len(vName) >= 3 && strings.EqualFold(vName[:3], "ora") {
				continue
			}
			t.Run(valName+"_"+ctName+"Ptr", func(t *testing.T) {
				t.Parallel()
				v := gen()
				//enableLogging(t)
				testBindPtr(v, ct, t)
			})
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE numberP38S0
// BIND SLICE numberP38S0Null
// BIND SLICE numberP16S15
// BIND SLICE numberP16S15Null
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_numeric(t *testing.T) {
	sc := ora.NewStmtCfg()

	generators := map[string](func() interface{}){
		"int64Slice":        func() interface{} { return gen_int64Slice() },
		"int32Slice":        func() interface{} { return gen_int32Slice() },
		"int16Slice":        func() interface{} { return gen_int16Slice() },
		"int8Slice":         func() interface{} { return gen_int8Slice() },
		"OraInt64Slice":     func() interface{} { return gen_OraInt64Slice(false) },
		"OraInt32Slice":     func() interface{} { return gen_OraInt32Slice(false) },
		"OraInt16Slice":     func() interface{} { return gen_OraInt16Slice(false) },
		"OraInt8Slice":      func() interface{} { return gen_OraInt8Slice(false) },
		"OraInt64SliceNull": func() interface{} { return gen_OraInt64Slice(true) },
		"OraInt32SliceNull": func() interface{} { return gen_OraInt32Slice(true) },
		"OraInt16SliceNull": func() interface{} { return gen_OraInt16Slice(true) },
		"OraInt8SliceNull":  func() interface{} { return gen_OraInt8Slice(true) },

		"uint64Slice":        func() interface{} { return gen_uint64Slice() },
		"uint32Slice":        func() interface{} { return gen_uint32Slice() },
		"uint16Slice":        func() interface{} { return gen_uint16Slice() },
		"uint8Slice":         func() interface{} { return gen_uint8Slice() },
		"OraUint64Slice":     func() interface{} { return gen_OraUint64Slice(false) },
		"OraUint32Slice":     func() interface{} { return gen_OraUint32Slice(false) },
		"OraUint16Slice":     func() interface{} { return gen_OraUint16Slice(false) },
		"OraUint8Slice":      func() interface{} { return gen_OraUint8Slice(false) },
		"OraUint64SliceNull": func() interface{} { return gen_OraUint64Slice(true) },
		"OraUint32SliceNull": func() interface{} { return gen_OraUint32Slice(true) },
		"OraUint16SliceNull": func() interface{} { return gen_OraUint16Slice(true) },
		"OraUint8SliceNull":  func() interface{} { return gen_OraUint8Slice(true) },

		"float64TruncSlice":        func() interface{} { return gen_float64TruncSlice() },
		"float32TruncSlice":        func() interface{} { return gen_float32TruncSlice() },
		"OraFloat64TruncSlice":     func() interface{} { return gen_OraFloat64TruncSlice(false) },
		"OraFloat32TruncSlice":     func() interface{} { return gen_OraFloat32TruncSlice(false) },
		"OraFloat64TruncSliceNull": func() interface{} { return gen_OraFloat64TruncSlice(true) },
		"OraFloat32TruncSliceNull": func() interface{} { return gen_OraFloat32TruncSlice(true) },

		"numStringTruncSlice": func() interface{} { return gen_NumStringTruncSlice() },
	}

	valName := "float32TruncSlice"
	enableLogging(t)
	testBindDefine(generators[valName](), _T_colType["numberP16S15"], t, sc)

	for valName, gen := range generators {
		valName := valName
		if gen == nil {
			continue
		}
		gen := gen
		for _, ctName := range _T_numericCols {
			if strings.HasSuffix(valName, "Null") && !strings.HasSuffix(ctName, "Null") {
				continue
			}
			ct := _T_colType[ctName]
			t.Run(valName+"_"+ctName, func(t *testing.T) {
				t.Parallel()
				v := gen()
				sc := sc
				if valName == "uint8Slice" {
					sc = sc.SetByteSlice(ora.U8)
				}
				testBindDefine(v, ct, t, sc)
			})
		}
	}
}

func TestMultiDefine_numeric(t *testing.T) {
	for _, ctName := range _T_numericCols {
		ct := _T_colType[ctName]
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(gen_int64(), ct, t)
		})
	}
}

func TestWorkload_numeric(t *testing.T) {
	for _, ctName := range _T_numericCols {
		ct := _T_colType[ctName]
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testWorkload(ct, t)
		})
	}
}

func TestBindDefine_numeric_nil(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, ctName := range _T_numericCols {
		if !strings.HasSuffix(ctName, "Null") {
			continue
		}
		ct := _T_colType[ctName]
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testBindDefine(nil, ct, t, sc)
		})
	}
}

func TestIssue144(t *testing.T) {
	oCfg := ora.Cfg()
	defer ora.SetCfg(oCfg)
	ora.SetCfg(ora.Cfg().SetNumberFloat(ora.S).SetFloat(ora.S))

	var s string
	var i int64
	var f float64
	for tN, dest := range []interface{}{
		&f,
		&i,
		&s,
	} {
		if err := testDb.QueryRow("SELECT 123456 FROM DUAL").Scan(dest); err != nil {
			t.Errorf("%d. %T: %v", tN, dest, err)
		}
		t.Logf("%d. %T: %v", tN, dest, reflect.ValueOf(dest).Elem().Interface())
	}
}
