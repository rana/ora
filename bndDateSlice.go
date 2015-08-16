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
)

type bndDateSlice struct {
	stmt     *Stmt
	ocibnd   *C.OCIBind
	ociDates []C.OCIDate
	zoneBuf  bytes.Buffer
	values   []Date
	times    []time.Time
	dtype    C.ub4
	arrHlp
}

func (bnd *bndDateSlice) bindOra(values []Date, position int, stmt *Stmt) (uint32, error) {
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

func (bnd *bndDateSlice) bind(values []time.Time, position int, stmt *Stmt) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, stmt.stmtType)
	if needAppend {
		values = append(values, time.Time{})
	}
	bnd.times = values
	if cap(bnd.ociDates) < C {
		bnd.ociDates = make([]C.OCIDate, L, C)
	} else {
		bnd.ociDates = bnd.ociDates[:L]
	}
	valueSz := C.ACTUAL_LENGTH_TYPE(C.sizeof_OCIDate)
	for n, timeValue := range values {
		ociSetDateTime(&bnd.ociDates[n], timeValue)
		bnd.alen[n] = valueSz
	}

	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p pos=%d cap=%d len=%d curlen=%d curlenp=%p value_sz=%d alen=%v",
		bnd, position, cap(bnd.ociDates), len(bnd.ociDates), bnd.curlen, curlenp,
		valueSz, bnd.alen)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                 //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),       //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,      //OCIError     *errhp,
		C.ub4(position),                  //ub4          position,
		unsafe.Pointer(&bnd.ociDates[0]), //void         *valuep,
		C.LENGTH_TYPE(valueSz),           //sb8          value_sz,
		C.SQLT_ODT,                       //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]), //void         *indp,
		&bnd.alen[0],                     //ub2          *alenp,
		&bnd.rcode[0],                    //ub2          *rcodep,
		C.ub4(C),                         //ub4          maxarr_len,
		curlenp,                          //ub4          *curelep,
		C.OCI_DEFAULT)                    //ub4          mode );
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
	bnd.times = bnd.times[:n]
	bnd.nullInds = bnd.nullInds[:n]
	if bnd.values != nil {
		bnd.values = bnd.values[:n]
	}
	for i, dt := range bnd.ociDates[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			bnd.times[i] = ociGetDateTime(dt)
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

func (bnd *bndDateSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.values = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxDateSlice, bnd)
	return nil
}
