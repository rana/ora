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

type intervalDSDefine struct {
	env         *Environment
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	isNull      C.sb2
}

func (d *intervalDSDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                             //OCIStmt     *stmtp,
		&d.ocidef,                           //OCIDefine   **defnpp,
		d.env.ocierr,                        //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&d.ociInterval),      //void        *valuep,
		C.sb8(unsafe.Sizeof(d.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_DS,                  //ub2         dty,
		unsafe.Pointer(&d.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}

func (d *intervalDSDefine) value() (value interface{}, err error) {
	intervalDS := IntervalDS{IsNull: d.isNull < 0}
	if !intervalDS.IsNull {
		var day C.sb4
		var hour C.sb4
		var minute C.sb4
		var second C.sb4
		var nanosecond C.sb4
		r := C.OCIIntervalGetDaySecond(
			unsafe.Pointer(d.env.ocienv), //void               *hndl,
			d.env.ocierr,                 //OCIError           *err,
			&day,                         //sb4                *dy,
			&hour,                        //sb4                *hr,
			&minute,                      //sb4                *mm,
			&second,                      //sb4                *ss,
			&nanosecond,                  //sb4                *fsec,
			d.ociInterval)                //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = d.env.ociError()
		}
		intervalDS.Day = int32(day)
		intervalDS.Hour = int32(hour)
		intervalDS.Minute = int32(minute)
		intervalDS.Second = int32(second)
		intervalDS.Nanosecond = int32(nanosecond)
	}
	return intervalDS, err
}

func (d *intervalDSDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_DS,                           //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (d *intervalDSDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_DS)       //timeDefine.descTypeCode)                //ub4      type );
}

func (d *intervalDSDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociInterval = nil
	d.isNull = C.sb2(0)
	d.env.intervalDSDefinePool.Put(d)
}
