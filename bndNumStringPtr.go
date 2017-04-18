// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndNumStringPtr struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
	value     *Num
	buf       [numStringLen]byte
	nullp
}

func (bnd *bndNumStringPtr) bind(value *Num, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	bnd.nullp.Set(value == nil || *value == "")
	//length := C.ub4(0)
	if value != nil && *value != "" {
		//length = C.ub4(copy(bnd.buf[:], string(*value)))
		//fmt.Printf("NumberFromtext %q [%d]\n", value, length)
		if err := bnd.stmt.ses.srv.env.numberFromText(&bnd.ociNumber[0], string(*value)); err != nil {
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
		unsafe.Pointer(&bnd.ociNumber[0]),   //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),   //sb8          value_sz,
		C.SQLT_VNU,                          //ub2          dty,
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

func (bnd *bndNumStringPtr) setPtr() error {
	if bnd.nullp.IsNull() {
		return nil
	}
	bufLen := C.ub4(numStringLen)
	r := C.OCINumberToText(
		bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
		&bnd.ociNumber[0],           //const OCINumber     *number,
		numberFmtC,
		C.ub4(numberFmtLen), //ub4                fmt_length,
		numberNLSC,          //CONST OraText      *nls_params,
		C.ub4(numberNLSLen), //ub4                nls_p_length,
		&bufLen,
		(*C.oratext)(unsafe.Pointer(&bnd.buf[0])), //void                *rsl );
	)
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	if bufLen > 0 && bnd.buf[0] == '.' {
		*bnd.value = Num(append(append(make([]byte, 0, int(bufLen)+1),
			'0'),
			bnd.buf[:int(bufLen)]...))
	} else {
		*bnd.value = Num(bnd.buf[:int(bufLen)])
	}
	return nil
}

func (bnd *bndNumStringPtr) close() (err error) {
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
	stmt.putBnd(bndIdxNumStringPtr, bnd)
	return nil
}
