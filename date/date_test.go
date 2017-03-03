// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package date_test

import (
	"bytes"
	"testing"

	"gopkg.in/rana/ora.v4/date"
)

func TestDate(t *testing.T) {
	format := "2006-01-02T15:04:05"
	dt := new(date.Date)
	for tN, tC := range dateTestData {
		tim := date.Date(tC.B).Get()
		if got := tim.Format(format); got != tC.S {
			t.Errorf("%d. got %q, want %q (from %v).", tN, got, tC.S, tC.B)
			continue
		}
		dt.Set(tim)
		if !bytes.Equal(dt[:], tC.B[:]) {
			t.Errorf("%d. got %v, want %v (from %q).", tN, dt[:], tC.B[:], tC.S)
		}
	}
}
func TestNull(t *testing.T) {
	var dt date.Date
	t.Log(dt.String())
	if !dt.IsNull() {
		t.Errorf("want NULL, got %t for %#v", dt.IsNull(), dt)
	}
}
