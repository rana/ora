// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type defString struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	buf        []byte
	isNullable bool
	rlen       C.ACTUAL_LENGTH_TYPE
	nullp
}

func (def *defString) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	//Log.Infof("defString position=%d columnSize=%d", position, columnSize)
	n := columnSize
	// AL32UTF8: one db "char" can be 4 bytes on wire, esp. if the database's
	// character set is not AL32UTF8 (e.g. some 8bit fixed width charset), and
	// the column is VARCHAR2 with byte semantics.
	//
	// For example when the db's charset is EE8ISO8859P2, then a VARCHAR2(1) can
	// contain an "ű", which is 2 bytes AL32UTF8.
	if !rset.stmt.ses.srv.dbIsUTF8 {
		n *= 2
	}
	if n == 0 {
		n = 2
	}
	if n%2 != 0 {
		n++
	}
	if c := cap(def.buf); c < n {
		i := 1
		if c > 0 && c&(c-1) == 0 { // c is power of 2.
			i = c
		}
		for i < n {
			i <<= 1 // double i
		}
		def.buf = make([]byte, i)
	}
	buf := def.buf[:n]
	// Create oci define handle
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                    //OCIStmt     *stmtp,
		&def.ocidef,                         //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,    //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&buf[0]),             //void        *valuep,
		C.LENGTH_TYPE(n),                    //sb8         value_sz,
		C.SQLT_CHR,                          //ub2         dty,
		unsafe.Pointer(def.nullp.Pointer()), //void        *indp,
		&def.rlen,                           //ub2         *rlenp,
		nil,                                 //ub2         *rcodep,
		C.OCI_DEFAULT)                       //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defString) value() (value interface{}, err error) {
	if def.isNullable {
		oraStringValue := String{IsNull: def.nullp.IsNull()}
		if !oraStringValue.IsNull {
			oraStringValue.Value = string(def.buf[:int(def.rlen)])
		}
		return oraStringValue, nil
	}
	if def.nullp.IsNull() {
		return "", nil
	}
	return string(def.buf[:int(def.rlen)]), nil
}

func (def *defString) alloc() error {
	return nil
}

func (def *defString) free() {
}

func (def *defString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.buf = nil
	def.nullp.Free()
	rset.putDef(defIdxString, def)
	return nil
}
