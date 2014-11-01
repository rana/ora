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
	env          *Environment
	ocidef       *C.OCIDefine
	isNull       C.sb2
	buffer       []byte
	returnLength C.ub4
	returnType   GoColumnType
}

func (d *longRawDefine) define(columnSize int, position int, returnType GoColumnType, longRawBufferSize uint32, ocistmt *C.OCIStmt) error {
	d.returnType = returnType
	d.buffer = make([]byte, int(longRawBufferSize))
	// Create oci define handle
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.buffer[0]), //void        *valuep,
		C.sb8(len(d.buffer)),         //sb8         value_sz,
		C.SQLT_LBI,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		&d.returnLength,              //ub4         *rlenp,
		nil,                          //ub2         *rcodep,
		C.OCI_DEFAULT)                //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *longRawDefine) value() (value interface{}, err error) {
	if d.returnType == Bits {
		if d.isNull > -1 {
			// Make a slice of length equal to the return length
			result := make([]byte, d.returnLength)
			// Copy returned data
			copyLength := copy(result, d.buffer)
			if C.ub4(copyLength) != d.returnLength {
				return nil, errNew("unable to copy LONG RAW result data from buffer")
			}
			value = result
		}
	} else {
		bytesValue := Bytes{IsNull: d.isNull < 0}
		if !bytesValue.IsNull {
			// Make a slice of length equal to the return length
			bytesValue.Value = make([]byte, d.returnLength)
			// Copy returned data
			copyLength := copy(bytesValue.Value, d.buffer)
			if C.ub4(copyLength) != d.returnLength {
				return nil, errNew("unable to copy LONG RAW result data from buffer")
			}
		}
		value = bytesValue
	}

	return value, err
}
func (d *longRawDefine) alloc() error {
	return nil
}
func (d *longRawDefine) free() {

}
func (d *longRawDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.returnLength = 0
	d.buffer = nil
	d.isNull = C.sb2(0)
	d.env.longRawDefinePool.Put(d)
}
