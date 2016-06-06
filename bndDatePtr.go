// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndDatePtr struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	value  *Date
	nullp
	datep
}

func (bnd *bndDatePtr) bind(value *Date, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	bnd.nullp.Set(value == nil)
	if value != nil {
		bnd.datep.Set(bnd.stmt.ses.srv.env, value.Value)
	}
	//bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind, "bind val=%#v null?=%t datep=%#v (%v)\n", bnd.value, bnd.nullp.IsNull(), bnd.datep, bnd.datep.Get())
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                    //OCIStmt      *stmtp,
		&bnd.ocibnd,                         //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,         //OCIError     *errhp,
		C.ub4(position),                     //ub4          position,
		unsafe.Pointer(bnd.datep.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.datep.Size()),     //sb8          value_sz,
		C.SQLT_ODT,                          //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndDatePtr) setPtr() (err error) {
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind, "setPtr val=%#v nullp=%#v datep=%#v (%v)\n", bnd.value, bnd.nullp, bnd.datep, bnd.datep.Get())

	if bnd.value == nil {
		return nil
	}
	if bnd.nullp.IsNull() {
		bnd.value.IsNull = true
		return nil
	}
	bnd.value.IsNull = false
	bnd.value.Value = bnd.datep.Get()
	return nil
}

func (bnd *bndDatePtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.nullp.Free()
	bnd.datep.Free()
	stmt.putBnd(bndIdxTimePtr, bnd)
	return nil
}
