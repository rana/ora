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

type defUint64 struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  [1]C.OCINumber
	isNullable bool
	nullp
}

func (def *defUint64) define(position int, isNullable bool, rset *Rset) error {
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

func (def *defUint64) value() (value interface{}, err error) {
	if def.isNullable {
		oraUint64Value := Uint64{IsNull: def.nullp.IsNull()}
		if !oraUint64Value.IsNull {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr,      //OCIError              *err,
				&def.ociNumber[0],                     //const OCINumber       *number,
				C.uword(8),                            //uword                 rsl_length,
				C.OCI_NUMBER_UNSIGNED,                 //uword                 rsl_flag,
				unsafe.Pointer(&oraUint64Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraUint64Value
	} else {
		if !def.nullp.IsNull() {
			var uint64Value uint64
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&def.ociNumber[0],                //const OCINumber       *number,
				C.uword(8),                       //uword                 rsl_length,
				C.OCI_NUMBER_UNSIGNED,            //uword                 rsl_flag,
				unsafe.Pointer(&uint64Value))     //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = uint64Value
		}
	}
	return value, err
}

func (def *defUint64) alloc() error {
	return nil
}

func (def *defUint64) free() {
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
	def.nullp.Free()
	rset.putDef(defIdxUint64, def)
	return nil
}
