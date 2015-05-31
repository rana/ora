// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"flag"
	"github.com/rana/ora"
	"github.com/rana/ora/glg"
)

// Sample logging produced by glog
//
//	I0523 17:31:41.702365   97708 drv.go:411] OpenEnv 1
//	I0523 17:31:41.728377   97708 drv.go:411] OpenEnv 2
//	I0523 17:31:41.728377   97708 env.go:115] E2] OpenSrv (dbname orcl)
//	I0523 17:31:41.741390   97708 env.go:150] E2] OpenSrv (srvId 1)
//	I0523 17:31:41.741390   97708 srv.go:113] E2S1] OpenSes (username test)
//	I0523 17:31:41.762366   97708 ses.go:163] E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL
//	I0523 17:31:41.762366   97708 rset.go:205] E2S1S1S1R0] open
//	I0523 17:31:41.762366   97708 ses.go:74] E2S1S1] Close
//	I0523 17:31:41.762366   97708 stmt.go:78] E2S1S1S1] Close
//	I0523 17:31:41.762366   97708 rset.go:57] E2S1S1S1R0] close
//	I0523 17:31:41.763365   97708 srv.go:63] E2S1] Close
//	I0523 17:31:41.763365   97708 env.go:68] E2] Close
//
func main() {

	// parse flags for glog (required)
	// consider specifying cmd line arg -alsologtostderr=true
	flag.Parse()

	// use the optional glog package for ora logging
	ora.Log = glg.Log

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
	_, err = ses.PrepAndQry("SELECT CURRENT_TIMESTAMP FROM DUAL")
}
