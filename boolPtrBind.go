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
	"bytes"
	"unsafe"
)

type boolPtrBind struct {
	env      *Environment
	ocibnd   *C.OCIBind
	value    *bool
	isNull   C.sb2
	buf      []byte
	trueRune rune
}

func (b *boolPtrBind) bind(value *bool, position int, trueRune rune, ocistmt *C.OCIStmt) error {
	b.value = value
	b.trueRune = trueRune
	if cap(b.buf) < 2 {
		b.buf = make([]byte, 2)
	}
	r := C.OCIBindByPos2(
		ocistmt,                   //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),  //OCIBind      **bindpp,
		b.env.ocierr,              //OCIError     *errhp,
		C.ub4(position),           //ub4          position,
		unsafe.Pointer(&b.buf[0]), //void         *valuep,
		C.sb8(len(b.buf)),         //sb8          value_sz,
		C.SQLT_CHR,                //ub2          dty,
		unsafe.Pointer(&b.isNull), //void         *indp,
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

func (b *boolPtrBind) setPtr() error {
	if b.isNull > -1 {
		*b.value = bytes.Runes(b.buf)[0] == b.trueRune
	}
	return nil
}

func (b *boolPtrBind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.isNull = C.sb2(0)
	b.value = nil
	clear(b.buf, 0)
	b.env.boolPtrBindPool.Put(b)
}
