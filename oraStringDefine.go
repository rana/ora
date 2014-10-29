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
	environment *Environment
	ocidef      *C.OCIDefine
	isNull      C.sb2
	buffer      []byte
}

func (oraStringDefine *oraStringDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	if cap(oraStringDefine.buffer) < columnSize {
		oraStringDefine.buffer = make([]byte, columnSize)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                                    //OCIStmt     *stmtp,
		&oraStringDefine.ocidef,                    //OCIDefine   **defnpp,
		oraStringDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                            //ub4         position,
		unsafe.Pointer(&oraStringDefine.buffer[0]), //void        *valuep,
		C.sb8(columnSize),                          //sb8         value_sz,
		C.SQLT_CHR,                                 //ub2         dty,
		unsafe.Pointer(&oraStringDefine.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraStringDefine.environment.ociError()
	}
	return nil
}
func (oraStringDefine *oraStringDefine) value() (value interface{}, err error) {
	stringValue := String{IsNull: oraStringDefine.isNull < 0}
	if !stringValue.IsNull {
		// Buffer is padded with Space char (32)
		stringValue.Value = stringTrimmed(oraStringDefine.buffer, 32)
	}
	value = stringValue
	return value, err
}
func (oraStringDefine *oraStringDefine) alloc() error {
	return nil
}
func (oraStringDefine *oraStringDefine) free() {

}
func (oraStringDefine *oraStringDefine) close() {
	defer func() {
		recover()
	}()
	oraStringDefine.ocidef = nil
	oraStringDefine.isNull = C.sb2(0)
	clear(oraStringDefine.buffer, 32)
	oraStringDefine.environment.oraStringDefinePool.Put(oraStringDefine)
}
