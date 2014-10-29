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

type uint8Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (uint8Define *uint8Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                //OCIStmt     *stmtp,
		&uint8Define.ocidef,                    //OCIDefine   **defnpp,
		uint8Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                        //ub4         position,
		unsafe.Pointer(&uint8Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8         value_sz,
		C.SQLT_VNU,                             //ub2         dty,
		unsafe.Pointer(&uint8Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return uint8Define.environment.ociError()
	}
	return nil
}

func (uint8Define *uint8Define) value() (value interface{}, err error) {
	if uint8Define.isNull > -1 {
		var uint8Value uint8
		r := C.OCINumberToInt(
			uint8Define.environment.ocierr, //OCIError              *err,
			&uint8Define.ociNumber,         //const OCINumber       *number,
			C.uword(1),                     //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,          //uword                 rsl_flag,
			unsafe.Pointer(&uint8Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = uint8Define.environment.ociError()
		}
		value = uint8Value
	}
	return value, err
}

func (uint8Define *uint8Define) alloc() error {
	return nil
}

func (uint8Define *uint8Define) free() {

}

func (uint8Define *uint8Define) close() {
	defer func() {
		recover()
	}()
	uint8Define.ocidef = nil
	uint8Define.isNull = C.sb2(0)
	uint8Define.environment.uint8DefinePool.Put(uint8Define)
}
