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
	"unsafe"
)

type defRowid struct {
	rset   *Rset
	ocidef *C.OCIDefine
	buf    []byte
}

func (def *defRowid) define(position int, rset *Rset) error {
	def.rset = rset
	// using a character host variable of width between 19
	// (18 bytes plus the null-terminator) and 4001 as the
	// host bind variable for universal ROWID.
	if len(def.buf) < 4001 {
		def.buf = make([]byte, 4001)
	}
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.buf[0]),      //void        *valuep,
		C.sb8(len(def.buf)),              //sb8         value_sz,
		C.SQLT_STR,                       //ub2         dty,
		nil,                              //void        *indp,
		nil,                              //ub2         *rlenp,
		nil,                              //ub2         *rcodep,
		C.OCI_DEFAULT)                    //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defRowid) value() (value interface{}, err error) {
	n := bytes.Index(def.buf, []byte{0})
	if n == -1 {
		n = len(def.buf)
	}
	value = string(def.buf[:n])
	return value, err
}

func (def *defRowid) alloc() error {
	return nil
}

func (def *defRowid) free() {
}

func (def *defRowid) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	clear(def.buf, 32)
	rset.putDef(defIdxRowid, def)
return nil
}
