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

type int32PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *int32
}

func (int32PtrBind *int32PtrBind) bind(value *int32, position int, ocistmt *C.OCIStmt) error {
	int32PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int32PtrBind.ocibnd),     //OCIBind      **bindpp,
		int32PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(&int32PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8          value_sz,
		C.SQLT_VNU,                              //ub2          dty,
		unsafe.Pointer(&int32PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return int32PtrBind.environment.ociError()
	}
	return nil
}

func (int32PtrBind *int32PtrBind) setPtr() error {
	if int32PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			int32PtrBind.environment.ocierr,    //OCIError              *err,
			&int32PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(4),                         //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,                //uword                 rsl_flag,
			unsafe.Pointer(int32PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return int32PtrBind.environment.ociError()
		}
	}
	return nil
}

func (int32PtrBind *int32PtrBind) close() {
	defer func() {
		recover()
	}()
	int32PtrBind.ocibnd = nil
	int32PtrBind.value = nil
	int32PtrBind.environment.int32PtrBindPool.Put(int32PtrBind)
}
