// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndInt32 struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
}

func (bnd *bndInt32) bind(value int32, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	r := C.OCINumberFromInt(
		bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),      //const void          *inum,
		byteWidth32,                 //uword               inum_length,
		C.OCI_NUMBER_SIGNED,         //uword               inum_s_flag,
		&bnd.ociNumber[0])           //OCINumber           *number );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r = C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ociNumber[0]), //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber), //sb8          value_sz,
		C.SQLT_VNU,                        //ub2          dty,
		nil,                               //void         *indp,
		nil,                               //ub2          *alenp,
		nil,                               //ub2          *rcodep,
		0,                                 //ub4          maxarr_len,
		nil,                               //ub4          *curelep,
		C.OCI_DEFAULT)                     //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndInt32) setPtr() error {
	return nil
}

func (bnd *bndInt32) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxInt32, bnd)
	return nil
}
