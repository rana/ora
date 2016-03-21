// Copyright 2014 Rana Ian. All rights reserved.
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

type bndInt32Ptr struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
	value     *int32
	nullp
}

func (bnd *bndInt32Ptr) bind(value *int32, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	bnd.nullp.Set(value == nil)
	if value != nil {
		r := C.OCINumberFromInt(
			bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
			unsafe.Pointer(value),       //const void          *inum,
			4,                   //uword               inum_length,
			C.OCI_NUMBER_SIGNED, //uword               inum_s_flag,
			&bnd.ociNumber[0])   //OCINumber           *number
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
			"Int32Ptr.bind(%d) value=%d => number=%#v", position, *value, bnd.ociNumber[0])
	}
	alen := C.ACTUAL_LENGTH_TYPE(4)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,         //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(&bnd.ociNumber[0]),   //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),   //sb8          value_sz,
		C.SQLT_VNU,                          //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
		&alen,         //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndInt32Ptr) setPtr() error {
	if !bnd.nullp.IsNull() {
		r := C.OCINumberToInt(
			bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
			&bnd.ociNumber[0],           //const OCINumber       *number,
			4,                           //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,         //uword                 rsl_flag,
			unsafe.Pointer(bnd.value))   //void                  *rsl );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
			"Int32Ptr.setPtr number=%#v => value=%d", bnd.ociNumber[0], *bnd.value)
	}
	return nil
}

func (bnd *bndInt32Ptr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind, "Int32Ptr.close value=%p", bnd.value)

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	stmt.putBnd(bndIdxInt32Ptr, bnd)
	return nil
}
