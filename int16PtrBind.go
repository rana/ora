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

type int16PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *int16
}

func (int16PtrBind *int16PtrBind) bind(value *int16, position int, ocistmt *C.OCIStmt) error {
	int16PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int16PtrBind.ocibnd),     //OCIBind      **bindpp,
		int16PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(&int16PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8          value_sz,
		C.SQLT_VNU,                              //ub2          dty,
		unsafe.Pointer(&int16PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return int16PtrBind.environment.ociError()
	}
	return nil
}

func (int16PtrBind *int16PtrBind) setPtr() error {
	if int16PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			int16PtrBind.environment.ocierr,    //OCIError              *err,
			&int16PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(2),                         //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,                //uword                 rsl_flag,
			unsafe.Pointer(int16PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return int16PtrBind.environment.ociError()
		}
	}
	return nil
}

func (int16PtrBind *int16PtrBind) close() {
	defer func() {
		recover()
	}()
	int16PtrBind.ocibnd = nil
	int16PtrBind.value = nil
	int16PtrBind.environment.int16PtrBindPool.Put(int16PtrBind)
}
