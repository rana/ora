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

type float32PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *float32
}

func (float32PtrBind *float32PtrBind) bind(value *float32, position int, ocistmt *C.OCIStmt) error {
	float32PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&float32PtrBind.ocibnd),     //OCIBind      **bindpp,
		float32PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                           //ub4          position,
		unsafe.Pointer(&float32PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                 //sb8          value_sz,
		C.SQLT_VNU,                                //ub2          dty,
		unsafe.Pointer(&float32PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return float32PtrBind.environment.ociError()
	}
	return nil
}

func (float32PtrBind *float32PtrBind) setPtr() error {
	if float32PtrBind.isNull > -1 {
		r := C.OCINumberToReal(
			float32PtrBind.environment.ocierr,    //OCIError              *err,
			&float32PtrBind.ociNumber,            //const OCINumber     *number,
			C.uword(4),                           //uword               rsl_length,
			unsafe.Pointer(float32PtrBind.value)) //void                *rsl );
		if r == C.OCI_ERROR {
			return float32PtrBind.environment.ociError()
		}
	}
	return nil
}

func (float32PtrBind *float32PtrBind) close() {
	defer func() {
		recover()
	}()
	float32PtrBind.ocibnd = nil
	float32PtrBind.value = nil
	float32PtrBind.environment.float32PtrBindPool.Put(float32PtrBind)
}
