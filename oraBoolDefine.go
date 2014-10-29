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
	environment *Environment
	resultSet   *ResultSet
	ocidef      *C.OCIDefine
	nameStr     string
	isNull      C.sb2
	buffer      []byte
}

func (oraBoolDefine *oraBoolDefine) define(columnSize int, position int, resultSet *ResultSet, ocistmt *C.OCIStmt) error {
	oraBoolDefine.resultSet = resultSet
	if cap(oraBoolDefine.buffer) < columnSize {
		oraBoolDefine.buffer = make([]byte, columnSize)
	}
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&oraBoolDefine.ocidef,                    //OCIDefine   **defnpp,
		oraBoolDefine.environment.ocierr,         //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&oraBoolDefine.buffer[0]), //void        *valuep,
		C.sb8(columnSize),                        //sb8         value_sz,
		C.SQLT_CHR,                               //ub2         dty,
		unsafe.Pointer(&oraBoolDefine.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraBoolDefine.environment.ociError()
	}
	return nil
}
func (oraBoolDefine *oraBoolDefine) value() (value interface{}, err error) {
	boolValue := Bool{IsNull: oraBoolDefine.isNull < 0}
	if !boolValue.IsNull {
		boolValue.Value = bytes.Runes(oraBoolDefine.buffer)[0] == oraBoolDefine.resultSet.Config.TrueRune
	}
	return boolValue, err
}
func (oraBoolDefine *oraBoolDefine) alloc() error {
	return nil
}
func (oraBoolDefine *oraBoolDefine) free() {

}
func (oraBoolDefine *oraBoolDefine) close() {
	defer func() {
		recover()
	}()
	oraBoolDefine.resultSet = nil
	oraBoolDefine.ocidef = nil
	oraBoolDefine.isNull = C.sb2(0)
	clear(oraBoolDefine.buffer, 0)
	oraBoolDefine.environment.oraBoolDefinePool.Put(oraBoolDefine)
}
