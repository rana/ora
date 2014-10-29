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

type int8Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (int8Define *int8Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                               //OCIStmt     *stmtp,
		&int8Define.ocidef,                    //OCIDefine   **defnpp,
		int8Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                       //ub4         position,
		unsafe.Pointer(&int8Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),             //sb8         value_sz,
		C.SQLT_VNU,                            //ub2         dty,
		unsafe.Pointer(&int8Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return int8Define.environment.ociError()
	}
	return nil
}

func (int8Define *int8Define) value() (value interface{}, err error) {
	if int8Define.isNull > -1 {
		var int8Value int8
		r := C.OCINumberToInt(
			int8Define.environment.ocierr, //OCIError              *err,
			&int8Define.ociNumber,         //const OCINumber       *number,
			C.uword(1),                    //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,           //uword                 rsl_flag,
			unsafe.Pointer(&int8Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = int8Define.environment.ociError()
		}
		value = int8Value
	}
	return value, err
}

func (int8Define *int8Define) alloc() error {
	return nil
}

func (int8Define *int8Define) free() {

}

func (int8Define *int8Define) close() {
	defer func() {
		recover()
	}()
	int8Define.ocidef = nil
	int8Define.isNull = C.sb2(0)
	int8Define.environment.int8DefinePool.Put(int8Define)
}
