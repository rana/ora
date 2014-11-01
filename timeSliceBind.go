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
	"bytes"
	"time"
	"unsafe"
)

type timeSliceBind struct {
	env          *Environment
	ocibnd       *C.OCIBind
	ociDateTimes []*C.OCIDateTime
	zoneBuffer   bytes.Buffer
}

func (b *timeSliceBind) bindOraTimeSlice(values []Time, position int, ocistmt *C.OCIStmt) error {
	timeValues := make([]time.Time, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			timeValues[n] = values[n].Value
		}
	}
	return b.bindTimeSlice(timeValues, nullInds, position, ocistmt)
}

func (b *timeSliceBind) bindTimeSlice(values []time.Time, nullInds []C.sb2, position int, ocistmt *C.OCIStmt) error {
	b.ociDateTimes = make([]*C.OCIDateTime, len(values))
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	for n, timeValue := range values {
		timezoneStr := zoneOffset(timeValue, &b.zoneBuffer)
		ctimezoneStrp := C.CString(timezoneStr)
		defer func() {
			C.free(unsafe.Pointer(ctimezoneStrp))
		}()
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(b.env.ocienv),                          //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&b.ociDateTimes[n])), //dvoid         **descpp,
			C.OCI_DTYPE_TIMESTAMP_TZ,                              //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return b.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci timestamp handle during bind")
		}
		r = C.OCIDateTimeConstruct(
			unsafe.Pointer(b.env.ocienv),                //dvoid         *hndl,
			b.env.ocierr,                                //OCIError      *err,
			b.ociDateTimes[n],                           //OCIDateTime   *datetime,
			C.sb2(timeValue.Year()),                     //sb2           year,
			C.ub1(int32(timeValue.Month())),             //ub1           month,
			C.ub1(timeValue.Day()),                      //ub1           day,
			C.ub1(timeValue.Hour()),                     //ub1           hour,
			C.ub1(timeValue.Minute()),                   //ub1           min,
			C.ub1(timeValue.Second()),                   //ub1           sec,
			C.ub4(timeValue.Nanosecond()),               //ub4           fsec,
			(*C.OraText)(unsafe.Pointer(ctimezoneStrp)), //OraText       *timezone,
			C.size_t(C.strlen(ctimezoneStrp)))           //size_t        timezone_length );
		if r == C.OCI_ERROR {
			return b.env.ociError()
		}
		alenp[n] = C.ub4(unsafe.Sizeof(b.ociDateTimes[n]))
	}

	r := C.OCIBindByPos2(
		ocistmt,                                 //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),                //OCIBind      **bindpp,
		b.env.ocierr,                            //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(&b.ociDateTimes[0]),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociDateTimes[0])), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                     //ub2          dty,
		unsafe.Pointer(&nullInds[0]),            //void         *indp,
		&alenp[0],                               //ub2          *alenp,
		&rcodep[0],                              //ub2          *rcodep,
		0,                                       //ub4          maxarr_len,
		nil,                                     //ub4          *curelep,
		C.OCI_DEFAULT)                           //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		b.ocibnd,
		b.env.ocierr,
		C.ub4(unsafe.Sizeof(b.ociDateTimes[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                     //ub4         indskip,
		C.ub4(C.sizeof_ub4),                     //ub4         alskip,
		C.ub4(C.sizeof_ub2))                     //ub4         rcskip
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *timeSliceBind) setPtr() error {
	return nil
}

func (b *timeSliceBind) close() {
	defer func() {
		recover()
	}()
	// release timestamp descriptor
	for n := 0; n < len(b.ociDateTimes); n++ {
		b.free(n)
	}
	b.ocibnd = nil
	b.ociDateTimes = nil
	b.env.timeSliceBindPool.Put(b)
}

func (b *timeSliceBind) free(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(b.ociDateTimes[n]), //void     *descp,
		C.OCI_DTYPE_TIMESTAMP_TZ)          //ub4      type );
}
