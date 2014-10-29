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

type oraIntervalDSSliceBind struct {
	environment  *Environment
	ocibnd       *C.OCIBind
	ociIntervals []*C.OCIInterval
}

func (intervalDSSliceBind *oraIntervalDSSliceBind) bind(values []IntervalDS, position int, ocistmt *C.OCIStmt) error {
	intervalDSSliceBind.ociIntervals = make([]*C.OCIInterval, len(values))
	nullInds := make([]C.sb2, len(values))
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	for n, value := range values {
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(intervalDSSliceBind.environment.ocienv),                  //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&intervalDSSliceBind.ociIntervals[n])), //dvoid         **descpp,
			C.OCI_DTYPE_INTERVAL_DS,                                                 //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return intervalDSSliceBind.environment.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci interval handle during bind")
		}
		r = C.OCIIntervalSetDaySecond(
			unsafe.Pointer(intervalDSSliceBind.environment.ocienv), //void               *hndl,
			intervalDSSliceBind.environment.ocierr,                 //OCIError           *err,
			C.sb4(value.Day),                                       //sb4                dy,
			C.sb4(value.Hour),                                      //sb4                hr,
			C.sb4(value.Minute),                                    //sb4                mm,
			C.sb4(value.Second),                                    //sb4                ss,
			C.sb4(value.Nanosecond),                                //sb4                fsec,
			intervalDSSliceBind.ociIntervals[n])                    //OCIInterval        *result );
		if r == C.OCI_ERROR {
			return intervalDSSliceBind.environment.ociError()
		}
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			nullInds[n] = C.sb2(0)
		}
		alenp[n] = C.ub4(unsafe.Sizeof(intervalDSSliceBind.ociIntervals[n]))
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&intervalDSSliceBind.ocibnd),                //OCIBind      **bindpp,
		intervalDSSliceBind.environment.ocierr,                    //OCIError     *errhp,
		C.ub4(position),                                           //ub4          position,
		unsafe.Pointer(&intervalDSSliceBind.ociIntervals[0]),      //void         *valuep,
		C.sb8(unsafe.Sizeof(intervalDSSliceBind.ociIntervals[0])), //sb8          value_sz,
		C.SQLT_INTERVAL_DS,                                        //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                              //void         *indp,
		&alenp[0],                                                 //ub2          *alenp,
		&rcodep[0],                                                //ub2          *rcodep,
		0,                                                         //ub4          maxarr_len,
		nil,                                                       //ub4          *curelep,
		C.OCI_DEFAULT)                                             //ub4          mode );
	if r == C.OCI_ERROR {
		return intervalDSSliceBind.environment.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		intervalDSSliceBind.ocibnd,
		intervalDSSliceBind.environment.ocierr,
		C.ub4(unsafe.Sizeof(intervalDSSliceBind.ociIntervals[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                                       //ub4         indskip,
		C.ub4(C.sizeof_ub4),                                       //ub4         alskip,
		C.ub4(C.sizeof_ub2))                                       //ub4         rcskip
	if r == C.OCI_ERROR {
		return intervalDSSliceBind.environment.ociError()
	}
	return nil
}

func (intervalDSSliceBind *oraIntervalDSSliceBind) setPtr() error {
	return nil
}

func (intervalDSSliceBind *oraIntervalDSSliceBind) close() {
	defer func() {
		recover()
	}()
	// release interval descriptor

	for n := 0; n < len(intervalDSSliceBind.ociIntervals); n++ {
		intervalDSSliceBind.freeDescriptor(n)
	}
	intervalDSSliceBind.ocibnd = nil
	intervalDSSliceBind.ociIntervals = nil
	intervalDSSliceBind.environment.oraIntervalDSSliceBindPool.Put(intervalDSSliceBind)
}

func (intervalDSSliceBind *oraIntervalDSSliceBind) freeDescriptor(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(intervalDSSliceBind.ociIntervals[n]), //void     *descp,
		C.OCI_DTYPE_INTERVAL_DS)                             //ub4      type );
}
