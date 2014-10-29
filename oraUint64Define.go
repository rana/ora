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

type oraUint64Define struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociNumber   C.OCINumber
	isNull      C.sb2
}

func (oraUint64Define *oraUint64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                    //OCIStmt     *stmtp,
		&oraUint64Define.ocidef,                    //OCIDefine   **defnpp,
		oraUint64Define.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                            //ub4         position,
		unsafe.Pointer(&oraUint64Define.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),                  //sb8         value_sz,
		C.SQLT_VNU,                                 //ub2         dty,
		unsafe.Pointer(&oraUint64Define.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraUint64Define.environment.ociError()
	}
	return nil
}

func (oraUint64Define *oraUint64Define) value() (value interface{}, err error) {
	uint64Value := Uint64{IsNull: oraUint64Define.isNull < 0}
	if !uint64Value.IsNull {
		r := C.OCINumberToInt(
			oraUint64Define.environment.ocierr, //OCIError              *err,
			&oraUint64Define.ociNumber,         //const OCINumber       *number,
			C.uword(8),                         //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(&uint64Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = oraUint64Define.environment.ociError()
		}
	}
	value = uint64Value
	return value, err
}

func (oraUint64Define *oraUint64Define) alloc() error {
	return nil
}

func (oraUint64Define *oraUint64Define) free() {

}

func (oraUint64Define *oraUint64Define) close() {
	defer func() {
		recover()
	}()
	oraUint64Define.ocidef = nil
	oraUint64Define.isNull = C.sb2(0)
	oraUint64Define.environment.oraUint64DefinePool.Put(oraUint64Define)
}
