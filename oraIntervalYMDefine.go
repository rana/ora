// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

type intervalYMDefine struct {
	env         *Environment
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	isNull      C.sb2
}

func (d *intervalYMDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                             //OCIStmt     *stmtp,
		&d.ocidef,                           //OCIDefine   **defnpp,
		d.env.ocierr,                        //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&d.ociInterval),      //void        *valuep,
		C.sb8(unsafe.Sizeof(d.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_YM,                  //ub2         dty,
		unsafe.Pointer(&d.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}

func (d *intervalYMDefine) value() (value interface{}, err error) {
	intervalYM := IntervalYM{IsNull: d.isNull < 0}
	if !intervalYM.IsNull {
		var year C.sb4
		var month C.sb4
		r := C.OCIIntervalGetYearMonth(
			unsafe.Pointer(d.env.ocienv), //void               *hndl,
			d.env.ocierr,                 //OCIError           *err,
			&year,                        //sb4                *yr,
			&month,                       //sb4                *mnth,
			d.ociInterval)                //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = d.env.ociError()
		}
		intervalYM.Year = int32(year)
		intervalYM.Month = int32(month)
	}
	return intervalYM, err
}

func (d *intervalYMDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_YM,                           //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (d *intervalYMDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)       //timeDefine.descTypeCode)                //ub4      type );
}

func (d *intervalYMDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociInterval = nil
	d.isNull = C.sb2(0)
	d.env.intervalYMDefinePool.Put(d)
}
