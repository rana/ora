// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"fmt"
	"github.com/ranaian/ora"
)

// Ses.PrepAndExe offers a convenient one-line call to Ses.Prep and Stmt.Exe.
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
	rowsAffected, err := ses.PrepAndExe("CREATE TABLE T1 (C1 NUMBER)")
	if err != nil {
		panic(err)
	}
	rowsAffected, err = ses.PrepAndExe("INSERT INTO T1 (C1) VALUES (3)")
	if err != nil {
		panic(err)
	}
	fmt.Println("rowsAffected: ", rowsAffected)
}
