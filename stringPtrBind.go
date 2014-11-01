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

type stringPtrBind struct {
	env    *Environment
	ocibnd *C.OCIBind
	value  *string
	isNull C.sb2
	buffer []byte
}

func (b *stringPtrBind) bind(value *string, position int, stringPtrBufferSize int, ocistmt *C.OCIStmt) error {
	b.value = value
	if cap(b.buffer) < stringPtrBufferSize {
		b.buffer = make([]byte, stringPtrBufferSize)
	}
	r := C.OCIBindByPos2(
		ocistmt,                      //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),     //OCIBind      **bindpp,
		b.env.ocierr,                 //OCIError     *errhp,
		C.ub4(position),              //ub4          position,
		unsafe.Pointer(&b.buffer[0]), //void         *valuep,
		C.sb8(len(b.buffer)),         //sb8          value_sz,
		C.SQLT_CHR,                   //ub2          dty,
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

func (b *stringPtrBind) setPtr() error {
	if b.isNull > -1 {
		// Buffer is padded with Space char (32)
		*b.value = stringTrimmed(b.buffer, 32)
	}
	return nil
}

func (b *stringPtrBind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.isNull = C.sb2(0)
	b.value = nil
	clear(b.buffer, 32)
	b.env.stringPtrBindPool.Put(b)
}
