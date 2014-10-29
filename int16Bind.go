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

type int16Bind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumber   C.OCINumber
}

func (int16Bind *int16Bind) bind(value int16, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		int16Bind.environment.ocierr, //OCIError            *err,
		unsafe.Pointer(&value),       //const void          *inum,
		2,                    //uword               inum_length,
		C.OCI_NUMBER_SIGNED,  //uword               inum_s_flag,
		&int16Bind.ociNumber) //OCINumber           *number );
	if r == C.OCI_ERROR {
		return int16Bind.environment.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&int16Bind.ocibnd),     //OCIBind      **bindpp,
		int16Bind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                      //ub4          position,
		unsafe.Pointer(&int16Bind.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),            //sb8          value_sz,
		C.SQLT_VNU,                           //ub2          dty,
		nil,                                  //void         *indp,
		nil,                                  //ub2          *alenp,
		nil,                                  //ub2          *rcodep,
		0,                                    //ub4          maxarr_len,
		nil,                                  //ub4          *curelep,
		C.OCI_DEFAULT)                        //ub4          mode );
	if r == C.OCI_ERROR {
		return int16Bind.environment.ociError()
	}
	return nil
}

func (int16Bind *int16Bind) setPtr() error {
	return nil
}

func (int16Bind *int16Bind) close() {
	defer func() {
		recover()
	}()
	int16Bind.ocibnd = nil
	int16Bind.environment.int16BindPool.Put(int16Bind)
}
