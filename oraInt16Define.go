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

type oraInt16Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraInt16Define *oraInt16Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                   //OCIStmt     *stmtp,
		&oraInt16Define.ocidef,                    //OCIDefine   **defnpp,
		oraInt16Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                           //ub4         position,
		unsafe.Pointer(&oraInt16Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                 //sb8         value_sz,
		C.SQLT_VNU,                                //ub2         dty,
		unsafe.Pointer(&oraInt16Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraInt16Define.environment.ociError()
	}
	return nil
}

func (oraInt16Define *oraInt16Define) value() (value interface{}, err error) {
	int16Value := Int16{IsNull: oraInt16Define.isNull < 0}
	if !int16Value.IsNull {
		r := C.OCINumberToInt(
			oraInt16Define.environment.ocierr, //OCIError              *err,
			&oraInt16Define.ociNumber,         //const OCINumber       *number,
			C.uword(2),                        //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(&int16Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraInt16Define.environment.ociError()
		}
	}
	value = int16Value
	return value, err
}

func (oraInt16Define *oraInt16Define) alloc() error {
	return nil
}

func (oraInt16Define *oraInt16Define) free() {

}

func (oraInt16Define *oraInt16Define) close() {
	defer func() {
		recover()
	}()
	oraInt16Define.ocidef = nil
	oraInt16Define.isNull = C.sb2(0)
	oraInt16Define.environment.oraInt16DefinePool.Put(oraInt16Define)
}
