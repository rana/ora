// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

type float64Define struct {
	env       *Environment
	ocidef    *C.OCIDefine
	ociNumber C.OCINumber
	isNull    C.sb2
}

func (d *float64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),    //sb8         value_sz,
		C.SQLT_VNU,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *float64Define) value() (value interface{}, err error) {
	if d.isNull > -1 {
		var float64Value float64
		r := C.OCINumberToReal(
			d.env.ocierr,                  //OCIError              *err,
			&d.ociNumber,                  //const OCINumber     *number,
			C.uword(8),                    //uword               rsl_length,
			unsafe.Pointer(&float64Value)) //void                *rsl );
		if r == C.OCI_ERROR {
			err = d.env.ociError()
		}
		value = float64Value
	}
	return value, err
}
func (d *float64Define) alloc() error {
	return nil
}
func (d *float64Define) free() {

}
func (d *float64Define) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.isNull = C.sb2(0)
	d.env.float64DefinePool.Put(d)
}
