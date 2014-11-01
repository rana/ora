// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"fmt"
	"github.com/ranaian/ora"
)

func main() {
	// example usage of the oracle package driver
	// connect to a server and open a session
	env := ora.NewEnv()
	env.Open()
	defer env.Close()
	srv, err := env.OpenServer("orcl")
	defer srv.Close()
	if err != nil {
		panic(err)
	}
	ses, err := srv.OpenSession("test", "test")
	defer ses.Close()
	if err != nil {
		panic(err)
	}

	// create table
	stmtTbl, err := ses.Prepare("create table t1 " +
		"(c1 number(19,0) generated always as identity (start with 1 increment by 1), " +
		"c2 varchar2(48 char))")
	defer stmtTbl.Close()
	if err != nil {
		panic(err)
	}
	rowsAffected, err := stmtTbl.Execute()
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// begin first transaction
	tx1, err := ses.BeginTransaction()
	if err != nil {
		panic(err)
	}

	// insert record
	var id uint64
	str := "Go is expressive, concise, clean, and efficient."
	stmtIns, err := ses.Prepare("insert into t1 (c2) values (:c2) returning c1 into :c1")
	defer stmtIns.Close()
	rowsAffected, err = stmtIns.Execute(str, &id)
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// insert nullable String slice
	a := make([]ora.String, 4)
	a[0] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	a[1] = ora.String{IsNull: true}
	a[2] = ora.String{Value: "It's a fast, statically typed, compiled"}
	a[3] = ora.String{Value: "One of Go's key design goals is code"}
	stmtSliceIns, err := ses.Prepare("insert into t1 (c2) values (:c2)")
	defer stmtSliceIns.Close()
	if err != nil {
		panic(err)
	}
	rowsAffected, err = stmtSliceIns.Execute(a)
	if err != nil {
		panic(err)
	}
	fmt.Println(rowsAffected)

	// fetch records
	stmtFetch, err := ses.Prepare("select c1, c2 from t1")
	defer stmtFetch.Close()
	if err != nil {
		panic(err)
	}
	rst, err := stmtFetch.Fetch()
	if err != nil {
		panic(err)
	}
	for rst.Next() {
		fmt.Println(rst.Row[0], rst.Row[1])
	}
	if rst.Err != nil {
		panic(rst.Err)
	}

	// commit first transaction
	err = tx1.Commit()
	if err != nil {
		panic(err)
	}

	// begin second transaction
	tx2, err := ses.BeginTransaction()
	if err != nil {
		panic(err)
	}
	// insert null String
	nullableStr := ora.String{IsNull: true}
	stmtTrans, err := ses.Prepare("insert into t1 (c2) values (:c2)")
	defer stmtTrans.Close()
	if err != nil {
		panic(err)
	}
	rowsAffected, err = stmtTrans.Execute(nullableStr)
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
	stmtCount, err := ses.Prepare("select count(c1) from t1 where c2 is null", ora.U8)
	defer stmtCount.Close()
	if err != nil {
		panic(err)
	}
	rst, err = stmtCount.Fetch()
	if err != nil {
		panic(err)
	}
	row := rst.NextRow()
	if row != nil {
		fmt.Println(row[0])
	}
	if rst.Err != nil {
		panic(rst.Err)
	}

	// create stored procedure with sys_refcursor
	stmtProcCreate, err := ses.Prepare(
		"create or replace procedure proc1(p1 out sys_refcursor) as begin " +
			"open p1 for select c1, c2 from t1 where c1 > 2 order by c1; " +
			"end proc1;")
	defer stmtProcCreate.Close()
	rowsAffected, err = stmtProcCreate.Execute()
	if err != nil {
		panic(err)
	}

	// call stored procedure
	// pass *ResultSet to Execute to receive the results of a sys_refcursor
	stmtProcCall, err := ses.Prepare("call proc1(:1)")
	defer stmtProcCall.Close()
	if err != nil {
		panic(err)
	}
	procResultSet := &ora.ResultSet{}
	rowsAffected, err = stmtProcCall.Execute(procResultSet)
	if err != nil {
		panic(err)
	}
	if procResultSet.IsOpen() {
		for procResultSet.Next() {
			fmt.Println(procResultSet.Row[0], procResultSet.Row[1])
		}
		if procResultSet.Err != nil {
			panic(procResultSet.Err)
		}
		fmt.Println(procResultSet.Len())
	}

	// Output:
	// 0
	// 1
	// 4
	// 1 Go is expressive, concise, clean, and efficient.
	// 2 Its concurrency mechanisms make it easy to
	// 3 <nil>
	// 4 It's a fast, statically typed, compiled
	// 5 One of Go's key design goals is code
	// 1
	// 1
	// 3 <nil>
	// 4 It's a fast, statically typed, compiled
	// 5 One of Go's key design goals is code
	// 3
}
