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

type oraUint32Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraUint32Define *oraUint32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                    //OCIStmt     *stmtp,
		&oraUint32Define.ocidef,                    //OCIDefine   **defnpp,
		oraUint32Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                            //ub4         position,
		unsafe.Pointer(&oraUint32Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                  //sb8         value_sz,
		C.SQLT_VNU,                                 //ub2         dty,
		unsafe.Pointer(&oraUint32Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraUint32Define.environment.ociError()
	}
	return nil
}

func (oraUint32Define *oraUint32Define) value() (value interface{}, err error) {
	uint32Value := Uint32{IsNull: oraUint32Define.isNull < 0}
	if !uint32Value.IsNull {
		r := C.OCINumberToInt(
			oraUint32Define.environment.ocierr, //OCIError              *err,
			&oraUint32Define.ociNumber,         //const OCINumber       *number,
			C.uword(4),                         //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(&uint32Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraUint32Define.environment.ociError()
		}
	}
	value = uint32Value
	return value, err
}

func (oraUint32Define *oraUint32Define) alloc() error {
	return nil
}

func (oraUint32Define *oraUint32Define) free() {

}

func (oraUint32Define *oraUint32Define) close() {
	defer func() {
		recover()
	}()
	oraUint32Define.ocidef = nil
	oraUint32Define.isNull = C.sb2(0)
	oraUint32Define.environment.oraUint32DefinePool.Put(oraUint32Define)
}
