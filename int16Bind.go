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
	env       *Environment
	ocibnd    *C.OCIBind
	ociNumber C.OCINumber
}

func (b *int16Bind) bind(value int16, position int, ocistmt *C.OCIStmt) error {
	r := C.OCINumberFromInt(
		b.env.ocierr,           //OCIError            *err,
		unsafe.Pointer(&value), //const void          *inum,
		2,                   //uword               inum_length,
		C.OCI_NUMBER_SIGNED, //uword               inum_s_flag,
		&b.ociNumber)        //OCINumber           *number );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt,                      //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),     //OCIBind      **bindpp,
		b.env.ocierr,                 //OCIError     *errhp,
		C.ub4(position),              //ub4          position,
		unsafe.Pointer(&b.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),    //sb8          value_sz,
		C.SQLT_VNU,                   //ub2          dty,
		nil,                          //void         *indp,
		nil,                          //ub2          *alenp,
		nil,                          //ub2          *rcodep,
		0,                            //ub4          maxarr_len,
		nil,                          //ub4          *curelep,
		C.OCI_DEFAULT)                //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *int16Bind) setPtr() error {
	return nil
}

func (b *int16Bind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.env.int16BindPool.Put(b)
}
