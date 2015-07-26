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
	ociNumber C.OCINumber
	isNull    C.sb2
	value     *int32
}

func (bnd *bndInt32Ptr) bind(value *int32, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	if value == nil {
		bnd.isNull = C.sb2(-1)
	} else {
		bnd.isNull = 0
		r := C.OCINumberFromInt(
			bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
			unsafe.Pointer(value),       //const void          *inum,
			4,                   //uword               inum_length,
			C.OCI_NUMBER_SIGNED, //uword               inum_s_flag,
			&bnd.ociNumber)      //OCINumber           *number
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
			"%p pos=%d value(%p)=%d => number=%#v", bnd, position, bnd.value, *value, bnd.ociNumber)
	}
	alen := C.ACTUAL_LENGTH_TYPE(4)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                  //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),        //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,       //OCIError     *errhp,
		C.ub4(position),                   //ub4          position,
		unsafe.Pointer(&bnd.ociNumber),    //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber), //sb8          value_sz,
		C.SQLT_VNU,                        //ub2          dty,
		unsafe.Pointer(&bnd.isNull),       //void         *indp,
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
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p value=%p isNull=%d number=%#v", bnd, bnd.value, bnd.isNull, bnd.ociNumber)
	if bnd.isNull > C.sb2(-1) {
		r := C.OCINumberToInt(
			bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
			&bnd.ociNumber,              //const OCINumber       *number,
			C.uword(4),                  //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,         //uword                 rsl_flag,
			unsafe.Pointer(bnd.value))   //void                  *rsl );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
			"Int32Ptr.setPtr number=%#v => value=%d", bnd.ociNumber, *bnd.value)
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
