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
	"fmt"
	"time"
	"unsafe"
)

type timeBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociDateTime *C.OCIDateTime
	cZone       *C.char
	zoneBuffer  bytes.Buffer
}

func (timeBind *timeBind) bind(value time.Time, position int, ocistmt *C.OCIStmt) error {
	timeBind.cZone = C.CString(zoneOffset(value, &timeBind.zoneBuffer))
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(timeBind.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&timeBind.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                                 //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return timeBind.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	r = C.OCIDateTimeConstruct(
		unsafe.Pointer(timeBind.environment.ocienv),  //dvoid         *hndl,
		timeBind.environment.ocierr,                  //OCIError      *err,
		timeBind.ociDateTime,                         //OCIDateTime   *datetime,
		C.sb2(value.Year()),                          //sb2           year,
		C.ub1(int32(value.Month())),                  //ub1           month,
		C.ub1(value.Day()),                           //ub1           day,
		C.ub1(value.Hour()),                          //ub1           hour,
		C.ub1(value.Minute()),                        //ub1           min,
		C.ub1(value.Second()),                        //ub1           sec,
		C.ub4(value.Nanosecond()),                    //ub4           fsec,
		(*C.OraText)(unsafe.Pointer(timeBind.cZone)), //OraText       *timezone,
		C.size_t(C.strlen(timeBind.cZone)))           //size_t        timezone_length );
	if r == C.OCI_ERROR {
		return timeBind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&timeBind.ocibnd),            //OCIBind      **bindpp,
		timeBind.environment.ocierr,                //OCIError     *errhp,
		C.ub4(position),                            //ub4          position,
		unsafe.Pointer(&timeBind.ociDateTime),      //void         *valuep,
		C.sb8(unsafe.Sizeof(timeBind.ociDateTime)), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                        //ub2          dty,
		nil,                                        //void         *indp,
		nil,                                        //ub2          *alenp,
		nil,                                        //ub2          *rcodep,
		0,                                          //ub4          maxarr_len,
		nil,                                        //ub4          *curelep,
		C.OCI_DEFAULT)                              //ub4          mode );
	if r == C.OCI_ERROR {
		return timeBind.environment.ociError()
	}
	return nil
}

func (timeBind *timeBind) setPtr() (err error) {
	return nil
}

func (timeBind *timeBind) close() {
	defer func() {
		recover()
	}()
	// cleanup bindTime
	if timeBind.cZone != nil {
		C.free(unsafe.Pointer(timeBind.cZone))
		timeBind.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(timeBind.ociDateTime), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)             //ub4      type );
	}
	timeBind.ocibnd = nil
	timeBind.ociDateTime = nil
	timeBind.environment.timeBindPool.Put(timeBind)
}

func zoneOffset(value time.Time, buffer *bytes.Buffer) string {
	buffer.Reset()
	_, zoneOffsetInSeconds := value.Zone()
	if zoneOffsetInSeconds < 0 {
		buffer.WriteRune('-')
		zoneOffsetInSeconds *= -1
	} else {
		buffer.WriteRune('+')
	}
	hourOffset := zoneOffsetInSeconds / 3600
	buffer.WriteString(fmt.Sprintf("%02d", hourOffset))
	buffer.WriteRune(':')
	zoneOffsetInSeconds -= hourOffset * 3600
	minuteOffset := zoneOffsetInSeconds / 60
	buffer.WriteString(fmt.Sprintf("%02d", minuteOffset))
	return buffer.String()
}
