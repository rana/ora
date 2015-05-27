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
}

func (bnd *bndIntervalYMSlice) bind(values []IntervalYM, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.ociIntervals = make([]*C.OCIInterval, len(values))
	nullInds := make([]C.sb2, len(values))
	alenp := make([]C.ACTUAL_LENGTH_TYPE, len(values))
	rcodep := make([]C.ub2, len(values))
	for n, value := range values {
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),             //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&bnd.ociIntervals[n])), //dvoid         **descpp,
			C.OCI_DTYPE_INTERVAL_YM,                                 //ub4           type,
			0,   //size_t        xtramem_sz,
			nil) //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci interval handle during bind")
		}
		r = C.OCIIntervalSetYearMonth(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv), //void               *hndl,
			bnd.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			C.sb4(value.Year),                           //sb4                yr,
			C.sb4(value.Month),                          //sb4                mnth,
			bnd.ociIntervals[n])                         //OCIInterval        *result );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			nullInds[n] = C.sb2(0)
		}
		alenp[n] = C.ACTUAL_LENGTH_TYPE(unsafe.Sizeof(bnd.ociIntervals[n]))
	}
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                                  //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                        //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                       //OCIError     *errhp,
		C.ub4(position),                                   //ub4          position,
		unsafe.Pointer(&bnd.ociIntervals[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociIntervals[0])), //sb8          value_sz,
		C.SQLT_INTERVAL_YM,                                //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                      //void         *indp,
		&alenp[0],                                         //ub2          *alenp,
		&rcodep[0],                                        //ub2          *rcodep,
		0,                                                 //ub4          maxarr_len,
		nil,                                               //ub4          *curelep,
		C.OCI_DEFAULT)                                     //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(unsafe.Sizeof(bnd.ociIntervals[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                       //ub4         indskip,
		C.ub4(C.sizeof_ub4),                       //ub4         alskip,
		C.ub4(C.sizeof_ub2))                       //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
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
			err = errRecover(value)
		}
	}()

	for n := range bnd.ociIntervals {
		bnd.free(n)
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociIntervals = nil
	stmt.putBnd(bndIdxIntervalYMSlice, bnd)
	return nil
}
