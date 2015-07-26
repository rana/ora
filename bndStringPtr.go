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
	alen   []C.ACTUAL_LENGTH_TYPE
	buf    []byte
}

func (bnd *bndStringPtr) bind(value *string, position int, stringPtrBufferSize int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	var length int
	if cap(bnd.buf) < stringPtrBufferSize {
		bnd.buf = make([]byte, 1, stringPtrBufferSize)
	}
	if value == nil {
		bnd.isNull = C.sb2(-1)
	} else {
		length = len(*value)
	}
	if length == 0 {
		bnd.buf = bnd.buf[:1] // to be able to address bnd.buf[0]
		bnd.buf[0] = 0
	} else {
		bnd.buf = bnd.buf[:length]
		copy(bnd.buf, []byte(*value))
	}
	if cap(bnd.alen) < 1 {
		bnd.alen = []C.ACTUAL_LENGTH_TYPE{C.ACTUAL_LENGTH_TYPE(length)}
	} else {
		bnd.alen = bnd.alen[:1]
		bnd.alen[0] = C.ACTUAL_LENGTH_TYPE(length)
	}
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"StringPtr.bind(%d) cap=%d len=%d alen=%d", position, cap(bnd.buf), len(bnd.buf), bnd.alen[0])
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),  //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position),             //ub4          position,
		unsafe.Pointer(&bnd.buf[0]), //void         *valuep,
		C.LENGTH_TYPE(cap(bnd.buf)), //sb8          value_sz,
		C.SQLT_CHR,                  //ub2          dty,
		unsafe.Pointer(&bnd.isNull), //void         *indp,
		&bnd.alen[0],                //ub2          *alenp,
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
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"StringPtr.setPtr isNull=%d alen=%d", bnd.isNull, bnd.alen[0])
	if bnd.isNull > C.sb2(-1) {
		*bnd.value = string(bnd.buf[:bnd.alen[0]])
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
	bnd.alen = bnd.alen[:0]
	bnd.buf = bnd.buf[:0]
	stmt.putBnd(bndIdxStringPtr, bnd)
	return nil
}
