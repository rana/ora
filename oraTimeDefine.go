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

type oraTimeDefine struct {
	environment *Environment
	ocidef      *C.OCIDefine
	ociDateTime *C.OCIDateTime
	isNull      C.sb2
}

func (oraTimeDefine *oraTimeDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                         //OCIStmt     *stmtp,
		&oraTimeDefine.ocidef,                           //OCIDefine   **defnpp,
		oraTimeDefine.environment.ocierr,                //OCIError    *errhp,
		C.ub4(position),                                 //ub4         position,
		unsafe.Pointer(&oraTimeDefine.ociDateTime),      //void        *valuep,
		C.sb8(unsafe.Sizeof(oraTimeDefine.ociDateTime)), //sb8         value_sz,
		C.SQLT_TIMESTAMP_TZ,                             //ub2         dty,
		unsafe.Pointer(&oraTimeDefine.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return oraTimeDefine.environment.ociError()
	}
	return nil
}

func (oraTimeDefine *oraTimeDefine) value() (value interface{}, err error) {
	timeValue := Time{IsNull: oraTimeDefine.isNull < 0}
	if !timeValue.IsNull {
		timeValue.Value, err = getTime(oraTimeDefine.environment, oraTimeDefine.ociDateTime)
	}
	value = timeValue
	return value, err
}

func (oraTimeDefine *oraTimeDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(oraTimeDefine.environment.ocienv),              //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&oraTimeDefine.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                                      //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return oraTimeDefine.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during define")
	}
	return nil

}
func (oraTimeDefine *oraTimeDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(oraTimeDefine.ociDateTime), //void     *descp,
		C.OCI_DTYPE_TIMESTAMP_TZ)                  //ub4      type );
}
func (oraTimeDefine *oraTimeDefine) close() {
	defer func() {
		recover()
	}()
	oraTimeDefine.ocidef = nil
	oraTimeDefine.ociDateTime = nil
	oraTimeDefine.isNull = C.sb2(0)
	oraTimeDefine.environment.oraTimeDefinePool.Put(oraTimeDefine)
}
