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

type uint64SliceBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	ociNumbers  []C.OCINumber
}

func (uint64SliceBind *uint64SliceBind) bindOra(values []Uint64, position int, ocistmt *C.OCIStmt) error {
	uint64Values := make([]uint64, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			uint64Values[n] = values[n].Value
		}
	}
	return uint64SliceBind.bind(uint64Values, nullInds, position, ocistmt)
}

func (uint64SliceBind *uint64SliceBind) bind(values []uint64, nullInds []C.sb2, position int, ocistmt *C.OCIStmt) error {
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	uint64SliceBind.ociNumbers = make([]C.OCINumber, len(values))
	for n, _ := range values {
		alenp[n] = C.ub4(C.sizeof_OCINumber)
		r := C.OCINumberFromInt(
			uint64SliceBind.environment.ocierr, //OCIError            *err,
			unsafe.Pointer(&values[n]),         //const void          *inum,
			8, //uword               inum_length,
			C.OCI_NUMBER_UNSIGNED,          //uword               inum_s_flag,
			&uint64SliceBind.ociNumbers[n]) //OCINumber           *number );
		if r == C.OCI_ERROR {
			return uint64SliceBind.environment.ociError()
		}
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&uint64SliceBind.ocibnd),         //OCIBind      **bindpp,
		uint64SliceBind.environment.ocierr,             //OCIError     *errhp,
		C.ub4(position),                                //ub4          position,
		unsafe.Pointer(&uint64SliceBind.ociNumbers[0]), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),                      //sb8          value_sz,
		C.SQLT_VNU,                                     //ub2          dty,
		unsafe.Pointer(&nullInds[0]),                   //void         *indp,
		&alenp[0],                                      //ub4          *alenp,
		&rcodep[0],                                     //ub2          *rcodep,
		0,                                              //ub4          maxarr_len,
		nil,                                            //ub4          *curelep,
		C.OCI_DEFAULT)                                  //ub4          mode );
	if r == C.OCI_ERROR {
		return uint64SliceBind.environment.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		uint64SliceBind.ocibnd,
		uint64SliceBind.environment.ocierr,
		C.ub4(C.sizeof_OCINumber), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),       //ub4         indskip,
		C.ub4(C.sizeof_ub4),       //ub4         alskip,
		C.ub4(C.sizeof_ub2))       //ub4         rcskip
	if r == C.OCI_ERROR {
		return uint64SliceBind.environment.ociError()
	}
	return nil
}

func (uint64SliceBind *uint64SliceBind) setPtr() error {
	return nil
}

func (uint64SliceBind *uint64SliceBind) close() {
	defer func() {
		recover()
	}()
	uint64SliceBind.ocibnd = nil
	uint64SliceBind.environment.uint64SliceBindPool.Put(uint64SliceBind)
}
