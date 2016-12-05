// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

// no bndBfileSlice: bfileSliceBind is unsupported due to OCILobFileSetName
// setting a single file name with OCIBindArrayOfStruct call

type bndBfile struct {
	stmt            *Stmt
	ocibnd          *C.OCIBind
	cDirectoryAlias *C.char
	cFilename       *C.char
	lobLocatorp
}

func (bnd *bndBfile) bind(value Bfile, position namedPos, stmt *Stmt) error {
	// DirectoryAlias must be specified to avoid error "ORA-24801: illegal parameter value in OCI lob function"
	// Raising a driver error clarifies the user error
	if value.DirectoryAlias == "" {
		return errNew("DirectoryAlias must be specified when binding a non-null Bfile")
	}
	// Filename must be specified to avoid error "ORA-24801: illegal parameter value in OCI lob function"
	// Raising a driver error clarifies the user error
	if value.Filename == "" {
		return errNew("Filename must be specified when binding a non-null Bfile")
	}

	bnd.stmt = stmt
	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),                  //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(bnd.lobLocatorp.Pointer())), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                             //ub4           type,
		0,                                                            //size_t        xtramem_sz,
		nil)                                                          //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during bind")
	}

	bnd.cDirectoryAlias = C.CString(value.DirectoryAlias)
	defer C.free(unsafe.Pointer(bnd.cDirectoryAlias))
	bnd.cFilename = C.CString(value.Filename)
	defer C.free(unsafe.Pointer(bnd.cFilename))
	r = C.OCILobFileSetName(
		bnd.stmt.ses.srv.env.ocienv,                       //OCIEnv             *envhp,
		bnd.stmt.ses.srv.env.ocierr,                       //OCIError           *errhp,
		bnd.lobLocatorp.Pointer(),                         //OCILobLocator      **filepp,
		(*C.OraText)(unsafe.Pointer(bnd.cDirectoryAlias)), //const OraText      *dir_alias,
		C.ub2(len(value.DirectoryAlias)),                  //ub2                d_length,
		(*C.OraText)(unsafe.Pointer(bnd.cFilename)),       //const OraText      *filename,
		C.ub2(len(value.Filename)))                        //ub2                f_length );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}

	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r = C.bindByNameOrPos(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		&bnd.ocibnd,                 //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(bnd.lobLocatorp.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.lobLocatorp.Size()),     //sb8          value_sz,
		C.SQLT_FILE,                               //ub2          dty,
		nil,                                       //void         *indp,
		nil,                                       //ub2          *alenp,
		nil,                                       //ub2          *rcodep,
		0,                                         //ub4          maxarr_len,
		nil,                                       //ub4          *curelep,
		C.OCI_DEFAULT)                             //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndBfile) setPtr() error {
	return nil
}

func (bnd *bndBfile) alloc() {
}

func (bnd *bndBfile) free() {
	bnd.lobLocatorp.Free()
}

func (bnd *bndBfile) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	if lob := bnd.lobLocatorp.Value(); lob != nil {
		C.OCIDescriptorFree(
			unsafe.Pointer(lob), //void     *descp,
			C.OCI_DTYPE_FILE)    //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxBfile, bnd)
	return nil
}
