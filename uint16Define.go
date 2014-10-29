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

type uint16Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (uint16Define *uint16Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                 //OCIStmt     *stmtp,
		&uint16Define.ocidef,                    //OCIDefine   **defnpp,
		uint16Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                         //ub4         position,
		unsafe.Pointer(&uint16Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8         value_sz,
		C.SQLT_VNU,                              //ub2         dty,
		unsafe.Pointer(&uint16Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return uint16Define.environment.ociError()
	}
	return nil
}

func (uint16Define *uint16Define) value() (value interface{}, err error) {
	if uint16Define.isNull > -1 {
		var uint16Value uint16
		r := C.OCINumberToInt(
			uint16Define.environment.ocierr, //OCIError              *err,
			&uint16Define.ociNumber,         //const OCINumber       *number,
			C.uword(2),                      //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,           //uword                 rsl_flag,
			unsafe.Pointer(&uint16Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = uint16Define.environment.ociError()
		}
		value = uint16Value
	}
	return value, err
}

func (uint16Define *uint16Define) alloc() error {
	return nil
}

func (uint16Define *uint16Define) free() {

}

func (uint16Define *uint16Define) close() {
	defer func() {
		recover()
	}()
	uint16Define.ocidef = nil
	uint16Define.isNull = C.sb2(0)
	uint16Define.environment.uint16DefinePool.Put(uint16Define)
}
