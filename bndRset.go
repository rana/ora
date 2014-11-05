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

type bndRset struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	ocistmt *C.OCIStmt
	isNull  C.sb2
	value   *Rset
}

func (bnd *bndRset) bind(value *Rset, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	// Allocate a statement handle
	ocistmt, err := bnd.stmt.ses.srv.env.allocOciHandle(C.OCI_HTYPE_STMT)
	bnd.ocistmt = (*C.OCIStmt)(ocistmt)
	if err != nil {
		return err
	}
	r := C.OCIBindByPos2(
		stmt.ocistmt,                 //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),   //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,  //OCIError     *errhp,
		C.ub4(position),              //ub4          position,
		unsafe.Pointer(&bnd.ocistmt), //void         *valuep,
		C.sb8(0),                     //sb8          value_sz,
		C.SQLT_RSET,                  //ub2          dty,
		unsafe.Pointer(&bnd.isNull),  //void         *indp,
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

func (bnd *bndRset) setPtr() error {
	err := bnd.value.open(bnd.stmt, bnd.ocistmt)
	bnd.stmt.rsets.PushBack(bnd.value)
	if err == nil {
		// open result set is successful; will be freed by Rset
		bnd.ocistmt = nil
	}

	return err
}

func (bnd *bndRset) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	// release ocistmt handle for failed Rset binding
	// Rset will release handle for successful bind
	if bnd.ocistmt != nil {
		bnd.stmt.ses.srv.env.freeOciHandle(unsafe.Pointer(bnd.ocistmt), C.OCI_HTYPE_STMT)
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ocistmt = nil
	bnd.value = nil
	stmt.putBnd(bndIdxRset, bnd)
return nil
}
