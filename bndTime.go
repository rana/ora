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
	"time"
	"unsafe"
)

type bndTime struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	dateTimep
}

func (bnd *bndTime) bind(value time.Time, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	if err := bnd.dateTimep.Set(bnd.stmt.ses.srv.env, value); err != nil {
		return err
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		&bnd.ocibnd,                 //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(bnd.dateTimep.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.dateTimep.Size()),     //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                     //ub2          dty,
		nil,                                     //void         *indp,
		nil,                                     //ub2          *alenp,
		nil,                                     //ub2          *rcodep,
		0,                                       //ub4          maxarr_len,
		nil,                                     //ub4          *curelep,
		C.OCI_DEFAULT)                           //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndTime) setPtr() (err error) {
	return nil
}

func (bnd *bndTime) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	bnd.dateTimep.Free()
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxTime, bnd)
	return nil
}
