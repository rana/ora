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
	env            *Environment
	ocidef         *C.OCIDefine
	isNull         C.sb2
	ociLobLocator  *C.OCILobLocator
	directoryAlias [30]byte
	filename       [255]byte
}

func (d *bfileDefine) define(columnSize int, position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                          //OCIStmt     *stmtp,
		&d.ocidef,                        //OCIDefine   **defnpp,
		d.env.ocierr,                     //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&d.ociLobLocator), //void        *valuep,
		C.sb8(columnSize),                //sb8         value_sz,
		C.SQLT_FILE,                      //ub2         dty,
		unsafe.Pointer(&d.isNull),        //void        *indp,
		nil,           //ub4         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *bfileDefine) value() (value interface{}, err error) {
	var bfileValue Bfile
	bfileValue.IsNull = d.isNull < 0
	if !bfileValue.IsNull {
		// Get directory alias and filename
		dLength := C.ub2(len(d.directoryAlias))
		fLength := C.ub2(len(d.filename))
		r := C.OCILobFileGetName(
			d.env.ocienv,                                       //OCIEnv                   *envhp,
			d.env.ocierr,                                       //OCIError                 *errhp,
			d.ociLobLocator,                                    //const OCILobLocator      *filep,
			(*C.OraText)(unsafe.Pointer(&d.directoryAlias[0])), //OraText                  *dir_alias,
			&dLength, //ub2                      *d_length,
			(*C.OraText)(unsafe.Pointer(&d.filename[0])), //OraText                  *filename,
			&fLength) //ub2                      *f_length );
		if r == C.OCI_ERROR {
			return value, d.env.ociError()
		}
		bfileValue.DirectoryAlias = string(d.directoryAlias[:int(dLength)])
		bfileValue.Filename = string(d.filename[:int(fLength)])
	}
	value = bfileValue
	return value, err
}
func (d *bfileDefine) alloc() error {
	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                        //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                    //ub4           type,
		0,                                                   //size_t        xtramem_sz,
		nil)                                                 //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}
func (d *bfileDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_FILE)                //ub4      type );
}
func (d *bfileDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ociLobLocator = nil
	d.isNull = C.sb2(0)
	for n, _ := range d.directoryAlias {
		d.directoryAlias[n] = 0
	}
	for n, _ := range d.filename {
		d.filename[n] = 0
	}
	d.env.bfileDefinePool.Put(d)
}
