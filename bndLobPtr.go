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
	stmt   *Stmt
	ocibnd *C.OCIBind
	value  *Lob
	sqlt   C.ub2
	lobLocatorp
}

func (bnd *bndLobPtr) bindLob(lob *Lob, position namedPos, lobBufferSize int, sqlt C.ub2, stmt *Stmt) (err error) {
	bnd.stmt = stmt
	bnd.value = lob
	bnd.sqlt = sqlt
	if lobBufferSize <= 0 {
		lobBufferSize = lobChunkSize
	}

	finish, err := bnd.allocTempLob()
	if err != nil {
		return err
	}

	if lob != nil && lob.Reader != nil {
		if err = writeLob(bnd.lobLocatorp.Value(), bnd.stmt, lob.Reader, lobBufferSize); err != nil {
			bnd.stmt.ses.Break()
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
	//Log.Infof("%s.setPtr()", bnd)
	if bnd.value == nil {
		return nil
	}
	//Log.Infof("setPtr OCILobOpen %p", bnd.ociLobLocator)
	lobLength, err := lobOpen(bnd.stmt.ses, bnd.lobLocatorp.Value(), C.OCI_LOB_READONLY)
	if err != nil {
		lobClose(bnd.stmt.ses, bnd.lobLocatorp.Value())
		return err
	}

	lr := &lobReader{
		ses:           bnd.stmt.ses,
		ociLobLocator: bnd.lobLocatorp.Value(),
		piece:         C.OCI_FIRST_PIECE,
		Length:        lobLength,
	}
	bnd.value.Reader, bnd.value.Closer = lr, lr
	return nil
}

func (bnd *bndLobPtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	// no need to clear bnd.buf
	if lob := bnd.lobLocatorp.Value(); lob != nil {
		// free temporary lob
		C.OCILobFreeTemporary(
			bnd.stmt.ses.ocisvcctx,      //OCISvcCtx          *svchp,
			bnd.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			lob) //OCILobLocator      *locp,
		// free lob locator handle
		C.OCIDescriptorFree(
			unsafe.Pointer(lob), //void     *descp,
			C.OCI_DTYPE_LOB)     //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.value = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxLobPtr, bnd)
	return nil
}

func (bnd *bndLobPtr) allocTempLob() (finish func(), err error) {
	var lob *C.OCILobLocator
	lob, finish, err = allocTempLob(bnd.stmt)
	if err == nil {
		*(bnd.lobLocatorp.Pointer()) = lob
	}
	return
}

func (bnd *bndLobPtr) bindByPos(position namedPos) error {
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
		unsafe.Pointer(bnd.lobLocatorp.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.lobLocatorp.Size()),     //sb8          value_sz,
		bnd.sqlt,      //ub2          dty,
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
