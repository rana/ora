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
	"bytes"
	"unsafe"
)

type boolDefine struct {
	env    *Environment
	rst    *ResultSet
	ocidef *C.OCIDefine
	isNull C.sb2
	buf    []byte
}

func (d *boolDefine) define(columnSize int, position int, rst *ResultSet, ocistmt *C.OCIStmt) error {
	d.rst = rst
	if cap(d.buf) < columnSize {
		d.buf = make([]byte, columnSize)
	}
	// Create oci define handle
	r := C.OCIDefineByPos2(
		ocistmt,                   //OCIStmt     *stmtp,
		&d.ocidef,                 //OCIDefine   **defnpp,
		d.env.ocierr,              //OCIError    *errhp,
		C.ub4(position),           //ub4         position,
		unsafe.Pointer(&d.buf[0]), //void        *valuep,
		C.sb8(columnSize),         //sb8         value_sz,
		C.SQLT_CHR,                //ub2         dty,
		unsafe.Pointer(&d.isNull), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *boolDefine) value() (value interface{}, err error) {
	if d.isNull > -1 {
		value = bytes.Runes(d.buf)[0] == d.rst.stmt.Config.ResultSet.TrueRune
	}
	return value, err
}
func (d *boolDefine) alloc() error {
	return nil
}
func (d *boolDefine) free() {

}
func (d *boolDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.isNull = C.sb2(0)
	clear(d.buf, 0)
	d.env.boolDefinePool.Put(d)
}
