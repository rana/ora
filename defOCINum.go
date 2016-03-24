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

	"gopkg.in/rana/ora.v3/num"
)

type defOCINum struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  [1]C.OCINumber
	isNullable bool
	nullp
}

func (def *defOCINum) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                    //OCIStmt     *stmtp,
		&def.ocidef,                         //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,    //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&def.ociNumber[0]),   //void        *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),   //sb8         value_sz,
		C.SQLT_VNU,                          //ub2         dty,
		unsafe.Pointer(def.nullp.Pointer()), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}
func (def *defOCINum) value() (value interface{}, err error) {
	if def.nullp.IsNull() {
		if def.isNullable {
			return String{IsNull: true}, nil
		}
		return "0", nil
	}
	length := int(def.ociNumber[0].OCINumberPart[0])
	num := num.OCINum(((*[C.OCI_NUMBER_SIZE]byte)(
		unsafe.Pointer(&def.ociNumber[0].OCINumberPart[1]),
	))[:length])
	if def.isNullable {
		return OraOCINum{Value: num}, nil
	}
	return OCINum{OCINum: num}, nil
}

func (def *defOCINum) alloc() error {
	return nil
}

func (def *defOCINum) free() {
}

func (def *defOCINum) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.nullp.Free()
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
