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

//// bytes
//longRaw     oracleColumnType = "long raw not null"
//longRawNull oracleColumnType = "long raw null"
//raw2000     oracleColumnType = "raw(2000) not null"
//raw2000Null oracleColumnType = "raw(2000) null"
//blob        oracleColumnType = "blob not null"
//blobNull    oracleColumnType = "blob null"
var _T_bytesCols = []string{
	"longRaw", "longRawNull",
	"raw2000", "raw2000Null",
	"blob", "blobNull",
}

func TestBindDefine_bytes(t *testing.T) {
	sc := ora.NewStmtCfg()
	type testCase struct {
		gen func() interface{}
		ct  oracleColumnType
		gct ora.GoColumnType
	}
	testCases := make(map[string]testCase)
	for _, ctName := range _T_bytesCols {
		for _, typName := range []string{
			"bytes", "OraBytes", "OraBytesLob",
			"bytes2000", "OraBytes2000",
		} {
			if strings.HasSuffix(ctName, "Null") && !strings.Contains(typName, "Ora") {
				continue
			}
			gct := ora.Bin
			if strings.Contains(typName, "Ora") {
				gct = ora.OraBin
			}
			testCases[typName+"_"+ctName] = testCase{
				gen: _T_bytesGen[typName],
				ct:  _T_colType[ctName],
				gct: gct,
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
			testBindDefine(tc.gen(), tc.ct, t, sc, tc.gct)
		})
	}
}

func TestBindSlice_bytes(t *testing.T) {
	sc := ora.NewStmtCfg()
	type testCase struct {
		ct  oracleColumnType
		gen func() interface{}
	}
	testCases := make(map[string]testCase)
	for _, typName := range []string{
		"bytesSlice", "OraBytesSlice",
		"bytesSlice2000", "OraBytesSlice2000",
	} {
		for _, ctName := range _T_bytesCols {
			typName := typName
			if strings.HasSuffix(ctName, "Null") {
				typName += "_null"
			}
			testCases[typName+"_"+ctName] = testCase{
				ct:  _T_colType[ctName],
				gen: _T_bytesGen[typName],
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
			testBindDefine(tc.gen(), tc.ct, t, sc)
		})
	}
}

func TestMultiDefine_bytes(t *testing.T) {
	for _, ctName := range _T_bytesCols {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			//enableLogging(t)
			testMultiDefine(gen_bytes(9), _T_colType[ctName], t)
		})
	}
}

func TestWorkload_bytes(t *testing.T) {
	for _, ctName := range []string{"raw2000", "raw2000Null", "blob", "blobNull"} {
		ct := _T_colType[ctName]
		t.Run(ctName, func(t *testing.T) {
			if !strings.Contains(ctName, "lob") {
				t.Parallel()
			}
			//if strings.Contains(ctName, "blob") {
			//enableLogging(t)
			//}
			testWorkload(ct, t)
		})
	}
}

func TestBindDefine_bytes_nil(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, ctName := range []string{"longRawNull", "raw2000Null", "blobNull"} {
		ct := _T_colType[ctName]
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testBindDefine(nil, ct, t, sc)
		})
	}
}

var _T_bytesGen = map[string](func() interface{}){
	"bytes":     func() interface{} { return gen_bytes(9) },
	"bytes2000": func() interface{} { return gen_bytes(2000) },

	"OraBytes":          func() interface{} { return gen_OraBytes(9, false) },
	"OraBytes2000":      func() interface{} { return gen_OraBytes(2000, false) },
	"OraBytes_null":     func() interface{} { return gen_OraBytes(9, true) },
	"OraBytes2000_null": func() interface{} { return gen_OraBytes(2000, true) },
	"OraBytesLob":       func() interface{} { return gen_OraBytesLob(9, false) },

	"bytesSlice":             func() interface{} { return gen_bytesSlice(9) },
	"bytesSlice2000":         func() interface{} { return gen_bytesSlice(2000) },
	"OraBytesSlice":          func() interface{} { return gen_OraBytesSlice(9, false) },
	"OraBytesSlice2000":      func() interface{} { return gen_OraBytesSlice(2000, false) },
	"OraBytesSlice_null":     func() interface{} { return gen_OraBytesSlice(9, true) },
	"OraBytesSlice2000_null": func() interface{} { return gen_OraBytesSlice(2000, true) },
}

//// Do not test workload of multiple Oracle LONG RAW types within the same table because
//// ORA-01754: a table may contain only one column of type LONG
//func TestWorkload_longRaw_session(t *testing.T) {
//	testWorkload(testWorkloadColumnCount, t, longRaw)
//}

//// Do not test workload of multiple Oracle LONG RAW types within the same table because
//// ORA-01754: a table may contain only one column of type LONG
//func TestWorkload_longRawNull_session(t *testing.T) {
//	testWorkload(testWorkloadColumnCount, t, longRawNull)
//}

func TestBindDefine_bytes_blob_size(t *testing.T) {
	sc := ora.NewStmtCfg()
	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)
	ora.SetCfg(cfg.SetLobBufferSize(1024))
	lbs := ora.Cfg().LobBufferSize()
	for _, size := range []int{
		lbs - 1,
		lbs,
		lbs + 1,
		lbs*3 - 1,
		lbs * 3,
		lbs*3 + 1,
	} {
		t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
			testBindDefine(gen_bytes(size), blob, t, sc, ora.Bin)
			testBindDefine(gen_OraBytesLob(size, false), blob, t, sc, ora.Bin)
			lob := gen_OraBytesLob(size, false)
			testBindDefine(&lob, blob, t, sc, ora.Bin)
		})
	}
}
