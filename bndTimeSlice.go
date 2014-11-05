// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"github.com/golang/glog"
	"time"
	"unsafe"
)

type bndTimeSlice struct {
	stmt         *Stmt
	ocibnd       *C.OCIBind
	ociDateTimes []*C.OCIDateTime
	zoneBuf      bytes.Buffer
}

func (bnd *bndTimeSlice) bindOra(values []Time, position int, stmt *Stmt) error {
	timeValues := make([]time.Time, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			timeValues[n] = values[n].Value
		}
	}
	return bnd.bind(timeValues, nullInds, position, stmt)
}

func (bnd *bndTimeSlice) bind(values []time.Time, nullInds []C.sb2, position int, stmt *Stmt) error {
	glog.Infoln("position: ", position)
	bnd.stmt = stmt
	bnd.ociDateTimes = make([]*C.OCIDateTime, len(values))
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	for n, timeValue := range values {
		timezoneStr := zoneOffset(timeValue, &bnd.zoneBuf)
		cTimezoneStr := C.CString(timezoneStr)
		defer func() {
			C.free(unsafe.Pointer(cTimezoneStr))
		}()
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),             //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&bnd.ociDateTimes[n])), //dvoid         **descpp,
			C.OCI_DTYPE_TIMESTAMP_TZ,                                //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci timestamp handle during bind")
		}
		r = C.OCIDateTimeConstruct(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv), //dvoid         *hndl,
			bnd.stmt.ses.srv.env.ocierr,                 //OCIError      *err,
			bnd.ociDateTimes[n],                         //OCIDateTime   *datetime,
			C.sb2(timeValue.Year()),                     //sb2           year,
			C.ub1(int32(timeValue.Month())),             //ub1           month,
			C.ub1(timeValue.Day()),                      //ub1           day,
			C.ub1(timeValue.Hour()),                     //ub1           hour,
			C.ub1(timeValue.Minute()),                   //ub1           min,
			C.ub1(timeValue.Second()),                   //ub1           sec,
			C.ub4(timeValue.Nanosecond()),               //ub4           fsec,
			(*C.OraText)(unsafe.Pointer(cTimezoneStr)),  //OraText       *timezone,
			C.size_t(len(timezoneStr)))                  //size_t        timezone_length );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		alenp[n] = C.ub4(unsafe.Sizeof(bnd.ociDateTimes[n]))
	}

	r := C.OCIBindByPos2(
		bnd.stmt.ocistmt,                          //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,               //OCIError     *errhp,
		C.ub4(position),                           //ub4          position,
		unsafe.Pointer(&bnd.ociDateTimes[0]),      //void         *valuep,
		C.sb8(unsafe.Sizeof(bnd.ociDateTimes[0])), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                       //ub2          dty,
		unsafe.Pointer(&nullInds[0]),              //void         *indp,
		&alenp[0],                                 //ub2          *alenp,
		&rcodep[0],                                //ub2          *rcodep,
		0,                                         //ub4          maxarr_len,
		nil,                                       //ub4          *curelep,
		C.OCI_DEFAULT)                             //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(unsafe.Sizeof(bnd.ociDateTimes[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                       //ub4         indskip,
		C.ub4(C.sizeof_ub4),                       //ub4         alskip,
		C.ub4(C.sizeof_ub2))                       //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndTimeSlice) setPtr() error {
	return nil
}

func (bnd *bndTimeSlice) free(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(bnd.ociDateTimes[n]), //void     *descp,
		C.OCI_DTYPE_TIMESTAMP_TZ)            //ub4      type );
}

func (bnd *bndTimeSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	for n := range bnd.ociDateTimes {
		bnd.free(n)
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociDateTimes = nil
	bnd.zoneBuf.Reset()
	stmt.putBnd(bndIdxTimeSlice, bnd)
	return nil
}
