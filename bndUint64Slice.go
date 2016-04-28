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

type bndUint64Slice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
	uints      []uint64
	values     *[]Uint64
	arrHlp
}

func (bnd *bndUint64Slice) bindOra(values *[]Uint64, position int, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	L, C := len(*values), cap(*values)
	bnd.values = values
	if cap(bnd.uints) < C {
		bnd.uints = make([]uint64, L, C)
	} else {
		bnd.uints = bnd.uints[:L]
	}
	if cap(bnd.nullInds) < C {
		bnd.nullInds = make([]C.sb2, L, C)
	} else {
		bnd.nullInds = bnd.nullInds[:L]
	}
	for n, v := range *values {
		if v.IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.uints[n] = v.Value
		}
	}
	return bnd.bind(bnd.uints, position, stmt, isAssocArray)
}

func (bnd *bndUint64Slice) bind(values []uint64, position int, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	// ensure we have at least 1 slot in the slice
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, 0)
	}
	bnd.uints = values
	if cap(bnd.ociNumbers) < C {
		bnd.ociNumbers = make([]C.OCINumber, L, C)
	} else {
		bnd.ociNumbers = bnd.ociNumbers[:L]
	}
	alen := C.ACTUAL_LENGTH_TYPE(C.sizeof_OCINumber)
	for n := range values {
		bnd.alen[n] = alen
	}
	if len(values) > 0 {
		if r := C.numberFromIntSlice(
			bnd.stmt.ses.srv.env.ocierr,
			unsafe.Pointer(&values[0]),
			8,
			C.OCI_NUMBER_UNSIGNED,
			&bnd.ociNumbers[0],
			C.ub4(len(values)),
		); r == C.OCI_ERROR {
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}
	}
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,        //OCIError     *errhp,
		C.ub4(position),                    //ub4          position,
		unsafe.Pointer(&bnd.ociNumbers[0]), //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),  //sb8          value_sz,
		C.SQLT_VNU,                         //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),   //void         *indp,
		&bnd.alen[0],                       //ub4          *alenp,
		&bnd.rcode[0],                      //ub2          *rcodep,
		C.ub4(C),                           //ub4          maxarr_len,
		curlenp,                            //ub4          *curelep,
		C.OCI_DEFAULT)                      //ub4          mode );
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

func (bnd *bndUint64Slice) setPtr() error {
	if !bnd.IsAssocArr() {
		return nil
	}
	n := int(bnd.curlen)
	bnd.uints = bnd.uints[:n]
	bnd.nullInds = bnd.nullInds[:n]
	if bnd.values != nil {
		if cap(*bnd.values) < n {
			*bnd.values = make([]Uint64, n)
		} else {
			*bnd.values = (*bnd.values)[:n]
		}
	}
	for i, number := range bnd.ociNumbers[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			arr := bnd.uints[i : i+1 : i+1]
			r := C.OCINumberToInt(
				bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
				&number,                     //const OCINumber     *number,
				C.uword(8),                  //uword               rsl_length,
				C.OCI_NUMBER_UNSIGNED,       //uword               rsl_flag,
				unsafe.Pointer(&arr[0]))     //void                *rsl );
			if r == C.OCI_ERROR {
				return bnd.stmt.ses.srv.env.ociError()
			}
			if bnd.values != nil {
				(*bnd.values)[i].IsNull = false
				(*bnd.values)[i].Value = bnd.uints[i]
			}
		} else if bnd.values != nil {
			(*bnd.values)[i].IsNull = true
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
	bnd.ociNumbers = nil
	bnd.arrHlp.close()
	bnd.values = nil
	bnd.uints = nil
	stmt.putBnd(bndIdxUint64Slice, bnd)
	return nil
}
