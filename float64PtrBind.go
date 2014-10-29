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

type float64PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *float64
}

func (float64PtrBind *float64PtrBind) bind(value *float64, position int, ocistmt *C.OCIStmt) error {
	float64PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&float64PtrBind.ocibnd),     //OCIBind      **bindpp,
		float64PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                           //ub4          position,
		unsafe.Pointer(&float64PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                 //sb8          value_sz,
		C.SQLT_VNU,                                //ub2          dty,
		unsafe.Pointer(&float64PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return float64PtrBind.environment.ociError()
	}
	return nil
}

func (float64PtrBind *float64PtrBind) setPtr() error {
	if float64PtrBind.isNull > -1 {
		r := C.OCINumberToReal(
			float64PtrBind.environment.ocierr,    //OCIError              *err,
			&float64PtrBind.ociNumber,            //const OCINumber     *number,
			C.uword(8),                           //uword               rsl_length,
			unsafe.Pointer(float64PtrBind.value)) //void                *rsl );
		if r == C.OCI_ERROR {
			return float64PtrBind.environment.ociError()
		}
	}
	return nil
}

func (float64PtrBind *float64PtrBind) close() {
	defer func() {
		recover()
	}()
	float64PtrBind.ocibnd = nil
	float64PtrBind.value = nil
	float64PtrBind.environment.float64PtrBindPool.Put(float64PtrBind)
}
