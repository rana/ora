/*
   Copyright 2013 Tamás Gulácsi

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package main in csvload is a csv -> table loader.
package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tgulacsi/go/term"
	"github.com/tgulacsi/go/text"
	"golang.org/x/text/encoding"
	"gopkg.in/inconshreveable/log15.v2"
)

var Log = log15.New()

const batchLen = 16

func main() {
	Log.SetHandler(log15.StderrHandler)
	flagConnect := flag.String("connect", os.Getenv("BRUNO_ID"), "database connection string")
	flagCharset := flag.String("charset", term.GetTTYEncodingName(), "input charset of the csv")
	flagTruncate := flag.Bool("truncate", false, "truncate table?")
	flagSep := flag.String("sep", ";", "csv field separator")
	flag.Parse()

	var enc encoding.Encoding
	if *flagCharset != "" {
		enc = text.GetEncoding(*flagCharset)
		if enc == nil {
			Log.Error("unknown charset " + *flagCharset)
			os.Exit(1)
		}
	}

	db, err := sql.Open("ora", *flagConnect)
	if err != nil {
		Log.Crit("connect to db", "dsn", *flagConnect, "error", err)
		os.Exit(1)
	}
	defer db.Close()

	fh, err := os.Open(flag.Arg(0))
	if err != nil {
		Log.Crit("open csv", "file", flag.Arg(0), "error", err)
		os.Exit(1)
	}
	defer fh.Close()
	r := io.Reader(fh)
	if enc != nil {
		Log.Debug("NewReader", "encoding", enc)
		r = text.NewReader(bufio.NewReaderSize(r, 1<<20), enc)
	}

	if *flagTruncate {
		if _, err = db.Exec("TRUNCATE TABLE " + flag.Arg(1)); err != nil {
			Log.Error("TRUNCATE", "table", flag.Arg(1), "error", err)
			os.Exit(1)
		}
	}
	if os.Getenv("GOMAXPROCS") == "" {
		Log.Info("Setting GOMAXPROCS", "numCPU", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cr := csv.NewReader(bufio.NewReaderSize(r, 16<<20))
	cr.Comma = ([]rune(*flagSep))[0]
	cr.TrimLeadingSpace = true
	cr.LazyQuotes = true
	if err := load(db, flag.Arg(1), cr); err != nil {
		Log.Error("load", "error", err)
		os.Exit(2)
	}
}

func load(db *sql.DB, tbl string, cr *csv.Reader) error {
	head, err := cr.Read()
	if err != nil {
		return err
	}
	cr.FieldsPerRecord = len(head)
	marks := make([]string, len(head))
	for i := range marks {
		marks[i] = "?"
	}
	qry := "INSERT INTO " + tbl + " (" + strings.Join(head, ",") + ") VALUES (" + strings.Join(marks, ",") + ")"
	Log.Info("insert", "qry", qry)

	var wg sync.WaitGroup
	conc := runtime.GOMAXPROCS(-1)
	blocks := make(chan [][]string, conc)
	errs := make(chan error, conc)
	rowCount := new(int32)

	R := func(f func() error) {
		defer wg.Done()
		errs <- f()
	}
	for i := 0; i < conc; i++ {
		wg.Add(1)
		go R(func() error {
			var (
				tx *sql.Tx
				st *sql.Stmt
			)
			n := 0
			values := make([]interface{}, len(marks))

			for block := range blocks {
				for _, row := range block {
					for i, v := range row {
						values[i] = v
					}
					if tx == nil {
						if st != nil {
							st.Close()
							st = nil
						}
						if tx, err = db.Begin(); err != nil {
							return err
						}
						if st, err = tx.Prepare(qry); err != nil {
							return err
						}
					}
					if _, err = st.Exec(values...); err != nil {
						return fmt.Errorf("error inserting %q with %q: %v", row, qry, err)
					}
					n++
					atomic.AddInt32(rowCount, 1)
					if n%1000 == 0 {
						if err = tx.Commit(); err != nil {
							return err
						}
						tx = nil
						Log.Info("commit", "n", n, "rowCount", atomic.LoadInt32(rowCount))
					}
				}
			}
			Log.Info("commit", "n", n, "rowCount", atomic.LoadInt32(rowCount))
			if st != nil {
				st.Close()
			}
			if tx != nil {
				return tx.Commit()
			}
			return nil
		})
	}

	var block [][]string
	t := time.Now()
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Error("read row", "error", err)
			continue
		}
		if block == nil {
			block = make([][]string, 0, batchLen)
		}
		block = append(block, row)
		if len(block) == batchLen {
			blocks <- block
			block = nil
		}
	}
	if len(block) > 0 {
		blocks <- block
	}
	close(blocks)
	wg.Wait()
	n, d := atomic.LoadInt32(rowCount), time.Since(t)
	fmt.Fprintf(os.Stderr, "Written %d rows under %s: %.3f rows/sec\n",
		n, d, float64(n)/float64(d/time.Second))
	close(errs)
	for err := range errs {
		if err == nil {
			continue
		}
		return err
	}

	return nil
}
