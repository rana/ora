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
import (
	"time"
	"unsafe"

	"gopkg.in/rana/ora.v4/date"
)

type bndDatePtr struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	value   *Date
	ocidate [1]date.Date
	nullp
}

func (bnd *bndDatePtr) bind(value *Date, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	bnd.nullp.Set(value == nil || value.IsNull())
	if value != nil {
		bnd.ocidate[0] = value.Date
	}
	//bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind, "bind val=%#v null?=%t datep=%#v (%v)\n", bnd.value, bnd.nullp.IsNull(), bnd.datep, bnd.datep.Get())
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		&bnd.ocibnd,                 //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ocidate[0]),     //void         *valuep,
		C.LENGTH_TYPE(7),                    //sb8          value_sz,
		C.SQLT_DAT,                          //ub2          dty,
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
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind, "setPtr val=%#v nullp=%#v datep=%#v (%v)\n", bnd.value, bnd.nullp, bnd.ocidate[0], bnd.ocidate[0].Get())

	if bnd.value == nil {
		return nil
	}
	if bnd.nullp.IsNull() {
		bnd.value.Set(time.Time{})
		return nil
	}
	bnd.value.Date = bnd.ocidate[0]
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
	stmt.putBnd(bndIdxDatePtr, bnd)
	return nil
}
