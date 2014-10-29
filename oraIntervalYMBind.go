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

type oraIntervalYMBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociInterval *C.OCIInterval
}

func (oraIntervalYMBind *oraIntervalYMBind) bind(value IntervalYM, position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(oraIntervalYMBind.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&oraIntervalYMBind.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_YM,                                           //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return oraIntervalYMBind.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during bind")
	}
	r = C.OCIIntervalSetYearMonth(
		unsafe.Pointer(oraIntervalYMBind.environment.ocienv), //void               *hndl,
		oraIntervalYMBind.environment.ocierr,                 //OCIError           *err,
		C.sb4(value.Year),                                    //sb4                yr,
		C.sb4(value.Month),                                   //sb4                mnth,
		oraIntervalYMBind.ociInterval)                        //OCIInterval        *result );
	if r == C.OCI_ERROR {
		return oraIntervalYMBind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&oraIntervalYMBind.ocibnd),            //OCIBind      **bindpp,
		oraIntervalYMBind.environment.ocierr,                //OCIError     *errhp,
		C.ub4(position),                                     //ub4          position,
		unsafe.Pointer(&oraIntervalYMBind.ociInterval),      //void         *valuep,
		C.sb8(unsafe.Sizeof(oraIntervalYMBind.ociInterval)), //sb8          value_sz,
		C.SQLT_INTERVAL_YM,                                  //ub2          dty,
		nil,                                                 //void         *indp,
		nil,                                                 //ub2          *alenp,
		nil,                                                 //ub2          *rcodep,
		0,                                                   //ub4          maxarr_len,
		nil,                                                 //ub4          *curelep,
		C.OCI_DEFAULT)                                       //ub4          mode );
	if r == C.OCI_ERROR {
		return oraIntervalYMBind.environment.ociError()
	}
	return nil
}

func (oraIntervalYMBind *oraIntervalYMBind) setPtr() error {
	return nil
}

func (oraIntervalYMBind *oraIntervalYMBind) close() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(oraIntervalYMBind.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)                       //timeDefine.descTypeCode)                //ub4      type );
	oraIntervalYMBind.ocibnd = nil
	oraIntervalYMBind.ociInterval = nil
	oraIntervalYMBind.environment.oraIntervalYMBindPool.Put(oraIntervalYMBind)
}
