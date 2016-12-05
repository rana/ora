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

type bndIntervalYMSlice struct {
	stmt         *Stmt
	ocibnd       *C.OCIBind
	ociIntervals []*C.OCIInterval
	arrHlp
}

func (bnd *bndIntervalYMSlice) bind(values []IntervalYM, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	// ensure we have at least 1 slot in the slice
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, IntervalYM{})
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
		C.OCI_DTYPE_INTERVAL_YM, //ub4           type,
		C.size_t(len(values)),   //size_t        xtramem_sz,
	); r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return iterations, errNew("unable to allocate oci interval handle during bind")
	}

	for n, value := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = C.sb2(0)
		}
		bnd.alen[n] = alen
		r := C.OCIIntervalSetYearMonth(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv), //void               *hndl,
			bnd.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			C.sb4(value.Year),                           //sb4                yr,
			C.sb4(value.Month),                          //sb4                mnth,
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
		C.SQLT_INTERVAL_YM,                                //ub2          dty,
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

func (bnd *bndIntervalYMSlice) setPtr() error {
	return nil
}

func (bnd *bndIntervalYMSlice) free(n int) {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(bnd.ociIntervals[n]), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)             //ub4      type );
}

func (bnd *bndIntervalYMSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	for n := range bnd.ociIntervals {
		bnd.free(n)
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociIntervals = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxIntervalYMSlice, bnd)
	return nil
}
