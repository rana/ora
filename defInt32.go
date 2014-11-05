// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"github.com/golang/glog"
	"unsafe"
)

type defInt32 struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  C.OCINumber
	null       C.sb2
	isNullable bool
}

func (def *defInt32) define(position int, isNullable bool, rset *Rset) error {
	glog.Infoln("position: ", position)
	def.rset = rset
	def.isNullable = isNullable
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,            //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.ociNumber),   //void        *valuep,
		C.sb8(C.sizeof_OCINumber),        //sb8         value_sz,
		C.SQLT_VNU,                       //ub2         dty,
		unsafe.Pointer(&def.null),        //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defInt32) value() (value interface{}, err error) {
	if def.isNullable {
		oraInt32Value := Int32{IsNull: def.null < 0}
		if !oraInt32Value.IsNull {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr,     //OCIError              *err,
				&def.ociNumber,                       //const OCINumber       *number,
				C.uword(4),                           //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,                  //uword                 rsl_flag,
				unsafe.Pointer(&oraInt32Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraInt32Value
	} else {
		if def.null > -1 {
			var int32Value int32
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&def.ociNumber,                   //const OCINumber       *number,
				C.uword(4),                       //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,              //uword                 rsl_flag,
				unsafe.Pointer(&int32Value))      //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
			value = int32Value
		}
	}
	return value, err
}

func (def *defInt32) alloc() error {
	return nil
}

func (def *defInt32) free() {

}

func (def *defInt32) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxInt32, def)
return nil
}
