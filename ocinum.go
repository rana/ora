// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
//import "C"
import "gopkg.in/rana/ora.v3/num"
import "database/sql/driver"

type OCINum struct {
	num.OCINum
}

// Value implements database/sql/driver's Valuer interface to return the number as string.
func (num OCINum) Value() (driver.Value, error) {
	return num.String(), nil
}
