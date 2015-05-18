// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"unsafe"
)

type defBfile struct {
	rset           *Rset
	ocidef         *C.OCIDefine
	null           C.sb2
	ociLobLocator  *C.OCILobLocator
	directoryAlias [30]byte
	filename       [255]byte
}

func (def *defBfile) define(position int, rset *Rset) error {
	def.rset = rset
	r := C.OCIDefineByPos2(
		def.rset.ocistmt,                   //OCIStmt     *stmtp,
		&def.ocidef,                             //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,        //OCIError    *errhp,
		C.ub4(position),                         //ub4         position,
		unsafe.Pointer(&def.ociLobLocator),      //void        *valuep,
		C.sb8(unsafe.Sizeof(def.ociLobLocator)), //sb8         value_sz,
		C.SQLT_FILE,                             //ub2         dty,
		unsafe.Pointer(&def.null),               //void        *indp,
		nil,           //ub4         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}
func (def *defBfile) value() (value interface{}, err error) {
	var bfileValue Bfile
	bfileValue.IsNull = def.null < 0
	if !bfileValue.IsNull {
		// Get directory alias and filename
		dLength := C.ub2(len(def.directoryAlias))
		fLength := C.ub2(len(def.filename))
		r := C.OCILobFileGetName(
			def.rset.stmt.ses.srv.env.ocienv,                     //OCIEnv                   *envhp,
			def.rset.stmt.ses.srv.env.ocierr,                     //OCIError                 *errhp,
			def.ociLobLocator,                                    //const OCILobLocator      *filep,
			(*C.OraText)(unsafe.Pointer(&def.directoryAlias[0])), //OraText                  *dir_alias,
			&dLength, //ub2                      *d_length,
			(*C.OraText)(unsafe.Pointer(&def.filename[0])), //OraText                  *filename,
			&fLength) //ub2                      *f_length );
		if r == C.OCI_ERROR {
			return value, def.rset.stmt.ses.srv.env.ociError()
		}
		bfileValue.DirectoryAlias = string(def.directoryAlias[:int(dLength)])
		bfileValue.Filename = string(def.filename[:int(fLength)])
	}
	value = bfileValue
	return value, err
}

func (def *defBfile) alloc() error {
	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&def.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_FILE,                                      //ub4           type,
		0,                                                     //size_t        xtramem_sz,
		nil)                                                   //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}

func (def *defBfile) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(def.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_FILE)                  //ub4      type );
}

func (def *defBfile) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	def.free()
	for n := range def.directoryAlias {
		def.directoryAlias[n] = 0
	}
	for n := range def.filename {
		def.filename[n] = 0
	}
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociLobLocator = nil
	rset.putDef(defIdxBfile, def)
return nil
}
