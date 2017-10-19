// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"

const ACTUAL_LENGTH_TYPE sof_OCIDate = sizeof(OCIDate *);
*/
import "C"
import (
	"bytes"
	"time"
	"unsafe"

	"gopkg.in/rana/ora.v4/date"
)

type bndDateSlice struct {
	stmt     *Stmt
	ocibnd   *C.OCIBind
	ociDates []date.Date
	zoneBuf  bytes.Buffer
	values   *[]Date
	times    *[]time.Time
	dtype    C.ub4
	timezone *time.Location
	isOra    bool
	arrHlp
}

func (bnd *bndDateSlice) bindOra(values *[]Date, position namedPos, stmt *Stmt, isAssocArray bool) (uint32, error) {
	if values == nil {
		values = &[]Date{}
	}
	bnd.values = values
	V := *values
	var T []time.Time
	if bnd.times == nil {
		bnd.times = &T
	} else {
		T = *bnd.times
	}
	if cap(T) < cap(V) {
		T = make([]time.Time, len(V), cap(V))
	} else {
		T = T[:len(V)]
	}
	if cap(bnd.nullInds) < cap(V) {
		bnd.nullInds = make([]C.sb2, len(V), cap(V))
	} else {
		bnd.nullInds = bnd.nullInds[:len(V)]
	}
	for n := range V {
		if V[n].IsNull() {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[0] = 0
			T[n] = V[n].Date.Get()
		}
	}
	*bnd.values = V
	*bnd.times = T
	bnd.isOra = true
	return bnd.bind(bnd.times, position, stmt, isAssocArray)
}

func (bnd *bndDateSlice) bind(values *[]time.Time, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	if bnd.timezone, err = bnd.stmt.ses.Timezone(); err != nil {
		return iterations, err
	}
	var V []time.Time
	if values == nil {
		values = &V
	} else {
		V = *values
	}
	L, C := len(V), cap(V)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		V = append(V, time.Time{})
	}
	bnd.times = values
	if cap(bnd.ociDates) < C {
		bnd.ociDates = make([]date.Date, L, C)
	} else {
		bnd.ociDates = bnd.ociDates[:L]
	}
	valueSz := C.ACTUAL_LENGTH_TYPE(7)
	for n, timeValue := range V {
		//arr := bnd.ociDates[n : n+1 : n+1]
		//ociSetDateTime(&arr[0], timeValue)
		bnd.ociDates[n].Set(timeValue)
		bnd.alen[n] = valueSz
	}
	if !bnd.isOra {
		for i := range bnd.nullInds {
			bnd.nullInds[i] = 0
		}
	}

	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v cap=%d len=%d curlen=%d curlenp=%p value_sz=%d alen=%v",
		bnd, position, cap(bnd.ociDates), len(bnd.ociDates), bnd.curlen, curlenp,
		valueSz, bnd.alen)
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		&bnd.ocibnd,                 //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ociDates[0]), //void         *valuep,
		C.LENGTH_TYPE(valueSz),           //sb8          value_sz,
		C.SQLT_DAT,                       //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]), //void         *indp,
		&bnd.alen[0],                     //ub2          *alenp,
		&bnd.rcode[0],                    //ub2          *rcodep,
		getMaxarrLen(C, isAssocArray),    //ub4          maxarr_len,
		curlenp,       //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(valueSz),                     //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                //ub4         indskip,
		C.ub4(C.sizeof_ACTUAL_LENGTH_TYPE), //ub4         alskip,
		C.ub4(C.sizeof_ub2))                //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	return iterations, nil
}

func (bnd *bndDateSlice) setPtr() error {
	if !bnd.isAssocArr {
		return nil
	}
	n := int(bnd.curlen)
	var T []time.Time
	if bnd.times == nil {
		bnd.times = &T
	} else {
		T = *bnd.times
	}
	T = T[:n]
	bnd.nullInds = bnd.nullInds[:n]
	var V []Date
	if bnd.values != nil {
		V = (*bnd.values)[:n]
	}
	for i, dt := range bnd.ociDates[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			//bnd.times[i] = ociGetDateTime(dt)
			T[i] = dt.GetIn(bnd.timezone)
			if V != nil {
				V[i].Date.Set(T[i])
			}
		} else if V != nil {
			V[i].Set(time.Time{})
		}
	}
	*bnd.times = T
	*bnd.values = V
	return nil
}

func (bnd *bndDateSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.times = nil
	bnd.values = nil
	bnd.isOra = false
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxDateSlice, bnd)
	return nil
}
