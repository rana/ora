// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include <string.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type defIntervalDS struct {
	ociDef
	intervals []*C.OCIInterval
}

func (def *defIntervalDS) define(position int, rset *Rset) error {
	def.rset = rset
	if def.intervals != nil {
		C.free(unsafe.Pointer(&def.intervals[0]))
	}
	def.intervals = (*((*[MaxFetchLen]*C.OCIInterval)(C.malloc(C.size_t(rset.fetchLen) * C.sof_Intervalp))))[:rset.fetchLen]
	def.ensureAllocatedLength(len(def.intervals))
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.intervals[0]), int(C.sof_Intervalp), C.SQLT_INTERVAL_DS)
}

func (def *defIntervalDS) value(offset int) (value interface{}, err error) {
	intervalDS := IntervalDS{IsNull: def.nullInds[offset] < 0}
	if !intervalDS.IsNull {
		var day C.sb4
		var hour C.sb4
		var minute C.sb4
		var second C.sb4
		var nanosecond C.sb4
		r := C.OCIIntervalGetDaySecond(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //void               *hndl,
			def.rset.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			&day,                  //sb4                *dy,
			&hour,                 //sb4                *hr,
			&minute,               //sb4                *mm,
			&second,               //sb4                *ss,
			&nanosecond,           //sb4                *fsec,
			def.intervals[offset]) //const OCIInterval  *interval );
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
	for i, p := range def.intervals {
		if p != nil {
			def.intervals[i] = nil
			//C.OCIDescriptorFree(unsafe.Pointer(p), C.OCI_DTYPE_INTERVAL_DS)
		}
		def.allocated[i] = false
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),     //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&def.intervals[i])), //dvoid         **descpp,
			C.OCI_DTYPE_INTERVAL_DS,                              //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return def.rset.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci interval handle during define")
		}
		def.allocated[i] = true
	}
	return nil
}

func (def *defIntervalDS) free() {
	for i, p := range def.intervals {
		if p == nil {
			continue
		}
		def.intervals[i] = nil
		if !def.allocated[i] {
			continue
		}
		C.OCIDescriptorFree(
			unsafe.Pointer(p),       //void     *descp,
			C.OCI_DTYPE_INTERVAL_DS) //timeDefine.descTypeCode)                //ub4      type );
	}
	def.arrHlp.close()
}

func (def *defIntervalDS) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	rset := def.rset
	def.rset = nil
	if def.intervals != nil {
		C.free(unsafe.Pointer(&def.intervals[0]))
		def.intervals = nil
	}
	def.ocidef = nil
	rset.putDef(defIdxIntervalDS, def)
	return nil
}
