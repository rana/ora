// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"fmt"
	"github.com/rana/ora"
)

// Ses.Sel offers a convenient one-line call to Ses.Prep and Stmt.Qry.
//
func main() {
	env, err := ora.GetDrv().OpenEnv()
	defer env.Close()
	if err != nil {
		panic(err)
	}
	srv, err := env.OpenSrv("orcl")
	defer srv.Close()
	if err != nil {
		panic(err)
	}
	ses, err := srv.OpenSes("test", "test")
	defer ses.Close()
	if err != nil {
		panic(err)
	}
	_, err = ses.PrepAndExe("CREATE TABLE T1 " +
		"(C1 NUMBER(20,0) GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1), C2 NUMBER(20,10), C3 NUMBER(20,10), " +
		"C4 NUMBER(20,10), C5 NUMBER(20,10), C6 NUMBER(20,10), " +
		"C7 NUMBER(20,10), C8 NUMBER(20,10), C9 NUMBER(20,10), " +
		"C10 NUMBER(20,10), C11 NUMBER(20,10), C12 NUMBER(20,10), " +
		"C13 NUMBER(20,10), C14 NUMBER(20,10), C15 NUMBER(20,10), " +
		"C16 NUMBER(20,10), C17 NUMBER(20,10), C18 NUMBER(20,10), " +
		"C19 NUMBER(20,10), C20 NUMBER(20,10), C21 NUMBER(20,10))")
	if err != nil {
		panic(err)
	}
	e := &Entity{}
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
	err = ses.Ins("T1",
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
	if err != nil {
		panic(err)
	}
	rset, err := ses.Sel("T1",
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

type Entity struct {
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
