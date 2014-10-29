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
	//	"fmt"
	"unsafe"
)

type oraFloat64Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraFloat64Define *oraFloat64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                     //OCIStmt     *stmtp,
		&oraFloat64Define.ocidef,                    //OCIDefine   **defnpp,
		oraFloat64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                             //ub4         position,
		unsafe.Pointer(&oraFloat64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                   //sb8         value_sz,
		C.SQLT_VNU,                                  //ub2         dty,
		unsafe.Pointer(&oraFloat64Define.isNull), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraFloat64Define.environment.ociError()
	}
	return nil
}

func (oraFloat64Define *oraFloat64Define) value() (value interface{}, err error) {
	float64Value := Float64{IsNull: oraFloat64Define.isNull < 0}
	if !float64Value.IsNull {
		r := C.OCINumberToReal(
			oraFloat64Define.environment.ocierr, //OCIError              *err,
			&oraFloat64Define.ociNumber,         //const OCINumber     *number,
			C.uword(8),                          //uword               rsl_length,
			unsafe.Pointer(&float64Value.Value)) //void                *rsl );
		if r == C.OCI_ERROR {
			err = oraFloat64Define.environment.ociError()
		}
	}
	value = float64Value
	return value, err
}

func (oraFloat64Define *oraFloat64Define) alloc() error {
	return nil
}

func (oraFloat64Define *oraFloat64Define) free() {

}

func (oraFloat64Define *oraFloat64Define) close() {
	defer func() {
		recover()
	}()
	oraFloat64Define.ocidef = nil
	oraFloat64Define.isNull = C.sb2(0)
	oraFloat64Define.environment.oraFloat64DefinePool.Put(oraFloat64Define)
}
