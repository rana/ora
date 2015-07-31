// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import (
	"bytes"
	"time"
	"unsafe"
)

type bndTimeSlice struct {
	stmt         *Stmt
	ocibnd       *C.OCIBind
	ociDateTimes []*C.OCIDateTime
	zoneBuf      bytes.Buffer
	values       []Time
	times        []time.Time
	arrHlp
}

func (bnd *bndTimeSlice) bindOra(values []Time, position int, stmt *Stmt) (uint32, error) {
	bnd.values = values
	if cap(bnd.times) < cap(values) {
		bnd.times = make([]time.Time, len(values), cap(values))
	} else {
		bnd.times = bnd.times[:len(values)]
	}
	if cap(bnd.nullInds) < cap(values) {
		bnd.nullInds = make([]C.sb2, len(values), cap(values))
	} else {
		bnd.nullInds = bnd.nullInds[:len(values)]
	}
	for n, _ := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[0] = 0
			bnd.times[n] = values[n].Value
		}
	}
	return bnd.bind(bnd.times, position, stmt)
}

func (bnd *bndTimeSlice) bind(values []time.Time, position int, stmt *Stmt) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, stmt.stmtType)
	if needAppend {
		values = append(values, time.Time{})
	}
	bnd.times = values
	if cap(bnd.ociDateTimes) < C {
		bnd.ociDateTimes = make([]*C.OCIDateTime, L, C)
	} else {
		bnd.ociDateTimes = bnd.ociDateTimes[:L]
	}
	for n, timeValue := range values {
		bnd.zoneBuf.Reset()
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
			return iterations, bnd.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return iterations, errNew("unable to allocate oci timestamp handle during bind")
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
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}
		bnd.alen[n] = C.ACTUAL_LENGTH_TYPE(unsafe.Sizeof(bnd.ociDateTimes[n]))
	}

	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p pos=%d cap=%d len=%d curlen=%d curlenp=%p value_sz=%d alen=%v",
		bnd, position, cap(bnd.ociDateTimes), len(bnd.ociDateTimes), bnd.curlen, curlenp,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociDateTimes[0])), //sb8          value_sz,
		bnd.alen)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                                  //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                        //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                       //OCIError     *errhp,
		C.ub4(position),                                   //ub4          position,
		unsafe.Pointer(&bnd.ociDateTimes[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociDateTimes[0])), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                               //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),                  //void         *indp,
		&bnd.alen[0],                                      //ub2          *alenp,
		&bnd.rcode[0],                                     //ub2          *rcodep,
		C.ACTUAL_LENGTH_TYPE(C),                           //ub4          maxarr_len,
		curlenp,                                           //ub4          *curelep,
		C.OCI_DEFAULT)                                     //ub4          mode );
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	/*
		r = C.OCIBindArrayOfStruct(
			bnd.ocibnd,
			bnd.stmt.ses.srv.env.ocierr,
			C.ub4(unsafe.Sizeof(bnd.ociDateTimes[0])), //ub4         pvskip,
			C.ub4(C.sizeof_sb2),                       //ub4         indskip,
			C.ub4(C.sizeof_ACTUAL_LENGTH_TYPE),        //ub4         alskip,
			C.ub4(C.sizeof_ub2))                       //ub4         rcskip
		if r == C.OCI_ERROR {
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}*/
	return iterations, nil
}

func (bnd *bndTimeSlice) setPtr() error {
	n := int(bnd.curlen)
	bnd.times = bnd.times[:n]
	var err error
	for i, dt := range bnd.ociDateTimes[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			if bnd.times[i], err = getTime(bnd.stmt.ses.srv.env, dt); err != nil {
				return err
			}
			if bnd.values != nil {
				bnd.values[i].IsNull = false
				bnd.values[i].Value = bnd.times[i]
			}
		} else if bnd.values != nil {
			bnd.values[i].IsNull = true
		}
	}
	return nil
}

func (bnd *bndTimeSlice) free(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(bnd.ociDateTimes[n]), //void     *descp,
		C.OCI_DTYPE_DATE)                    //ub4      type );
}

func (bnd *bndTimeSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	for n := range bnd.ociDateTimes {
		bnd.free(n)
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.zoneBuf.Reset()
	bnd.values = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxTimeSlice, bnd)
	return nil
}
