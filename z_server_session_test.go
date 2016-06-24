// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
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
	pool := env.NewPool(testSrvCfg, testSesCfg, 4)
	defer pool.Close()

	for i := 0; i < 100; i++ {
		ses, err := pool.Get()
		testErr(err, t)
		pool.Put(ses)
	}

	for i := 0; i < 100; i++ {
		ses, err := pool.Get()
		testErr(err, t)
		ses.Close()
	}
}
