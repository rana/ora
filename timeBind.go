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
	env         *Environment
	ocibnd      *C.OCIBind
	ociDateTime *C.OCIDateTime
	cZone       *C.char
	zoneBuffer  bytes.Buffer
}

func (b *timeBind) bind(value time.Time, position int, ocistmt *C.OCIStmt) error {
	b.cZone = C.CString(zoneOffset(value, &b.zoneBuffer))
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(b.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&b.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	r = C.OCIDateTimeConstruct(
		unsafe.Pointer(b.env.ocienv),          //dvoid         *hndl,
		b.env.ocierr,                          //OCIError      *err,
		b.ociDateTime,                         //OCIDateTime   *datetime,
		C.sb2(value.Year()),                   //sb2           year,
		C.ub1(int32(value.Month())),           //ub1           month,
		C.ub1(value.Day()),                    //ub1           day,
		C.ub1(value.Hour()),                   //ub1           hour,
		C.ub1(value.Minute()),                 //ub1           min,
		C.ub1(value.Second()),                 //ub1           sec,
		C.ub4(value.Nanosecond()),             //ub4           fsec,
		(*C.OraText)(unsafe.Pointer(b.cZone)), //OraText       *timezone,
		C.size_t(C.strlen(b.cZone)))           //size_t        timezone_length );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt,                             //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),            //OCIBind      **bindpp,
		b.env.ocierr,                        //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(&b.ociDateTime),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociDateTime)), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                 //ub2          dty,
		nil,                                 //void         *indp,
		nil,                                 //ub2          *alenp,
		nil,                                 //ub2          *rcodep,
		0,                                   //ub4          maxarr_len,
		nil,                                 //ub4          *curelep,
		C.OCI_DEFAULT)                       //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *timeBind) setPtr() (err error) {
	return nil
}

func (b *timeBind) close() {
	defer func() {
		recover()
	}()
	// cleanup bindTime
	if b.cZone != nil {
		C.free(unsafe.Pointer(b.cZone))
		b.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(b.ociDateTime), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)      //ub4      type );
	}
	b.ocibnd = nil
	b.ociDateTime = nil
	b.env.timeBindPool.Put(b)
}

func zoneOffset(value time.Time, buf *bytes.Buffer) string {
	buf.Reset()
	_, zoneOffsetInSeconds := value.Zone()
	if zoneOffsetInSeconds < 0 {
		buf.WriteRune('-')
		zoneOffsetInSeconds *= -1
	} else {
		buf.WriteRune('+')
	}
	hourOffset := zoneOffsetInSeconds / 3600
	buf.WriteString(fmt.Sprintf("%02d", hourOffset))
	buf.WriteRune(':')
	zoneOffsetInSeconds -= hourOffset * 3600
	minuteOffset := zoneOffsetInSeconds / 60
	buf.WriteString(fmt.Sprintf("%02d", minuteOffset))
	return buf.String()
}
