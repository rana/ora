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

// Generate all the def[IU]int{8,16,32,64}.go from defInt64.go
//
//go:generate go run gen.go

type defInt64 struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defInt64) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}

func (def *defInt64) value(offset int) (value interface{}, err error) {
	if def.isNullable {
		oraInt64Value := Int64{IsNull: def.nullInds[offset] < 0}
		if !oraInt64Value.IsNull {
			on := def.ociNumber[offset]
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&on,                                  //const OCINumber       *number,
				byteWidth64,                          //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,                  //uword                 rsl_flag,
				unsafe.Pointer(&oraInt64Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraInt64Value
	} else {
		var int64Value int64
		//if !def.nullp.IsNull() {
		if def.nullInds[offset] >= 0 {
			on := def.ociNumber[offset]
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&on,                         //const OCINumber       *number,
				byteWidth64,                 //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,         //uword                 rsl_flag,
				unsafe.Pointer(&int64Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = int64Value
	}
	return value, err
}

func (def *defInt64) alloc() error {
	return nil
}

func (def *defInt64) free() {
}

func (def *defInt64) close() (err error) {
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
	def.arrHlp.close()
	rset.putDef(defIdxInt64, def)
	return nil
}
