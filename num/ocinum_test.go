// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package num

import (
	"bytes"
	"testing"
)

var testNums = []struct {
	await string
	num   []byte
}{
	{"0", []byte{128}},
	{"1", []byte{193, 2}},
	{"-1", []byte{62, 100, 102}},
	{"12", []byte{193, 13}},
	{"-12", []byte{62, 89, 102}},
	{"123", []byte{194, 2, 24}},
	{"-123", []byte{61, 100, 78, 102}},
	{"123456789012345678901234567890123456789", []byte{212, 2, 24, 46, 68, 90, 2, 24, 46, 68, 90, 2, 24, 46, 68, 90, 2, 24, 46, 68, 90}},
	{"-123456789012345678901234567890123456789", []byte{43, 100, 78, 56, 34, 12, 100, 78, 56, 34, 12, 100, 78, 56, 34, 12, 100, 78, 56, 34, 12}},

	{"1000", []byte{194, 11}},
	{"-1000", []byte{61, 91, 102}},
	{"0.1", []byte{192, 11}},
	{"-0.1", []byte{63, 91, 102}},
	{"0.01", []byte{192, 2}},
	{"-0.01", []byte{63, 100, 102}},
	{"0.12", []byte{192, 13}},
	{"-0.12", []byte{63, 89, 102}},
	{"0.012", []byte{192, 2, 21}},
	{"-0.012", []byte{63, 100, 81, 102}},
}

func TestOCINumPrint(t *testing.T) {
	var b []byte
	for eltNum, elt := range testNums {
		b = OCINum(elt.num).Print(b)
		if !bytes.Equal(b, []byte(elt.await)) {
			t.Errorf("%d. % v\ngot\n\t%s (% v)\nawaited\n\t%s (% v).", eltNum, elt.num, b, b, elt.await, []byte(elt.await))
		}
	}
}

func TestOCINumSet(t *testing.T) {
	var num OCINum
	for eltNum, elt := range testNums {
		if err := num.SetString(elt.await); err != nil {
			t.Errorf("%d. %s: %v", eltNum, elt.await, err)
			continue
		}
		if !bytes.Equal(num, elt.num) {
			t.Errorf("%d. %s:\ngot\n\t%v\nawaited\n\t%v", eltNum, elt.await, []byte(num), elt.num)
		}
	}
}
