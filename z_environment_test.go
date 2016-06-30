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

func TestEnv_SrvCfg(t *testing.T) {
	env, err := ora.OpenEnv(nil)
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

	srvCfg := srv.Cfg()
	old := srvCfg.StmtCfg.Rset.NumberBigFloat()
	defer srvCfg.StmtCfg.Rset.SetNumberBigFloat(old)

	x := ora.F64
	srvCfg.StmtCfg.Rset.SetNumberBigFloat(x)
	if y := srvCfg.StmtCfg.Rset.NumberBigFloat(); y != x {
		t.Errorf("srvCfg: wanted %s, got %s", x, y)
	}

	stmt, err := ses.Prep("SELECT COUNT(0) FROM user_objects")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if y := stmt.Cfg().Rset.NumberBigFloat(); y != x {
		t.Errorf("stmt.Cfg: wanted %v, got %s (default: %s)", x, y, old)
	}
}

func TestEnv_SesCfg(t *testing.T) {
	env, err := ora.OpenEnv(nil)
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
	old := sesCfg.StmtCfg.Rset.NumberBigFloat()
	defer sesCfg.StmtCfg.Rset.SetNumberBigFloat(old)

	x := ora.F64
	sesCfg.StmtCfg.Rset.SetNumberBigFloat(x)
	if y := sesCfg.StmtCfg.Rset.NumberBigFloat(); y != x {
		t.Errorf("srvCfg: wanted %s, got %s", x, y)
	}

	stmt, err := ses.Prep("SELECT COUNT(0) FROM user_objects")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if y := stmt.Cfg().Rset.NumberBigFloat(); y != x {
		t.Errorf("stmt.Cfg: wanted %v, got %s (default: %s)", x, y, old)
	}
}
