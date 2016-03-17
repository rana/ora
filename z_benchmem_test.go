// Copyright 2016 Tamás Gulácsi, Valentin Kuznetsov. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"runtime"
	"testing"

	"gopkg.in/rana/ora.v3"
)

// BenchmarkMemory usage for querying rows.
//
// go test -c && ./ora.v3.test -test.run=^$ -test.bench=Memory -test.memprofilerate=1 -test.memprofile=/tmp/mem.prof && go tool pprof --alloc_space ora.v3.test /tmp/mem.prof
func BenchmarkMemoryNumString(b *testing.B) {
	benchMem(b, `SELECT
		TO_NUMBER('123456789012345678901234567890') bn01
		, TO_NUMBER('223456789012345678901234567890') bn02
		, TO_NUMBER('323456789012345678901234567890') bn03
		, TO_NUMBER('423456789012345678901234567890') bn04
		, TO_NUMBER('523456789012345678901234567890') bn05
		, TO_NUMBER('623456789012345678901234567890') bn06
		, TO_NUMBER('723456789012345678901234567890') bn07
		, TO_NUMBER('823456789012345678901234567890') bn08
		, TO_NUMBER('923456789012345678901234567890') bn09
		, TO_NUMBER('023456789012345678901234567890') bn10
	FROM ALL_OBJECTS B, all_objects A WHERE ROWNUM <= :1`)
}

func BenchmarkMemoryString(b *testing.B) {
	benchMem(b, `SELECT
		'123456789012345678901234567890' bs01
		, '223456789012345678901234567890' bs02
		, '323456789012345678901234567890' bs03
		, '423456789012345678901234567890' bs04
		, '523456789012345678901234567890' bs05
		, '623456789012345678901234567890' bs06
		, '723456789012345678901234567890' bs07
		, '823456789012345678901234567890' bs08
		, '923456789012345678901234567890' bs09
		, '023456789012345678901234567890' bs10
	FROM ALL_OBJECTS B, all_objects A WHERE ROWNUM <= :1`)
}

func benchMem(b *testing.B, qry string) {
	columns, err := ora.DescribeQuery(testDb, qry)
	if err != nil {
		b.Fatal(err)
	}
	b.Logf("columns: %#v", columns)

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
			b.Fatalf("ERROR: DB.Query, query='%s' args='%v' error=%v", qry, args, err)
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
				b.Fatalf("ERROR: rows.Scan, dest='%v', error=%v", vals, err)
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
				b.Logf("One record's length: %d", length)
			}
			b.SetBytes(int64(length))
		}
		if err = rows.Err(); err != nil {
			b.Fatal(err)
		}
		return out
	}

	var ostat, nstat runtime.MemStats
	runtime.ReadMemStats(&ostat)
	for j := 0; j < 1; j++ {
		results := execute(qry, cols, args...)
		runtime.ReadMemStats(&nstat)
		fmt.Printf("test %d, nres=%d, allocated %d bytes\n", j, len(results), nstat.TotalAlloc-ostat.TotalAlloc)
		ostat = nstat
	}
}
