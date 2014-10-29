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

type int64Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (int64Define *int64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                //OCIStmt     *stmtp,
		&int64Define.ocidef,                    //OCIDefine   **defnpp,
		int64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                        //ub4         position,
		unsafe.Pointer(&int64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8         value_sz,
		C.SQLT_VNU,                             //ub2         dty,
		unsafe.Pointer(&int64Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return int64Define.environment.ociError()
	}
	return nil
}

func (int64Define *int64Define) value() (value interface{}, err error) {
	if int64Define.isNull > -1 {
		var int64Value int64
		r := C.OCINumberToInt(
			int64Define.environment.ocierr, //OCIError              *err,
			&int64Define.ociNumber,         //const OCINumber       *number,
			C.uword(8),                     //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,            //uword                 rsl_flag,
			unsafe.Pointer(&int64Value))    //void                  *rsl );
		if r == C.OCI_ERROR {
			err = int64Define.environment.ociError()
		}
		value = int64Value
	}
	return value, err
}

func (int64Define *int64Define) alloc() error {
	return nil
}

func (int64Define *int64Define) free() {

}

func (int64Define *int64Define) close() {
	defer func() {
		recover()
	}()
	int64Define.ocidef = nil
	int64Define.isNull = C.sb2(0)
	int64Define.environment.int64DefinePool.Put(int64Define)
}
