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

type oraFloat32Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraFloat32Define *oraFloat32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                     //OCIStmt     *stmtp,
		&oraFloat32Define.ocidef,                    //OCIDefine   **defnpp,
		oraFloat32Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                             //ub4         position,
		unsafe.Pointer(&oraFloat32Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                   //sb8         value_sz,
		C.SQLT_VNU,                                  //ub2         dty,
		unsafe.Pointer(&oraFloat32Define.isNull), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraFloat32Define.environment.ociError()
	}
	return nil
}

func (oraFloat32Define *oraFloat32Define) value() (value interface{}, err error) {
	float32Value := Float32{IsNull: oraFloat32Define.isNull < 0}
	if !float32Value.IsNull {
		r := C.OCINumberToReal(
			oraFloat32Define.environment.ocierr, //OCIError              *err,
			&oraFloat32Define.ociNumber,         //const OCINumber     *number,
			C.uword(4),                          //uword               rsl_length,
			unsafe.Pointer(&float32Value.Value)) //void                *rsl );
		if r == C.OCI_ERROR {
			err = oraFloat32Define.environment.ociError()
		}
	}
	value = float32Value
	return value, err
}

func (oraFloat32Define *oraFloat32Define) alloc() error {
	return nil
}

func (oraFloat32Define *oraFloat32Define) free() {

}

func (oraFloat32Define *oraFloat32Define) close() {
	defer func() {
		recover()
	}()
	oraFloat32Define.ocidef = nil
	oraFloat32Define.isNull = C.sb2(0)
	oraFloat32Define.environment.oraFloat32DefinePool.Put(oraFloat32Define)
}
