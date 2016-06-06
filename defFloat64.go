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
import (
	"unsafe"
)

// defFloat32.go is generated from defFloat64.go!

type defFloat64 struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defFloat64) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}

func (def *defFloat64) value(offset int) (value interface{}, err error) {
	if def.isNullable {
		oraFloat64Value := Float64{IsNull: def.nullInds[offset] < 0}
		if !oraFloat64Value.IsNull {
			var float64Value float64
			on := def.ociNumber[offset]
			r := C.OCINumberToReal(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&on,                                    //const OCINumber     *number,
				byteWidth64,                            //uword               rsl_length,
				unsafe.Pointer(&oraFloat64Value.Value)) //void                *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = float64Value
		}
		value = oraFloat64Value
	} else {
		if def.nullInds[offset] > -1 {
			var float64Value float64
			on := def.ociNumber[offset]
			r := C.OCINumberToReal(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&on,                           //const OCINumber     *number,
				byteWidth64,                   //uword               rsl_length,
				unsafe.Pointer(&float64Value)) //void                *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = float64Value
		}
	}
	return value, err
}

func (def *defFloat64) alloc() error {
	return nil
}

func (def *defFloat64) free() {
}

func (def *defFloat64) close() (err error) {
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
	rset.putDef(defIdxFloat64, def)
	return nil
}
