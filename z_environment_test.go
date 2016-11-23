// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v4"
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

	x := ora.F64
	srvCfg.StmtCfg = srvCfg.StmtCfg.SetNumberBigFloat(x)
	srv.SetCfg(srvCfg)
	if y := srvCfg.NumberBigFloat(); y != x {
		t.Fatalf("srvCfg: wanted %s, got %s (%v)", x, y, srvCfg.Err)
	}
	sesCfg := testSesCfg
	sesCfg.StmtCfg = srvCfg.StmtCfg
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ses.Close()
	if y := ses.Cfg().NumberBigFloat(); y != x {
		t.Fatalf("sesCfg: wanted %s, got %s", x, y)
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

	old := ses.Cfg()
	defer ses.SetCfg(old)

	x := ora.F64
	sesCfg := old
	//enableLogging(t)
	if sesCfg = sesCfg.SetNumberBigFloat(x); sesCfg.Err != nil {
		t.Fatal(err)
	}
	ses.SetCfg(sesCfg)
	if y := ses.Cfg().NumberBigFloat(); y != x {
		t.Fatalf("sesCfg: wanted %s, got %s", x, y)
	}
	t.Logf(" sesCfg=%#v", ses.Cfg())
	stmt, err := ses.Prep("SELECT COUNT(0) FROM user_objects")
	t.Logf("stmtCfg=%#v", stmt.Cfg())
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if y := stmt.Cfg().NumberBigFloat(); y != x {
		t.Errorf("stmt.Cfg: wanted %s=%d, got %s=%d (default: %s)", x, x, y, y, old.NumberBigFloat())
	}
}
