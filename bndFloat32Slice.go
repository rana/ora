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

type bndFloat32Slice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
	values     []Float32
	floats     []float32
	arrHlp
}

func (bnd *bndFloat32Slice) bindOra(values []Float32, position int, stmt *Stmt) (uint32, error) {
	if cap(bnd.floats) < cap(values) {
		bnd.floats = make([]float32, len(values), cap(values))
	} else {
		bnd.floats = bnd.floats[:len(values)]
	}
	if cap(bnd.nullInds) < cap(values) {
		bnd.nullInds = make([]C.sb2, len(values), cap(values))
	} else {
		bnd.nullInds = bnd.nullInds[:len(values)]
	}
	bnd.values = values
	for n := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = 0
			bnd.floats[n] = values[n].Value
		}
	}
	return bnd.bind(bnd.floats, position, stmt)
}

func (bnd *bndFloat32Slice) bind(values []float32, position int, stmt *Stmt) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, stmt.stmtType)
	if needAppend {
		values = append(values, 0)
	}
	bnd.floats = values
	if cap(bnd.ociNumbers) < C {
		bnd.ociNumbers = make([]C.OCINumber, L, C)
	} else {
		bnd.ociNumbers = bnd.ociNumbers[:L]
	}
	for n := range values {
		bnd.alen[n] = C.ACTUAL_LENGTH_TYPE(C.sizeof_OCINumber)
		r := C.OCINumberFromReal(
			bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
			unsafe.Pointer(&values[n]),  //const void          *rnum,
			4,                  //uword               rnum_length,
			&bnd.ociNumbers[n]) //OCINumber           *number );
		if r == C.OCI_ERROR {
			return iterations, bnd.stmt.ses.srv.env.ociError()
		}
	}
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p pos=%d cap=%d len=%d curlen=%d curlenp=%p iterations=%d",
		bnd, position, cap(bnd.ociNumbers), len(bnd.ociNumbers), bnd.curlen, curlenp, iterations)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                          //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,               //OCIError     *errhp,
		C.ub4(position),                           //ub4          position,
		unsafe.Pointer(&bnd.ociNumbers[0]),        //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),         //sb8          value_sz,
		C.SQLT_VNU,                                //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]),          //void         *indp,
		&bnd.alen[0],                              //ub4          *alenp,
		&bnd.rcode[0],                             //ub2          *rcodep,
		C.ACTUAL_LENGTH_TYPE(cap(bnd.ociNumbers)), //ub4          maxarr_len,
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
	bnd.floats = bnd.floats[:n]
	bnd.nullInds = bnd.nullInds[:n]
	if bnd.values != nil {
		bnd.values = bnd.values[:n]
	}
	for i, number := range bnd.ociNumbers[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			r := C.OCINumberToReal(
				bnd.stmt.ses.srv.env.ocierr,    //OCIError              *err,
				&number,                        //const OCINumber     *number,
				C.uword(4),                     //uword               rsl_length,
				unsafe.Pointer(&bnd.floats[i])) //void                *rsl );
			if r == C.OCI_ERROR {
				return bnd.stmt.ses.srv.env.ociError()
			}
			if bnd.values != nil {
				bnd.values[i].IsNull = false
				bnd.values[i].Value = bnd.floats[i]
			}
		} else if bnd.values != nil {
			bnd.values[i].IsNull = true
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
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxFloat32Slice, bnd)
	return nil
}
