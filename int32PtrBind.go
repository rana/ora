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

type int32PtrBind struct {
	env       *Environment
	ocibnd    *C.OCIBind
	ociNumber C.OCINumber
	isNull    C.sb2
	value     *int32
}

func (b *int32PtrBind) bind(value *int32, position int, ocistmt *C.OCIStmt) error {
	b.value = value
	r := C.OCIBindByPos2(
		ocistmt,                      //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),     //OCIBind      **bindpp,
		b.env.ocierr,                 //OCIError     *errhp,
		C.ub4(position),              //ub4          position,
		unsafe.Pointer(&b.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),    //sb8          value_sz,
		C.SQLT_VNU,                   //ub2          dty,
		unsafe.Pointer(&b.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *int32PtrBind) setPtr() error {
	if b.isNull > -1 {
		r := C.OCINumberToInt(
			b.env.ocierr,            //OCIError              *err,
			&b.ociNumber,            //const OCINumber       *number,
			C.uword(4),              //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,     //uword                 rsl_flag,
			unsafe.Pointer(b.value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			return b.env.ociError()
		}
	}
	return nil
}

func (b *int32PtrBind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.value = nil
	b.env.int32PtrBindPool.Put(b)
}
