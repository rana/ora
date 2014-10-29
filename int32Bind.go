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

type int32Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (int32Bind *int32Bind) bind(value int32, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		int32Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),       //const void          *inum,
		4,                    //uword               inum_length,
		C.OCI_NUMBER_SIGNED,  //uword               inum_s_flag,
		&int32Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return int32Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int32Bind.ocibnd),     //OCIBind      **bindpp,
		int32Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                      //ub4          position,
		unsafe.Pointer(&int32Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),            //sb8          value_sz,
		C.SQLT_VNU,                           //ub2          dty,
		nil,                                  //void         *indp,
		nil,                                  //ub2          *alenp,
		nil,                                  //ub2          *rcodep,
		0,                                    //ub4          maxarr_len,
		nil,                                  //ub4          *curelep,
		C.OCI_DEFAULT)                        //ub4          mode );
	if r == C.OCI_ERROR {
		return int32Bind.environment.ociError()
	}
	return nil
}

func (int32Bind *int32Bind) setPtr() error {
	return nil
}

func (int32Bind *int32Bind) close() {
	defer func() {
		recover()
	}()
	int32Bind.ocibnd = nil
	int32Bind.environment.int32BindPool.Put(int32Bind)
}
