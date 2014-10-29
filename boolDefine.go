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
	environment *Environment
	resultSet   *ResultSet
	ocidef      *C.OCIDefine
	isNull      C.sb2
	buffer      []byte
}

func (boolDefine *boolDefine) define(columnSize int, position int, resultSet *ResultSet, ocistmt *C.OCIStmt) error {
	boolDefine.resultSet = resultSet
	if cap(boolDefine.buffer) < columnSize {
		boolDefine.buffer = make([]byte, columnSize)
	}
	// Create oci define handle
	r := C.OCIDefineByPos2(
		ocistmt,                               //OCIStmt     *stmtp,
		&boolDefine.ocidef,                    //OCIDefine   **defnpp,
		boolDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                       //ub4         position,
		unsafe.Pointer(&boolDefine.buffer[0]), //void        *valuep,
		C.sb8(columnSize),                     //sb8         value_sz,
		C.SQLT_CHR,                            //ub2         dty,
		unsafe.Pointer(&boolDefine.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return boolDefine.environment.ociError()
	}
	return nil
}
func (boolDefine *boolDefine) value() (value interface{}, err error) {
	if boolDefine.isNull > -1 {
		value = bytes.Runes(boolDefine.buffer)[0] == boolDefine.resultSet.Config.TrueRune
	}
	return value, err
}
func (boolDefine *boolDefine) alloc() error {
	return nil
}
func (boolDefine *boolDefine) free() {

}
func (boolDefine *boolDefine) close() {
	defer func() {
		recover()
	}()
	boolDefine.ocidef = nil
	boolDefine.isNull = C.sb2(0)
	clear(boolDefine.buffer, 0)
	boolDefine.environment.boolDefinePool.Put(boolDefine)
}
