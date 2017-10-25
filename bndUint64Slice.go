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

type bndUint64Slice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
	values     *[]Uint64
	ints       *[]uint64
	isOra      bool
	arrHlp
}

func (bnd *bndUint64Slice) bindOra(values *[]Uint64, position namedPos, stmt *Stmt, isAssocArray bool) (uint32, error) {
	L, C := len(*values), cap(*values)
	var ints []uint64
	if bnd.ints == nil {
		bnd.ints = &ints
	} else {
		ints = *bnd.ints
	}
	if cap(ints) < C {
		ints = make([]uint64, L, C)
	} else {
		ints = ints[:L]
	}
	if cap(bnd.nullInds) < C {
		bnd.nullInds = make([]C.sb2, L, C)
	} else {
		bnd.nullInds = bnd.nullInds[:L]
	}
	bnd.values = values
	for n, v := range *values {
		if v.IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = 0
			ints[n] = v.Value
		}
	}
	*bnd.ints = ints
	bnd.isOra = true
	return bnd.bind(bnd.ints, position, stmt, isAssocArray)
}

func (bnd *bndUint64Slice) bind(values *[]uint64, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	V := *values
	L, C := len(V), cap(V)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		V = append(V, 0)
	}
	if cap(bnd.ociNumbers) < C {
		bnd.ociNumbers = make([]C.OCINumber, L, C)
	} else {
		bnd.ociNumbers = bnd.ociNumbers[:L]
	}
	alen := C.ACTUAL_LENGTH_TYPE(C.sizeof_OCINumber)
	for n := range V {
		bnd.alen[n] = alen
	}
	*values = V
	bnd.ints = values
	if len(V) > 0 {
		if r := C.numberFromIntSlice(
			bnd.stmt.ses.srv.env.ocierr,
			unsafe.Pointer(&V[0]),
			byteWidth64,
			C.OCI_NUMBER_UNSIGNED,
			&bnd.ociNumbers[0],
			C.ub4(len(V)),
		); r == C.OCI_ERROR {
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}
	}
	if !bnd.isOra {
		for i := range bnd.nullInds {
			bnd.nullInds[i] = 0
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
		C.ub4(C.sizeof_OCINumber),          //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                //ub4         indskip,
		C.ub4(C.sizeof_ACTUAL_LENGTH_TYPE), //ub4         alskip,
		C.ub4(C.sizeof_ub2))                //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	return iterations, nil
}

func (bnd *bndUint64Slice) setPtr() error {
	if !bnd.IsAssocArr() {
		return nil
	}
	n := int(bnd.curlen)
	ints := (*bnd.ints)[:n]
	bnd.nullInds = bnd.nullInds[:n]
	*bnd.ints = ints
	var V []Uint64
	if bnd.values != nil {
		V := *bnd.values
		if cap(V) < n {
			V = make([]Uint64, n)
		} else {
			V = V[:n]
		}
		*bnd.values = V
	}
	for i, number := range bnd.ociNumbers[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			r := C.OCINumberToInt(
				bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&number,                     //const OCINumber     *number,
				byteWidth64,                 //uword               rsl_length,
				C.OCI_NUMBER_UNSIGNED,         //uword               rsl_flag,
				unsafe.Pointer(&ints[i]))    //void                *rsl );
			if r == C.OCI_ERROR {
				return bnd.stmt.ses.srv.env.ociError()
			}
			if V != nil {
				V[i].IsNull = false
				V[i].Value = ints[i]
			}
		} else if V != nil {
			V[i].IsNull = true
		}
	}
	return nil
}

func (bnd *bndUint64Slice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.values = nil
	bnd.ints = nil
	bnd.isOra = false
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxUint64Slice, bnd)
	return nil
}
