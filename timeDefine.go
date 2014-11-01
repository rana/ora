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
	"math"
	"strings"
	"time"
	"unsafe"
)

type timeDefine struct {
	env         *Environment
	ocidef      *C.OCIDefine
	ociDateTime *C.OCIDateTime
	isNull      C.sb2
}

func (d *timeDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                             //OCIStmt     *stmtp,
		&d.ocidef,                           //OCIDefine   **defnpp,
		d.env.ocierr,                        //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&d.ociDateTime),      //void        *valuep,
		C.sb8(unsafe.Sizeof(d.ociDateTime)), //sb8         value_sz,
		C.SQLT_TIMESTAMP_TZ,                 //defineTypeCode,                               //ub2         dty,
		unsafe.Pointer(&d.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *timeDefine) value() (value interface{}, err error) {
	if d.isNull > -1 {
		value, err = getTime(d.env, d.ociDateTime)
	}
	return value, err
}
func (d *timeDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during define")
	}
	return nil

}
func (d *timeDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociDateTime), //void     *descp,
		C.OCI_DTYPE_TIMESTAMP_TZ)      //ub4      type );
}
func (d *timeDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociDateTime = nil
	d.isNull = C.sb2(0)
	d.env.timeDefinePool.Put(d)
}

func getTime(env *Environment, ociDateTime *C.OCIDateTime) (result time.Time, err error) {
	var year C.sb2
	var month C.ub1
	var day C.ub1
	var hour C.ub1
	var minute C.ub1
	var second C.ub1
	var fsec C.ub4
	var location *time.Location
	r := C.OCIDateTimeGetDate(
		unsafe.Pointer(env.ocienv), //void               *hndl,
		env.ocierr,                 //OCIError           *err,
		ociDateTime,                //const OCIDateTime  *datetime,
		&year,                      //sb2                *year,
		&month,                     //ub1                *month,
		&day)                       //ub1                *day );
	if r == C.OCI_ERROR {
		return result, env.ociError()
	}
	r = C.OCIDateTimeGetTime(
		unsafe.Pointer(env.ocienv), //void               *hndl,
		env.ocierr,                 //OCIError           *err,
		ociDateTime,                //OCIDateTime  *datetime,
		&hour,                      //ub1           *hour,
		&minute,                    //ub1           *min,
		&second,                    //ub1           *sec,
		&fsec)                      //ub4           *fsec );
	if r == C.OCI_ERROR {
		return result, env.ociError()
	}
	var buf [32]byte
	var buflen C.ub4 = 32
	r = C.OCIDateTimeGetTimeZoneName(
		unsafe.Pointer(env.ocienv), //void               *hndl,
		env.ocierr,                 //OCIError           *err,
		ociDateTime,                //const OCIDateTime  *datetime,
		(*C.ub1)(&buf[0]),          //ub1                *buf,
		&buflen)                    //ub4                *buflen, );
	if r != C.OCI_ERROR {
		var buffer bytes.Buffer
		for n := 0; n < int(buflen); n++ {
			buffer.WriteByte(buf[n])
		}
		locName := buffer.String()
		// timestamp_ltz returns numeric offset
		// time.Time's lookup for numeric offset is unknown;
		// therefore, create a fixed location for the offset
		var offsetHour C.sb1
		var offsetMinute C.sb1
		if strings.ContainsAny(locName, "-0123456789") {
			r = C.OCIDateTimeGetTimeZoneOffset(
				unsafe.Pointer(env.ocienv), //void               *hndl,
				env.ocierr,                 //OCIError           *err,
				ociDateTime,                //const OCIDateTime  *datetime,
				&offsetHour,                //sb1                *hour,
				&offsetMinute)              //sb1                *min, );
			if r == C.OCI_ERROR {
				return result, env.ociError()
			}
			seconds := math.Abs(float64(offsetHour)) * 60 * 60
			seconds += math.Abs(float64(offsetMinute)) * 60
			if offsetHour < 0 {
				seconds *= -1
			}
			location = time.FixedZone(locName, int(seconds))
		} else {
			location, err = time.LoadLocation(locName)
			if err != nil {
				return result, err
			}
		}
	} else {
		// Date Oracle type doesn't have timezone info
		// no timezone information available from server
		location = time.Local
	}
	result = time.Date(int(year), time.Month(int(month)), int(day), int(hour), int(minute), int(second), int(fsec), location)
	return result, nil
}
