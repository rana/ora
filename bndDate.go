// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <orl.h>
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import (
	"time"
	"unsafe"
)

type bndDate struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	datep
}

func (bnd *bndDate) bind(value time.Time, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	if err := bnd.datep.Set(bnd.stmt.ses.srv.env, value); err != nil {
		return err
	}
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                    //OCIStmt      *stmtp,
		&bnd.ocibnd,                         //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,         //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(bnd.datep.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.datep.Size()),     //sb8          value_sz,
		C.SQLT_ODT,                          //ub2          dty,
		nil,                                 //void         *indp,
		nil,                                 //ub2          *alenp,
		nil,                                 //ub2          *rcodep,
		0,                                   //ub4          maxarr_len,
		nil,                                 //ub4          *curelep,
		C.OCI_DEFAULT)                       //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndDate) setPtr() (err error) {
	return nil
}

func (bnd *bndDate) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	bnd.datep.Free()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxDate, bnd)
	return nil
}
