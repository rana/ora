// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"bytes"
	"github.com/golang/glog"
	"unsafe"
)

type defBool struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	null       C.sb2
	isNullable bool
	buf        []byte
}

func (def *defBool) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	glog.Infoln("position: ", position)
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

func (def *defBool) value() (value interface{}, err error) {
	if def.isNullable {
		oraBoolValue := Bool{IsNull: def.null < 0}
		if !oraBoolValue.IsNull {
			oraBoolValue.Value = bytes.Runes(def.buf)[0] == def.rset.stmt.Config.Rset.TrueRune
		}
		value = oraBoolValue
	} else {
		if def.null > -1 {
			value = bytes.Runes(def.buf)[0] == def.rset.stmt.Config.Rset.TrueRune
		}
	}

	return value, err
}

func (def *defBool) alloc() error {
	return nil
}

func (def *defBool) free() {

}

func (def *defBool) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	clear(def.buf, 0)
	rset.putDef(defIdxBool, def)
	return nil
}
