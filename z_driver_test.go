// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"testing"
)

func TestDriver_OpenCloseConnection(t *testing.T) {
	drv := &Driver{env: NewEnv()}
	defer drv.env.Close()
	conn, err := drv.Open(testConnectionStr)
	defer conn.Close()
	testErr(err, t)
	if conn == nil {
		t.Fatal("connection is nil")
	}
}
