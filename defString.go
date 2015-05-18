// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"unsafe"
)

type defString struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	null       C.sb2
	isNullable bool
	buf        []byte
}

func (def *defString) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if cap(def.buf) < columnSize {
		def.buf = make([]byte, columnSize)
	}
	// Create oci define handle
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.buf[0]),      //void        *valuep,
		C.sb8(columnSize),                //sb8         value_sz,
		C.SQLT_CHR,                       //ub2         dty,
		unsafe.Pointer(&def.null),        //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defString) value() (value interface{}, err error) {
	// Buffer is padded with Space char (32)
	if def.isNullable {
		oraStringValue := String{IsNull: def.null < 0}
		if !oraStringValue.IsNull {
			oraStringValue.Value = stringTrimmed(def.buf, 32)
		}
		value = oraStringValue
	} else {
		if def.null < 0 {
			value = ""
		} else {
			value = stringTrimmed(def.buf, 32)
		}
	}
	return value, err
}

func (def *defString) alloc() error {
	return nil
}

func (def *defString) free() {

}

func (def *defString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	clear(def.buf, 32)
	rset.putDef(defIdxString, def)
	return nil
}
