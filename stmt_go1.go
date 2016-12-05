// +build !go1.8

//Copyright 2016 Tamás Gulácsi. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
import "C"

// NumInput returns the number of placeholders in a sql statement.
func (stmt *Stmt) NumInput() int {
	bc, err := stmt.attr(4, C.OCI_ATTR_BIND_COUNT)
	if err != nil {
		return 0
	}
	bindCount := int(*((*C.ub4)(bc)))
	C.free(bc)
	return bindCount
}

func nameAndValue(v interface{}) (string, interface{}) {
	return "", v
}
