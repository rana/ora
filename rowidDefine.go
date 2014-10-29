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
	environment *Environment
	ocidef      *C.OCIDefine
	buffer      []byte
}

func (rowidDefine *rowidDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	// using a character host variable of width between 19
	// (18 bytes plus the null-terminator) and 4001 as the
	// host bind variable for universal ROWID.
	if len(rowidDefine.buffer) < 4001 {
		rowidDefine.buffer = make([]byte, 4001)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                                //OCIStmt     *stmtp,
		&rowidDefine.ocidef,                    //OCIDefine   **defnpp,
		rowidDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                        //ub4         position,
		unsafe.Pointer(&rowidDefine.buffer[0]), //void        *valuep,
		C.sb8(len(rowidDefine.buffer)),         //sb8         value_sz,
		C.SQLT_STR,                             //ub2         dty,
		nil,                                    //void        *indp,
		nil,                                    //ub2         *rlenp,
		nil,                                    //ub2         *rcodep,
		C.OCI_DEFAULT)                          //ub4         mode );
	if r == C.OCI_ERROR {
		return rowidDefine.environment.ociError()
	}
	return nil
}
func (rowidDefine *rowidDefine) value() (value interface{}, err error) {
	n := bytes.Index(rowidDefine.buffer, []byte{0})
	if n == -1 {
		n = len(rowidDefine.buffer)
	}
	value = string(rowidDefine.buffer[:n])
	return value, err
}
func (rowidDefine *rowidDefine) alloc() error {
	return nil
}
func (rowidDefine *rowidDefine) free() {
}
func (rowidDefine *rowidDefine) close() {
	defer func() {
		recover()
	}()
	rowidDefine.ocidef = nil
	clear(rowidDefine.buffer, 32)
	rowidDefine.environment.rowidDefinePool.Put(rowidDefine)
}
