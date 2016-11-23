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

	"gopkg.in/rana/ora.v4/num"
)

type defOCINum struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defOCINum) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}
func (def *defOCINum) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return OraOCINum{IsNull: true}, nil
		}
		return nil, nil
	}
	length := int(def.ociNumber[offset].OCINumberPart[0])
	num := num.OCINum(((*[C.OCI_NUMBER_SIZE]byte)(
		unsafe.Pointer(&def.ociNumber[offset].OCINumberPart[1]),
	))[:length])
	if def.isNullable {
		return OraOCINum{Value: num}, nil
	}
	return OCINum{OCINum: num}, nil
}

func (def *defOCINum) alloc() error { return nil }
func (def *defOCINum) free()        {}

func (def *defOCINum) close() (err error) {
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
	rset.putDef(defIdxOCINum, def)
	return nil
}

func (env *Env) numberToText(dest []byte, number C.OCINumber) ([]byte, error) {
	if cap(dest) < numStringLen {
		dest = make([]byte, numStringLen)
	} else {
		dest = dest[:numStringLen]
	}
	bufSize := C.ub4(len(dest))
	r := C.OCINumberToText(
		env.ocierr, //OCIError              *err,
		&number,    //const OCINumber     *number,
		numberFmtC,
		C.ub4(numberFmtLen), //ub4                fmt_length,
		numberNLSC,          //CONST OraText      *nls_params,
		C.ub4(numberNLSLen), //ub4                nls_p_length,
		&bufSize,            //ub4 ,
		(*C.oratext)(unsafe.Pointer(&dest[0])), //OraText                *rsl );
	)
	if r == C.OCI_ERROR {
		return dest, env.ociError()
	}
	return dest[:bufSize], nil
}
