// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"

const ACTUAL_LENGTH_TYPE sof_OCIDateTime = sizeof(OCIDateTime *);
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

var checkDateTime = false

type bndTimeSlice struct {
	stmt         *Stmt
	ocibnd       *C.OCIBind
	ociDateTimes []*C.OCIDateTime
	values       []Time
	times        []time.Time
	isOra        bool
	arrHlp
}

// FIXME(tgulacsi): somewhere here we leak a lot of memory!!!

func (bnd *bndTimeSlice) bindOra(values []Time, position namedPos, stmt *Stmt, isAssocArray bool) (uint32, error) {
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
	for n := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[0] = 0
			bnd.times[n] = values[n].Value
		}
	}
	bnd.isOra = true
	return bnd.bind(bnd.times, position, stmt, isAssocArray)
}

func (bnd *bndTimeSlice) bind(values []time.Time, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, time.Time{})
	}
	bnd.times = values
	if cap(bnd.ociDateTimes) < C {
		bnd.ociDateTimes = make([]*C.OCIDateTime, L, C)
	} else {
		bnd.ociDateTimes = bnd.ociDateTimes[:L]
	}
	valueSz := C.ACTUAL_LENGTH_TYPE(C.sof_OCIDateTime)
	for n, timeValue := range values {
		arr := bnd.ociDateTimes[n : n+1 : n+1]
		if err := (&dateTimep{p: arr}).Set(bnd.stmt.ses.srv.env, timeValue); err != nil {
			return iterations, err
		}
		bnd.alen[n] = valueSz

		if checkDateTime {
			var valid C.ub4
			r := C.OCIDateTimeCheck(unsafe.Pointer(bnd.stmt.ses.ocises),
				bnd.stmt.ses.srv.env.ocierr,
				bnd.ociDateTimes[n],
				&valid)
			if r == C.OCI_ERROR {
				return iterations, bnd.stmt.ses.srv.env.ociError()
			}
			if valid != 0 {
				return iterations, fmt.Errorf("%s given bad date: %d", timeValue, valid)
			}
		}
	}
	if !bnd.isOra {
		for i := range bnd.nullInds {
			bnd.nullInds[i] = 0
		}
	}

	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v cap=%d len=%d curlen=%d curlenp=%p value_sz=%d alen=%v",
		bnd, position, cap(bnd.ociDateTimes), len(bnd.ociDateTimes), bnd.curlen, curlenp,
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
		unsafe.Pointer(&bnd.ociDateTimes[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociDateTimes[0])), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                               //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),                  //void         *indp,
		&bnd.alen[0],                                      //ub2          *alenp,
		&bnd.rcode[0],                                     //ub2          *rcodep,
		getMaxarrLen(C, isAssocArray),                     //ub4          maxarr_len,
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

func (bnd *bndTimeSlice) setPtr() error {
	if bnd.isAssocArr {
		return nil
	}
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
	free := func(p *C.OCIDateTime) {
		defer func() {
			recover()
		}()
		C.OCIDescriptorFree(
			unsafe.Pointer(p),        //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ) //ub4      type );
	}
	for i := 0; i < n && i < len(bnd.ociDateTimes); i++ {
		arr := bnd.ociDateTimes[i : i+1 : i+1]
		if arr[0] == nil {
			continue
		}
		free(arr[0])
	}
}

func (bnd *bndTimeSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.free(len(bnd.ociDateTimes))
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.values = nil
	bnd.ociDateTimes = nil
	bnd.isOra = false
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxTimeSlice, bnd)
	return nil
}
