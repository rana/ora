// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

type oraIntervalYMBind struct {
	env         *Environment
	ocibnd      *C.OCIBind
	ociInterval *C.OCIInterval
}

func (b *oraIntervalYMBind) bind(value IntervalYM, position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(b.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&b.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_YM,                           //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during bind")
	}
	r = C.OCIIntervalSetYearMonth(
		unsafe.Pointer(b.env.ocienv), //void               *hndl,
		b.env.ocierr,                 //OCIError           *err,
		C.sb4(value.Year),            //sb4                yr,
		C.sb4(value.Month),           //sb4                mnth,
		b.ociInterval)                //OCIInterval        *result );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt,                             //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),            //OCIBind      **bindpp,
		b.env.ocierr,                        //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(&b.ociInterval),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociInterval)), //sb8          value_sz,
		C.SQLT_INTERVAL_YM,                  //ub2          dty,
		nil,                                 //void         *indp,
		nil,                                 //ub2          *alenp,
		nil,                                 //ub2          *rcodep,
		0,                                   //ub4          maxarr_len,
		nil,                                 //ub4          *curelep,
		C.OCI_DEFAULT)                       //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *oraIntervalYMBind) setPtr() error {
	return nil
}

func (b *oraIntervalYMBind) close() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(b.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)       //timeDefine.descTypeCode)                //ub4      type );
	b.ocibnd = nil
	b.ociInterval = nil
	b.env.oraIntervalYMBindPool.Put(b)
}
