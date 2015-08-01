// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type defDate struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociDate    C.OCIDate
	null       C.sb2
	isNullable bool
}

func (def *defDate) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.ociDate),     //void        *valuep,
		C.LENGTH_TYPE(C.sizeof_OCIDate),  //sb8         value_sz,
		C.SQLT_ODT,                       //defineTypeCode,                               //ub2         dty,
		unsafe.Pointer(&def.null),        //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defDate) value() (value interface{}, err error) {
	if def.isNullable {
		oraTimeValue := Date{IsNull: def.null < C.sb2(0)}
		if !oraTimeValue.IsNull {
			oraTimeValue.Value = ociGetDateTime(def.ociDate)
		}
		return oraTimeValue, nil
	}
	return Date{Value: ociGetDateTime(def.ociDate)}, nil
}

func (def *defDate) alloc() error { return nil }
func (def *defDate) free()        {}

func (def *defDate) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxDate, def)
	return nil
}
