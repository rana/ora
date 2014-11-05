// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"github.com/golang/glog"
	"unsafe"
)

type bndStringPtr struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	isNull C.sb2
	value  *string
	buf    []byte
}

func (bnd *bndStringPtr) bind(value *string, position int, stringPtrBufferSize int, stmt *Stmt) error {
	glog.Infoln("position: ", position)
	bnd.stmt = stmt
	bnd.value = value
	if cap(bnd.buf) < stringPtrBufferSize {
		bnd.buf = make([]byte, stringPtrBufferSize)
	}
	r := C.OCIBindByPos2(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),  //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position),             //ub4          position,
		unsafe.Pointer(&bnd.buf[0]), //void         *valuep,
		C.sb8(len(bnd.buf)),         //sb8          value_sz,
		C.SQLT_CHR,                  //ub2          dty,
		unsafe.Pointer(&bnd.isNull), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndStringPtr) setPtr() error {
	if bnd.isNull > -1 {
		// Buffer is padded with Space char (32)
		*bnd.value = stringTrimmed(bnd.buf, 32)
	}
	return nil
}

func (bnd *bndStringPtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	clear(bnd.buf, 32)
	stmt.putBnd(bndIdxStringPtr, bnd)
return nil
}
