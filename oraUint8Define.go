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

type oraUint8Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraUint8Define *oraUint8Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                   //OCIStmt     *stmtp,
		&oraUint8Define.ocidef,                    //OCIDefine   **defnpp,
		oraUint8Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                           //ub4         position,
		unsafe.Pointer(&oraUint8Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                 //sb8         value_sz,
		C.SQLT_VNU,                                //ub2         dty,
		unsafe.Pointer(&oraUint8Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraUint8Define.environment.ociError()
	}
	return nil
}

func (oraUint8Define *oraUint8Define) value() (value interface{}, err error) {
	uint8Value := Uint8{IsNull: oraUint8Define.isNull < 0}
	if !uint8Value.IsNull {
		r := C.OCINumberToInt(
			oraUint8Define.environment.ocierr, //OCIError              *err,
			&oraUint8Define.ociNumber,         //const OCINumber       *number,
			C.uword(1),                        //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,             //uword                 rsl_flag,
			unsafe.Pointer(&uint8Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraUint8Define.environment.ociError()
		}
	}
	value = uint8Value
	return value, err
}

func (oraUint8Define *oraUint8Define) alloc() error {
	return nil
}

func (oraUint8Define *oraUint8Define) free() {

}

func (oraUint8Define *oraUint8Define) close() {
	defer func() {
		recover()
	}()
	oraUint8Define.ocidef = nil
	oraUint8Define.isNull = C.sb2(0)
	oraUint8Define.environment.oraUint8DefinePool.Put(oraUint8Define)

}
