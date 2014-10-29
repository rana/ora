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

type float32Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (float32Bind *float32Bind) bind(value float32, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromReal(
		float32Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),         //const void          *rnum,
		4, //uword               rnum_length,
		&float32Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return float32Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&float32Bind.ocibnd),     //OCIBind      **bindpp,
		float32Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                        //ub4          position,
		unsafe.Pointer(&float32Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),              //sb8          value_sz,
		C.SQLT_VNU,                             //ub2          dty,
		nil,                                    //void         *indp,
		nil,                                    //ub2          *alenp,
		nil,                                    //ub2          *rcodep,
		0,                                      //ub4          maxarr_len,
		nil,                                    //ub4          *curelep,
		C.OCI_DEFAULT)                          //ub4          mode );
	if r == C.OCI_ERROR {
		return float32Bind.environment.ociError()
	}
	return nil
}

func (float32Bind *float32Bind) setPtr() error {
	return nil
}

func (float32Bind *float32Bind) close() {
	defer func() {
		recover()
	}()
	float32Bind.ocibnd = nil
	float32Bind.environment.float32BindPool.Put(float32Bind)
}
