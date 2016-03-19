// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package num

//go:generate go-fuzz-build gopkg.in/rana/ora.v3/num

// Fuzz:
// go-fuzz -bin=./num-fuzz.zip -workdir=/tmp/fuzz
func Fuzz(p []byte) int {
	pS := string(p)
	var q [22]byte
	n := OCINum(q[:0])
	if err := n.SetString(pS); err != nil {
		return 0
	}
	s := n.String()
	if s != pS {
		return 1
	}
	return 0
}
