// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"github.com/ranaian/ora"
	"github.com/ranaian/ora/lg15"
)

// Sample logging produced by log15
//
//	t=2015-05-23T17:08:32-0700 lvl=info msg="OpenEnv 1" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="OpenEnv 2" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] OpenSrv (dbname orcl)" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] OpenSrv (srvId 1)" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1] OpenSes (username test)" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1R0] open" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1] Close" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1] Close" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1R0] close" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1] Close" lib=ora
//	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] Close" lib=ora
//
func main() {

	// use the optional log15 package for ora logging
	ora.Log = lg15.Log

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
