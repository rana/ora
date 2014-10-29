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

type int32Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (int32Define *int32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                //OCIStmt     *stmtp,
		&int32Define.ocidef,                    //OCIDefine   **defnpp,
		int32Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                        //ub4         position,
		unsafe.Pointer(&int32Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8         value_sz,
		C.SQLT_VNU,                             //ub2         dty,
		unsafe.Pointer(&int32Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return int32Define.environment.ociError()
	}
	return nil
}

func (int32Define *int32Define) value() (value interface{}, err error) {
	if int32Define.isNull > -1 {
		var int32Value int32
		r := C.OCINumberToInt(
			int32Define.environment.ocierr, //OCIError              *err,
			&int32Define.ociNumber,         //const OCINumber       *number,
			C.uword(4),                     //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,            //uword                 rsl_flag,
			unsafe.Pointer(&int32Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = int32Define.environment.ociError()
		}
		value = int32Value
	}
	return value, err
}

func (int32Define *int32Define) alloc() error {
	return nil
}

func (int32Define *int32Define) free() {

}

func (int32Define *int32Define) close() {
	defer func() {
		recover()
	}()
	int32Define.ocidef = nil
	int32Define.isNull = C.sb2(0)
	int32Define.environment.int32DefinePool.Put(int32Define)
}
