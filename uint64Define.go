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

type uint64Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (uint64Define *uint64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                 //OCIStmt     *stmtp,
		&uint64Define.ocidef,                    //OCIDefine   **defnpp,
		uint64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                         //ub4         position,
		unsafe.Pointer(&uint64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8         value_sz,
		C.SQLT_VNU,                              //ub2         dty,
		unsafe.Pointer(&uint64Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return uint64Define.environment.ociError()
	}
	return nil
}

func (uint64Define *uint64Define) value() (value interface{}, err error) {
	if uint64Define.isNull > -1 {
		var uint64Value uint64
		r := C.OCINumberToInt(
			uint64Define.environment.ocierr, //OCIError              *err,
			&uint64Define.ociNumber,         //const OCINumber       *number,
			C.uword(8),                      //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,           //uword                 rsl_flag,
			unsafe.Pointer(&uint64Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = uint64Define.environment.ociError()
		}
		value = uint64Value
	}
	return value, err
}

func (uint64Define *uint64Define) alloc() error {
	return nil
}

func (uint64Define *uint64Define) free() {

}

func (uint64Define *uint64Define) close() {
	defer func() {
		recover()
	}()
	uint64Define.ocidef = nil
	uint64Define.isNull = C.sb2(0)
	uint64Define.environment.uint64DefinePool.Put(uint64Define)
}
