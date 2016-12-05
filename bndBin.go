// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndBin struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
}

func (bnd *bndBin) bind(value []byte, position namedPos, stmt *Stmt) (err error) {
	bnd.stmt = stmt
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
		unsafe.Pointer(&value[0]), //void         *valuep,
		C.LENGTH_TYPE(len(value)), //sb8          value_sz,
		C.SQLT_LBI,                //ub2          dty,
		nil,                       //void         *indp,
		nil,                       //ub2          *alenp,
		nil,                       //ub2          *rcodep,
		0,                         //ub4          maxarr_len,
		nil,                       //ub4          *curelep,
		C.OCI_DEFAULT)             //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndBin) setPtr() error {
	return nil
}

func (bnd *bndBin) free() {
}

func (bnd *bndBin) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt := bnd.stmt
	bnd.stmt = nil
	stmt.putBnd(bndIdxBin, bnd)
	return nil
}
