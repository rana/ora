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

type uint32PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *uint32
}

func (uint32PtrBind *uint32PtrBind) bind(value *uint32, position int, ocistmt *C.OCIStmt) error {
	uint32PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint32PtrBind.ocibnd),     //OCIBind      **bindpp,
		uint32PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                          //ub4          position,
		unsafe.Pointer(&uint32PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8          value_sz,
		C.SQLT_VNU,                               //ub2          dty,
		unsafe.Pointer(&uint32PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return uint32PtrBind.environment.ociError()
	}
	return nil
}

func (uint32PtrBind *uint32PtrBind) setPtr() error {
	if uint32PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			uint32PtrBind.environment.ocierr,    //OCIError              *err,
			&uint32PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(4),                          //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(uint32PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return uint32PtrBind.environment.ociError()
		}
	}
	return nil
}

func (uint32PtrBind *uint32PtrBind) close() {
	defer func() {
		recover()
	}()
	uint32PtrBind.ocibnd = nil
	uint32PtrBind.value = nil
	uint32PtrBind.environment.uint32PtrBindPool.Put(uint32PtrBind)
}
