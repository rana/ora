// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"testing"
)

func TestServer_OpenCloseSession(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenServer(testServerName)
	defer srv.Close()
	testErr(err, t)

	ses, err := srv.OpenSession(testUsername, testPassword)
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
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenServer(testServerName)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSession(testUsername, testPassword)
	defer ses.Close()
	testErr(err, t)

	err = srv.Ping()
	testErr(err, t)
}

func TestServer_Version(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)
	srv, err := env.OpenServer(testServerName)
	defer srv.Close()
	testErr(err, t)
	ses, err := srv.OpenSession(testUsername, testPassword)
	defer ses.Close()
	testErr(err, t)

	version, err := srv.Version()
	testErr(err, t)
	if version == "" {
		t.Fatal("Version is empty.")
	}
}
