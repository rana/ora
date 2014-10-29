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

type uint16Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (uint16Bind *uint16Bind) bind(value uint16, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		uint16Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),        //const void          *inum,
		2, //uword               inum_length,
		C.OCI_NUMBER_UNSIGNED, //uword               inum_s_flag,
		&uint16Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return uint16Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint16Bind.ocibnd),     //OCIBind      **bindpp,
		uint16Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                       //ub4          position,
		unsafe.Pointer(&uint16Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),             //sb8          value_sz,
		C.SQLT_VNU,                            //ub2          dty,
		nil,                                   //void         *indp,
		nil,                                   //ub2          *alenp,
		nil,                                   //ub2          *rcodep,
		0,                                     //ub4          maxarr_len,
		nil,                                   //ub4          *curelep,
		C.OCI_DEFAULT)                         //ub4          mode );
	if r == C.OCI_ERROR {
		return uint16Bind.environment.ociError()
	}
	return nil
}

func (uint16Bind *uint16Bind) setPtr() error {
	return nil
}

func (uint16Bind *uint16Bind) close() {
	defer func() {
		recover()
	}()
	uint16Bind.ocibnd = nil
	uint16Bind.environment.uint16BindPool.Put(uint16Bind)
}
