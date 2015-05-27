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

type defUint16 struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  C.OCINumber
	null       C.sb2
	isNullable bool
}

func (def *defUint16) define(position int, isNullable bool, rset *Rset) error {
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

func (def *defUint16) value() (value interface{}, err error) {
	if def.isNullable {
		oraUint16Value := Uint16{IsNull: def.null < 0}
		if !oraUint16Value.IsNull {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr,      //OCIError              *err,
				&def.ociNumber,                        //const OCINumber       *number,
				C.uword(2),                            //uword                 rsl_length,
				C.OCI_NUMBER_UNSIGNED,                 //uword                 rsl_flag,
				unsafe.Pointer(&oraUint16Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraUint16Value
	} else {
		if def.null > -1 {
			var uint16Value uint16
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&def.ociNumber,                   //const OCINumber       *number,
				C.uword(2),                       //uword                 rsl_length,
				C.OCI_NUMBER_UNSIGNED,            //uword                 rsl_flag,
				unsafe.Pointer(&uint16Value))     //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = uint16Value
		}
	}
	return value, err
}

func (def *defUint16) alloc() error {
	return nil
}

func (def *defUint16) free() {

}

func (def *defUint16) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxUint16, def)
	return nil
}
