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
	"unsafe"
)

type bndBoolSlice struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	buf    bytes.Buffer
	bytes  []byte
}

func (bnd *bndBoolSlice) bindOra(values []Bool, position namedPos, falseRune rune, trueRune rune, stmt *Stmt) error {
	boolValues := make([]bool, len(values))
	nullInds := make([]C.sb2, len(values))
	for n := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			boolValues[n] = values[n].Value
		}
	}
	return bnd.bind(boolValues, nullInds, position, falseRune, trueRune, stmt)
}

func (bnd *bndBoolSlice) bind(values []bool, nullInds []C.sb2, position namedPos, falseRune rune, trueRune rune, stmt *Stmt) (err error) {
	bnd.stmt = stmt
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ACTUAL_LENGTH_TYPE, len(values))
	rcodep := make([]C.ub2, len(values))
	var maxLen = 1
	for n, bValue := range values {
		if bValue {
			_, err = bnd.buf.WriteRune(trueRune)
			if err != nil {
				return err
			}
		} else {
			_, err = bnd.buf.WriteRune(falseRune)
			if err != nil {
				return err
			}
		}
		alenp[n] = 1
	}
	bnd.bytes = bnd.buf.Bytes()

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
		unsafe.Pointer(&bnd.bytes[0]), //void         *valuep,
		C.LENGTH_TYPE(maxLen),         //sb8          value_sz,
		C.SQLT_CHR,                    //ub2          dty,
		unsafe.Pointer(&nullInds[0]),  //void         *indp,
		&alenp[0],                     //ub4          *alenp,
		&rcodep[0],                    //ub2          *rcodep,
		0,                             //ub4          maxarr_len,
		nil,                           //ub4          *curelep,
		C.OCI_DEFAULT)                 //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}

	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,                  //OCIBind     *bindp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(maxLen),               //ub4         pvskip,
		C.ub4(C.sizeof_sb2),         //ub4         indskip,
		C.ub4(C.sizeof_ub4),         //ub4         alskip,
		C.ub4(C.sizeof_ub2))         //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}

	return nil
}

func (bnd *bndBoolSlice) setPtr() error {
	return nil
}

func (bnd *bndBoolSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.bytes = nil
	bnd.buf.Reset()
	stmt.putBnd(bndIdxBoolSlice, bnd)
	return nil
}
