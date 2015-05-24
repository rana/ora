// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"flag"
	"github.com/ranaian/ora"
	"github.com/ranaian/ora/glg"
)

func main() {

	// parse flags for glog (required)
	// consider specifying cmd line arg -alsologtostderr=true
	flag.Parse()

	// use the optional glog package for ora logging
	ora.Log = glg.Log

	env, _ := ora.GetDrv().OpenEnv()
	defer env.Close()
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
