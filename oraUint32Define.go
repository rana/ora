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
	env       *Environment
	ocidef    *C.OCIDefine
	ociNumber C.OCINumber
	isNull    C.sb2
}

func (d *oraUint32Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),    //sb8         value_sz,
		C.SQLT_VNU,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}

func (d *oraUint32Define) value() (value interface{}, err error) {
	uint32Value := Uint32{IsNull: d.isNull < 0}
	if !uint32Value.IsNull {
		r := C.OCINumberToInt(
			d.env.ocierr,                       //OCIError              *err,
			&d.ociNumber,                       //const OCINumber       *number,
			C.uword(4),                         //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(&uint32Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = d.env.ociError()
		}
	}
	value = uint32Value
	return value, err
}

func (d *oraUint32Define) alloc() error {
	return nil
}

func (d *oraUint32Define) free() {

}

func (d *oraUint32Define) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.isNull = C.sb2(0)
	d.env.oraUint32DefinePool.Put(d)
}
