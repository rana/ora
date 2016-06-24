// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"gopkg.in/rana/ora.v3"
)

func TestServer_OpenCloseSession(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	testErr(err, t)

	ses, err := srv.OpenSes(testSesCfg)
	testErr(err, t)
	if ses == nil {
		t.Fatal("session is nil")
	} else {
		err = ses.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestServer_Ping(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSes(testSesCfg)
	defer ses.Close()
	testErr(err, t)

	err = ses.Ping()
	testErr(err, t)
}

func TestServer_Version(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenSrv(testSrvCfg)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSes(testSesCfg)
	defer ses.Close()
	testErr(err, t)

	version, err := srv.Version()
	testErr(err, t)
	if version == "" {
		t.Fatal("Version is empty.")
	}
}

func TestPool(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	defer env.Close()
	const idleSize = 2
	pool := env.NewPool(testSrvCfg, testSesCfg, idleSize)
	defer pool.Close()

	getProcCount := func() int {
		ses, err := pool.Get()
		testErr(err, t)
		defer pool.Put(ses)
		rset, err := ses.PrepAndQry("SELECT COUNT(0) FROM v$process")
		if err != nil {
			t.Log(err)
			return -1
		}
		rset.Next()
		var c int
		switch x := rset.Row[0].(type) {
		case float64:
			c = int(x)
		case ora.OCINum:
			c, _ = strconv.Atoi(x.String())
		default:
			c, _ = strconv.Atoi(fmt.Sprintf("%v", x))
		}
		for rset.Next() {
		}
		return c
	}

	var wg sync.WaitGroup
	c1 := getProcCount()
	for i := 0; i < 2*idleSize+1; i++ {
		wg.Add(1)
		go func(c bool) {
			defer wg.Done()
			ses, err := pool.Get()
			testErr(err, t)
			if c {
				ses.Close()
			} else {
				pool.Put(ses)
			}
		}(i%2 == 0)
	}
	wg.Wait()

	c2 := getProcCount()
	t.Logf("c1=%d c2=%d", c1, c2)
	if c2-c1 > 2 {
		t.Errorf("process count went to %d from %d!", c2, c1)
	}
}
