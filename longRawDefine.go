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

type longRawDefine struct {
	environment  *Environment
	ocidef       *C.OCIDefine
	isNull       C.sb2
	buffer       []byte
	returnLength C.ub4
	returnType   GoColumnType
}

func (longRawDefine *longRawDefine) define(columnSize int, position int, returnType GoColumnType, longRawBufferSize uint32, ocistmt *C.OCIStmt) error {
	longRawDefine.returnType = returnType
	longRawDefine.buffer = make([]byte, int(longRawBufferSize))
	// Create oci define handle
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&longRawDefine.ocidef,                    //OCIDefine   **defnpp,
		longRawDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&longRawDefine.buffer[0]), //void        *valuep,
		C.sb8(len(longRawDefine.buffer)),         //sb8         value_sz,
		C.SQLT_LBI,                               //ub2         dty,
		unsafe.Pointer(&longRawDefine.isNull),    //void        *indp,
		&longRawDefine.returnLength,              //ub4         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return longRawDefine.environment.ociError()
	}
	return nil
}
func (longRawDefine *longRawDefine) value() (value interface{}, err error) {
	if longRawDefine.returnType == Bits {
		if longRawDefine.isNull > -1 {
			// Make a slice of length equal to the return length
			result := make([]byte, longRawDefine.returnLength)
			// Copy returned data
			copyLength := copy(result, longRawDefine.buffer)
			if C.ub4(copyLength) != longRawDefine.returnLength {
				return nil, errNew("unable to copy LONG RAW result data from buffer")
			}
			value = result
		}
	} else {
		bytesValue := Bytes{IsNull: longRawDefine.isNull < 0}
		if !bytesValue.IsNull {
			// Make a slice of length equal to the return length
			bytesValue.Value = make([]byte, longRawDefine.returnLength)
			// Copy returned data
			copyLength := copy(bytesValue.Value, longRawDefine.buffer)
			if C.ub4(copyLength) != longRawDefine.returnLength {
				return nil, errNew("unable to copy LONG RAW result data from buffer")
			}
		}
		value = bytesValue
	}

	return value, err
}
func (longRawDefine *longRawDefine) alloc() error {
	return nil
}
func (longRawDefine *longRawDefine) free() {

}
func (longRawDefine *longRawDefine) close() {
	defer func() {
		recover()
	}()
	longRawDefine.ocidef = nil
	longRawDefine.returnLength = 0
	longRawDefine.buffer = nil
	longRawDefine.isNull = C.sb2(0)
	longRawDefine.environment.longRawDefinePool.Put(longRawDefine)
}
