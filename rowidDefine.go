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

type rowidDefine struct {
	env    *Environment
	ocidef *C.OCIDefine
	buffer []byte
}

func (d *rowidDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	// using a character host variable of width between 19
	// (18 bytes plus the null-terminator) and 4001 as the
	// host bind variable for universal ROWID.
	if len(d.buffer) < 4001 {
		d.buffer = make([]byte, 4001)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.buffer[0]), //void        *valuep,
		C.sb8(len(d.buffer)),         //sb8         value_sz,
		C.SQLT_STR,                   //ub2         dty,
		nil,                          //void        *indp,
		nil,                          //ub2         *rlenp,
		nil,                          //ub2         *rcodep,
		C.OCI_DEFAULT)                //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *rowidDefine) value() (value interface{}, err error) {
	n := bytes.Index(d.buffer, []byte{0})
	if n == -1 {
		n = len(d.buffer)
	}
	value = string(d.buffer[:n])
	return value, err
}
func (d *rowidDefine) alloc() error {
	return nil
}
func (d *rowidDefine) free() {
}
func (d *rowidDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	clear(d.buffer, 32)
	d.env.rowidDefinePool.Put(d)
}
