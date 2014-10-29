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
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (float64Define *float64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&float64Define.ocidef,                    //OCIDefine   **defnpp,
		float64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&float64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8         value_sz,
		C.SQLT_VNU,                               //ub2         dty,
		unsafe.Pointer(&float64Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return float64Define.environment.ociError()
	}
	return nil
}
func (float64Define *float64Define) value() (value interface{}, err error) {
	if float64Define.isNull > -1 {
		var float64Value float64
		r := C.OCINumberToReal(
			float64Define.environment.ocierr, //OCIError              *err,
			&float64Define.ociNumber,         //const OCINumber     *number,
			C.uword(8),                       //uword               rsl_length,
			unsafe.Pointer(&float64Value))    //void                *rsl );
		if r == C.OCI_ERROR {
			err = float64Define.environment.ociError()
		}
		value = float64Value
	}
	return value, err
}
func (float64Define *float64Define) alloc() error {
	return nil
}
func (float64Define *float64Define) free() {

}
func (float64Define *float64Define) close() {
	defer func() {
		recover()
	}()
	float64Define.ocidef = nil
	float64Define.isNull = C.sb2(0)
	float64Define.environment.float64DefinePool.Put(float64Define)
}
