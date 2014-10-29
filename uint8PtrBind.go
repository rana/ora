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

type uint8PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *uint8
}

func (uint8PtrBind *uint8PtrBind) bind(value *uint8, position int, ocistmt *C.OCIStmt) error {
	uint8PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint8PtrBind.ocibnd),     //OCIBind      **bindpp,
		uint8PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(&uint8PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),               //sb8          value_sz,
		C.SQLT_VNU,                              //ub2          dty,
		unsafe.Pointer(&uint8PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return uint8PtrBind.environment.ociError()
	}
	return nil
}

func (uint8PtrBind *uint8PtrBind) setPtr() error {
	if uint8PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			uint8PtrBind.environment.ocierr,    //OCIError              *err,
			&uint8PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(1),                         //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,              //uword                 rsl_flag,
			unsafe.Pointer(uint8PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return uint8PtrBind.environment.ociError()
		}
	}
	return nil
}

func (uint8PtrBind *uint8PtrBind) close() {
	defer func() {
		recover()
	}()
	uint8PtrBind.ocibnd = nil
	uint8PtrBind.value = nil
	uint8PtrBind.environment.uint8PtrBindPool.Put(uint8PtrBind)
}
