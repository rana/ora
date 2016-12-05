// +build go1.8

//Copyright 2016 Tamás Gulácsi. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora

// NumInput returns the number of placeholders in a sql statement.
//
// This returns a constant -1, as named params can be less, then positional params.
func (stmt *Stmt) NumInput() int {
	if bindNames, _, duplicates, err := stmt.getBindInfo(); err == nil {
		n := len(bindNames)
		for _, d := range duplicates {
			if d {
				n--
			}
		}
		return n
	}
	return -1
}
