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

type defInt16 struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  C.OCINumber
	null       C.sb2
	isNullable bool
}

func (def *defInt16) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                  //OCIStmt     *stmtp,
		&def.ocidef,                       //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,  //OCIError    *errhp,
		C.ub4(position),                   //ub4         position,
		unsafe.Pointer(&def.ociNumber),    //void        *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber), //sb8         value_sz,
		C.SQLT_VNU,                        //ub2         dty,
		unsafe.Pointer(&def.null),         //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defInt16) alloc() error {
	return nil
}

func (def *defInt16) free() {

}

func (def *defInt16) value() (value interface{}, err error) {
	if def.isNullable {
		oraInt16Value := Int16{IsNull: def.null < C.sb2(0)}
		if !oraInt16Value.IsNull {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr,     //OCIError              *err,
				&def.ociNumber,                       //const OCINumber       *number,
				C.uword(2),                           //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,                  //uword                 rsl_flag,
				unsafe.Pointer(&oraInt16Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraInt16Value
	} else {
		if def.null > C.sb2(-1) {
			var int16Value int16
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&def.ociNumber,                   //const OCINumber       *number,
				C.uword(2),                       //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,              //uword                 rsl_flag,
				unsafe.Pointer(&int16Value))      //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = int16Value
		}
	}
	return value, err
}

func (def *defInt16) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxInt16, def)
	return nil
}
