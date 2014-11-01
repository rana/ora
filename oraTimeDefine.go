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
	env         *Environment
	ocidef      *C.OCIDefine
	ociDateTime *C.OCIDateTime
	isNull      C.sb2
}

func (d *oraTimeDefine) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                             //OCIStmt     *stmtp,
		&d.ocidef,                           //OCIDefine   **defnpp,
		d.env.ocierr,                        //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&d.ociDateTime),      //void        *valuep,
		C.sb8(unsafe.Sizeof(d.ociDateTime)), //sb8         value_sz,
		C.SQLT_TIMESTAMP_TZ,                 //ub2         dty,
		unsafe.Pointer(&d.isNull),           //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}

func (d *oraTimeDefine) value() (value interface{}, err error) {
	timeValue := Time{IsNull: d.isNull < 0}
	if !timeValue.IsNull {
		timeValue.Value, err = getTime(d.env, d.ociDateTime)
	}
	value = timeValue
	return value, err
}

func (d *oraTimeDefine) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociDateTime)), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                          //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during define")
	}
	return nil

}
func (d *oraTimeDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociDateTime), //void     *descp,
		C.OCI_DTYPE_TIMESTAMP_TZ)      //ub4      type );
}
func (d *oraTimeDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociDateTime = nil
	d.isNull = C.sb2(0)
	d.env.oraTimeDefinePool.Put(d)
}
