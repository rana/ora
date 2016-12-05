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
	"bytes"
	"unicode/utf8"
	"unsafe"
)

type bndBoolPtr struct {
	stmt     *Stmt
	ocibnd   *C.OCIBind
	value    *bool
	buf      []byte
	trueRune rune
	nullp
}

func (bnd *bndBoolPtr) bind(value *bool, position namedPos, trueRune rune, stmt *Stmt) error {
	//Log.Infof("%v.bind(%t, %d)", bnd, value, position)
	bnd.stmt = stmt
	bnd.value = value
	bnd.trueRune = trueRune
	if cap(bnd.buf) < 2 {
		bnd.buf = make([]byte, 2)
	}
	if value != nil && *value {
		if _, err := bytes.NewBuffer(bnd.buf).WriteRune(trueRune); err != nil {
			return err
		}
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.buf[0]),         //void         *valuep,
		C.LENGTH_TYPE(len(bnd.buf)),         //sb8          value_sz,
		C.SQLT_CHR,                          //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
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
	//Log.Infof("%s.setPtr()", bnd)
	if !bnd.nullp.IsNull() {
		r, _ := utf8.DecodeRune(bnd.buf)
		*bnd.value = r == bnd.trueRune
	} else {
		bnd.value = nil
	}
	return nil
}

func (bnd *bndBoolPtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.nullp.Free()
	clear(bnd.buf, 0)
	stmt.putBnd(bndIdxBoolPtr, bnd)
	return nil
}
