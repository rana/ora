// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v3"
)

func TestEnv_OpenClose(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
	testErr(err, t)
	err = env.Close()
	testErr(err, t)
}

func TestEnv_IsOpen_opened(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
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
	t.Parallel()
	env, err := ora.OpenEnv()
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
	t.Parallel()
	env, err := ora.OpenEnv()
	testErr(err, t)
	defer env.Close()
	testErr(err, t)

	srv, err := env.OpenSrv(testSrvCfg)
	testErr(err, t)

	err = srv.Close()
	testErr(err, t)
}

func TestEnv_OpenCloseCon(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
	testErr(err, t)
	defer env.Close()
	testErr(err, t)

	conn, err := env.OpenCon(testConStr)
	testErr(err, t)

	err = conn.Close()
	testErr(err, t)
}

func TestEnv_SrvCfg(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	srvCfg := srv.Cfg()
	old := srvCfg.NumberBigFloat()
	defer srvCfg.SetNumberBigFloat(old)

	x := ora.F64
	srvCfg.SetNumberBigFloat(x)
	if y := srvCfg.NumberBigFloat(); y != x {
		t.Errorf("srvCfg: wanted %s, got %s", x, y)
	}
	sesCfg := testSesCfg
	sesCfg.StmtCfg = ora.NewStmtCfg()
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Close()
	if y := ses.Cfg().NumberBigFloat(); y != x {
		t.Errorf("sesCfg: wanted %s, got %s", x, y)
	}

	stmt, err := ses.Prep("SELECT COUNT(0) FROM user_objects")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if y := stmt.Cfg().NumberBigFloat(); y != x {
		t.Errorf("stmt.Cfg: wanted %v, got %s (default: %s)", x, y, old)
	}
}

func TestEnv_SesCfg(t *testing.T) {
	t.Parallel()
	env, err := ora.OpenEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer env.Close()
	srv, err := env.OpenSrv(testSrvCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ses, err := srv.OpenSes(testSesCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Close()

	sesCfg := ses.Cfg()
	old := sesCfg.NumberBigFloat()
	defer sesCfg.SetNumberBigFloat(old)

	x := ora.F64
	sesCfg.SetNumberBigFloat(x)
	if y := sesCfg.NumberBigFloat(); y != x {
		t.Errorf("srvCfg: wanted %s, got %s", x, y)
	}
	ses.SetCfg(sesCfg)

	stmt, err := ses.Prep("SELECT COUNT(0) FROM user_objects")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if y := stmt.Cfg().NumberBigFloat(); y != x {
		t.Errorf("stmt.Cfg: wanted %v, got %s (default: %s)", x, y, old)
	}
}
