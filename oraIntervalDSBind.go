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

type oraIntervalDSBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociInterval *C.OCIInterval
}

func (oraIntervalDSBind *oraIntervalDSBind) bind(value IntervalDS, position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(oraIntervalDSBind.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&oraIntervalDSBind.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_DS,                                           //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return oraIntervalDSBind.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during bind")
	}
	r = C.OCIIntervalSetDaySecond(
		unsafe.Pointer(oraIntervalDSBind.environment.ocienv), //void               *hndl,
		oraIntervalDSBind.environment.ocierr,                 //OCIError           *err,
		C.sb4(value.Day),                                     //sb4                dy,
		C.sb4(value.Hour),                                    //sb4                hr,
		C.sb4(value.Minute),                                  //sb4                mm,
		C.sb4(value.Second),                                  //sb4                ss,
		C.sb4(value.Nanosecond),                              //sb4                fsec,
		oraIntervalDSBind.ociInterval)                        //OCIInterval        *result );
	if r == C.OCI_ERROR {
		return oraIntervalDSBind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&oraIntervalDSBind.ocibnd),            //OCIBind      **bindpp,
		oraIntervalDSBind.environment.ocierr,                //OCIError     *errhp,
		C.ub4(position),                                     //ub4          position,
		unsafe.Pointer(&oraIntervalDSBind.ociInterval),      //void         *valuep,
		C.sb8(unsafe.Sizeof(oraIntervalDSBind.ociInterval)), //sb8          value_sz,
		C.SQLT_INTERVAL_DS,                                  //ub2          dty,
		nil,                                                 //void         *indp,
		nil,                                                 //ub2          *alenp,
		nil,                                                 //ub2          *rcodep,
		0,                                                   //ub4          maxarr_len,
		nil,                                                 //ub4          *curelep,
		C.OCI_DEFAULT)                                       //ub4          mode );
	if r == C.OCI_ERROR {
		return oraIntervalDSBind.environment.ociError()
	}
	return nil
}

func (oraIntervalDSBind *oraIntervalDSBind) setPtr() error {
	return nil
}

func (oraIntervalDSBind *oraIntervalDSBind) close() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(oraIntervalDSBind.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_DS)                       //timeDefine.descTypeCode)
	oraIntervalDSBind.ocibnd = nil
	oraIntervalDSBind.ociInterval = nil
	oraIntervalDSBind.environment.oraIntervalDSBindPool.Put(oraIntervalDSBind)
}
