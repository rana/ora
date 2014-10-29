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
	environment *Environment
	ocidef      *C.OCIDefine
	isNull      C.sb2
	buffer      []byte
}

func (stringDefine *stringDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	if cap(stringDefine.buffer) < columnSize {
		stringDefine.buffer = make([]byte, columnSize)
	}
	// Create oci define handle
	r := C.OCIDefineByPos2(
		ocistmt,                                 //OCIStmt     *stmtp,
		&stringDefine.ocidef,                    //OCIDefine   **defnpp,
		stringDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                         //ub4         position,
		unsafe.Pointer(&stringDefine.buffer[0]), //void        *valuep,
		C.sb8(columnSize),                       //sb8         value_sz,
		C.SQLT_CHR,                              //ub2         dty,
		unsafe.Pointer(&stringDefine.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return stringDefine.environment.ociError()
	}
	return nil
}
func (stringDefine *stringDefine) value() (value interface{}, err error) {
	if stringDefine.isNull > -1 {
		// Buffer is padded with Space char (32)
		value = stringTrimmed(stringDefine.buffer, 32)
	}
	return value, err
}
func (stringDefine *stringDefine) alloc() error {
	return nil
}
func (stringDefine *stringDefine) free() {

}
func (stringDefine *stringDefine) close() {
	defer func() {
		recover()
	}()
	stringDefine.ocidef = nil
	stringDefine.isNull = C.sb2(0)
	clear(stringDefine.buffer, 32)
	stringDefine.environment.stringDefinePool.Put(stringDefine)
}
