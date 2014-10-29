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

type float32Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (float32Define *float32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&float32Define.ocidef,                    //OCIDefine   **defnpp,
		float32Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&float32Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8         value_sz,
		C.SQLT_VNU,                               //ub2         dty,
		unsafe.Pointer(&float32Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return float32Define.environment.ociError()
	}
	return nil
}
func (float32Define *float32Define) value() (value interface{}, err error) {
	if float32Define.isNull > -1 {
		var float32Value float32
		r := C.OCINumberToReal(
			float32Define.environment.ocierr, //OCIError              *err,
			&float32Define.ociNumber,         //const OCINumber     *number,
			C.uword(4),                       //uword               rsl_length,
			unsafe.Pointer(&float32Value))    //void                *rsl );
		if r == C.OCI_ERROR {
			err = float32Define.environment.ociError()
		}
		value = float32Value
	}
	return value, err
}
func (float32Define *float32Define) alloc() error {
	return nil
}
func (float32Define *float32Define) free() {

}
func (float32Define *float32Define) close() {
	defer func() {
		recover()
	}()
	float32Define.ocidef = nil
	float32Define.isNull = C.sb2(0)
	float32Define.environment.float32DefinePool.Put(float32Define)
}
