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

type bfileDefine struct {
	environment    *Environment
	ocidef         *C.OCIDefine
	isNull         C.sb2
	ociLobLocator  *C.OCILobLocator
	directoryAlias [30]byte
	filename       [255]byte
}

func (bfileDefine *bfileDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                                    //OCIStmt     *stmtp,
		&bfileDefine.ocidef,                        //OCIDefine   **defnpp,
		bfileDefine.environment.ocierr,             //OCIError    *errhp,
		C.ub4(position),                            //ub4         position,
		unsafe.Pointer(&bfileDefine.ociLobLocator), //void        *valuep,
		C.sb8(columnSize),                          //sb8         value_sz,
		C.SQLT_FILE,                                //ub2         dty,
		unsafe.Pointer(&bfileDefine.isNull),        //void        *indp,
		nil,           //ub4         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return bfileDefine.environment.ociError()
	}
	return nil
}
func (bfileDefine *bfileDefine) value() (value interface{}, err error) {
	var bfileValue Bfile
	bfileValue.IsNull = bfileDefine.isNull < 0
	if !bfileValue.IsNull {
		// Get directory alias and filename
		dLength := C.ub2(len(bfileDefine.directoryAlias))
		fLength := C.ub2(len(bfileDefine.filename))
		r := C.OCILobFileGetName(
			bfileDefine.environment.ocienv,                               //OCIEnv                   *envhp,
			bfileDefine.environment.ocierr,                               //OCIError                 *errhp,
			bfileDefine.ociLobLocator,                                    //const OCILobLocator      *filep,
			(*C.OraText)(unsafe.Pointer(&bfileDefine.directoryAlias[0])), //OraText                  *dir_alias,
			&dLength, //ub2                      *d_length,
			(*C.OraText)(unsafe.Pointer(&bfileDefine.filename[0])), //OraText                  *filename,
			&fLength) //ub2                      *f_length );
		if r == C.OCI_ERROR {
			return value, bfileDefine.environment.ociError()
		}
		bfileValue.DirectoryAlias = string(bfileDefine.directoryAlias[:int(dLength)])
		bfileValue.Filename = string(bfileDefine.filename[:int(fLength)])
	}
	value = bfileValue
	return value, err
}
func (bfileDefine *bfileDefine) alloc() error {
	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(bfileDefine.environment.ocienv),                //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&bfileDefine.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                              //ub4           type,
		0,                                                             //size_t        xtramem_sz,
		nil)                                                           //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return bfileDefine.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}
func (bfileDefine *bfileDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(bfileDefine.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_FILE)                          //ub4      type );
}
func (bfileDefine *bfileDefine) close() {
	defer func() {
		recover()
	}()
	bfileDefine.ocidef = nil
	bfileDefine.ociLobLocator = nil
	bfileDefine.isNull = C.sb2(0)
	for n, _ := range bfileDefine.directoryAlias {
		bfileDefine.directoryAlias[n] = 0
	}
	for n, _ := range bfileDefine.filename {
		bfileDefine.filename[n] = 0
	}
	bfileDefine.environment.bfileDefinePool.Put(bfileDefine)
}
