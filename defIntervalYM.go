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

type defIntervalYM struct {
	rset        *Rset
	ocidef      *C.OCIDefine
	ociInterval *C.OCIInterval
	null        C.sb2
}

func (def *defIntervalYM) define(position int, rset *Rset) error {
	glog.Infoln("position: ", position)
	def.rset = rset
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,                      //OCIStmt     *stmtp,
		&def.ocidef,                           //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,      //OCIError    *errhp,
		C.ub4(position),                       //ub4         position,
		unsafe.Pointer(&def.ociInterval),      //void        *valuep,
		C.sb8(unsafe.Sizeof(def.ociInterval)), //sb8         value_sz,
		C.SQLT_INTERVAL_YM,                    //ub2         dty,
		unsafe.Pointer(&def.null),             //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defIntervalYM) value() (value interface{}, err error) {
	intervalYM := IntervalYM{IsNull: def.null < 0}
	if !intervalYM.IsNull {
		var year C.sb4
		var month C.sb4
		r := C.OCIIntervalGetYearMonth(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //void               *hndl,
			def.rset.stmt.ses.srv.env.ocierr,                 //OCIError           *err,
			&year,           //sb4                *yr,
			&month,          //sb4                *mnth,
			def.ociInterval) //const OCIInterval  *interval );
		if r == C.OCI_ERROR {
			err = def.rset.stmt.ses.srv.env.ociError()
		}
		intervalYM.Year = int32(year)
		intervalYM.Month = int32(month)
	}
	return intervalYM, err
}

func (def *defIntervalYM) alloc() error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),    //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&def.ociInterval)), //dvoid         **descpp,
		C.OCI_DTYPE_INTERVAL_YM,                             //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci interval handle during define")
	}
	return nil
}

func (def *defIntervalYM) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(def.ociInterval), //void     *descp,
		C.OCI_DTYPE_INTERVAL_YM)         //timeDefine.descTypeCode)                //ub4      type );
}

func (def *defIntervalYM) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociInterval = nil
	rset.putDef(defIdxIntervalYM, def)
	return nil
}
