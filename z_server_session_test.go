// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"gopkg.in/rana/ora.v4"
)

func TestServer_OpenCloseSession(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
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
	t.Parallel()
	env, err := ora.OpenEnv()
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
	t.Parallel()
	env, err := ora.OpenEnv()
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
	t.Parallel()
	env, err := ora.OpenEnv()
	testErr(err, t)
	defer env.Close()
	const idleSize = 2
	pool := env.NewPool(testSrvCfg, testSesCfg, idleSize)
	defer pool.Close()

	getCounts := func() (p, s map[string]struct{}) {
		ses, err := pool.Get()
		testErr(err, t)
		defer pool.Put(ses)
		for _, tbl := range []string{"v$process", "v$session"} {
			fld := "addr"
			if tbl == "v$session" {
				fld = "paddr"
			}
			rset, err := ses.PrepAndQry("SELECT " + fld + " FROM " + tbl)
			if err != nil {
				t.Log(err)
				continue
			}
			addrs := make(map[string]struct{}, 128)
			for rset.Next() {
				addrs[string(rset.Row[0].([]uint8))] = struct{}{}
			}
			if tbl == "v$session" {
				s = addrs
			} else {
				p = addrs
			}
		}
		return
	}

	diffCounts := func(a, b map[string]struct{}) int {
		var n int
		for k := range b {
			if _, ok := a[k]; !ok {
				n++
			}
		}
		return n
	}

	var wg sync.WaitGroup
	p1, s1 := getCounts()
	for i := 0; i < 2*idleSize+1; i++ {
		wg.Add(1)
		go func(c bool) {
			defer wg.Done()
			ses, err := pool.Get()
			testErr(err, t)
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			if c {
				ses.Close()
			} else {
				pool.Put(ses)
			}
		}(i%2 == 0)
	}
	wg.Wait()

	T := func(name string, p1, s1 map[string]struct{}) (p2, s2 map[string]struct{}) {
		p2, s2 = getCounts()
		dp, ds := diffCounts(p1, p2), diffCounts(s1, s2)
		t.Logf("%s: (%d,%d) -> (%d,%d)", name, len(p1), len(s1), len(p2), len(s2))
		if dp > 2*idleSize {
			t.Errorf("%s process count went up %d!", name, dp)
		}
		if ds > 2*idleSize {
			t.Errorf("%s session count went up %d!", name, ds)
		}
		return p2, s2
	}

	p2, s2 := T("After work", p1, s1)

	pool.Close()
	T("Pool close", p2, s2)
}
