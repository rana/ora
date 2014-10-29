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

type oraInt8Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraInt8Define *oraInt8Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&oraInt8Define.ocidef,                    //OCIDefine   **defnpp,
		oraInt8Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&oraInt8Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8         value_sz,
		C.SQLT_VNU,                               //ub2         dty,
		unsafe.Pointer(&oraInt8Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraInt8Define.environment.ociError()
	}
	return nil
}

func (oraInt8Define *oraInt8Define) value() (value interface{}, err error) {
	int8Value := Int8{IsNull: oraInt8Define.isNull < 0}
	if !int8Value.IsNull {
		r := C.OCINumberToInt(
			oraInt8Define.environment.ocierr, //OCIError              *err,
			&oraInt8Define.ociNumber,         //const OCINumber       *number,
			C.uword(1),                       //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(&int8Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraInt8Define.environment.ociError()
		}
	}
	value = int8Value
	return value, err
}

func (oraInt8Define *oraInt8Define) alloc() error {
	return nil
}

func (oraInt8Define *oraInt8Define) free() {

}

func (oraInt8Define *oraInt8Define) close() {
	defer func() {
		recover()
	}()
	oraInt8Define.ocidef = nil
	oraInt8Define.isNull = C.sb2(0)
	oraInt8Define.environment.oraInt8DefinePool.Put(oraInt8Define)
}
