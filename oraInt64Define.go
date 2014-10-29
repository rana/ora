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

type oraInt64Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraInt64Define *oraInt64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                   //OCIStmt     *stmtp,
		&oraInt64Define.ocidef,                    //OCIDefine   **defnpp,
		oraInt64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                           //ub4         position,
		unsafe.Pointer(&oraInt64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                 //sb8         value_sz,
		C.SQLT_VNU,                                //ub2         dty,
		unsafe.Pointer(&oraInt64Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraInt64Define.environment.ociError()
	}
	return nil
}

func (oraInt64Define *oraInt64Define) value() (value interface{}, err error) {
	int64Value := Int64{IsNull: oraInt64Define.isNull < 0}
	if !int64Value.IsNull {
		r := C.OCINumberToInt(
			oraInt64Define.environment.ocierr, //OCIError              *err,
			&oraInt64Define.ociNumber,         //const OCINumber       *number,
			C.uword(8),                        //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(&int64Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraInt64Define.environment.ociError()
		}
	}
	value = int64Value
	return value, err
}

func (oraInt64Define *oraInt64Define) alloc() error {
	return nil
}

func (oraInt64Define *oraInt64Define) free() {
	defer func() {
		recover()
	}()
}

func (oraInt64Define *oraInt64Define) close() {
	defer func() {
		recover()
	}()
	oraInt64Define.ocidef = nil
	oraInt64Define.isNull = C.sb2(0)
	oraInt64Define.environment.oraInt64DefinePool.Put(oraInt64Define)
}
