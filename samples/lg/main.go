// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"github.com/ranaian/ora"
	"github.com/ranaian/ora/lg"
)

// Sample logging produced using the standard Go log package
//
//	ORA I 2015/05/23 16:54:44.615462 drv.go:411: OpenEnv 1
//	ORA I 2015/05/23 16:54:44.626443 drv.go:411: OpenEnv 2
//	ORA I 2015/05/23 16:54:44.627465 env.go:115: E2] OpenSrv (dbname orcl)
//	ORA I 2015/05/23 16:54:44.643449 env.go:150: E2] OpenSrv (srvId 1)
//	ORA I 2015/05/23 16:54:44.643449 srv.go:113: E2S1] OpenSes (username test)
//	ORA I 2015/05/23 16:54:44.665451 ses.go:163: E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL
//	ORA I 2015/05/23 16:54:44.666451 rset.go:205: E2S1S1S1R0] open
//	ORA I 2015/05/23 16:54:44.666451 ses.go:74: E2S1S1] Close
//	ORA I 2015/05/23 16:54:44.666451 stmt.go:78: E2S1S1S1] Close
//	ORA I 2015/05/23 16:54:44.666451 rset.go:57: E2S1S1S1R0] close
//	ORA I 2015/05/23 16:54:44.666451 srv.go:63: E2S1] Close
//	ORA I 2015/05/23 16:54:44.667451 env.go:68: E2] Close
//
func main() {

	// use the optional lg package for ora logging
	ora.Log = lg.Log

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
