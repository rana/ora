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

type bndNumStringSlice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
	arrHlp
}

func (bnd *bndNumStringSlice) bindOra(values []OraNum, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	stringValues := make([]Num, len(values))
	if cap(bnd.nullInds) < len(values) {
		bnd.nullInds = make([]C.sb2, len(values))
	} else {
		bnd.nullInds = bnd.nullInds[:len(values)]
	}
	for n := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			stringValues[n] = Num(values[n].Value)
		}
	}
	return bnd.bind(stringValues, bnd.nullInds, position, stmt, isAssocArray)
}

func (bnd *bndNumStringSlice) bind(values []Num, nullInds []C.sb2, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	if nullInds != nil {
		bnd.nullInds = nullInds
	}
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, Num("0"))
	}
	bnd.ociNumbers = make([]C.OCINumber, len(values))
	alen := C.ACTUAL_LENGTH_TYPE(C.sizeof_OCINumber)
	for n := range values {
		bnd.alen[n] = alen
		numbers := bnd.ociNumbers[n : n+1 : n+1] // against _cgoCheckPointer0
		if err := bnd.stmt.ses.srv.env.numberFromText(&numbers[0], string(values[n])); err != nil {
			return iterations, err
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
		unsafe.Pointer(&bnd.ociNumbers[0]), //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),  //sb8          value_sz,
		C.SQLT_VNU,                         //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),   //void         *indp,
		&bnd.alen[0],                       //ub4          *alenp,
		&bnd.rcode[0],                      //ub2          *rcodep,
		getMaxarrLen(C, isAssocArray),      //ub4          maxarr_len,
		curlenp,       //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(C.sizeof_OCINumber), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),       //ub4         indskip,
		C.ub4(C.sizeof_ub4),       //ub4         alskip,
		C.ub4(C.sizeof_ub2))       //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	return iterations, nil
}

func (bnd *bndNumStringSlice) setPtr() error {
	return nil
}

func (bnd *bndNumStringSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociNumbers = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxNumStringSlice, bnd)
	return nil
}
