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
	environment *Environment
	ocibnd      *C.OCIBind
	ociDateTime *C.OCIDateTime
	valuePtr    *time.Time
	cZone       *C.char
	isNull      C.sb2
}

func (timePtrBind *timePtrBind) bind(value *time.Time, position int, ocistmt *C.OCIStmt) error {
	timePtrBind.valuePtr = value
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(timePtrBind.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&timePtrBind.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                                    //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return timePtrBind.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&timePtrBind.ocibnd),            //OCIBind      **bindpp,
		timePtrBind.environment.ocierr,                //OCIError     *errhp,
		C.ub4(position),                               //ub4          position,
		unsafe.Pointer(&timePtrBind.ociDateTime),      //void         *valuep,
		C.sb8(unsafe.Sizeof(timePtrBind.ociDateTime)), //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                           //ub2          dty,
		unsafe.Pointer(&timePtrBind.isNull),           //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return timePtrBind.environment.ociError()
	}
	return nil
}

func (timePtrBind *timePtrBind) setPtr() (err error) {
	if timePtrBind.valuePtr != nil && timePtrBind.isNull > -1 {
		*timePtrBind.valuePtr, err = getTime(timePtrBind.environment, timePtrBind.ociDateTime)
	}
	return err
}

func (timePtrBind *timePtrBind) close() {
	defer func() {
		recover()
	}()
	// cleanup bindTime
	if timePtrBind.cZone != nil {
		C.free(unsafe.Pointer(timePtrBind.cZone))
		timePtrBind.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(timePtrBind.ociDateTime), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)                //ub4      type );

	}
	timePtrBind.ocibnd = nil
	timePtrBind.ociDateTime = nil
	timePtrBind.valuePtr = nil
	timePtrBind.isNull = C.sb2(0)
	timePtrBind.environment.timePtrBindPool.Put(timePtrBind)
}
