// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type bndIntervalDSSlice struct {
	stmt         *Stmt
	ocibnd       *C.OCIBind
	ociIntervals []*C.OCIInterval
	arrHlp
}

func (bnd *bndIntervalDSSlice) bind(values []IntervalDS, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	// ensure we have at least 1 slot in the slice
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, IntervalDS{})
	}
	if cap(bnd.ociIntervals) < C {
		bnd.ociIntervals = make([]*C.OCIInterval, L, C)
	} else {
		bnd.ociIntervals = bnd.ociIntervals[:L]
	}
	alen := C.ACTUAL_LENGTH_TYPE(unsafe.Sizeof(bnd.ociIntervals[0]))

	if r := C.decriptorAllocSlice(
		bnd.stmt.ses.srv.env.ocienv,          //CONST dvoid   *parenth,
		unsafe.Pointer(&bnd.ociIntervals[0]), //dvoid         **descpp,
		C.ub4(alen),
		C.OCI_DTYPE_INTERVAL_DS, //ub4           type,
		C.size_t(len(values)),   //size_t        xtramem_sz,
	); r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return iterations, errNew("unable to allocate oci interval handle during bind")
	}

	for n, value := range values {
		bnd.alen[n] = alen
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = C.sb2(0)
		}
		r := C.OCIIntervalSetDaySecond(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv), //void               *hndl,
			bnd.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			C.sb4(value.Day),                            //sb4                dy,
			C.sb4(value.Hour),                           //sb4                hr,
			C.sb4(value.Minute),                         //sb4                mm,
			C.sb4(value.Second),                         //sb4                ss,
			C.sb4(value.Nanosecond),                     //sb4                fsec,
			bnd.ociIntervals[n])                         //OCIInterval        *result );
		if r == C.OCI_ERROR {
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ociIntervals[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociIntervals[0])), //sb8          value_sz,
		C.SQLT_INTERVAL_DS,                                //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),                  //void         *indp,
		&bnd.alen[0],                                      //ub2          *alenp,
		&bnd.rcode[0],                                     //ub2          *rcodep,
		getMaxarrLen(C, isAssocArray),
		curlenp,       //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(unsafe.Sizeof(bnd.ociIntervals[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                       //ub4         indskip,
		C.ub4(C.sizeof_ub4),                       //ub4         alskip,
		C.ub4(C.sizeof_ub2))                       //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	return iterations, nil
}

func (bnd *bndIntervalDSSlice) setPtr() error {
	return nil
}

func (bnd *bndIntervalDSSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	for n := 0; n < len(bnd.ociIntervals); n++ {
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.ociIntervals[n]), //void     *descp,
			C.OCI_DTYPE_INTERVAL_DS)             //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociIntervals = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxIntervalDSSlice, bnd)
	return nil
}
