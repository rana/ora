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
	"unsafe"
)

type defIntervalYM struct {
	ociDef
	intervals []*C.OCIInterval
}

func (def *defIntervalYM) define(position int, rset *Rset) error {
	def.rset = rset
	if def.intervals != nil {
		C.free(unsafe.Pointer(&def.intervals[0]))
	}
	def.intervals = (*((*[MaxFetchLen]*C.OCIInterval)(C.malloc(C.size_t(rset.fetchLen) * C.sof_Intervalp))))[:rset.fetchLen]
	def.ensureAllocatedLength(len(def.intervals))
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.intervals[0]), int(C.sof_Intervalp), C.SQLT_INTERVAL_YM)
}

func (def *defIntervalYM) value(offset int) (value interface{}, err error) {
	intervalYM := IntervalYM{IsNull: def.nullInds[offset] < 0}
	if !intervalYM.IsNull {
		var year C.sb4
		var month C.sb4
		r := C.OCIIntervalGetYearMonth(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //void               *hndl,
			def.rset.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			&year,                 //sb4                *yr,
			&month,                //sb4                *mnth,
			def.intervals[offset]) //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = def.rset.stmt.ses.srv.env.ociError()
		}
		intervalYM.Year = int32(year)
		intervalYM.Month = int32(month)
	}
	return intervalYM, err
}

func (def *defIntervalYM) alloc() error {
	for i, p := range def.intervals {
		if p != nil {
			def.intervals[i] = nil
			//C.OCIDescriptorFree(unsafe.Pointer(p), C.OCI_DTYPE_INTERVAL_YM)
		}
		def.allocated[i] = false
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),     //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&def.intervals[i])), //dvoid         **descpp,
			C.OCI_DTYPE_INTERVAL_YM,                              //ub4           type,
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

func (def *defIntervalYM) free() {
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
			C.OCI_DTYPE_INTERVAL_YM) //timeDefine.descTypeCode)                //ub4      type );
	}
	def.arrHlp.close()
}

func (def *defIntervalYM) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	if def.intervals != nil {
		C.free(unsafe.Pointer(&def.intervals[0]))
		def.intervals = nil
	}
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxIntervalYM, def)
	return nil
}
