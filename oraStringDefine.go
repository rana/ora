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

type oraStringDefine struct {
	env    *Environment
	ocidef *C.OCIDefine
	isNull C.sb2
	buffer []byte
}

func (d *oraStringDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	if cap(d.buffer) < columnSize {
		d.buffer = make([]byte, columnSize)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.buffer[0]), //void        *valuep,
		C.sb8(columnSize),            //sb8         value_sz,
		C.SQLT_CHR,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *oraStringDefine) value() (value interface{}, err error) {
	stringValue := String{IsNull: d.isNull < 0}
	if !stringValue.IsNull {
		// Buffer is padded with Space char (32)
		stringValue.Value = stringTrimmed(d.buffer, 32)
	}
	value = stringValue
	return value, err
}
func (d *oraStringDefine) alloc() error {
	return nil
}
func (d *oraStringDefine) free() {

}
func (d *oraStringDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.isNull = C.sb2(0)
	clear(d.buffer, 32)
	d.env.oraStringDefinePool.Put(d)
}
