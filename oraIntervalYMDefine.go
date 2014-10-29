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
	environment *Environment
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	isNull      C.sb2
}

func (intervalYMDefine *intervalYMDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                            //OCIStmt     *stmtp,
		&intervalYMDefine.ocidef,                           //OCIDefine   **defnpp,
		intervalYMDefine.environment.ocierr,                //OCIError    *errhp,
		C.ub4(position),                                    //ub4         position,
		unsafe.Pointer(&intervalYMDefine.ociInterval),      //void        *valuep,
		C.sb8(unsafe.Sizeof(intervalYMDefine.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_YM,                                 //ub2         dty,
		unsafe.Pointer(&intervalYMDefine.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return intervalYMDefine.environment.ociError()
	}
	return nil
}

func (intervalYMDefine *intervalYMDefine) value() (value interface{}, err error) {
	intervalYM := IntervalYM{IsNull: intervalYMDefine.isNull < 0}
	if !intervalYM.IsNull {
		var year C.sb4
		var month C.sb4
		r := C.OCIIntervalGetYearMonth(
			unsafe.Pointer(intervalYMDefine.environment.ocienv), //void               *hndl,
			intervalYMDefine.environment.ocierr,                 //OCIError           *err,
			&year,  //sb4                *yr,
			&month, //sb4                *mnth,
			intervalYMDefine.ociInterval) //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = intervalYMDefine.environment.ociError()
		}
		intervalYM.Year = int32(year)
		intervalYM.Month = int32(month)
	}
	return intervalYM, err
}

func (intervalYMDefine *intervalYMDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(intervalYMDefine.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&intervalYMDefine.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_YM,                                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return intervalYMDefine.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (intervalYMDefine *intervalYMDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(intervalYMDefine.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)                      //timeDefine.descTypeCode)                //ub4      type );
}

func (intervalYMDefine *intervalYMDefine) close() {
	defer func() {
		recover()
	}()
	intervalYMDefine.ocidef = nil
	intervalYMDefine.ociInterval = nil
	intervalYMDefine.isNull = C.sb2(0)
	intervalYMDefine.environment.intervalYMDefinePool.Put(intervalYMDefine)
}
