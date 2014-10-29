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

type uint16PtrBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
	isNull      C.sb2
	value       *uint16
}

func (uint16PtrBind *uint16PtrBind) bind(value *uint16, position int, ocistmt *C.OCIStmt) error {
	uint16PtrBind.value = value
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint16PtrBind.ocibnd),     //OCIBind      **bindpp,
		uint16PtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                          //ub4          position,
		unsafe.Pointer(&uint16PtrBind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                //sb8          value_sz,
		C.SQLT_VNU,                               //ub2          dty,
		unsafe.Pointer(&uint16PtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return uint16PtrBind.environment.ociError()
	}
	return nil
}

func (uint16PtrBind *uint16PtrBind) setPtr() error {
	if uint16PtrBind.isNull > -1 {
		r := C.OCINumberToInt(
			uint16PtrBind.environment.ocierr,    //OCIError              *err,
			&uint16PtrBind.ociNumber,            //const OCINumber       *number,
			C.uword(2),                          //uword                 rsl_length,
			C.OCI_NUMBER_UNSIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(uint16PtrBind.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return uint16PtrBind.environment.ociError()
		}
	}
	return nil
}

func (uint16PtrBind *uint16PtrBind) close() {
	defer func() {
		recover()
	}()
	uint16PtrBind.ocibnd = nil
	uint16PtrBind.value = nil
	uint16PtrBind.environment.uint16PtrBindPool.Put(uint16PtrBind)
}
