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

type rawDefine struct {
	env        *Environment
	ocidef     *C.OCIDefine
	ociRaw     *C.OCIRaw
	buffer     []byte
	isNull     C.sb2
	returnType GoColumnType
}

func (d *rawDefine) define(columnSize int, position int, returnType GoColumnType, ocistmt *C.OCIStmt) error {
	d.returnType = returnType
	d.buffer = make([]byte, columnSize)
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.buffer[0]), //void        *valuep,
		C.sb8(columnSize),            //sb8         value_sz,
		C.SQLT_BIN,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *rawDefine) value() (value interface{}, err error) {
	if d.returnType == Bits {
		if d.isNull > -1 {
			value = d.buffer
		}
	} else {
		bytesValue := Bytes{IsNull: d.isNull < 0}
		if !bytesValue.IsNull {
			bytesValue.Value = d.buffer
		}
		value = bytesValue
	}

	return value, err
}
func (d *rawDefine) alloc() error {
	return nil
}
func (d *rawDefine) free() {

}
func (d *rawDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociRaw = nil
	d.buffer = nil
	d.isNull = C.sb2(0)
	d.returnType = D
	d.env.rawDefinePool.Put(d)
}
