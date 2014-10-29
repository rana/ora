// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

type float32SliceBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumbers  []C.OCINumber
}

func (float32SliceBind *float32SliceBind) bindOra(values []Float32, position int, ocistmt *C.OCIStmt) error {
	float32Values := make([]float32, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			float32Values[n] = values[n].Value
		}
	}
	return float32SliceBind.bind(float32Values, nullInds, position, ocistmt)
}

func (float32SliceBind *float32SliceBind) bind(values []float32, nullInds []C.sb2, position int, ocistmt *C.OCIStmt) error {
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	float32SliceBind.ociNumbers = make([]C.OCINumber, len(values))
	for n, _ := range values {
		alenp[n] = C.ub4(C.sizeof_OCINumber)
		r := C.OCINumberFromReal(
			float32SliceBind.environment.ocierr, //OCIError            *err,
			unsafe.Pointer(&values[n]),          //const void          *rnum,
			4, //uword               rnum_length,
			&float32SliceBind.ociNumbers[n]) //OCINumber           *number );
		if r == C.OCI_ERROR {
			return float32SliceBind.environment.ociError()
		}
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&float32SliceBind.ocibnd),         //OCIBind      **bindpp,
		float32SliceBind.environment.ocierr,             //OCIError     *errhp,
		C.ub4(position),                                 //ub4          position,
		unsafe.Pointer(&float32SliceBind.ociNumbers[0]), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                       //sb8          value_sz,
		C.SQLT_VNU,                                      //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                    //void         *indp,
		&alenp[0],                                       //ub4          *alenp,
		&rcodep[0],                                      //ub2          *rcodep,
		0,                                               //ub4          maxarr_len,
		nil,                                             //ub4          *curelep,
		C.OCI_DEFAULT)                                   //ub4          mode );
	if r == C.OCI_ERROR {
		return float32SliceBind.environment.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		float32SliceBind.ocibnd,
		float32SliceBind.environment.ocierr,
		C.ub4(C.sizeof_OCINumber), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),       //ub4         indskip,
		C.ub4(C.sizeof_ub4),       //ub4         alskip,
		C.ub4(C.sizeof_ub2))       //ub4         rcskip
	if r == C.OCI_ERROR {
		return float32SliceBind.environment.ociError()
	}
	return nil
}

func (float32SliceBind *float32SliceBind) setPtr() error {
	return nil
}

func (float32SliceBind *float32SliceBind) close() {
	defer func() {
		recover()
	}()
	float32SliceBind.ocibnd = nil
	float32SliceBind.environment.float32SliceBindPool.Put(float32SliceBind)
}
