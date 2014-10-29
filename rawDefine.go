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
	environment *Environment
	ocidef      *C.OCIDefine
	ociRaw      *C.OCIRaw
	buffer      []byte
	isNull      C.sb2
	returnType  GoColumnType
}

func (rawDefine *rawDefine) define(columnSize int, position int, returnType GoColumnType, ocistmt *C.OCIStmt) error {
	rawDefine.returnType = returnType
	rawDefine.buffer = make([]byte, columnSize)
	r := C.OCIDefineByPos2(
		ocistmt,                              //OCIStmt     *stmtp,
		&rawDefine.ocidef,                    //OCIDefine   **defnpp,
		rawDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                      //ub4         position,
		unsafe.Pointer(&rawDefine.buffer[0]), //void        *valuep,
		C.sb8(columnSize),                    //sb8         value_sz,
		C.SQLT_BIN,                           //ub2         dty,
		unsafe.Pointer(&rawDefine.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return rawDefine.environment.ociError()
	}
	return nil
}
func (rawDefine *rawDefine) value() (value interface{}, err error) {
	if rawDefine.returnType == Bits {
		if rawDefine.isNull > -1 {
			value = rawDefine.buffer
		}
	} else {
		bytesValue := Bytes{IsNull: rawDefine.isNull < 0}
		if !bytesValue.IsNull {
			bytesValue.Value = rawDefine.buffer
		}
		value = bytesValue
	}

	return value, err
}
func (rawDefine *rawDefine) alloc() error {
	return nil
}
func (rawDefine *rawDefine) free() {

}
func (rawDefine *rawDefine) close() {
	defer func() {
		recover()
	}()
	rawDefine.ocidef = nil
	rawDefine.ociRaw = nil
	rawDefine.buffer = nil
	rawDefine.isNull = C.sb2(0)
	rawDefine.returnType = D
	rawDefine.environment.rawDefinePool.Put(rawDefine)
}
