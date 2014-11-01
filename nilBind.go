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
	env    *Environment
	ocibnd *C.OCIBind
}

func (b *nilBind) bind(position int, sqlt C.ub2, ocistmt *C.OCIStmt) error {
	indp := C.sb2(-1)
	r := C.OCIBindByPos2(
		ocistmt,                  //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd), //OCIBind      **bindpp,
		b.env.ocierr,             //OCIError     *errhp,
		C.ub4(position),          //ub4          position,
		nil,                      //void         *valuep,
		C.sb8(0),                 //sb8          value_sz,
		sqlt,                     //C.SQLT_CHR,                                          //ub2          dty,
		unsafe.Pointer(&indp), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}

	return nil
}

func (b *nilBind) setPtr() error {
	return nil
}

func (b *nilBind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.env.nilBindPool.Put(b)
}
