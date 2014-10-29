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

type oraUint16Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraUint16Define *oraUint16Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                    //OCIStmt     *stmtp,
		&oraUint16Define.ocidef,                    //OCIDefine   **defnpp,
		oraUint16Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                            //ub4         position,
		unsafe.Pointer(&oraUint16Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                  //sb8         value_sz,
		C.SQLT_VNU,                                 //ub2         dty,
		unsafe.Pointer(&oraUint16Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraUint16Define.environment.ociError()
	}
	return nil
}

func (oraUint16Define *oraUint16Define) value() (value interface{}, err error) {
	uint16Value := Uint16{IsNull: oraUint16Define.isNull < 0}
	if !uint16Value.IsNull {
		r := C.OCINumberToInt(
			oraUint16Define.environment.ocierr, //OCIError              *err,
			&oraUint16Define.ociNumber,         //const OCINumber       *number,
			C.uword(2),                         //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(&uint16Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraUint16Define.environment.ociError()
		}
	}
	value = uint16Value
	return value, err
}

func (oraUint16Define *oraUint16Define) alloc() error {
	return nil
}

func (oraUint16Define *oraUint16Define) free() {

}

func (oraUint16Define *oraUint16Define) close() {
	defer func() {
		recover()
	}()
	oraUint16Define.ocidef = nil
	oraUint16Define.isNull = C.sb2(0)
	oraUint16Define.environment.oraUint16DefinePool.Put(oraUint16Define)
}
