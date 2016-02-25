// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import "testing"

// TestNumericColumnType tests RsetCfg.numericColumnType.
func TestNumericColumnType(t *testing.T) {
	c := NewRsetCfg()
	// be exact
	c.float, c.numberFloat, c.numberInt = F32, F64, I64
	for i, tc := range []struct {
		typ, precision, scale int
		want                  GoColumnType
	}{
		{2, 6, 3, F64},
		{2, 3, 0, I64},
		{100, 0, -127, F32},
		{2, 1, 0, I64},
		{2, 0, -127, N},
	} {
		got := c.numericColumnType(tc.typ, tc.precision, tc.scale)
		if got != tc.want {
			t.Errorf("%d. (%d,%d,%d) got %s, want %s.",
				i, tc.typ, tc.precision, tc.scale, GctName(got), GctName(tc.want))
		}
	}
}
