// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"bytes"
	"github.com/golang/glog"
	"unsafe"
)

type bndBoolPtr struct {
	stmt     *Stmt
	ocibnd   *C.OCIBind
	isNull   C.sb2
	value    *bool
	buf      []byte
	trueRune rune
}

func (bnd *bndBoolPtr) bind(value *bool, position int, trueRune rune, stmt *Stmt) error {
	glog.Infoln("position: ", position)
	bnd.stmt = stmt
	bnd.value = value
	bnd.trueRune = trueRune
	if cap(bnd.buf) < 2 {
		bnd.buf = make([]byte, 2)
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

func (bnd *bndBoolPtr) setPtr() error {
	if bnd.isNull > -1 {
		*bnd.value = bytes.Runes(bnd.buf)[0] == bnd.trueRune
	}
	return nil
}

func (bnd *bndBoolPtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	clear(bnd.buf, 0)
	stmt.putBnd(bndIdxBoolPtr, bnd)
	return nil
}
