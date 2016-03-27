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
	stmt    *Stmt
	ocibnd  *C.OCIBind
	ociDate C.OCIDate
}

func (bnd *bndDate) bind(value time.Time, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	ociSetDateTime(&bnd.ociDate, value)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),      //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,     //OCIError     *errhp,
		C.ub4(position),                 //ub4          position,
		unsafe.Pointer(&bnd.ociDate),    //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCIDate), //sb8          value_sz,
		C.SQLT_ODT,                      //ub2          dty,
		nil,                             //void         *indp,
		nil,                             //ub2          *alenp,
		nil,                             //ub2          *rcodep,
		0,                               //ub4          maxarr_len,
		nil,                             //ub4          *curelep,
		C.OCI_DEFAULT)                   //ub4          mode );
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

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxDate, bnd)
	return nil
}

func ociSetDateTime(ociDate *C.OCIDate, value time.Time) {
	value = value.Local()
	//OCIDateSetDate and OCIDateSetTime are just macros, don't play well with cgo
	ociDate.OCIDateYYYY = C.sb2(value.Year())
	ociDate.OCIDateMM = C.ub1(int32(value.Month()))
	ociDate.OCIDateDD = C.ub1(value.Day())
	ociDate.OCIDateTime.OCITimeHH = C.ub1(value.Hour())
	ociDate.OCIDateTime.OCITimeMI = C.ub1(value.Minute())
	ociDate.OCIDateTime.OCITimeSS = C.ub1(value.Second())
}

func ociGetDateTime(ociDate C.OCIDate) time.Time {
	//OCIDateGetDate and OCIDateGetTime are just macros, don't play well with cgo
	return time.Date(
		int(ociDate.OCIDateYYYY),
		time.Month(ociDate.OCIDateMM),
		int(ociDate.OCIDateDD),
		int(ociDate.OCIDateTime.OCITimeHH),
		int(ociDate.OCIDateTime.OCITimeMI),
		int(ociDate.OCIDateTime.OCITimeSS),
		0,
		time.Local)
}
