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
	environment *Environment
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	isNull      C.sb2
}

func (intervalDSDefine *intervalDSDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                            //OCIStmt     *stmtp,
		&intervalDSDefine.ocidef,                           //OCIDefine   **defnpp,
		intervalDSDefine.environment.ocierr,                //OCIError    *errhp,
		C.ub4(position),                                    //ub4         position,
		unsafe.Pointer(&intervalDSDefine.ociInterval),      //void        *valuep,
		C.sb8(unsafe.Sizeof(intervalDSDefine.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_DS,                                 //ub2         dty,
		unsafe.Pointer(&intervalDSDefine.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return intervalDSDefine.environment.ociError()
	}
	return nil
}

func (intervalDSDefine *intervalDSDefine) value() (value interface{}, err error) {
	intervalDS := IntervalDS{IsNull: intervalDSDefine.isNull < 0}
	if !intervalDS.IsNull {
		var day C.sb4
		var hour C.sb4
		var minute C.sb4
		var second C.sb4
		var nanosecond C.sb4
		r := C.OCIIntervalGetDaySecond(
			unsafe.Pointer(intervalDSDefine.environment.ocienv), //void               *hndl,
			intervalDSDefine.environment.ocierr,                 //OCIError           *err,
			&day,                         //sb4                *dy,
			&hour,                        //sb4                *hr,
			&minute,                      //sb4                *mm,
			&second,                      //sb4                *ss,
			&nanosecond,                  //sb4                *fsec,
			intervalDSDefine.ociInterval) //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = intervalDSDefine.environment.ociError()
		}
		intervalDS.Day = int32(day)
		intervalDS.Hour = int32(hour)
		intervalDS.Minute = int32(minute)
		intervalDS.Second = int32(second)
		intervalDS.Nanosecond = int32(nanosecond)
	}
	return intervalDS, err
}

func (intervalDSDefine *intervalDSDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(intervalDSDefine.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&intervalDSDefine.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_DS,                                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return intervalDSDefine.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (intervalDSDefine *intervalDSDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(intervalDSDefine.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_DS)                      //timeDefine.descTypeCode)                //ub4      type );
}

func (intervalDSDefine *intervalDSDefine) close() {
	defer func() {
		recover()
	}()
	intervalDSDefine.ocidef = nil
	intervalDSDefine.ociInterval = nil
	intervalDSDefine.isNull = C.sb2(0)
	intervalDSDefine.environment.intervalDSDefinePool.Put(intervalDSDefine)
}
