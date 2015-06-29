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
	"unsafe"
)

// no bndBfileSlice: bfileSliceBind is unsupported due to OCILobFileSetName
// setting a single file name with OCIBindArrayOfStruct call

type bndBfile struct {
	stmt            *Stmt
	ocibnd          *C.OCIBind
	ociLobLocator   *C.OCILobLocator
	cDirectoryAlias *C.char
	cFilename       *C.char
}

func (bnd *bndBfile) bind(value Bfile, position int, stmt *Stmt) error {
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
		unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),           //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&bnd.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                      //ub4           type,
		0,                                                     //size_t        xtramem_sz,
		nil)                                                   //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during bind")
	}

	bnd.cDirectoryAlias = C.CString(value.DirectoryAlias)
	bnd.cFilename = C.CString(value.Filename)
	r = C.OCILobFileSetName(
		bnd.stmt.ses.srv.env.ocienv,                       //OCIEnv             *envhp,
		bnd.stmt.ses.srv.env.ocierr,                       //OCIError           *errhp,
		&bnd.ociLobLocator,                                //OCILobLocator      **filepp,
		(*C.OraText)(unsafe.Pointer(bnd.cDirectoryAlias)), //const OraText      *dir_alias,
		C.ub2(len(value.DirectoryAlias)),                  //ub2                d_length,
		(*C.OraText)(unsafe.Pointer(bnd.cFilename)),       //const OraText      *filename,
		C.ub2(len(value.Filename)))                        //ub2                f_length );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                                //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),                      //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,                     //OCIError     *errhp,
		C.ub4(position),                                 //ub4          position,
		unsafe.Pointer(&bnd.ociLobLocator),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocator)), //sb8          value_sz,
		C.SQLT_FILE,   //ub2          dty,
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

func (bnd *bndBfile) setPtr() error {
	return nil
}

func (bnd *bndBfile) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	if bnd.cDirectoryAlias != nil {
		C.free(unsafe.Pointer(bnd.cDirectoryAlias))
	}
	if bnd.cFilename != nil {
		C.free(unsafe.Pointer(bnd.cFilename))
	}
	if bnd.ociLobLocator != nil {
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.ociLobLocator), //void     *descp,
			C.OCI_DTYPE_FILE)                  //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociLobLocator = nil
	bnd.cDirectoryAlias = nil
	bnd.cFilename = nil
	stmt.putBnd(bndIdxBfile, bnd)
	return nil
}
