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
	environment *Environment
	ocibnd      *C.OCIBind
	value       *string
	isNull      C.sb2
	buffer      []byte
}

func (stringPtrBind *stringPtrBind) bind(value *string, position int, stringPtrBufferSize int, ocistmt *C.OCIStmt) error {
	stringPtrBind.value = value
	if cap(stringPtrBind.buffer) < stringPtrBufferSize {
		stringPtrBind.buffer = make([]byte, stringPtrBufferSize)
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&stringPtrBind.ocibnd),     //OCIBind      **bindpp,
		stringPtrBind.environment.ocierr,         //OCIError     *errhp,
		C.ub4(position),                          //ub4          position,
		unsafe.Pointer(&stringPtrBind.buffer[0]), //void         *valuep,
		C.sb8(len(stringPtrBind.buffer)),         //sb8          value_sz,
		C.SQLT_CHR,                               //ub2          dty,
		unsafe.Pointer(&stringPtrBind.isNull),    //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return stringPtrBind.environment.ociError()
	}
	return nil
}

func (stringPtrBind *stringPtrBind) setPtr() error {
	if stringPtrBind.isNull > -1 {
		// Buffer is padded with Space char (32)
		*stringPtrBind.value = stringTrimmed(stringPtrBind.buffer, 32)
	}
	return nil
}

func (stringPtrBind *stringPtrBind) close() {
	defer func() {
		recover()
	}()
	stringPtrBind.ocibnd = nil
	stringPtrBind.isNull = C.sb2(0)
	stringPtrBind.value = nil
	clear(stringPtrBind.buffer, 32)
	stringPtrBind.environment.stringPtrBindPool.Put(stringPtrBind)
}
