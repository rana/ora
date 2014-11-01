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

// bfileSliceBind is unsupported due to OCILobFileSetName setting a single file name with OCIBindArrayOfStruct call

type bfileBind struct {
	env             *Environment
	ocibnd          *C.OCIBind
	ociLobLocator   *C.OCILobLocator
	cDirectoryAlias *C.char
	cFilename       *C.char
}

func (b *bfileBind) bind(value Bfile, position int, ocistmt *C.OCIStmt) error {
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

	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(b.env.ocienv),                        //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&b.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                    //ub4           type,
		0,                                                   //size_t        xtramem_sz,
		nil)                                                 //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during bind")
	}

	b.cDirectoryAlias = C.CString(value.DirectoryAlias)
	b.cFilename = C.CString(value.Filename)
	r = C.OCILobFileSetName(
		b.env.ocienv,                                    //OCIEnv             *envhp,
		b.env.ocierr,                                    //OCIError           *errhp,
		&b.ociLobLocator,                                //OCILobLocator      **filepp,
		(*C.OraText)(unsafe.Pointer(b.cDirectoryAlias)), //const OraText      *dir_alias,
		C.ub2(C.strlen(b.cDirectoryAlias)),              //ub2                d_length,
		(*C.OraText)(unsafe.Pointer(b.cFilename)),       //const OraText      *filename,
		C.ub2(C.strlen(b.cFilename)))                    //ub2                f_length );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindByPos2(
		ocistmt,                               //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),              //OCIBind      **bindpp,
		b.env.ocierr,                          //OCIError     *errhp,
		C.ub4(position),                       //ub4          position,
		unsafe.Pointer(&b.ociLobLocator),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociLobLocator)), //sb8          value_sz,
		C.SQLT_FILE,                           //ub2          dty,
		nil,                                   //void         *indp,
		nil,                                   //ub2          *alenp,
		nil,                                   //ub2          *rcodep,
		0,                                     //ub4          maxarr_len,
		nil,                                   //ub4          *curelep,
		C.OCI_DEFAULT)                         //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *bfileBind) setPtr() error {
	return nil
}

func (b *bfileBind) close() {
	defer func() {
		recover()
	}()
	// free c strings
	if b.cDirectoryAlias != nil {
		C.free(unsafe.Pointer(b.cDirectoryAlias))
	}
	if b.cFilename != nil {
		C.free(unsafe.Pointer(b.cFilename))
	}
	// free lob locator handle
	if b.ociLobLocator != nil {
		C.OCIDescriptorFree(
			unsafe.Pointer(b.ociLobLocator), //void     *descp,
			C.OCI_DTYPE_FILE)                //ub4      type );
	}
	b.ocibnd = nil
	b.ociLobLocator = nil
	b.cDirectoryAlias = nil
	b.cFilename = nil
	b.env.bfileBindPool.Put(b)
}
