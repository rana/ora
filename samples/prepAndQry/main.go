// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"fmt"
	"github.com/ranaian/ora"
)

// Ses.PrepAndQry offers a convenient one-line call to Ses.Prep and Stmt.Qry.
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
	rset, err := ses.PrepAndQry("SELECT CURRENT_TIMESTAMP FROM DUAL")
	if err != nil {
		panic(err)
	}
	row := rset.NextRow()
	if row != nil && len(row) > 0 {
		fmt.Println("CURRENT_TIMESTAMP: ", row[0])
	}
}
