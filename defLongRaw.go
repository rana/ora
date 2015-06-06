// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type defLongRaw struct {
	rset         *Rset
	ocidef       *C.OCIDefine
	null         C.sb2
	isNullable   bool
	returnLength C.ACTUAL_LENGTH_TYPE
	buf          []byte
}

func (def *defLongRaw) define(position int, bufSize uint32, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	def.buf = make([]byte, int(bufSize))
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.buf[0]),      //void        *valuep,
		C.LENGTH_TYPE(len(def.buf)),      //sb8         value_sz,
		C.SQLT_LBI,                       //ub2         dty,
		unsafe.Pointer(&def.null),        //void        *indp,
		&def.returnLength,                //ub4         *rlenp,
		nil,                              //ub2         *rcodep,
		C.OCI_DEFAULT)                    //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defLongRaw) value() (value interface{}, err error) {
	if def.isNullable {
		bytesValue := Raw{IsNull: def.null < 0}
		if !bytesValue.IsNull {
			// Make a slice of length equal to the return length
			bytesValue.Value = make([]byte, def.returnLength)
			// Copy returned data
			copyLength := copy(bytesValue.Value, def.buf)
			if C.ACTUAL_LENGTH_TYPE(copyLength) != def.returnLength {
				return nil, errNew("unable to copy LONG RAW result data from buffer")
			}
		}
		value = bytesValue
	} else {
		// Make a slice of length equal to the return length
		result := make([]byte, def.returnLength)
		// Copy returned data
		copyLength := copy(result, def.buf)
		if C.ACTUAL_LENGTH_TYPE(copyLength) != def.returnLength {
			return nil, errNew("unable to copy LONG RAW result data from buffer")
		}
		value = result
	}
	return value, err
}

func (def *defLongRaw) alloc() error {
	return nil
}

func (def *defLongRaw) free() {

}

func (def *defLongRaw) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.buf = nil
	rset.putDef(defIdxLongRaw, def)
	return nil
}
