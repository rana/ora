// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"io"
	"unsafe"
)

type bndLobSlice struct {
	stmt           *Stmt
	ocibnd         *C.OCIBind
	ociLobLocators []*C.OCILobLocator
	buf            []byte
}

func (bnd *bndLobSlice) bindOra(values []Lob, position int, lobBufferSize int, stmt *Stmt) error {
	binValues := make([]io.Reader, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].Reader == nil {
			nullInds[n] = C.sb2(-1)
		} else {
			binValues[n] = values[n].Reader
		}
	}
	return bnd.bindReaders(binValues, nullInds, position, lobBufferSize, stmt)
}

func (bnd *bndLobSlice) bindReaders(
	values []io.Reader,
	nullInds []C.sb2,
	position int,
	lobBufferSize int,
	stmt *Stmt,
) (
	err error,
) {
	bnd.stmt = stmt
	bnd.ociLobLocators = make([]*C.OCILobLocator, len(values))
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ACTUAL_LENGTH_TYPE, len(values))
	rcodep := make([]C.ub2, len(values))
	if len(bnd.buf) < lobBufferSize {
		bnd.buf = make([]byte, lobBufferSize)
	}
	finishers := make([]func(), len(values))
	defer func() {
		if err == nil {
			return
		}
		for _, finish := range finishers {
			if finish == nil {
				continue
			}
			finish()
		}
	}()

	for i, r := range values {
		bnd.ociLobLocators[i], finishers[i], err = allocTempLob(bnd.stmt)
		if err != nil {
			return err
		}

		alenp[i] = C.ACTUAL_LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocators[i]))
		if nullInds[i] <= C.sb2(-1) {
			continue
		}
		if err = writeLob(bnd.ociLobLocators[i], bnd.stmt, r, lobBufferSize); err != nil {
			bnd.stmt.ses.srv.Break()
			return err
		}
	}

	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                                    //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                          //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                         //OCIError     *errhp,
		C.ub4(position),                                     //ub4          position,
		unsafe.Pointer(&bnd.ociLobLocators[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocators[0])), //sb8          value_sz,
		C.SQLT_BLOB,                  //ub2          dty,
		unsafe.Pointer(&nullInds[0]), //void         *indp,
		&alenp[0],                    //ub4          *alenp,
		&rcodep[0],                   //ub2          *rcodep,
		0,                            //ub4          maxarr_len,
		nil,                          //ub4          *curelep,
		C.OCI_DEFAULT)                //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}

	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(unsafe.Sizeof(bnd.ociLobLocators[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                         //ub4         indskip,
		C.ub4(C.sizeof_ub4),                         //ub4         alskip,
		C.ub4(C.sizeof_ub2))                         //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}

	return nil
}

func (bnd *bndLobSlice) setPtr() error {
	return nil
}

func (bnd *bndLobSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	for n := 0; n < len(bnd.ociLobLocators); n++ {
		// free temporary lob
		C.OCILobFreeTemporary(
			bnd.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
			bnd.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			bnd.ociLobLocators[n])       //OCILobLocator      *locp,
		// free lob locator handle
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.ociLobLocators[n]), //void     *descp,
			C.OCI_DTYPE_LOB)                       //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociLobLocators = nil
	stmt.putBnd(bndIdxBinSlice, bnd)
	return nil
}
