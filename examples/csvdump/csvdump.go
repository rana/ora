/*
   Package main in csvdump represents a cursor->csv dumper

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
package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/errgo.v1"
	"gopkg.in/rana/ora.v4/examples/connect"
)

func getQuery(table, where string, columns []string) string {
	if strings.HasPrefix(table, "SELECT ") {
		return table
	}
	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}
	if where == "" {
		return "SELECT " + cols + " FROM " + table
	}
	return "SELECT " + cols + " FROM " + table + " WHERE " + where
}

func dump(w io.Writer, qry string) error {
	db, err := connect.GetConnection("")
	if err != nil {
		return errgo.Notef(err, "connect to database")
	}
	defer db.Close()
	columns, err := GetColumns(db, qry)
	if err != nil {
		return errgo.Notef(err, "get column converters", err)
	}
	log.Printf("columns: %#v", columns)
	row := make([]interface{}, len(columns))
	rowP := make([]interface{}, len(row))
	for i := range row {
		rowP[i] = &row[i]
	}

	rows, err := db.Query(qry)
	if err != nil {
		return errgo.Newf("error executing %q: %s", qry, err)
	}
	defer rows.Close()

	bw := bufio.NewWriterSize(w, 65536)
	defer bw.Flush()
	for i, col := range columns {
		if i > 0 {
			bw.Write([]byte{';'})
		}
		bw.Write([]byte{'"'})
		bw.WriteString(col.Name)
		bw.Write([]byte{'"'})
	}
	bw.Write([]byte{'\n'})
	n := 0
	for rows.Next() {
		if err = rows.Scan(rowP...); err != nil {
			return errgo.Notef(err, "scan %d. row", n+1)
		}
		for i, data := range row {
			if i > 0 {
				bw.Write([]byte{';'})
			}
			if data == nil {
				continue
			}
			bw.WriteString(columns[i].String(data))
		}
		bw.Write([]byte{'\n'})
		n++
	}
	log.Printf("written %d rows.", n)
	return rows.Err()
}

// QueryColumn is the described column.
type QueryColumn struct {
	Schema, Name                   string
	Type, Length, Precision, Scale int
	Nullable                       bool
	CharsetID, CharsetForm         int
}

type execer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

// DescribeQuery describes the columns in the qry string,
// using DBMS_SQL.PARSE + DBMS_SQL.DESCRIBE_COLUMNS2.
//
// This can help using unknown-at-compile-time, a.k.a.
// dynamic queries.
func DescribeQuery(db execer, qry string) ([]QueryColumn, error) {
	//res := strings.Repeat("\x00", 32767)
	res := make([]byte, 32767)
	if _, err := db.Exec(`DECLARE
  c INTEGER;
  col_cnt INTEGER;
  rec_tab DBMS_SQL.DESC_TAB;
  a DBMS_SQL.DESC_REC;
  v_idx PLS_INTEGER;
  res VARCHAR2(32767);
BEGIN
  c := DBMS_SQL.OPEN_CURSOR;
  BEGIN
    DBMS_SQL.PARSE(c, :1, DBMS_SQL.NATIVE);
    DBMS_SQL.DESCRIBE_COLUMNS(c, col_cnt, rec_tab);
    v_idx := rec_tab.FIRST;
    WHILE v_idx IS NOT NULL LOOP
      a := rec_tab(v_idx);
      res := res||a.col_schema_name||' '||a.col_name||' '||a.col_type||' '||
                  a.col_max_len||' '||a.col_precision||' '||a.col_scale||' '||
                  (CASE WHEN a.col_null_ok THEN 1 ELSE 0 END)||' '||
                  a.col_charsetid||' '||a.col_charsetform||
                  CHR(10);
      v_idx := rec_tab.NEXT(v_idx);
    END LOOP;
	--Loop ended, close cursor
    DBMS_SQL.CLOSE_CURSOR(c);
  EXCEPTION WHEN OTHERS THEN NULL;
    --Error happened, close cursor anyway!
    DBMS_SQL.CLOSE_CURSOR(c);
	RAISE;
  END;
  :2 := UTL_RAW.CAST_TO_RAW(res);
END;`, qry, &res,
	); err != nil {
		return nil, err
	}
	if i := bytes.IndexByte(res, 0); i >= 0 {
		res = res[:i]
	}
	lines := bytes.Split(res, []byte{'\n'})
	cols := make([]QueryColumn, 0, len(lines))
	var nullable int
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var col QueryColumn
		switch j := bytes.IndexByte(line, ' '); j {
		case -1:
			continue
		case 0:
			line = line[1:]
		default:
			col.Schema, line = string(line[:j]), line[j+1:]
		}
		if n, err := fmt.Sscanf(string(line), "%s %d %d %d %d %d %d %d",
			&col.Name, &col.Type, &col.Length, &col.Precision, &col.Scale, &nullable, &col.CharsetID, &col.CharsetForm,
		); err != nil {
			return cols, errors.Wrapf(err, "parsing %q (parsed: %d)", line, n)
		}
		col.Nullable = nullable != 0
		cols = append(cols, col)
	}
	return cols, nil
}

type ColConverter func(interface{}) string

type Column struct {
	Name   string
	String ColConverter
}

func GetColumns(db *sql.DB, qry string) (cols []Column, err error) {
	desc, err := DescribeQuery(db, qry)
	if err != nil {
		return nil, errgo.Newf("error getting description for %q: %s", qry, err)
	}
	log.Printf("desc: %#v", desc)
	var ok bool
	cols = make([]Column, len(desc))
	for i, col := range desc {
		cols[i].Name = col.Name
		if cols[i].String, ok = converters[col.Type]; !ok {
			cols[i].String = defaultConverter
			log.Printf("no converter for type %d (column name: %s)", col.Type, col.Name)
		}
	}
	return cols, nil
}

func defaultConverter(data interface{}) string { return fmt.Sprintf("%v", data) }

var converters = map[int]ColConverter{
	1: func(data interface{}) string { //VARCHAR2
		return fmt.Sprintf("%q", data.(string))
	},
	6: func(data interface{}) string { //NUMBER
		return fmt.Sprintf("%v", data)
	},
	96: func(data interface{}) string { //CHAR
		return fmt.Sprintf("%q", data.(string))
	},
	156: func(data interface{}) string { //DATE
		return `"` + data.(time.Time).Format(time.RFC3339) + `"`
	},
}

func main() {
	var (
		where   string
		columns []string
	)

	flag.Parse()
	if flag.NArg() > 1 {
		where = flag.Arg(1)
		if flag.NArg() > 2 {
			columns = flag.Args()[2:]
		}
	}
	qry := getQuery(flag.Arg(0), where, columns)
	if err := dump(os.Stdout, qry); err != nil {
		log.Printf("error dumping: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
