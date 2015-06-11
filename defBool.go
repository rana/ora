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
	"unicode/utf8"
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
	def.rset = rset
	def.isNullable = isNullable
	if cap(def.buf) < columnSize {
		def.buf = make([]byte, columnSize)
	}
	Log.Infof("defBool.define(position=%d, columnSize=%d)", position, columnSize)
	// Create oci define handle
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.buf[0]),      //void        *valuep,
		C.LENGTH_TYPE(columnSize),        //sb8         value_sz,
		C.SQLT_AFC,                       //ub2         dty,
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
	Log.Infof("%v.value", def)
	if def.isNullable {
		oraBoolValue := Bool{IsNull: def.null < C.sb2(0)}
		if !oraBoolValue.IsNull {
			r, _ := utf8.DecodeRune(def.buf)
			oraBoolValue.Value = r == def.rset.stmt.Cfg.Rset.TrueRune
		}
		return oraBoolValue, nil
	}
	if def.null > C.sb2(-1) {
		r, _ := utf8.DecodeRune(def.buf)
		return r == def.rset.stmt.Cfg.Rset.TrueRune, nil
	}
	// NULL is false, too
	return false, nil
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

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	clear(def.buf, 0)
	rset.putDef(defIdxBool, def)
	return nil
}
