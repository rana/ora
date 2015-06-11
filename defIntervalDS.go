// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type defIntervalDS struct {
	rset        *Rset
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	null        C.sb2
}

func (def *defIntervalDS) define(position int, rset *Rset) error {
	def.rset = rset
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                              //OCIStmt     *stmtp,
		&def.ocidef,                                   //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,              //OCIError    *errhp,
		C.ub4(position),                               //ub4         position,
		unsafe.Pointer(&def.ociInterval),              //void        *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(def.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_DS,                            //ub2         dty,
		unsafe.Pointer(&def.null),                     //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defIntervalDS) value() (value interface{}, err error) {
	intervalDS := IntervalDS{IsNull: def.null < C.sb2(0)}
	if !intervalDS.IsNull {
		var day C.sb4
		var hour C.sb4
		var minute C.sb4
		var second C.sb4
		var nanosecond C.sb4
		r := C.OCIIntervalGetDaySecond(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //void               *hndl,
			def.rset.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			&day,            //sb4                *dy,
			&hour,           //sb4                *hr,
			&minute,         //sb4                *mm,
			&second,         //sb4                *ss,
			&nanosecond,     //sb4                *fsec,
			def.ociInterval) //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = def.rset.stmt.ses.srv.env.ociError()
		}
		intervalDS.Day = int32(day)
		intervalDS.Hour = int32(hour)
		intervalDS.Minute = int32(minute)
		intervalDS.Second = int32(second)
		intervalDS.Nanosecond = int32(nanosecond)
	}
	return intervalDS, err
}

func (def *defIntervalDS) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),    //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&def.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_DS,                             //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (def *defIntervalDS) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(def.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_DS)         //timeDefine.descTypeCode)                //ub4      type );
}

func (def *defIntervalDS) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociInterval = nil
	rset.putDef(defIdxIntervalDS, def)
	return nil
}
