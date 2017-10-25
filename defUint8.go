// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

// Generate all the def[IU]int{8,16,32,8}.go from defUint8.go
//
// Generated from defUint8.go by go run gen.go

type defUint8 struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defUint8) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}

func (def *defUint8) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Uint8{IsNull: true}, nil
		}
		return nil, nil
	}
	var uint8Value uint8
	on := def.ociNumber[offset]
	r := C.OCINumberToInt(
		def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
		&on,                         //const OCINumber       *number,
		byteWidth8,                 //uword                 rsl_length,
		C.OCI_NUMBER_UNSIGNED,         //uword                 rsl_flag,
		unsafe.Pointer(&uint8Value)) //void                  *rsl );
	if r == C.OCI_ERROR {
		err = def.rset.stmt.ses.srv.env.ociError()
	}
	if def.isNullable {
		return Uint8{Value: uint8Value}, err
	}
	return uint8Value, err
}

func (def *defUint8) alloc() error { return nil }
func (def *defUint8) free() {
	def.arrHlp.close()
}

func (def *defUint8) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
		def.ociNumber = nil
	}
	rset.putDef(defIdxUint8, def)
	return nil
}
