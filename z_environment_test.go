// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v3"
)

func TestEnv_OpenClose(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	err = env.Close()
	testErr(err, t)
}

func TestEnv_IsOpen_opened(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	defer env.Close()
	testErr(err, t)

	var expected bool = true
	var actual bool = env.IsOpen()
	if actual != expected {
		t.Fatalf("actual(%v), expected(%v)", actual, expected)
	}
}

func TestEnv_IsOpen_opened_closed(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	testErr(err, t)
	err = env.Close()
	testErr(err, t)

	var expected bool = false
	var actual bool = env.IsOpen()
	if actual != expected {
		t.Fatalf("actual(%v), expected(%v)", actual, expected)
	}
}

func TestEnv_OpenCloseServer(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	defer env.Close()
	testErr(err, t)

	srv, err := env.OpenSrv(testSrvCfg)
	testErr(err, t)

	err = srv.Close()
	testErr(err, t)
}

func TestEnv_OpenCloseCon(t *testing.T) {
	env, err := ora.OpenEnv(nil)
	testErr(err, t)
	defer env.Close()
	testErr(err, t)

	conn, err := env.OpenCon(testConStr)
	testErr(err, t)

	err = conn.Close()
	testErr(err, t)
}
