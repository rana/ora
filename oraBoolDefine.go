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

type oraBoolDefine struct {
	env     *Environment
	rst     *ResultSet
	ocidef  *C.OCIDefine
	nameStr string
	isNull  C.sb2
	buffer  []byte
}

func (d *oraBoolDefine) define(columnSize int, position int, rst *ResultSet, ocistmt *C.OCIStmt) error {
	d.rst = rst
	if cap(d.buffer) < columnSize {
		d.buffer = make([]byte, columnSize)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.buffer[0]), //void        *valuep,
		C.sb8(columnSize),            //sb8         value_sz,
		C.SQLT_CHR,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *oraBoolDefine) value() (value interface{}, err error) {
	boolValue := Bool{IsNull: d.isNull < 0}
	if !boolValue.IsNull {
		boolValue.Value = bytes.Runes(d.buffer)[0] == d.rst.stmt.Config.ResultSet.TrueRune
	}
	return boolValue, err
}
func (d *oraBoolDefine) alloc() error {
	return nil
}
func (d *oraBoolDefine) free() {

}
func (d *oraBoolDefine) close() {
	defer func() {
		recover()
	}()
	d.rst = nil
	d.ocidef = nil
	d.isNull = C.sb2(0)
	clear(d.buffer, 0)
	d.env.oraBoolDefinePool.Put(d)
}
