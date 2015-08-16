// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type bndStringPtr struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	isNull C.sb2
	value  *string
	alen   C.ACTUAL_LENGTH_TYPE
	buf    []byte
}

func (bnd *bndStringPtr) bind(value *string, position int, stringPtrBufferSize int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	if stringPtrBufferSize < 2 {
		stringPtrBufferSize = 2
	} else if stringPtrBufferSize%2 == 1 {
		stringPtrBufferSize++
	}
	L, C := len(bnd.buf), cap(bnd.buf)
	if C < stringPtrBufferSize {
		bnd.buf = make([]byte, L, stringPtrBufferSize)
		C = stringPtrBufferSize
	}
	if value == nil {
		bnd.isNull = C.sb2(-1)
		bnd.alen = 0
		bnd.buf = bnd.buf[:2]
	} else {
		if len(*value) == 0 {
			bnd.buf = bnd.buf[:2] // to be able to address bnd.buf[0]
			bnd.buf[0], bnd.buf[1] = 0, 0
		} else {
			L = len(*value)
			if L < 2 {
				L = 2
			} else if L%2 == 0 {
				L++
			}
			bnd.buf = bnd.buf[:L]
			bnd.buf[L-1] = 0
			copy(bnd.buf, []byte(*value))
		}
		bnd.alen = C.ACTUAL_LENGTH_TYPE(len(*value))
	}
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p pos=%d cap=%d len=%d alen=%d bufSize=%d", bnd, position, cap(bnd.buf), len(bnd.buf), bnd.alen, stringPtrBufferSize)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),  //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position),             //ub4          position,
		unsafe.Pointer(&bnd.buf[0]), //void         *valuep,
		C.LENGTH_TYPE(cap(bnd.buf)), //sb8          value_sz,
		C.SQLT_CHR,                  //ub2          dty,
		unsafe.Pointer(&bnd.isNull), //void         *indp,
		&bnd.alen,                   //ub2          *alenp,
		nil,                         //ub2          *rcodep,
		0,                           //ub4          maxarr_len,
		nil,                         //ub4          *curelep,
		C.OCI_DEFAULT)               //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndStringPtr) setPtr() error {
	if bnd.value == nil {
		return nil
	}
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"StringPtr.setPtr isNull=%d alen=%d", bnd.isNull, bnd.alen)
	if bnd.isNull > C.sb2(-1) {
		*bnd.value = string(bnd.buf[:bnd.alen])
	} else {
		*bnd.value = ""
	}

	return nil
}

func (bnd *bndStringPtr) close() (err error) {
	/*
		defer func() {
			if value := recover(); value != nil {
				err = errR(value)
			}
		}()
	*/
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.alen = 0
	bnd.buf = bnd.buf[:0]
	stmt.putBnd(bndIdxStringPtr, bnd)
	return nil
}
