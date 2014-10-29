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

type int16Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (int16Define *int16Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                //OCIStmt     *stmtp,
		&int16Define.ocidef,                    //OCIDefine   **defnpp,
		int16Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                        //ub4         position,
		unsafe.Pointer(&int16Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8         value_sz,
		C.SQLT_VNU,                             //ub2         dty,
		unsafe.Pointer(&int16Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return int16Define.environment.ociError()
	}
	return nil
}

func (int16Define *int16Define) value() (value interface{}, err error) {
	if int16Define.isNull > -1 {
		var int16Value int16
		r := C.OCINumberToInt(
			int16Define.environment.ocierr, //OCIError              *err,
			&int16Define.ociNumber,         //const OCINumber       *number,
			C.uword(2),                     //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,            //uword                 rsl_flag,
			unsafe.Pointer(&int16Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = int16Define.environment.ociError()
		}
		value = int16Value
	}
	return value, err
}

func (int16Define *int16Define) alloc() error {
	return nil
}

func (int16Define *int16Define) free() {

}

func (int16Define *int16Define) close() {
	defer func() {
		recover()
	}()
	int16Define.ocidef = nil
	int16Define.isNull = C.sb2(0)
	int16Define.environment.int16DefinePool.Put(int16Define)
}
