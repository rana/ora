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

type nilBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
}

func (nilBind *nilBind) bind(position int, sqlt C.ub2, ocistmt *C.OCIStmt) error {
	indp := C.sb2(-1)
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&nilBind.ocibnd), //OCIBind      **bindpp,
		nilBind.environment.ocierr,     //OCIError     *errhp,
		C.ub4(position),                //ub4          position,
		nil,                            //void         *valuep,
		C.sb8(0),                       //sb8          value_sz,
		sqlt,                           //C.SQLT_CHR,                                          //ub2          dty,
		unsafe.Pointer(&indp), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return nilBind.environment.ociError()
	}

	return nil
}

func (nilBind *nilBind) setPtr() error {
	return nil
}

func (nilBind *nilBind) close() {
	defer func() {
		recover()
	}()
	nilBind.ocibnd = nil
	nilBind.environment.nilBindPool.Put(nilBind)
}
