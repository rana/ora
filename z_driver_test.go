// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"testing"
)

func TestDriver_OpenCloseCon(t *testing.T) {
	drv := GetDrv()
	con, err := drv.Open(testConStr)
	defer con.Close()
	testErr(err, t)
	if con == nil {
		t.Fatal("Con is nil")
	}
}
