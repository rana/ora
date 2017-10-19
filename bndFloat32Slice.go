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

// http://docs.oracle.com/database/121/LNOCI/oci05bnd.htm#sthref868

type bndFloat32Slice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
	values     *[]Float32
	floats     *[]float32
	isOra      bool
	arrHlp
}

func (bnd *bndFloat32Slice) bindOra(values *[]Float32, position namedPos, stmt *Stmt, isAssocArray bool) (uint32, error) {
	L, C := len(*values), cap(*values)
	var V []float32
	if bnd.floats == nil {
		bnd.floats = &V
	} else {
		V = *bnd.floats
	}
	if cap(V) < C {
		V = make([]float32, L, C)
	} else {
		V = V[:L]
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
			V[n] = v.Value
		}
	}
	*bnd.floats = V
	bnd.isOra = true
	return bnd.bind(bnd.floats, position, stmt, isAssocArray)
}

func (bnd *bndFloat32Slice) bind(values *[]float32, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	// ensure we have at least 1 slot in the slice
	var V []float32
	if values == nil {
		values = &V
	} else {
		V = *values
	}
	L, C := len(V), cap(V)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		V = append(V, 0)
	}
	*values = V
	bnd.floats = values
	if cap(bnd.ociNumbers) < C {
		bnd.ociNumbers = make([]C.OCINumber, L, C)
	} else {
		bnd.ociNumbers = bnd.ociNumbers[:L]
	}
	alen := C.ACTUAL_LENGTH_TYPE(C.sizeof_OCINumber)
	for n := range V {
		bnd.alen[n] = alen
	}
	if len(V) > 0 {
		if r := C.numberFromFloatSlice(
			bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
			unsafe.Pointer(&V[0]),       //const void          *rnum,
			byteWidth32,                 //uword               rnum_length,
			&bnd.ociNumbers[0],          //OCINumber           *number
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
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v cap=%d len=%d curlen=%d curlenp=%p", bnd, position, cap(bnd.ociNumbers), len(bnd.ociNumbers), bnd.curlen, curlenp)
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

func (bnd *bndFloat32Slice) setPtr() error {
	if !bnd.IsAssocArr() {
		return nil
	}
	n := int(bnd.curlen)
	var F []float32
	if bnd.floats == nil {
		bnd.floats = &F
	} else {
		F = *bnd.floats
	}
	F = F[:n]
	*bnd.floats = F
	bnd.nullInds = bnd.nullInds[:n]
	var V []Float32
	if bnd.values != nil {
		V = *bnd.values
		if cap(V) < n {
			V = make([]Float32, n)
		} else {
			V = V[:n]
		}
		*bnd.values = V
	}
	for i, number := range bnd.ociNumbers[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			arr := F[i : i+1 : i+1]
			r := C.OCINumberToReal(
				bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&number,                     //const OCINumber     *number,
				byteWidth32,                 //uword               rsl_length,
				unsafe.Pointer(&arr[0]))     //void                *rsl );
			if r == C.OCI_ERROR {
				return bnd.stmt.ses.srv.env.ociError()
			}
			if bnd.values != nil {
				V[i].IsNull = false
				V[i].Value = F[i]
			}
		} else if bnd.values != nil {
			V[i].IsNull = true
		}
	}
	return nil
}

func (bnd *bndFloat32Slice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.values = nil
	bnd.floats = nil
	bnd.isOra = false
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxFloat32Slice, bnd)
	return nil
}
