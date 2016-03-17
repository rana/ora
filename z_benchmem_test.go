// Copyright 2016 Tamás Gulácsi, Valentin Kuznetsov. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"runtime"
	"testing"

	"gopkg.in/rana/ora.v3"
)

// BenchmarkMemory usage for querying rows.
//
// go test -c && ./ora.v3.test -test.run=^$ -test.bench=Memory -test.memprofilerate=1 -test.memprofile=/tmp/mem.prof && go tool pprof --alloc_space ora.v3.test /tmp/mem.prof
func TestMemoryNumString(t *testing.T) {
	benchMem(t, 1794456, `SELECT
		TO_NUMBER('123456789012345678') bn01
		, TO_NUMBER('223456789012345678') bn02
		, TO_NUMBER('323456789012345678') bn03
		, TO_NUMBER('423456789012345678') bn04
		, TO_NUMBER('523456789012345678') bn05
		, TO_NUMBER('623456789012345678') bn06
		, TO_NUMBER('723456789012345678') bn07
		, TO_NUMBER('823456789012345678') bn08
		, TO_NUMBER('923456789012345678') bn09
		, TO_NUMBER('023456789012345678') bn10
	FROM ALL_OBJECTS B, all_objects A WHERE ROWNUM <= :1`)
}
func TestMemoryNumStringI64(t *testing.T) {
	drvCfg := ora.NewDrvCfg()
	drvCfg.Env.StmtCfg.Rset.SetNumberBigInt(ora.I64)
	drvCfg.Env.StmtCfg.Rset.SetNumberBigFloat(ora.I64)
	ora.SetCfg(*drvCfg)
	benchMem(t, 1349992, `SELECT
		TO_NUMBER('123456789012345678') bn01
		, TO_NUMBER('223456789012345678') bn02
		, TO_NUMBER('323456789012345678') bn03
		, TO_NUMBER('423456789012345678') bn04
		, TO_NUMBER('523456789012345678') bn05
		, TO_NUMBER('623456789012345678') bn06
		, TO_NUMBER('723456789012345678') bn07
		, TO_NUMBER('823456789012345678') bn08
		, TO_NUMBER('923456789012345678') bn09
		, TO_NUMBER('023456789012345678') bn10
	FROM ALL_OBJECTS B, all_objects A WHERE ROWNUM <= :1`)
	ora.SetCfg(*ora.NewDrvCfg())
}

func TestMemoryString(t *testing.T) {
	benchMem(t, 1424968, `SELECT
		'123456789012345678' bs01
		, '223456789012345678' bs02
		, '323456789012345678' bs03
		, '423456789012345678' bs04
		, '523456789012345678' bs05
		, '623456789012345678' bs06
		, '723456789012345678' bs07
		, '823456789012345678' bs08
		, '923456789012345678' bs09
		, '023456789012345678' bs10
	FROM ALL_OBJECTS B, all_objects A WHERE ROWNUM <= :1`)
}

func benchMem(tb testing.TB, maxBytes uint64, qry string) {
	columns, err := ora.DescribeQuery(testDb, qry)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Logf("columns: %#v", columns)

	cols := make([]string, len(columns))
	for i, c := range columns {
		cols[i] = c.Name
	}
	args := []interface{}{1000}

	type Record map[string]interface{}

	execute := func(qry string, cols []string, args ...interface{}) []Record {
		var out []Record

		rows, err := testDb.Query(qry, args...)
		if err != nil {
			tb.Fatalf("ERROR: DB.Query, query='%s' args='%v' error=%v", qry, args, err)
		}
		defer rows.Close()

		count := len(cols)
		vals := make([]interface{}, count)
		valPtrs := make([]interface{}, count)
		for i, _ := range cols {
			valPtrs[i] = &vals[i]
		}
		// loop over rows
		for rows.Next() {
			err := rows.Scan(valPtrs...)
			//        err := rows.Scan(vals...)
			if err != nil {
				tb.Fatalf("ERROR: rows.Scan, dest='%v', error=%v", vals, err)
			}
			rec := make(Record)
			length := 0
			for i, _ := range cols {
				rec[cols[i]] = vals[i]
				if s, ok := vals[i].(string); ok {
					length += len(s)
				}
			}
			out = append(out, rec)
			if len(out) == 1 {
				tb.Logf("One record's length: %d", length)
			}
		}
		if err = rows.Err(); err != nil {
			tb.Fatal(err)
		}
		return out
	}

	var ostat, nstat runtime.MemStats
	runtime.ReadMemStats(&ostat)
	for j := 0; j < 1; j++ {
		results := execute(qry, cols, args...)
		runtime.ReadMemStats(&nstat)
		d := nstat.TotalAlloc - ostat.TotalAlloc
		tb.Logf("test %d, nres=%d, allocated %d bytes\n", j, len(results), d)
		if maxBytes > 0 && d > maxBytes {
			tb.Errorf("test %d, nres=%d, allocated %d bytes (max: %d)", j, len(results), d, maxBytes)
		}
		ostat = nstat
	}
}
