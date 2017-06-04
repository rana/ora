// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type defRset struct {
	ociDef
	ocistmt []*C.OCIStmt
	result  []*Rset
}

func (def *defRset) define(position int, rset *Rset) error {
	def.rset = rset
	if def.ocistmt != nil {
		C.free(unsafe.Pointer(&def.ocistmt[0]))
	}
	def.ocistmt = (*((*[MaxFetchLen]*C.OCIStmt)(C.malloc(C.size_t(rset.fetchLen) * C.sof_Stmtp))))[:rset.fetchLen]
	def.result = make([]*Rset, len(def.ocistmt))

	// create result set
	for i := range def.result {
		result := _drv.rsetPool.Get().(*Rset)
		if result.id == 0 {
			result.id = _drv.rsetId.nextId()
		}
		result.autoClose = true
		result.env = def.rset.env
		result.stmt = rset.stmt
		result.ocistmt = rset.ocistmt
		def.result[i] = result

		upOciStmt, err := def.rset.stmt.ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
		if err != nil {
			return errE(err)
		}
		def.ocistmt[i] = (*C.OCIStmt)(upOciStmt)
	}

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ocistmt[0]), int(C.sof_Stmtp), C.SQLT_RSET)
}

func (def *defRset) value(offset int) (value interface{}, err error) {
	rst := def.result[offset]

	err = rst.open(rst.stmt, def.ocistmt[offset])
	rst.stmt.openRsets.add(rst)

	return rst, err
}

func (def *defRset) alloc() error {
	return nil
}

func (def *defRset) free() {
	def.arrHlp.close()
	for i, p := range def.ocistmt {
		if p == nil {
			continue
		}
		def.ocistmt[i] = nil
		def.rset.stmt.ses.srv.env.freeOciHandle(unsafe.Pointer(p), C.OCI_HTYPE_STMT)
	}
}

func (def *defRset) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	if def.ocistmt != nil {
		C.free(unsafe.Pointer(&def.ocistmt[0]))
		def.ocistmt = nil
	}
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxRset, def)
	return nil
}
