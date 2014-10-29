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

type oraIntervalYMSliceBind struct {
	environment  *Environment
	ocibnd       *C.OCIBind
	ociIntervals []*C.OCIInterval
}

func (oraIntervalYMSliceBind *oraIntervalYMSliceBind) bind(values []IntervalYM, position int, ocistmt *C.OCIStmt) error {
	oraIntervalYMSliceBind.ociIntervals = make([]*C.OCIInterval, len(values))
	nullInds := make([]C.sb2, len(values))
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	for n, value := range values {
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(oraIntervalYMSliceBind.environment.ocienv),                  //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&oraIntervalYMSliceBind.ociIntervals[n])), //dvoid         **descpp,
			C.OCI_DTYPE_INTERVAL_YM,                                                    //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return oraIntervalYMSliceBind.environment.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci interval handle during bind")
		}
		r = C.OCIIntervalSetYearMonth(
			unsafe.Pointer(oraIntervalYMSliceBind.environment.ocienv), //void               *hndl,
			oraIntervalYMSliceBind.environment.ocierr,                 //OCIError           *err,
			C.sb4(value.Year),                                         //sb4                yr,
			C.sb4(value.Month),                                        //sb4                mnth,
			oraIntervalYMSliceBind.ociIntervals[n])                    //OCIInterval        *result );
		if r == C.OCI_ERROR {
			return oraIntervalYMSliceBind.environment.ociError()
		}
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			nullInds[n] = C.sb2(0)
		}
		alenp[n] = C.ub4(unsafe.Sizeof(oraIntervalYMSliceBind.ociIntervals[n]))
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&oraIntervalYMSliceBind.ocibnd),                //OCIBind      **bindpp,
		oraIntervalYMSliceBind.environment.ocierr,                    //OCIError     *errhp,
		C.ub4(position),                                              //ub4          position,
		unsafe.Pointer(&oraIntervalYMSliceBind.ociIntervals[0]),      //void         *valuep,
		C.sb8(unsafe.Sizeof(oraIntervalYMSliceBind.ociIntervals[0])), //sb8          value_sz,
		C.SQLT_INTERVAL_YM,                                           //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                                 //void         *indp,
		&alenp[0],                                                    //ub2          *alenp,
		&rcodep[0],                                                   //ub2          *rcodep,
		0,                                                            //ub4          maxarr_len,
		nil,                                                          //ub4          *curelep,
		C.OCI_DEFAULT)                                                //ub4          mode );
	if r == C.OCI_ERROR {
		return oraIntervalYMSliceBind.environment.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		oraIntervalYMSliceBind.ocibnd,
		oraIntervalYMSliceBind.environment.ocierr,
		C.ub4(unsafe.Sizeof(oraIntervalYMSliceBind.ociIntervals[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                                          //ub4         indskip,
		C.ub4(C.sizeof_ub4),                                          //ub4         alskip,
		C.ub4(C.sizeof_ub2))                                          //ub4         rcskip
	if r == C.OCI_ERROR {
		return oraIntervalYMSliceBind.environment.ociError()
	}
	return nil
}

func (oraIntervalYMSliceBind *oraIntervalYMSliceBind) setPtr() error {
	return nil
}

func (oraIntervalYMSliceBind *oraIntervalYMSliceBind) close() {
	defer func() {
		recover()
	}()
	// release interval descriptor
	for n := 0; n < len(oraIntervalYMSliceBind.ociIntervals); n++ {
		oraIntervalYMSliceBind.freeDescriptor(n)
	}
	oraIntervalYMSliceBind.ocibnd = nil
	oraIntervalYMSliceBind.ociIntervals = nil
	oraIntervalYMSliceBind.environment.oraIntervalYMSliceBindPool.Put(oraIntervalYMSliceBind)
}

func (oraIntervalYMSliceBind *oraIntervalYMSliceBind) freeDescriptor(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(oraIntervalYMSliceBind.ociIntervals[n]), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)                                //ub4      type );
}
