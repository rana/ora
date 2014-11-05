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

type bndFloat32Ptr struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber C.OCINumber
	isNull    C.sb2
	value     *float32
}

func (bnd *bndFloat32Ptr) bind(value *float32, position int, stmt *Stmt) error {
	glog.Infoln("position: ", position)
	bnd.stmt = stmt
	bnd.value = value
	r := C.OCIBindByPos2(
		bnd.stmt.ocistmt,               //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),     //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,    //OCIError     *errhp,
		C.ub4(position),                //ub4          position,
		unsafe.Pointer(&bnd.ociNumber), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),      //sb8          value_sz,
		C.SQLT_VNU,                     //ub2          dty,
		unsafe.Pointer(&bnd.isNull),    //void         *indp,
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

func (bnd *bndFloat32Ptr) setPtr() error {
	if bnd.isNull > -1 {
		r := C.OCINumberToReal(
			bnd.stmt.ses.srv.env.ocierr, //OCIError              *err,
			&bnd.ociNumber,              //const OCINumber     *number,
			C.uword(4),                  //uword               rsl_length,
			unsafe.Pointer(bnd.value))   //void                *rsl );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
	}
	return nil
}

func (bnd *bndFloat32Ptr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	stmt.putBnd(bndIdxFloat32Ptr, bnd)
	return nil
}
