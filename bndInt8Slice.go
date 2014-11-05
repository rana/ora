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

type bndInt8Slice struct {
	stmt       *Stmt
	ocibnd     *C.OCIBind
	ociNumbers []C.OCINumber
}

func (bnd *bndInt8Slice) bindOra(values []Int8, position int, stmt *Stmt) error {
	int8Values := make([]int8, len(values))
	nullInds := make([]C.sb2, len(values))
	for n := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			int8Values[n] = values[n].Value
		}
	}
	return bnd.bind(int8Values, nullInds, position, stmt)
}

func (bnd *bndInt8Slice) bind(values []int8, nullInds []C.sb2, position int, stmt *Stmt) error {
	glog.Infoln("position: ", position)
	bnd.stmt = stmt
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	bnd.ociNumbers = make([]C.OCINumber, len(values))
	for n := range values {
		alenp[n] = C.ub4(C.sizeof_OCINumber)
		r := C.OCINumberFromInt(
			bnd.stmt.ses.srv.env.ocierr, //OCIError            *err,
			unsafe.Pointer(&values[n]),  //const void          *inum,
			1,                   //uword               inum_length,
			C.OCI_NUMBER_SIGNED, //uword               inum_s_flag,
			&bnd.ociNumbers[n])  //OCINumber           *number );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
	}
	r := C.OCIBindByPos2(
		bnd.stmt.ocistmt,                   //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),         //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,        //OCIError     *errhp,
		C.ub4(position),                    //ub4          position,
		unsafe.Pointer(&bnd.ociNumbers[0]), //void         *valuep,
		C.sb8(C.sizeof_OCINumber),          //sb8          value_sz,
		C.SQLT_VNU,                         //ub2          dty,
		unsafe.Pointer(&nullInds[0]),       //void         *indp,
		&alenp[0],                          //ub4          *alenp,
		&rcodep[0],                         //ub2          *rcodep,
		0,                                  //ub4          maxarr_len,
		nil,                                //ub4          *curelep,
		C.OCI_DEFAULT)                      //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(C.sizeof_OCINumber), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),       //ub4         indskip,
		C.ub4(C.sizeof_ub4),       //ub4         alskip,
		C.ub4(C.sizeof_ub2))       //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndInt8Slice) setPtr() error {
	return nil
}

func (bnd *bndInt8Slice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	glog.Infoln("close")
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociNumbers = nil
	stmt.putBnd(bndIdxInt8Slice, bnd)
	return nil
}
