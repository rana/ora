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

type int64PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *int64
}

func (int64PtrBind *int64PtrBind) bind(value *int64, position int, ocistmt *C.OCIStmt) error {
	int64PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int64PtrBind.ocibnd),     //OCIBind      **bindpp,
		int64PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(&int64PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8          value_sz,
		C.SQLT_VNU,                              //ub2          dty,
		unsafe.Pointer(&int64PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return int64PtrBind.environment.ociError()
	}
	return nil
}

func (int64PtrBind *int64PtrBind) setPtr() error {
	if int64PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			int64PtrBind.environment.ocierr,    //OCIError              *err,
			&int64PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(8),                         //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,                //uword                 rsl_flag,
			unsafe.Pointer(int64PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return int64PtrBind.environment.ociError()
		}
	}
	return nil
}

func (int64PtrBind *int64PtrBind) close() {
	defer func() {
		recover()
	}()
	int64PtrBind.ocibnd = nil
	int64PtrBind.value = nil
	int64PtrBind.environment.int64PtrBindPool.Put(int64PtrBind)
}
