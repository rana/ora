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

type stringDefine struct {
	env    *Environment
	ocidef *C.OCIDefine
	isNull C.sb2
	buffer []byte
}

func (d *stringDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	if cap(d.buffer) < columnSize {
		d.buffer = make([]byte, columnSize)
	}
	// Create oci define handle
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
func (d *stringDefine) value() (value interface{}, err error) {
	if d.isNull > -1 {
		// Buffer is padded with Space char (32)
		value = stringTrimmed(d.buffer, 32)
	}
	return value, err
}
func (d *stringDefine) alloc() error {
	return nil
}
func (d *stringDefine) free() {

}
func (stringDefine *stringDefine) close() {
	defer func() {
		recover()
	}()
	stringDefine.ocidef = nil
	stringDefine.isNull = C.sb2(0)
	clear(stringDefine.buffer, 32)
	stringDefine.env.stringDefinePool.Put(stringDefine)
}
