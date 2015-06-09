// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndLobPtr struct {
	stmt          *Stmt
	ocibnd        *C.OCIBind
	ociLobLocator *C.OCILobLocator
	value         *Lob
}

func (bnd *bndLobPtr) bindLob(lob *Lob, position int, lobBufferSize int, stmt *Stmt) (err error) {
	bnd.stmt = stmt
	bnd.value = lob
	if lobBufferSize <= 0 {
		lobBufferSize = lobChunkSize
	}

	finish, err := bnd.allocTempLob()
	if err != nil {
		return err
	}

	if lob != nil && lob.Reader != nil {
		if err = writeLob(bnd.ociLobLocator, bnd.stmt, lob.Reader, lobBufferSize); err != nil {
			bnd.stmt.ses.srv.Break()
			finish()
			return err
		}
	}

	if err = bnd.bindByPos(position); err != nil {
		finish()
		return err
	}
	return nil
}

func (bnd *bndLobPtr) setPtr() error {
	Log.Infof("%s.setPtr()", bnd)
	if bnd.value == nil {
		return nil
	}
	Log.Infof("setPtr OCILobOpen %p", bnd.ociLobLocator)
	lobLength, err := lobOpen(bnd.stmt.ses.srv, bnd.ociLobLocator, C.OCI_LOB_READONLY)
	if err != nil {
		lobClose(bnd.stmt.ses.srv, bnd.ociLobLocator)
		bnd.ociLobLocator = nil
		return err
	}

	lr := &lobReader{
		srv:           bnd.stmt.ses.srv,
		ociLobLocator: bnd.ociLobLocator,
		piece:         C.OCI_FIRST_PIECE,
		Length:        lobLength,
	}
	bnd.value.Reader, bnd.value.Closer = lr, lr
	bnd.ociLobLocator = nil
	return nil
}

func (bnd *bndLobPtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	// no need to clear bnd.buf
	// free temporary lob
	C.OCILobFreeTemporary(
		bnd.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		bnd.ociLobLocator)           //OCILobLocator      *locp,
	// free lob locator handle
	C.OCIDescriptorFree(
		unsafe.Pointer(bnd.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                   //ub4      type );
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.value = nil
	bnd.ocibnd = nil
	bnd.ociLobLocator = nil
	stmt.putBnd(bndIdxLob, bnd)
	return nil
}

func (bnd *bndLobPtr) allocTempLob() (finish func(), err error) {
	bnd.ociLobLocator, finish, err = allocTempLob(bnd.stmt)
	return
}

func (bnd *bndLobPtr) bindByPos(position int) error {
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                                //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                      //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                     //OCIError     *errhp,
		C.ub4(position),                                 //ub4          position,
		unsafe.Pointer(&bnd.ociLobLocator),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocator)), //sb8          value_sz,
		C.SQLT_BLOB,   //ub2          dty,
		nil,           //void         *indp,
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
