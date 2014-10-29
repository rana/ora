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

type float64SliceBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumbers  []C.OCINumber
}

func (float64SliceBind *float64SliceBind) bindOra(values []Float64, position int, ocistmt *C.OCIStmt) error {
	float64Values := make([]float64, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			float64Values[n] = values[n].Value
		}
	}
	return float64SliceBind.bind(float64Values, nullInds, position, ocistmt)
}

func (float64SliceBind *float64SliceBind) bind(values []float64, nullInds []C.sb2, position int, ocistmt *C.OCIStmt) error {
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	float64SliceBind.ociNumbers = make([]C.OCINumber, len(values))
	for n, _ := range values {
		alenp[n] = C.ub4(C.sizeof_OCINumber)
		r := C.OCINumberFromReal(
			float64SliceBind.environment.ocierr, //OCIError            *err,
			unsafe.Pointer(&values[n]),          //const void          *rnum,
			8, //uword               rnum_length,
			&float64SliceBind.ociNumbers[n]) //OCINumber           *number );
		if r == C.OCI_ERROR {
			return float64SliceBind.environment.ociError()
		}
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&float64SliceBind.ocibnd),         //OCIBind      **bindpp,
		float64SliceBind.environment.ocierr,             //OCIError     *errhp,
		C.ub4(position),                                 //ub4          position,
		unsafe.Pointer(&float64SliceBind.ociNumbers[0]), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                       //sb8          value_sz,
		C.SQLT_VNU,                                      //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                    //void         *indp,
		&alenp[0],                                       //ub4          *alenp,
		&rcodep[0],                                      //ub2          *rcodep,
		0,                                               //ub4          maxarr_len,
		nil,                                             //ub4          *curelep,
		C.OCI_DEFAULT)                                   //ub4          mode );
	if r == C.OCI_ERROR {
		return float64SliceBind.environment.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		float64SliceBind.ocibnd,
		float64SliceBind.environment.ocierr,
		C.ub4(C.sizeof_OCINumber), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),       //ub4         indskip,
		C.ub4(C.sizeof_ub4),       //ub4         alskip,
		C.ub4(C.sizeof_ub2))       //ub4         rcskip
	if r == C.OCI_ERROR {
		return float64SliceBind.environment.ociError()
	}
	return nil
}

func (float64SliceBind *float64SliceBind) setPtr() error {
	return nil
}

func (float64SliceBind *float64SliceBind) close() {
	defer func() {
		recover()
	}()
	float64SliceBind.ocibnd = nil
	float64SliceBind.environment.float64SliceBindPool.Put(float64SliceBind)
}
