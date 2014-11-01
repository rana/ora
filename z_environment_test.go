// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"testing"
)

func TestEnvironment_OpenClose(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	testErr(err, t)
	err = env.Close()
	testErr(err, t)
}

func TestEnvironment_IsOpen_unopened(t *testing.T) {
	env := NewEnv()
	var expected bool = false
	var actual bool = env.IsOpen()
	if actual != expected {
		t.Fatalf("actual(%v), expected(%v)", actual, expected)
	}
}

func TestEnvironment_IsOpen_opened(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)

	var expected bool = true
	var actual bool = env.IsOpen()
	if actual != expected {
		t.Fatalf("actual(%v), expected(%v)", actual, expected)
	}
}

func TestEnvironment_IsOpen_opened_closed(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	testErr(err, t)
	err = env.Close()
	testErr(err, t)

	var expected bool = false
	var actual bool = env.IsOpen()
	if actual != expected {
		t.Fatalf("actual(%v), expected(%v)", actual, expected)
	}
}

func TestEnvironment_OpenCloseServer(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)

	srv, err := env.OpenServer(testServerName)
	testErr(err, t)

	err = srv.Close()
	testErr(err, t)
}

func TestEnvironment_OpenCloseConnection(t *testing.T) {
	env := NewEnv()
	err := env.Open()
	defer env.Close()
	testErr(err, t)

	conn, err := env.OpenConnection(testConnectionStr)
	testErr(err, t)

	err = conn.Close()
	testErr(err, t)
}
