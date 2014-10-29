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

type uint32Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (uint32Define *uint32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                 //OCIStmt     *stmtp,
		&uint32Define.ocidef,                    //OCIDefine   **defnpp,
		uint32Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                         //ub4         position,
		unsafe.Pointer(&uint32Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8         value_sz,
		C.SQLT_VNU,                              //ub2         dty,
		unsafe.Pointer(&uint32Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return uint32Define.environment.ociError()
	}
	return nil
}

func (uint32Define *uint32Define) value() (value interface{}, err error) {
	if uint32Define.isNull > -1 {
		var uint32Value uint32
		r := C.OCINumberToInt(
			uint32Define.environment.ocierr, //OCIError              *err,
			&uint32Define.ociNumber,         //const OCINumber       *number,
			C.uword(4),                      //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,           //uword                 rsl_flag,
			unsafe.Pointer(&uint32Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = uint32Define.environment.ociError()
		}
		value = uint32Value
	}
	return value, err
}

func (uint32Define *uint32Define) alloc() error {
	return nil
}

func (uint32Define *uint32Define) free() {

}

func (uint32Define *uint32Define) close() {
	defer func() {
		recover()
	}()
	uint32Define.ocidef = nil
	uint32Define.isNull = C.sb2(0)
	uint32Define.environment.uint32DefinePool.Put(uint32Define)
}
