// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type bndString struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	cString *C.char
}

func (bnd *bndString) bind(value string, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.cString = C.CString(value)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),  //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position),             //ub4          position,
		unsafe.Pointer(bnd.cString), //void         *valuep,
		C.LENGTH_TYPE(len(value)),   //sb8          value_sz,
		C.SQLT_CHR,                  //ub2          dty,
		nil,                         //void         *indp,
		nil,                         //ub2          *alenp,
		nil,                         //ub2          *rcodep,
		0,                           //ub4          maxarr_len,
		nil,                         //ub4          *curelep,
		C.OCI_DEFAULT)               //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndString) setPtr() error {
	return nil
}

func (bnd *bndString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()
	C.free(unsafe.Pointer(bnd.cString))
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.cString = nil
	stmt.putBnd(bndIdxString, bnd)
	return nil
}
