// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"github.com/golang/glog"
	"unsafe"
)

type defRaw struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociRaw     *C.OCIRaw
	null       C.sb2
	isNullable bool
	buf        []byte
}

func (def *defRaw) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	glog.Infoln("position: ", position)
	def.rset = rset
	def.isNullable = isNullable
	def.buf = make([]byte, columnSize)
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.buf[0]),      //void        *valuep,
		C.sb8(columnSize),                //sb8         value_sz,
		C.SQLT_BIN,                       //ub2         dty,
		unsafe.Pointer(&def.null),        //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defRaw) value() (value interface{}, err error) {
	if def.isNullable {
		bytesValue := Binary{IsNull: def.null < 0}
		if !bytesValue.IsNull {
			bytesValue.Value = def.buf
		}
		value = bytesValue
	} else {
		if def.null > -1 {
			value = def.buf
		}
	}
	return value, err
}

func (def *defRaw) alloc() error {
	return nil
}

func (def *defRaw) free() {

}

func (def *defRaw) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociRaw = nil
	def.buf = nil
	rset.putDef(defIdxRaw, def)
	return nil
}
