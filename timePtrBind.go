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
	"time"
	"unsafe"
)

type timePtrBind struct {
	env         *Environment
	ocibnd      *C.OCIBind
	ociDateTime *C.OCIDateTime
	valuePtr    *time.Time
	cZone       *C.char
	isNull      C.sb2
}

func (b *timePtrBind) bind(value *time.Time, position int, ocistmt *C.OCIStmt) error {
	b.valuePtr = value
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(b.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&b.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	r = C.OCIBindByPos2(
		ocistmt,                             //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),            //OCIBind      **bindpp,
		b.env.ocierr,                        //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(&b.ociDateTime),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociDateTime)), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                 //ub2          dty,
		unsafe.Pointer(&b.isNull),           //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *timePtrBind) setPtr() (err error) {
	if b.valuePtr != nil && b.isNull > -1 {
		*b.valuePtr, err = getTime(b.env, b.ociDateTime)
	}
	return err
}

func (b *timePtrBind) close() {
	defer func() {
		recover()
	}()
	// cleanup bindTime
	if b.cZone != nil {
		C.free(unsafe.Pointer(b.cZone))
		b.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(b.ociDateTime), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)      //ub4      type );

	}
	b.ocibnd = nil
	b.ociDateTime = nil
	b.valuePtr = nil
	b.isNull = C.sb2(0)
	b.env.timePtrBindPool.Put(b)
}
