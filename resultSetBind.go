// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

type resultSetBind struct {
	env     *Environment
	stmt    *Statement
	ocistmt *C.OCIStmt
	ocibnd  *C.OCIBind
	isNull  C.sb2
	value   *ResultSet
}

func (b *resultSetBind) bind(value *ResultSet, position int, stmt *Statement) error {
	b.stmt = stmt
	b.value = value
	// Allocate a statement handle
	ocistmt, err := b.env.allocateOciHandle(C.OCI_HTYPE_STMT)
	b.ocistmt = (*C.OCIStmt)(ocistmt)
	if err != nil {
		return err
	}
	r := C.OCIBindByPos2(
		stmt.ocistmt,               //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),   //OCIBind      **bindpp,
		b.env.ocierr,               //OCIError     *errhp,
		C.ub4(position),            //ub4          position,
		unsafe.Pointer(&b.ocistmt), //void         *valuep,
		C.sb8(0),                   //sb8          value_sz,
		C.SQLT_RSET,                //ub2          dty,
		unsafe.Pointer(&b.isNull),  //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}

	return nil
}

func (b *resultSetBind) setPtr() error {
	err := b.value.open(b.stmt, b.ocistmt)
	b.stmt.rsts.PushBack(b.value)
	if err == nil {
		// open result set is successful; will be freed by ResultSet
		b.ocistmt = nil
	}

	return err
}

func (b *resultSetBind) close() {
	defer func() {
		recover()
	}()
	// release ocistmt handle for failed ResultSet binding
	// ResultSet will release handle for successful bind
	if b.ocistmt != nil {
		b.env.freeOciHandle(unsafe.Pointer(b.ocistmt), C.OCI_HTYPE_STMT)
	}
	b.stmt = nil
	b.ocistmt=nil
	b.ocistmt = nil
	b.isNull = C.sb2(0)
	b.value = nil
	b.env.resultSetBindPool.Put(b)
}
