// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"bytes"
	"math"
	"strings"
	"time"
	"unsafe"
)

type defTime struct {
	ociDef
	isNullable bool
	dates      []*C.OCIDateTime
}

func (def *defTime) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.dates != nil {
		C.free(unsafe.Pointer(&def.dates[0]))
	}
	def.dates = (*((*[MaxFetchLen]*C.OCIDateTime)(C.malloc(C.size_t(rset.fetchLen) * C.sof_DateTimep))))[:rset.fetchLen]
	def.ensureAllocatedLength(len(def.dates))
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.dates[0]), int(C.sof_DateTimep), C.SQLT_TIMESTAMP_TZ)
}

func (def *defTime) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Time{IsNull: true}, nil
		}
		return nil, nil
	}
	t, err := getTime(def.rset.stmt.ses.srv.env, def.dates[offset])
	if def.isNullable {
		return Time{Value: t}, err
	}
	return t, err
}

func (def *defTime) alloc() error {
	for i := range def.dates {
		def.allocated[i] = false
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&def.dates[i])), //dvoid         **descpp,
			C.OCI_DTYPE_TIMESTAMP_TZ,                         //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return def.rset.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci timestamp handle during define")
		}
		def.allocated[i] = true
	}
	return nil

}

func (def *defTime) free() {
	for i, d := range def.dates {
		if d == nil {
			continue
		}
		def.dates[i] = nil
		if !def.allocated[i] {
			continue
		}
		C.OCIDescriptorFree(
			unsafe.Pointer(d),        //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ) //timeDefine.descTypeCode)                //ub4      type );
	}
	def.arrHlp.close()
}

func (def *defTime) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	rset := def.rset
	def.rset = nil
	if def.dates != nil {
		C.free(unsafe.Pointer(&def.dates[0]))
		def.dates = nil
	}
	def.ocidef = nil
	rset.putDef(defIdxTime, def)
	return nil
}

func getTime(env *Env, ociDateTime *C.OCIDateTime) (result time.Time, err error) {
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
		_drv.locationsMu.RLock()
		location = _drv.locations[locName]
		_drv.locationsMu.RUnlock()
		if location == nil {
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
			// stored location for future reference
			// important that FixedZone is called as few times as possible
			// to reduce significant memory allocation
			_drv.locationsMu.Lock()
			_drv.locations[locName] = location
			_drv.locationsMu.Unlock()
		}
	} else {
		// Date Oracle type doesn't have timezone info
		// no timezone information available from server
		location = time.Local
	}
	result = time.Date(int(year), time.Month(int(month)), int(day), int(hour), int(minute), int(second), int(fsec), location)
	return result, nil
}
