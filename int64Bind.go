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

type int64Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (int64Bind *int64Bind) bind(value int64, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		int64Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),       //const void          *inum,
		8,                    //uword               inum_length,
		C.OCI_NUMBER_SIGNED,  //uword               inum_s_flag,
		&int64Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return int64Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int64Bind.ocibnd),     //OCIBind      **bindpp,
		int64Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                      //ub4          position,
		unsafe.Pointer(&int64Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),            //sb8          value_sz,
		C.SQLT_VNU,                           //ub2          dty,
		nil,                                  //void         *indp,
		nil,                                  //ub2          *alenp,
		nil,                                  //ub2          *rcodep,
		0,                                    //ub4          maxarr_len,
		nil,                                  //ub4          *curelep,
		C.OCI_DEFAULT)                        //ub4          mode );
	if r == C.OCI_ERROR {
		return int64Bind.environment.ociError()
	}
	return nil
}

func (int64Bind *int64Bind) setPtr() error {
	return nil
}

func (int64Bind *int64Bind) close() {
	defer func() {
		recover()
	}()
	int64Bind.ocibnd = nil
	int64Bind.environment.int64BindPool.Put(int64Bind)
}
