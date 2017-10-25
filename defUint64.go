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

// Generate all the def[IU]int{8,16,32,64}.go from defUint64.go
//
// Generated from defUint64.go by go run gen.go

type defUint64 struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defUint64) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}

func (def *defUint64) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Uint64{IsNull: true}, nil
		}
		return nil, nil
	}
	var uint64Value uint64
	on := def.ociNumber[offset]
	r := C.OCINumberToInt(
		def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
		&on,                         //const OCINumber       *number,
		byteWidth64,                 //uword                 rsl_length,
		C.OCI_NUMBER_UNSIGNED,         //uword                 rsl_flag,
		unsafe.Pointer(&uint64Value)) //void                  *rsl );
	if r == C.OCI_ERROR {
		err = def.rset.stmt.ses.srv.env.ociError()
	}
	if def.isNullable {
		return Uint64{Value: uint64Value}, err
	}
	return uint64Value, err
}

func (def *defUint64) alloc() error { return nil }
func (def *defUint64) free() {
	def.arrHlp.close()
}

func (def *defUint64) close() (err error) {
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
	rset.putDef(defIdxUint64, def)
	return nil
}
