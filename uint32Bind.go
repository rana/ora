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

type uint32Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (uint32Bind *uint32Bind) bind(value uint32, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		uint32Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),        //const void          *inum,
		4, //uword               inum_length,
		C.OCI_NUMBER_UNSIGNED, //uword               inum_s_flag,
		&uint32Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return uint32Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint32Bind.ocibnd),     //OCIBind      **bindpp,
		uint32Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                       //ub4          position,
		unsafe.Pointer(&uint32Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),             //sb8          value_sz,
		C.SQLT_VNU,                            //ub2          dty,
		nil,                                   //void         *indp,
		nil,                                   //ub2          *alenp,
		nil,                                   //ub2          *rcodep,
		0,                                     //ub4          maxarr_len,
		nil,                                   //ub4          *curelep,
		C.OCI_DEFAULT)                         //ub4          mode );
	if r == C.OCI_ERROR {
		return uint32Bind.environment.ociError()
	}
	return nil
}

func (uint32Bind *uint32Bind) setPtr() error {
	return nil
}

func (uint32Bind *uint32Bind) close() {
	defer func() {
		recover()
	}()
	uint32Bind.ocibnd = nil
	uint32Bind.environment.uint32BindPool.Put(uint32Bind)
}
