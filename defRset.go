// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type defRset struct {
	rset    *Rset
	ocidef  *C.OCIDefine
	ocistmt *C.OCIStmt
	result  *Rset
}

func (def *defRset) define(position int, rset *Rset) error {
	def.rset = rset
	def.result = rset

	upOciStmt, err := def.rset.stmt.ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
	if err != nil {
		return errE(err)
	}
	def.ocistmt = (*C.OCIStmt)(upOciStmt)

	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                 //OCIStmt     *stmtp,
		&def.ocidef,                      //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&def.ocistmt),     //void        *valuep,
		C.LENGTH_TYPE(C.sizeof_dvoid),    //sb8         value_sz,
		C.SQLT_RSET,                      //ub2         dty,
		nil,                              //void        *indp,
		nil,                              //ub2         *rlenp,
		nil,                              //ub2         *rcodep,
		C.OCI_DEFAULT)                    //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defRset) value() (value interface{}, err error) {
	err = def.result.open(def.result.stmt, def.ocistmt)
	def.result.stmt.openRsets.add(def.result)
	if err == nil {
		err = def.result.stmt.setPrefetchSize()
		// open result set is successful; will be freed by Rset
		def.ocistmt = nil
	}
	return def.result, err
}

func (def *defRset) alloc() error {
	return nil
}

func (def *defRset) free() {
}

func (def *defRset) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ocistmt = nil
	def.result = nil
	rset.putDef(defIdxRset, def)
	return nil
}
