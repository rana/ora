// Copyright 2017 The Ora Authors. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type bndRset struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	ocistmt [1]*C.OCIStmt
	value   *Rset
	nullp
}

func (bnd *bndRset) bind(value *Rset, position namedPos, stmt *Stmt) error {
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind, "%p pos=%v", bnd, position)

	bnd.stmt = stmt
	bnd.value = value
	// Allocate a statement handle
	ocistmt, err := bnd.stmt.ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
	bnd.ocistmt[0] = (*C.OCIStmt)(ocistmt)
	if err != nil {
		return err
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ocistmt[0]), //void         *valuep,
		0,                                   //sb8          value_sz,
		C.SQLT_RSET,                         //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
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

func (bnd *bndRset) setPtr() error {
	if bnd.IsNull() || bnd.ocistmt[0] == nil {
		return nil
	}
	err := bnd.value.open(bnd.stmt, bnd.ocistmt[0])
	bnd.ocistmt[0] = nil
	if err != nil {
		if cerr, ok := err.(interface {
			Code() int
		}); ok && cerr.Code() == 24337 { // statement is not prepared
			bnd.value = nil
			return nil
		}
		return err
	}
	// open result set is successful; will be freed by Rset
	bnd.stmt.openRsets.add(bnd.value)
	return bnd.stmt.setPrefetchSize()
}

func (bnd *bndRset) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ocistmt[0] = nil
	bnd.value = nil
	bnd.nullp.Free()
	stmt.putBnd(bndIdxRset, bnd)
	return nil
}
