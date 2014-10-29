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
	"unsafe"
)

type stringBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	cstringp    *C.char
}

func (stringBind *stringBind) bind(value string, position int, ocistmt *C.OCIStmt) error {
	stringBind.cstringp = C.CString(value)
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&stringBind.ocibnd),    //OCIBind      **bindpp,
		stringBind.environment.ocierr,        //OCIError     *errhp,
		C.ub4(position),                      //ub4          position,
		unsafe.Pointer(stringBind.cstringp),  //void         *valuep,
		C.sb8(C.strlen(stringBind.cstringp)), //sb8          value_sz,
		C.SQLT_CHR,                           //ub2          dty,
		nil,                                  //void         *indp,
		nil,                                  //ub2          *alenp,
		nil,                                  //ub2          *rcodep,
		0,                                    //ub4          maxarr_len,
		nil,                                  //ub4          *curelep,
		C.OCI_DEFAULT)                        //ub4          mode );
	if r == C.OCI_ERROR {
		return stringBind.environment.ociError()
	}
	return nil
}

func (stringBind *stringBind) setPtr() error {
	return nil
}

func (stringBind *stringBind) close() {
	defer func() {
		recover()
	}()
	// free c-string memory
	C.free(unsafe.Pointer(stringBind.cstringp))
	stringBind.ocibnd = nil
	stringBind.environment.stringBindPool.Put(stringBind)
}
