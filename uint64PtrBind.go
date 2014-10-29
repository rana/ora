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

type uint64PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *uint64
}

func (uint64PtrBind *uint64PtrBind) bind(value *uint64, position int, ocistmt *C.OCIStmt) error {
	uint64PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint64PtrBind.ocibnd),     //OCIBind      **bindpp,
		uint64PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                          //ub4          position,
		unsafe.Pointer(&uint64PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8          value_sz,
		C.SQLT_VNU,                               //ub2          dty,
		unsafe.Pointer(&uint64PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return uint64PtrBind.environment.ociError()
	}
	return nil
}

func (uint64PtrBind *uint64PtrBind) setPtr() error {
	if uint64PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			uint64PtrBind.environment.ocierr,    //OCIError              *err,
			&uint64PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(8),                          //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(uint64PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return uint64PtrBind.environment.ociError()
		}
	}
	return nil
}

func (uint64PtrBind *uint64PtrBind) close() {
	defer func() {
		recover()
	}()
	uint64PtrBind.ocibnd = nil
	uint64PtrBind.value = nil
	uint64PtrBind.environment.uint64PtrBindPool.Put(uint64PtrBind)
}
