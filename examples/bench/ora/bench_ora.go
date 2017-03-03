// Copyright 2016 Tamás Gulácsi, Jia Lu. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	_ "gopkg.in/rana/ora.v4"
)

var (
	DB         *sql.DB
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	N          = flag.Int("N", 1000, "row count")
)

func init() {
	flag.Parse()
	drv, dsn := "ora", os.ExpandEnv(flag.Arg(0))
	var err error
	if DB, err = sql.Open(drv, dsn); err != nil {
		log.Fatalf("cannot connect with %q to %q: %v", drv, dsn, err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("cannot connect with %q to %q: %v", drv, dsn, err)
	}
}

func main() {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Write CPU profile to %q.", f.Name())
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	result := testing.Benchmark(BenchmarkIter)
	log.Println(result)
}

func BenchmarkIter(b *testing.B) {
	rows, err := DB.Query("SELECT A.object_name from all_objects, all_objects A")
	if err != nil {
		b.Fatal(err)
	}
	defer rows.Close()
	b.ReportAllocs()
	start := time.Now()
	b.ResetTimer()
	i := 0
	N := *N
	for rows.Next() && i < N {
		i++
	}
	b.StopTimer()
	d := time.Since(start)
	b.SetBytes(int64(i))
	log.Printf("Iterated %d rows in %s: %.3f row/s.",
		i, d, float64(i)/(float64(d)/float64(time.Second)))
}
