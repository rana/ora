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

type int8PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *int8
}

func (int8PtrBind *int8PtrBind) bind(value *int8, position int, ocistmt *C.OCIStmt) error {
	int8PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int8PtrBind.ocibnd),     //OCIBind      **bindpp,
		int8PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                        //ub4          position,
		unsafe.Pointer(&int8PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8          value_sz,
		C.SQLT_VNU,                             //ub2          dty,
		unsafe.Pointer(&int8PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return int8PtrBind.environment.ociError()
	}
	return nil
}

func (int8PtrBind *int8PtrBind) setPtr() error {
	if int8PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			int8PtrBind.environment.ocierr,    //OCIError              *err,
			&int8PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(1),                        //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(int8PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return int8PtrBind.environment.ociError()
		}
	}
	return nil
}

func (int8PtrBind *int8PtrBind) close() {
	defer func() {
		recover()
	}()
	int8PtrBind.ocibnd = nil
	int8PtrBind.value = nil
	int8PtrBind.environment.int8PtrBindPool.Put(int8PtrBind)
}
