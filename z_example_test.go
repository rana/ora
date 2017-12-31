// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/rana/ora.v4"
)

func dbName() string {
	db := testConStr[strings.LastIndex(testConStr, "@")+1:]
	if db != "" {
		return db
	}
	return os.Getenv("GO_ORA_DRV_TEST_DB")
}

func ExampleDrvStmt_Exec_insert() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	db.Exec(fmt.Sprintf("create table %v (c1 number)", tableName))

	// placeholder ':c1' is bound by position; ':c1' may be any name
	var value int64 = 9
	result, err := db.Exec(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName), value)
	if err != nil {
		panic(err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleDrvStmt_Exec_insert_return_identity() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	qry := "create table " + tableName + " (c1 number(19,0) generated always as identity (start with 1 increment by 1), c2 varchar2(48 char))"
	if _, err := db.Exec(qry); err != nil {
		qry = strings.Replace(qry, "generated always as identity (start with 1 increment by 1)", "DEFAULT 1", 1)
		if _, err = db.Exec(qry); err != nil {
			fmt.Fprintf(os.Stderr, "error creating table with %q: %v", qry, err)
			return
		}
	}

	// use a 'returning into' SQL clause and specify a nil parameter to Exec
	// placeholder ':c1' is bound by position; ':c1' may be any name
	result, err := db.Exec(fmt.Sprintf("insert into %v (c2) values ('go') returning c1 /*lastinsertid*/ into :c1", tableName), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting 'go' with returning: %v", err)
		return
	}
	id, _ := result.LastInsertId()
	fmt.Println(id)
	// Output: 1
}

func ExampleDrvStmt_Exec_insert_bool() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	db.Exec(fmt.Sprintf("create table %v (c1 char(1 byte))", tableName))

	// default false symbol is '0'
	// default true symbol is '1'
	// placeholder ':c1' is bound by position; ':c1' may be any name
	var value bool = true
	result, _ := db.Exec(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName), value)
	rowsAffected, _ := result.RowsAffected()
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleDrvStmt_Exec_update() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	db.Exec(fmt.Sprintf("create table %v (c1 number)", tableName))
	db.Exec(fmt.Sprintf("insert into %v (c1) values (9)", tableName))

	// placeholder ':three' and ':nine' are bound by position; ':three' and ':nine' may be any name
	var three int64 = 3
	var nine int64 = 9
	result, _ := db.Exec(fmt.Sprintf("update %v set c1 = :three where c1 = :nine", tableName), three, nine)
	rowsAffected, _ := result.RowsAffected()
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleDrvStmt_Exec_delete() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	db.Exec(fmt.Sprintf("create table %v (c1 number)", tableName))
	db.Exec(fmt.Sprintf("insert into %v (c1) values (9)", tableName))

	// placeholder ':1' is bound by position; ':1' may be any name
	var v int64 = 9
	result, _ := db.Exec(fmt.Sprintf("delete from %v where c1 = :1", tableName), v)
	rowsAffected, _ := result.RowsAffected()
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleDrvStmt_Exec_Query() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	db.Exec(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char), c3 char(1 byte))", tableName))
	db.Exec(fmt.Sprintf("insert into %v (c1, c2, c3) values (3, 'slice', '0')", tableName))
	db.Exec(fmt.Sprintf("insert into %v (c1, c2, c3) values (7, 'map', '1')", tableName))
	db.Exec(fmt.Sprintf("insert into %v (c1, c2, c3) values (9, 'channel', '1')", tableName))

	// placeholder ':p' is bound by position; ':p' may be any name
	var value int64 = 8
	rows, _ := db.Query(fmt.Sprintf("select c1, c2, c3 from %v where c1 > :p", tableName), value)
	defer rows.Close()
	for rows.Next() {
		var c1 int64
		var c2 string
		var c3 string
		rows.Scan(&c1, &c2, &c3)
		fmt.Printf("%v %v %v", c1, c2, c3)
	}
	// Output: 9 channel 1
}

// TODO: Fix QueryRow
func ExampleDrvStmt_Exec_QueryRow() {
	db, _ := sql.Open("ora", testConStr)
	defer db.Close()

	tableName := tableName()
	qry := fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char))", tableName)
	if _, err := db.Exec(qry); err != nil {
		log.Fatal(errors.Wrap(err, qry))
	}
	qry = fmt.Sprintf("insert into %v (c1, c2) values (9, 'go')", tableName)
	if _, err := db.Exec(qry); err != nil {
		log.Fatal(errors.Wrap(err, qry))
	}

	// placeholder ':p' is bound by position; ':p' may be any name
	var c1 int64 = 9
	var c2 string
	qry = fmt.Sprintf("select c2 from %v where c1 = :p", tableName)
	if err := db.QueryRow(qry, c1).Scan(&c2); err != nil {
		log.Fatal(errors.Wrap(err, qry))
	}
	fmt.Println(c2)
	// Output: go
}

func ExampleStmt_Exe_insert() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert record
	var value int64 = 9
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(value)
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleStmt_Exe_insert_return_identity() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	qry := "create table " + tableName + " (c1 number(19,0)"
	if ver, _ := srv.Version(); strings.Contains(ver, " 12.") {
		qry += " generated always as identity (start with 1 increment by 1)"
	} else {
		qry += " default 1"
	}
	qry += ", c2 varchar2(48 char))"
	stmt, _ := ses.Prep(qry)
	defer stmt.Close()
	stmt.Exe()

	// insert record
	var id int64
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c2) values ('go') returning c1 into :c1", tableName))
	defer stmt.Close()
	// pass a numeric pointer to rereive a database generated identity value
	stmt.Exe(&id)
	fmt.Println(id)
	// Output: 1
}

func ExampleStmt_Exe_insert_return_rowid() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert record
	var rowid string
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (9) returning rowid into :r", tableName))
	defer stmt.Close()
	// pass a string pointer to rereive a rowid
	stmt.Exe(&rowid)
	if rowid != "" {
		fmt.Println("Retrieved rowid")
	}
	// Output: Retrieved rowid
}

func ExampleStmt_Exe_insert_fetch_bool() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert 'false' record
	var falseValue bool = false
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(falseValue)
	// insert 'true' record
	var trueValue bool = true
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(trueValue)

	// fetch inserted records
	stmt, _ = ses.Prep("select c1 from "+tableName, ora.B)
	defer stmt.Close()
	rset, err := stmt.Qry()
	if err != nil {
		log.Fatal(err)
	}
	for rset.Next() {
		fmt.Printf("%v ", rset.Row[0])
	}
	// Output: false true
}

func ExampleStmt_Exe_insert_fetch_bool_alternate() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// Update StmtCfg to change the FalseRune and TrueRune inserted into the database
	// insert 'false' record
	var falseValue bool = false
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmtCfg := stmt.Cfg()
	stmtCfg.FalseRune = 'N'
	stmt.SetCfg(stmtCfg)
	stmt.Exe(falseValue)
	// insert 'true' record
	var trueValue bool = true
	stmt, _ = ses.Prep("insert into "+tableName+" (c1) values (:c1)", ora.B)
	defer stmt.Close()
	stmtCfg.TrueRune = 'Y'
	stmt.SetCfg(stmtCfg)
	stmt.Exe(trueValue)

	// Update RsetCfg to change the TrueRune
	// used to translate an Oracle char to a Go bool
	// fetch inserted records
	stmt, _ = ses.Prep("select c1 from "+tableName, ora.B)
	defer stmt.Close()
	stmtCfg.TrueRune = 'Y'
	stmt.SetCfg(stmtCfg)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v ", rset.Row[0])
	}
	// Output: false true
}

func ExampleStmt_Exe_update() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// update record
	var a int64 = 3
	var b int64 = 9
	stmt, _ = ses.Prep(fmt.Sprintf("update %v set c1 = :three where c1 = :nine", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(a, b)
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleStmt_Exe_delete() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// delete record
	var value int64 = 9
	stmt, _ = ses.Prep(fmt.Sprintf("delete from %v where c1 = :1", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(value)
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleStmt_Exe_insert_slice() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert one million rows with single round-trip to server
	values := make([]int64, 1000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(values)
	fmt.Println(rowsAffected)
	// Output: 1000
}

func ExampleStmt_Exe_insert_nullable() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char), c3 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// create nullable Go types for inserting null
	// insert record
	a := ora.Int64{IsNull: true}
	b := ora.String{IsNull: true}
	c := ora.Bool{IsNull: true}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (:c1, :c2, :c3)", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(a, b, c)
	fmt.Println(rowsAffected)
	// Output: 1
}

func ExampleStmt_Exe_insert_fetch_blob() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 blob)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// by default, byte slices are expected to be bound and retrieved
	// to/from a binary column such as a blob
	// insert record
	a := make([]byte, 10)
	for n, _ := range a {
		a[n] = byte(n)
	}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(a)
	fmt.Println(rowsAffected)

	// fetch record
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	defer stmt.Close()
	rset, _ := stmt.Qry()
	row := rset.NextRow()
	if err := rset.Err(); err != nil {
		fmt.Printf("ERROR: %v", err)
	} else {
		fmt.Println(row[0].([]byte))
	}

	// Output:
	// 1
	// [0 1 2 3 4 5 6 7 8 9]
}

func ExampleStmt_Exe_insert_fetch_byteSlice() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// note the NUMBER column
	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// Specify stmt.Cfg.SetByteSlice(U8)
	// Specify byte slice to be inserted into a NUMBER column
	// insert records
	a := make([]byte, 10)
	for n, _ := range a {
		a[n] = byte(n)
	}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmtCfg := stmt.Cfg()
	stmtCfg = stmtCfg.SetByteSlice(ora.U8)
	stmt.SetCfg(stmtCfg)
	rowsAffected, _ := stmt.Exe(a)
	fmt.Println(rowsAffected)

	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	defer stmt.Close()
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v, ", rset.Row[0])
	}

	// Output:
	// 10
	// 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
}

func ExampleStmt_Qry() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	cfg := ses.Cfg()
	defer ses.SetCfg(cfg)
	cfg = cfg.SetChar1(ora.B)
	ses.SetCfg(cfg)

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char), c3 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (3, 'slice', '0')", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (7, 'map', '1')", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (9, 'channel', '1')", tableName))
	defer stmt.Close()
	stmt.Exe()

	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1, c2, c3 from %v", tableName))
	defer stmt.Close()
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v %v %v, ", rset.Row[0], rset.Row[1], rset.Row[2])
	}
	// Output: 3 slice false, 7 map true, 9 channel true,
}

func ExampleStmt_Qry_nullable() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char), c3 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (null, 'slice', '0')", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (7, null, '1')", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (9, 'channel', null)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// Specify nullable return types to the Prep method
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1, c2, c3 from %v", tableName), ora.OraI64, ora.OraS, ora.OraB)
	defer stmt.Close()
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v %v %v, ", rset.Row[0], rset.Row[1], rset.Row[2])
	}
	// Output: {true 0} slice {false false}, {false 7}  {false true}, {false 9} channel {true false},
}

func ExampleStmt_Qry_numerics() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()
	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// Specify various numeric return types to the Prep method
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1, c1, c1, c1, c1, c1, c1, c1, c1, c1 from %v", tableName), ora.I64, ora.I32, ora.I16, ora.I8, ora.U64, ora.U32, ora.U16, ora.U8, ora.F64, ora.F32)
	defer stmt.Close()
	rset, _ := stmt.Qry()
	row := rset.NextRow()
	fmt.Printf("%v %v %v %v %v %v %v %v %v %v",
		reflect.TypeOf(row[0]).Name(),
		reflect.TypeOf(row[1]).Name(),
		reflect.TypeOf(row[2]).Name(),
		reflect.TypeOf(row[3]).Name(),
		reflect.TypeOf(row[4]).Name(),
		reflect.TypeOf(row[5]).Name(),
		reflect.TypeOf(row[6]).Name(),
		reflect.TypeOf(row[7]).Name(),
		reflect.TypeOf(row[8]).Name(),
		reflect.TypeOf(row[9]).Name())
	// Output: int64 int32 int16 int8 uint64 uint32 uint16 uint8 float64 float32
}

func ExampleRset_Next() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert records
	a := make([]uint16, 5)
	for n, _ := range a {
		a[n] = uint16(n)
	}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	rowsAffected, _ := stmt.Exe(a)
	fmt.Println(rowsAffected)

	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.U16)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v, ", rset.Row[0])
	}
	// Output:
	// 5
	// 0, 1, 2, 3, 4,
}

func ExampleRset_NextRow() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char), c3 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert record
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2, c3) values (7, 'go', '1')", tableName))
	defer stmt.Close()
	stmt.Exe()

	// fetch record
	stmt, _ = ses.Prep(fmt.Sprintf("select c1, c2, c3 from %v", tableName))
	rset, _ := stmt.Qry()
	row := rset.NextRow()
	fmt.Printf("%v %v %v", row[0], row[1], row[2])
	// Output: 7 go 1
}

func ExampleRset_cursor_single() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert records
	a := make([]int64, 3)
	a[0] = 5
	a[1] = 7
	a[2] = 9
	b := make([]string, 3)
	b[0] = "Go is expressive, concise, clean, and efficient."
	b[1] = "Its concurrency mechanisms make it easy to"
	b[2] = "Go compiles quickly to machine code yet has"
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2) values (:1, :2)", tableName))
	stmt.Exe(a, b)

	// create proc
	stmt, _ = ses.Prep(fmt.Sprintf("create or replace procedure proc1(p1 out sys_refcursor) as begin open p1 for select c1, c2 from %v order by c1; end proc1;", tableName))
	defer stmt.Close()
	stmt.Exe()

	// pass *ora.Rset to Exec for an out sys_refcursor
	// call proc
	stmt, _ = ses.Prep("call proc1(:1)")
	defer stmt.Close()
	rset := &ora.Rset{}
	if _, err := stmt.Exe(rset); err != nil {
		log.Fatal(err)
	}
	if rset.IsOpen() {
		for rset.Next() {
			fmt.Println(rset.Row[0], rset.Row[1])
		}
	}
	// Output:
	// 5 Go is expressive, concise, clean, and efficient.
	// 7 Its concurrency mechanisms make it easy to
	// 9 Go compiles quickly to machine code yet has
}

func ExampleRset_cursor_multiple() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number, c2 varchar2(48 char))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert records
	a := make([]int64, 3)
	a[0] = 5
	a[1] = 7
	a[2] = 9
	b := make([]string, 3)
	b[0] = "Go is expressive, concise, clean, and efficient."
	b[1] = "Its concurrency mechanisms make it easy to"
	b[2] = "Go compiles quickly to machine code yet has"
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1, c2) values (:1, :2)", tableName))
	stmt.Exe(a, b)

	// create proc
	stmt, _ = ses.Prep(fmt.Sprintf("create or replace procedure proc1(p1 out sys_refcursor, p2 out sys_refcursor) as begin open p1 for select c1 from %v order by c1; open p2 for select c2 from %v order by c2; end proc1;", tableName, tableName))
	defer stmt.Close()
	stmt.Exe()

	// pass *ora.Rset to Exec for an out sys_refcursor
	// call proc
	stmt, _ = ses.Prep("call proc1(:1, :2)")
	defer stmt.Close()
	rsetC1 := &ora.Rset{}
	rsetC2 := &ora.Rset{}
	stmt.Exe(rsetC1, rsetC2)
	fmt.Println("--- first result set ---")
	if rsetC1.IsOpen() {
		for rsetC1.Next() {
			fmt.Println(rsetC1.Row[0])
		}
	}
	fmt.Println("--- second result set ---")
	if rsetC2.IsOpen() {
		for rsetC2.Next() {
			fmt.Println(rsetC2.Row[0])
		}
	}
	// Output:
	// --- first result set ---
	// 5
	// 7
	// 9
	// --- second result set ---
	// Go compiles quickly to machine code yet has
	// Go is expressive, concise, clean, and efficient.
	// Its concurrency mechanisms make it easy to
}

func ExampleSrv_Ping() {
	// setup
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	// open a session before calling Ping
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- ses.Ping()
	}()

	select {
	case <-time.After(10 * time.Second):
		ses.Break()
		fmt.Println("Ping timed out!")
	case err := <-done:
		if err == nil {
			fmt.Println("Ping successful")
		} else {
			fmt.Println("Ping ERROR:", err)
		}
	}
	// Output: Ping successful
}

func ExampleSrv_Version() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()

	// open a session before calling Version
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	version, err := srv.Version()
	if version != "" && err == nil {
		fmt.Println("Received version from server")
	}
	// Output: Received version from server
}

func ExampleInt64() {
	// setup
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, err := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exe(); err != nil {
		panic(err)
	}

	// insert ora.Int64 slice
	a := make([]ora.Int64, 5)
	a[0] = ora.Int64{Value: -9}
	a[1] = ora.Int64{Value: -1}
	a[2] = ora.Int64{IsNull: true}
	a[3] = ora.Int64{Value: 1}
	a[4] = ora.Int64{Value: 9}
	if stmt, err = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName)); err != nil {
		panic(err)
	}
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraI64 to Prep method to return nullable ora.Int64 values
	// fetch records
	if stmt, err = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraI64); err != nil {
		panic(err)
	}
	rset, err := stmt.Qry()
	if err != nil {
		panic(err)
	}
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	if err := rset.Err(); err != nil {
		panic(err)
	}
	// Output:
	// {false -9}
	// {false -1}
	// {true 0}
	// {false 1}
	// {false 9}
}

func ExampleInt32() {
	// setup
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, err := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Int32 slice
	a := make([]ora.Int32, 5)
	a[0] = ora.Int32{Value: -9}
	a[1] = ora.Int32{Value: -1}
	a[2] = ora.Int32{IsNull: true}
	a[3] = ora.Int32{Value: 1}
	a[4] = ora.Int32{Value: 9}
	if stmt, err = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName)); err != nil {
		panic(err)
	}
	defer stmt.Close()
	if _, err = stmt.Exe(a); err != nil {
		panic(err)
	}

	// Specify ora.OraI32 to Prep method to return nullable ora.Int32 values
	// fetch records
	if stmt, err = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraI32); err != nil {
		panic(err)
	}
	rset, err := stmt.Qry()
	if err != nil {
		panic(err)
	}
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	if err := rset.Err(); err != nil {
		panic(err)
	}
	// Output:
	// {false -9}
	// {false -1}
	// {true 0}
	// {false 1}
	// {false 9}
}

func ExampleInt16() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Int16 slice
	a := make([]ora.Int16, 5)
	a[0] = ora.Int16{Value: -9}
	a[1] = ora.Int16{Value: -1}
	a[2] = ora.Int16{IsNull: true}
	a[3] = ora.Int16{Value: 1}
	a[4] = ora.Int16{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraI16 to Prep method to return nullable ora.Int16 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraI16)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false -9}
	// {false -1}
	// {true 0}
	// {false 1}
	// {false 9}
}

func ExampleInt8() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Int8 slice
	a := make([]ora.Int8, 5)
	a[0] = ora.Int8{Value: -9}
	a[1] = ora.Int8{Value: -1}
	a[2] = ora.Int8{IsNull: true}
	a[3] = ora.Int8{Value: 1}
	a[4] = ora.Int8{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraI8 to Prep method to return nullable ora.Int8 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraI8)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false -9}
	// {false -1}
	// {true 0}
	// {false 1}
	// {false 9}
}

func ExampleUint64() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Uint64 slice
	a := make([]ora.Uint64, 5)
	a[0] = ora.Uint64{Value: 0}
	a[1] = ora.Uint64{Value: 3}
	a[2] = ora.Uint64{IsNull: true}
	a[3] = ora.Uint64{Value: 7}
	a[4] = ora.Uint64{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraU64 to Prep method to return nullable ora.Uint64 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraU64)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false 0}
	// {false 3}
	// {true 0}
	// {false 7}
	// {false 9}
}

func ExampleUint32() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Uint32 slice
	a := make([]ora.Uint32, 5)
	a[0] = ora.Uint32{Value: 0}
	a[1] = ora.Uint32{Value: 3}
	a[2] = ora.Uint32{IsNull: true}
	a[3] = ora.Uint32{Value: 7}
	a[4] = ora.Uint32{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraU32 to Prep method to return nullable ora.Uint32 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraU32)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false 0}
	// {false 3}
	// {true 0}
	// {false 7}
	// {false 9}
}

func ExampleUint16() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Uint16 slice
	a := make([]ora.Uint16, 5)
	a[0] = ora.Uint16{Value: 0}
	a[1] = ora.Uint16{Value: 3}
	a[2] = ora.Uint16{IsNull: true}
	a[3] = ora.Uint16{Value: 7}
	a[4] = ora.Uint16{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraU16 to Prep method to return nullable ora.Uint16 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraU16)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false 0}
	// {false 3}
	// {true 0}
	// {false 7}
	// {false 9}
}

func ExampleUint8() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(10,0))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Uint8 slice
	a := make([]ora.Uint8, 5)
	a[0] = ora.Uint8{Value: 0}
	a[1] = ora.Uint8{Value: 3}
	a[2] = ora.Uint8{IsNull: true}
	a[3] = ora.Uint8{Value: 7}
	a[4] = ora.Uint8{Value: 9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraU8 to Prep method to return nullable ora.Uint8 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraU8)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false 0}
	// {false 3}
	// {true 0}
	// {false 7}
	// {false 9}
}

func ExampleFloat64() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number(16,15))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Float64 slice
	a := make([]ora.Float64, 5)
	a[0] = ora.Float64{Value: -float64(6.28318)}
	a[1] = ora.Float64{Value: -float64(3.14159)}
	a[2] = ora.Float64{IsNull: true}
	a[3] = ora.Float64{Value: float64(3.14159)}
	a[4] = ora.Float64{Value: float64(6.28318)}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraF64 to Prep method to return nullable ora.Float64 values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraF64)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false -6.28318}
	// {false -3.14159}
	// {true 0}
	// {false 3.14159}
	// {false 6.28318}
}

func ExampleFloat32() {
	// setup
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, err := ses.Prep(fmt.Sprintf("create table %v (c1 number(16,15))", tableName))
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exe(); err != nil {
		panic(err)
	}

	// insert ora.Float32 slice
	a := make([]ora.Float32, 5)
	a[0] = ora.Float32{Value: -float32(6.28318)}
	a[1] = ora.Float32{Value: -float32(3.14159)}
	a[2] = ora.Float32{IsNull: true}
	a[3] = ora.Float32{Value: float32(3.14159)}
	a[4] = ora.Float32{Value: float32(6.28318)}
	if stmt, err = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName)); err != nil {
		panic(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exe(a); err != nil {
		panic(err)
	}

	// Specify ora.OraF32 to Prep method to return nullable ora.Float32 values
	// fetch records
	if stmt, err = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraF32); err != nil {
		panic(err)
	}
	rset, err := stmt.Qry()
	if err != nil {
		panic(err)
	}
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	if err := rset.Err(); err != nil {
		panic(err)
	}
	// Output:
	// {false -6.28318}
	// {false -3.14159}
	// {true 0}
	// {false 3.14159}
	// {false 6.28318}
}

func ExampleString() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 varchar2(48 char))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.String slice
	a := make([]ora.String, 5)
	a[0] = ora.String{Value: "Go is expressive, concise, clean, and efficient."}
	a[1] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	a[2] = ora.String{IsNull: true}
	a[3] = ora.String{Value: "It's a fast, statically typed, compiled"}
	a[4] = ora.String{Value: "One of Go's key design goals is code"}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraS to Prep method to return nullable ora.String values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraS)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// Go is expressive, concise, clean, and efficient.
	// Its concurrency mechanisms make it easy to
	//
	// It's a fast, statically typed, compiled
	// One of Go's key design goals is code
}

func ExampleBool() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 char(1 byte))", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Bool slice
	a := make([]ora.Bool, 5)
	a[0] = ora.Bool{Value: true}
	a[1] = ora.Bool{Value: false}
	a[2] = ora.Bool{IsNull: true}
	a[3] = ora.Bool{Value: false}
	a[4] = ora.Bool{Value: true}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// Specify ora.OraB to Prep method to return nullable ora.Bool values
	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.OraB)
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// {false true}
	// {false false}
	// {true false}
	// {false false}
	// {false true}
}

func ExampleTime() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 timestamp)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Time slice
	a := []ora.Time{
		{Value: time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)},
		{Value: time.Date(2001, 2, 3, 4, 5, 6, 0, testDbsessiontimezone)},
		{IsNull: true},
		{Value: time.Date(2003, 4, 5, 6, 7, 8, 0, testDbsessiontimezone)},
		{Value: time.Date(2004, 5, 6, 7, 8, 9, 0, testDbsessiontimezone)},
	}
	stmt, err := ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exe(a); err != nil {
		log.Fatal(err)
	}

	//ora.Cfg().Log.Rset.BeginRow = true
	//ora.Cfg().Log.Logger = lg.Log

	// Specify ora.OraT to Prep method to return nullable ora.Time values
	// fetch records
	stmt, _ = ses.Prep("select c1 from "+tableName, ora.OraT)
	rset, err := stmt.Qry()
	if err != nil {
		log.Fatal(err)
	}
	for rset.Next() {
		t := rset.Row[0].(ora.Time)
		fmt.Printf("%v %v-%v-%v %v:%v:%v\n", t.IsNull, t.Value.Year(), t.Value.Month(), t.Value.Day(), t.Value.Hour(), t.Value.Minute(), t.Value.Second())
	}
	// Output:
	// false 2000-January-2 3:4:5
	// false 2001-February-3 4:5:6
	// true 1-January-1 0:0:0
	// false 2003-April-5 6:7:8
	// false 2004-May-6 7:8:9
}

func ExampleIntervalYM() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 interval year to month)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.IntervalYM slice
	a := make([]ora.IntervalYM, 5)
	a[0] = ora.IntervalYM{Year: 1, Month: 1}
	a[1] = ora.IntervalYM{Year: 99, Month: 9}
	a[2] = ora.IntervalYM{IsNull: true}
	a[3] = ora.IntervalYM{Year: -1, Month: -1}
	a[4] = ora.IntervalYM{Year: -99, Month: -9}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// fetch ora.IntervalYM
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v, ", rset.Row[0])
	}
	// Output: 0001-01, 0099-09, , -001--1, -099--9,
}

func ExampleIntervalYM_ShiftTime() {
	interval := ora.IntervalYM{Year: 1, Month: 1}
	actual := interval.ShiftTime(time.Date(2000, time.January, 0, 0, 0, 0, 0, time.Local))
	fmt.Println(actual.Year(), actual.Month(), actual.Day())
	// returns normalized date per time.AddDate
	// Output: 2001 January 31
}

func ExampleIntervalDS() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 interval day to second)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.IntervalDS slice
	a := make([]ora.IntervalDS, 5)
	a[0] = ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	a[1] = ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}
	a[2] = ora.IntervalDS{IsNull: true}
	a[3] = ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}
	a[4] = ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// fetch ora.IntervalDS
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v, ", rset.Row[0])
	}
	// {false 1 1 1 1 123457000}, {false 59 59 59 59 123457000}, {true 0 0 0 0 0}, {false -1 -1 -1 -1 -123457000}, {false -59 -59 -59 -59 -123457000},
	// Output: 01d 01:01:01.123457000, 59d 59:59:59.123457000, , -1d -1:-1:-1.-123457000, -59d -59:-59:-59.-123457000,
}

func ExampleIntervalDS_ShiftTime() {
	interval := ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	actual := interval.ShiftTime(time.Date(2000, time.Month(1), 1, 0, 0, 0, 0, time.Local))
	fmt.Println(actual.Day(), actual.Hour(), actual.Minute(), actual.Second(), actual.Nanosecond())
	// Output: 2 1 1 1 123456789
}

func ExampleBytes() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 blob)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert Binary slice
	a := make([]ora.Raw, 5)
	b := make([]byte, 10)
	for n, _ := range b {
		b[n] = byte(n)
	}
	a[0] = ora.Raw{Value: b}
	b = make([]byte, 10)
	for n, _ := range b {
		b[n] = byte(n * 2)
	}
	a[1] = ora.Raw{Value: b}
	a[2] = ora.Raw{IsNull: true}
	b = make([]byte, 10)
	for n, _ := range b {
		b[n] = byte(n * 3)
	}
	a[3] = ora.Raw{Value: b}
	b = make([]byte, 10)
	for n, _ := range b {
		b[n] = byte(n * 4)
	}
	a[4] = ora.Raw{Value: b}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	//ora.Cfg().Log.Rset.BeginRow = true
	//ora.Cfg().Log.Logger = lg.Log

	// Specify OraBin to Prep method to return Binary values
	// fetch records
	stmt, _ = ses.Prep("select c1 from "+tableName, ora.OraBin)
	rset, _ := stmt.Qry()
	for rset.Next() {
		raw := rset.Row[0].(ora.Raw)
		fmt.Println(raw.Value)
	}
	// Output:
	// [0 1 2 3 4 5 6 7 8 9]
	// [0 2 4 6 8 10 12 14 16 18]
	// []
	// [0 3 6 9 12 15 18 21 24 27]
	// [0 4 8 12 16 20 24 28 32 36]
}

func ExampleWriteLOB() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	qry := fmt.Sprintf("create table %v (c1 blob)", tableName)
	stmt, err := ses.Prep(qry)
	if err != nil {
		log.Fatalf("%q: %v", qry, err)
	}
	defer stmt.Close()
	if _, err = stmt.Exe(); err != nil {
		log.Fatalf("%q: %v", qry, err)
	}
	n := 32767 + 1

	// insert Binary slice
	qry = fmt.Sprintf("insert into %v (c1) values (:c1)", tableName)
	blob := &ora.Lob{Reader: bytes.NewReader(make([]byte, n))}
	if stmt, err = ses.Prep(qry); err != nil {
		log.Fatalf("%q: %v", qry, err)
	}
	defer stmt.Close()
	if _, err = stmt.Exe(blob); err != nil {
		log.Fatalf("%q: %v", qry, err)
	}

	fmt.Println(n)

	// Specify OraBin to Prep method to return Binary values
	// fetch records
	qry = fmt.Sprintf("select c1 from %v", tableName)
	if stmt, err = ses.Prep(qry, ora.D); err != nil {
		log.Fatalf("%q: %v", qry, err)
	}
	defer stmt.Close()
	rset, err := stmt.Qry()
	if err != nil {
		log.Fatalf("%q: %v", qry, err)
	}
	for rset.Next() {
		b := rset.Row[0].([]byte)
		fmt.Println(len(b))
	}
	// Output:
	// 32768
	// 32768
}

func ExampleBfile() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 bfile)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// insert ora.Bfile
	a := ora.Bfile{IsNull: false, DirectoryAlias: "TEMP_DIR", Filename: "test.txt"}
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	defer stmt.Close()
	stmt.Exe(a)

	// fetch ora.Bfile
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Printf("%v", rset.Row[0])
	}
	// Output: {false TEMP_DIR test.txt}
}

func ExampleTx() {
	// setup
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()

	// create table
	tableName := tableName()
	stmt, _ := ses.Prep(fmt.Sprintf("create table %v (c1 number)", tableName))
	defer stmt.Close()
	stmt.Exe()

	// rollback
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (3)", tableName))
	tx, _ := ses.StartTx()
	stmt.Exe()
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (5)", tableName))
	stmt.Exe()
	tx.Rollback()

	// commit
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (7)", tableName))
	tx, _ = ses.StartTx()
	stmt.Exe()
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (9)", tableName))
	stmt.Exe()
	tx.Commit()

	// check that auto commit is reenabled
	stmt, _ = ses.Prep(fmt.Sprintf("insert into %v (c1) values (11)", tableName))
	stmt.Exe()

	// fetch records
	stmt, _ = ses.Prep(fmt.Sprintf("select c1 from %v", tableName))
	rset, _ := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// 7
	// 9
	// 11
}

func Example() {
	// example usage of the ora package driver
	// connect to a server and open a session
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	if err != nil {
		panic(err)
	}
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()

	// create table
	tableName := "t1"
	ses.PrepAndExe("DROP TABLE " + tableName)
	qry := "CREATE TABLE " + tableName + "(C1 NUMBER(19,0)"
	ver, _ := srv.Version()
	var autoC1 int
	if strings.Contains(ver, " 12.") {
		qry += " GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1)"
	} else {
		autoC1 = 1
	}
	qry += ", C2 VARCHAR2(48 CHAR))"
	stmtTbl, err := ses.Prep(qry)
	if err != nil {
		panic(err)
	}
	defer stmtTbl.Close()
	rowsAffected, err := stmtTbl.Exe()
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// begin first transaction
	tx1, err := ses.StartTx()
	if err != nil {
		panic(err)
	}

	// insert record
	var id uint64
	str := "Go is expressive, concise, clean, and efficient."
	qry = "(C2) VALUES (:C2)"
	if autoC1 > 0 {
		qry = fmt.Sprintf("(C1,C2) VALUES (%d,:C2)", autoC1)
		autoC1++
	}
	stmtIns, err := ses.Prep(fmt.Sprintf(
		"INSERT INTO %v "+qry+" RETURNING C1 INTO :C1", tableName))
	if err != nil {
		panic(err)
	}
	defer stmtIns.Close()
	rowsAffected, err = stmtIns.Exe(str, &id)
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// insert nullable ora.String slice
	a := make([]ora.String, 4)
	a[0] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	a[1] = ora.String{IsNull: true}
	a[2] = ora.String{Value: "It's a fast, statically typed, compiled"}
	a[3] = ora.String{Value: "One of Go's key design goals is code"}
	if autoC1 > 0 {
		qry = "(C1,C2) VALUES (:C1,:C2)"
	}
	stmtSliceIns, err := ses.Prep(fmt.Sprintf(
		"INSERT INTO %v "+qry, tableName))
	defer stmtSliceIns.Close()
	if err != nil {
		panic(err)
	}
	if autoC1 == 0 {
		rowsAffected, err = stmtSliceIns.Exe(a)
	} else {
		b := make([]ora.Int32, len(a))
		for i := range b {
			b[i] = ora.Int32{Value: int32(autoC1)}
			autoC1++
		}
		rowsAffected, err = stmtSliceIns.Exe(b, a)
	}
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// fetch records
	stmtQry, err := ses.Prep(fmt.Sprintf(
		"SELECT C1, C2 FROM %v", tableName))
	defer stmtQry.Close()
	if err != nil {
		panic(err)
	}
	rset, err := stmtQry.Qry()
	if err != nil {
		panic(err)
	}
	for rset.Next() {
		fmt.Println(rset.Row[0], emptyString(rset.Row[1].(string)))
	}
	if rset.Err() != nil {
		panic(rset.Err())
	}

	// commit first transaction
	err = tx1.Commit()
	if err != nil {
		panic(err)
	}

	// begin second transaction
	tx2, err := ses.StartTx()
	if err != nil {
		panic(err)
	}
	// insert null ora.String
	nullableStr := ora.String{IsNull: true}
	stmtTrans, err := ses.Prep(fmt.Sprintf(
		"INSERT INTO %v (C2) VALUES (:C2)", tableName))
	if err != nil {
		panic(err)
	}
	defer stmtTrans.Close()
	rowsAffected, err = stmtTrans.Exe(nullableStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)
	// rollback second transaction
	err = tx2.Rollback()
	if err != nil {
		panic(err)
	}

	// fetch and specify return type
	stmtCount, err := ses.Prep(fmt.Sprintf(
		"SELECT COUNT(C1) FROM %v WHERE C2 IS NULL", tableName), ora.U8)
	defer stmtCount.Close()
	if err != nil {
		panic(err)
	}
	rset, err = stmtCount.Qry()
	if err != nil {
		panic(err)
	}
	row := rset.NextRow()
	if row != nil {
		fmt.Println(row[0])
	}
	if rset.Err() != nil {
		panic(rset.Err())
	}

	// create stored procedure with sys_refcursor
	stmtProcCreate, err := ses.Prep(fmt.Sprintf(
		"CREATE OR REPLACE PROCEDURE PROC1(P1 OUT SYS_REFCURSOR) AS BEGIN "+
			"OPEN P1 FOR SELECT C1, C2 FROM %v WHERE C1 > 2 ORDER BY C1; "+
			"END PROC1;",
		tableName))
	if err != nil {
		panic(err)
	}
	defer stmtProcCreate.Close()
	_, err = stmtProcCreate.Exe()
	if err != nil {
		panic(err)
	}

	// call stored procedure
	// pass *ora.Rset to Exec to receive the results of a sys_refcursor
	stmtProcCall, err := ses.Prep("CALL PROC1(:1)")
	if err != nil {
		panic(err)
	}
	defer stmtProcCall.Close()
	if err != nil {
		panic(err)
	}
	procRset := &ora.Rset{}
	_, err = stmtProcCall.Exe(procRset)
	if err != nil {
		panic(err)
	}
	if procRset.IsOpen() {
		for procRset.Next() {
			fmt.Println(procRset.Row[0], emptyString(procRset.Row[1].(string)))
		}
		if procRset.Err() != nil {
			panic(procRset.Err())
		}
		fmt.Println(procRset.Len())
	}

	// Output:
	// 0
	// 1
	// 4
	// 1 Go is expressive, concise, clean, and efficient.
	// 2 Its concurrency mechanisms make it easy to
	// 3 <empty>
	// 4 It's a fast, statically typed, compiled
	// 5 One of Go's key design goals is code
	// 1
	// 1
	// 3 <empty>
	// 4 It's a fast, statically typed, compiled
	// 5 One of Go's key design goals is code
	// 3
}

func ExampleSes_PrepAndExe() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to %q: %v", dbName(), err)
		return
	}
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v (C1 NUMBER)", tableName))
	rowsAffected, _ := ses.PrepAndExe(fmt.Sprintf("INSERT INTO %v (C1) VALUES (3)", tableName))
	fmt.Println(rowsAffected)
	// Output:
	// 1
}

func ExampleSes_PrepAndQry() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v (C1 NUMBER)", tableName))
	ses.PrepAndExe(fmt.Sprintf("INSERT INTO %v (C1) VALUES (3)", tableName))
	rset, _ := ses.PrepAndQry(fmt.Sprintf("SELECT C1 FROM %v", tableName))
	row := rset.NextRow()
	fmt.Println(row[0])
	// Output:
	// 3
}

func ExampleSes_Ins() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ident := "DEFAULT 1"
	if ver, _ := srv.Version(); strings.Contains(ver, " 12.") {
		ident = "GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1)"
	}
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v "+
		"(C1 NUMBER(20,0) "+ident+", C2 NUMBER(20,10), C3 NUMBER(20,10), "+
		"C4 NUMBER(20,10), C5 NUMBER(20,10), C6 NUMBER(20,10), "+
		"C7 NUMBER(20,10), C8 NUMBER(20,10), C9 NUMBER(20,10), "+
		"C10 NUMBER(20,10), C11 NUMBER(20,10), C12 NUMBER(20,10), "+
		"C13 NUMBER(20,10), C14 NUMBER(20,10), C15 NUMBER(20,10), "+
		"C16 NUMBER(20,10), C17 NUMBER(20,10), C18 NUMBER(20,10), "+
		"C19 NUMBER(20,10), C20 NUMBER(20,10), C21 NUMBER(20,10))", tableName))
	e := &testEntity{}
	e.C2 = 2.2
	e.C3 = 3
	e.C4 = 4
	e.C5 = 5
	e.C6 = 6
	e.C7 = 7
	e.C8 = 8
	e.C9 = 9
	e.C10 = 10
	e.C11 = 11.11
	e.C12 = 12.12
	e.C13 = 13
	e.C14 = 14
	e.C15 = 15
	e.C16 = 16
	e.C17 = 17
	e.C18 = 18
	e.C19 = 19
	e.C20 = 20
	e.C21 = 21.21
	ses.Ins(tableName,
		"C2", e.C2,
		"C3", e.C3,
		"C4", e.C4,
		"C5", e.C5,
		"C6", e.C6,
		"C7", e.C7,
		"C8", e.C8,
		"C9", e.C9,
		"C10", e.C10,
		"C11", e.C11,
		"C12", e.C12,
		"C13", e.C13,
		"C14", e.C14,
		"C15", e.C15,
		"C16", e.C16,
		"C17", e.C17,
		"C18", e.C18,
		"C19", e.C19,
		"C20", e.C20,
		"C21", e.C21,
		"C1", &e.C1)
	fmt.Println(e.C1)
	// Output:
	// 1
}

func ExampleSes_Upd() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ident := "DEFAULT 1"
	if ver, _ := srv.Version(); strings.Contains(ver, " 12.") {
		ident = "GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1)"
	}
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v "+
		"(C1 NUMBER(20,0) "+ident+", C2 NUMBER(20,10), C3 NUMBER(20,10), "+
		"C4 NUMBER(20,10), C5 NUMBER(20,10), C6 NUMBER(20,10), "+
		"C7 NUMBER(20,10), C8 NUMBER(20,10), C9 NUMBER(20,10), "+
		"C10 NUMBER(20,10), C11 NUMBER(20,10), C12 NUMBER(20,10), "+
		"C13 NUMBER(20,10), C14 NUMBER(20,10), C15 NUMBER(20,10), "+
		"C16 NUMBER(20,10), C17 NUMBER(20,10), C18 NUMBER(20,10), "+
		"C19 NUMBER(20,10), C20 NUMBER(20,10), C21 NUMBER(20,10))", tableName))
	e := &testEntity{}
	e.C2 = 2.2
	e.C3 = 3
	e.C4 = 4
	e.C5 = 5
	e.C6 = 6
	e.C7 = 7
	e.C8 = 8
	e.C9 = 9
	e.C10 = 10
	e.C11 = 11.11
	e.C12 = 12.12
	e.C13 = 13
	e.C14 = 14
	e.C15 = 15
	e.C16 = 16
	e.C17 = 17
	e.C18 = 18
	e.C19 = 19
	e.C20 = 20
	e.C21 = 21.21
	ses.Ins(tableName,
		"C2", e.C2,
		"C3", e.C3,
		"C4", e.C4,
		"C5", e.C5,
		"C6", e.C6,
		"C7", e.C7,
		"C8", e.C8,
		"C9", e.C9,
		"C10", e.C10,
		"C11", e.C11,
		"C12", e.C12,
		"C13", e.C13,
		"C14", e.C14,
		"C15", e.C15,
		"C16", e.C16,
		"C17", e.C17,
		"C18", e.C18,
		"C19", e.C19,
		"C20", e.C20,
		"C21", e.C21,
		"C1", &e.C1)
	err := ses.Upd(tableName,
		"C2", e.C2*2,
		"C3", e.C3*2,
		"C4", e.C4*2,
		"C5", e.C5*2,
		"C6", e.C6*2,
		"C7", e.C7*2,
		"C8", e.C8*2,
		"C9", e.C9*2,
		"C10", e.C10*2,
		"C11", e.C11*2,
		"C12", e.C12*2,
		"C13", e.C13*2,
		"C14", e.C14*2,
		"C15", e.C15*2,
		"C16", e.C16*2,
		"C17", e.C17*2,
		"C18", e.C18*2,
		"C19", e.C19*2,
		"C20", e.C20*2,
		"C21", e.C21*2,
		"C1", e.C1)
	if err == nil {
		fmt.Println("success")
	}
	// Output:
	// success
}

func ExampleSes_Sel() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, _ := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	ident := "DEFAULT 1"
	if ver, _ := srv.Version(); strings.Contains(ver, " 12.") {
		ident = " GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1)"
	}
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v "+
		"(C1 NUMBER(20,0) "+ident+","+
		"C2 NUMBER(20,10), C3 NUMBER(20,10), "+
		"C4 NUMBER(20,10), C5 NUMBER(20,10), C6 NUMBER(20,10), "+
		"C7 NUMBER(20,10), C8 NUMBER(20,10), C9 NUMBER(20,10), "+
		"C10 NUMBER(20,10), C11 NUMBER(20,10), C12 NUMBER(20,10), "+
		"C13 NUMBER(20,10), C14 NUMBER(20,10), C15 NUMBER(20,10), "+
		"C16 NUMBER(20,10), C17 NUMBER(20,10), C18 NUMBER(20,10), "+
		"C19 NUMBER(20,10), C20 NUMBER(20,10), C21 NUMBER(20,10))", tableName))
	e := &testEntity{}
	e.C2 = 2.2
	e.C3 = 3
	e.C4 = 4
	e.C5 = 5
	e.C6 = 6
	e.C7 = 7
	e.C8 = 8
	e.C9 = 9
	e.C10 = 10
	e.C11 = 11.11
	e.C12 = 12.12
	e.C13 = 13
	e.C14 = 14
	e.C15 = 15
	e.C16 = 16
	e.C17 = 17
	e.C18 = 18
	e.C19 = 19
	e.C20 = 20
	e.C21 = 21.21
	ses.Ins(tableName,
		"C2", e.C2,
		"C3", e.C3,
		"C4", e.C4,
		"C5", e.C5,
		"C6", e.C6,
		"C7", e.C7,
		"C8", e.C8,
		"C9", e.C9,
		"C10", e.C10,
		"C11", e.C11,
		"C12", e.C12,
		"C13", e.C13,
		"C14", e.C14,
		"C15", e.C15,
		"C16", e.C16,
		"C17", e.C17,
		"C18", e.C18,
		"C19", e.C19,
		"C20", e.C20,
		"C21", e.C21,
		"C1", &e.C1)
	rset, _ := ses.Sel(tableName,
		"C1", ora.U64,
		"C2", ora.F64,
		"C3", ora.I8,
		"C4", ora.I16,
		"C5", ora.I32,
		"C6", ora.I64,
		"C7", ora.U8,
		"C8", ora.U16,
		"C9", ora.U32,
		"C10", ora.U64,
		"C11", ora.F32,
		"C12", ora.F64,
		"C13", ora.I8,
		"C14", ora.I16,
		"C15", ora.I32,
		"C16", ora.I64,
		"C17", ora.U8,
		"C18", ora.U16,
		"C19", ora.U32,
		"C20", ora.U64,
		"C21", ora.F32)
	for rset.Next() {
		for n := 0; n < len(rset.Row); n++ {
			fmt.Printf("R%v %v\n", n, rset.Row[n])
		}
	}
	// Output:
	//R0 1
	//R1 2.2
	//R2 3
	//R3 4
	//R4 5
	//R5 6
	//R6 7
	//R7 8
	//R8 9
	//R9 10
	//R10 11.11
	//R11 12.120000000000001
	//R12 13
	//R13 14
	//R14 15
	//R15 16
	//R16 17
	//R17 18
	//R18 19
	//R19 20
	//R20 21.21
}

type testEntity struct {
	C1  uint64
	C2  float64
	C3  int8
	C4  int16
	C5  int32
	C6  int64
	C7  uint8
	C8  uint16
	C9  uint32
	C10 uint64
	C11 float32
	C12 float64
	C13 int8
	C14 int16
	C15 int32
	C16 int64
	C17 uint8
	C18 uint16
	C19 uint32
	C20 uint64
	C21 float32
}

func emptyString(s string) string {
	if s == "" {
		return "<empty>"
	}
	return s
}

func ExampleSes_InsertBatchDirect() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to %q: %v", dbName(), err)
		return
	}
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	ses.PrepAndExe("CREATE TABLE " + tableName + " (C1 NUMBER)")
	rowsAffected, _ := ses.PrepAndExe("INSERT INTO "+tableName+" (C1) VALUES (:1)", []int64{1, 2})
	fmt.Println(rowsAffected)
	// Output:
	// 2
}

func ExampleSes_InsertBatchPlsql() {
	env, _ := ora.OpenEnv()
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to %q: %v", dbName(), err)
		return
	}
	defer srv.Close()
	ses, _ := srv.OpenSes(testSesCfg)
	defer ses.Close()
	tableName := tableName()
	procName := tableName + "_ins"
	ses.PrepAndExe("CREATE OR REPLACE PROCEDURE " + procName + `(p_num IN NUMBER) AS
BEGIN
  INSERT INTO ` + tableName + ` VALUES (p_num);
END;`)
	ses.PrepAndExe(fmt.Sprintf("CREATE TABLE %v (C1 NUMBER)", tableName))
	//ora.Cfg().Log.Logger = lg.Log
	if _, err = ses.PrepAndExe("BEGIN "+procName+"(:1); END;", []int64{1, 2}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	rset, err := ses.Sel(tableName, "C1", ora.U64)
	if err != nil {
		return
	}
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}
	// Output:
	// 1
	// 2
}
