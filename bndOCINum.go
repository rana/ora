// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

// FromC converts from the given C.OCINumber.
func (num *OCINum) FromC(x C.OCINumber) {
	a := *(*[22]byte)(unsafe.Pointer(&x))
	length := int(a[0])
	if length < 0 || length > 21 {
		num.OCINum = num.OCINum[:0]
		return
	}

	if cap(num.OCINum) < length {
		num.OCINum = make([]byte, length, 22-1)
	} else {
		num.OCINum = num.OCINum[:length]
	}
	copy(num.OCINum[:length], a[1:1+length])
}

// ToC converts the OCINum into the given *C.OCINumber.
func (num OCINum) ToC(x *C.OCINumber) {
	a := ((*[22]byte)(unsafe.Pointer(x)))
	a[0] = byte(len(num.OCINum))
	copy(a[1:1+len(num.OCINum)], num.OCINum)
	for i := 1 + len(num.OCINum); i < 22; i++ {
		a[i] = 0
	}
}

type bndOCINum struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
}

func (bnd *bndOCINum) bind(value OCINum, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	value.ToC(&bnd.ociNumber[0])
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
		unsafe.Pointer(&bnd.ociNumber[0]), //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber), //sb8          value_sz,
		C.SQLT_VNU,                        //ub2          dty,
		nil,                               //void         *indp,
		nil,                               //ub2          *alenp,
		nil,                               //ub2          *rcodep,
		0,                                 //ub4          maxarr_len,
		nil,                               //ub4          *curelep,
		C.OCI_DEFAULT)                     //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndOCINum) setPtr() error {
	return nil
}

func (bnd *bndOCINum) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxOCINum, bnd)
	return nil
}
