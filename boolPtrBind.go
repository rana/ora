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
	environment *Environment
	ocibnd      *C.OCIBind
	value       *bool
	isNull      C.sb2
	buffer      []byte
	trueRune    rune
}

func (boolPtrBind *boolPtrBind) bind(value *bool, position int, trueRune rune, ocistmt *C.OCIStmt) error {
	boolPtrBind.value = value
	boolPtrBind.trueRune = trueRune
	if cap(boolPtrBind.buffer) < 2 {
		boolPtrBind.buffer = make([]byte, 2)
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&boolPtrBind.ocibnd),     //OCIBind      **bindpp,
		boolPtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                        //ub4          position,
		unsafe.Pointer(&boolPtrBind.buffer[0]), //void         *valuep,
		C.sb8(len(boolPtrBind.buffer)),         //sb8          value_sz,
		C.SQLT_CHR,                             //ub2          dty,
		unsafe.Pointer(&boolPtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return boolPtrBind.environment.ociError()
	}
	return nil
}

func (boolPtrBind *boolPtrBind) setPtr() error {
	if boolPtrBind.isNull > -1 {
		*boolPtrBind.value = bytes.Runes(boolPtrBind.buffer)[0] == boolPtrBind.trueRune
	}
	return nil
}

func (boolPtrBind *boolPtrBind) close() {
	defer func() {
		recover()
	}()
	boolPtrBind.ocibnd = nil
	boolPtrBind.isNull = C.sb2(0)
	boolPtrBind.value = nil
	clear(boolPtrBind.buffer, 0)
	boolPtrBind.environment.boolPtrBindPool.Put(boolPtrBind)
}
