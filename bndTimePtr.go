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

type bndTimePtr struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	value  *time.Time
	dateTimep
	nullp
}

func (bnd *bndTimePtr) bind(value *time.Time, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.nullp.Set(value == nil || value.IsZero())
	if err := bnd.dateTimep.Alloc(bnd.stmt.ses.srv.env); err != nil {
		return err
	}
	bnd.value = value
	if value != nil {
		if err := bnd.dateTimep.Set(bnd.stmt.ses.srv.env, *value); err != nil {
			return err
		}
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(bnd.dateTimep.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.dateTimep.Size()),     //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                     //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()),     //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndTimePtr) setPtr() (err error) {
	if bnd.value == nil { // cannot set on a nil pointer
		return nil
	}
	if bnd.nullp.IsNull() {
		*bnd.value = time.Time{} // zero time
		return nil
	}
	*bnd.value, err = getTime(bnd.stmt.ses.srv.env, bnd.dateTimep.Value())
	return err
}

func (bnd *bndTimePtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.dateTimep.Free()
	bnd.nullp.Free()
	stmt.putBnd(bndIdxTimePtr, bnd)
	return nil
}
