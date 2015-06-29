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
	stmt        *Stmt
	ocibnd      *C.OCIBind
	ociDateTime *C.OCIDateTime
	isNull      C.sb2
	value       *time.Time
	cZone       *C.char
}

func (bnd *bndTimePtr) bind(value *time.Time, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),         //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&bnd.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                            //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	r = C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                              //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                    //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                   //OCIError     *errhp,
		C.ub4(position),                               //ub4          position,
		unsafe.Pointer(&bnd.ociDateTime),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociDateTime)), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                           //ub2          dty,
		unsafe.Pointer(&bnd.isNull),                   //void         *indp,
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
	if bnd.value != nil && bnd.isNull > C.sb2(-1) {
		*bnd.value, err = getTime(bnd.stmt.ses.srv.env, bnd.ociDateTime)
	}
	return err
}

func (bnd *bndTimePtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	if bnd.cZone != nil {
		C.free(unsafe.Pointer(bnd.cZone))
		bnd.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.ociDateTime), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)        //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociDateTime = nil
	bnd.value = nil
	stmt.putBnd(bndIdxTimePtr, bnd)
	return nil
}
