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

type defBfile struct {
	ociDef
	directoryAlias [30]byte
	filename       [255]byte
	lobs           []*C.OCILobLocator
}

func (def *defBfile) define(position int, rset *Rset) error {
	def.rset = rset
	if def.lobs != nil {
		C.free(unsafe.Pointer(&def.lobs[0]))
	}
	def.lobs = (*((*[MaxFetchLen]*C.OCILobLocator)(C.malloc(C.size_t(rset.fetchLen) * C.sof_LobLocatorp))))[:rset.fetchLen]
	def.ensureAllocatedLength(len(def.lobs))
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.lobs[0]), int(C.sof_LobLocatorp), C.SQLT_FILE)
}
func (def *defBfile) value(offset int) (value interface{}, err error) {
	var bfileValue Bfile
	bfileValue.IsNull = def.nullInds[offset] < 0
	if bfileValue.IsNull {
		return bfileValue, nil
	}
	// Get directory alias and filename
	dLength := C.ub2(len(def.directoryAlias))
	fLength := C.ub2(len(def.filename))
	r := C.OCILobFileGetName(
		def.rset.stmt.ses.srv.env.ocienv,                     //OCIEnv                   *envhp,
		def.rset.stmt.ses.srv.env.ocierr,                     //OCIError                 *errhp,
		def.lobs[offset],                                     //const OCILobLocator      *filep,
		(*C.OraText)(unsafe.Pointer(&def.directoryAlias[0])), //OraText                  *dir_alias,
		&dLength, //ub2                      *d_length,
		(*C.OraText)(unsafe.Pointer(&def.filename[0])), //OraText                  *filename,
		&fLength) //ub2                      *f_length );
	if r == C.OCI_ERROR {
		return value, def.rset.stmt.ses.srv.env.ociError()
	}
	bfileValue.DirectoryAlias = string(def.directoryAlias[:int(dLength)])
	bfileValue.Filename = string(def.filename[:int(fLength)])
	return bfileValue, err
}

func (def *defBfile) alloc() error {
	// Allocate lob locator handle
	// For a LOB define, the buffer pointer must be a pointer to a LOB locator of type OCILobLocator, allocated by the OCIDescriptorAlloc() call.
	for i := range def.lobs {
		def.allocated[i] = false
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&def.lobs[i])),  //dvoid         **descpp,
			C.OCI_DTYPE_FILE,                                 //ub4           type,
			0,                                                //size_t        xtramem_sz,
			nil)                                              //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return def.rset.stmt.ses.srv.env.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci lob handle during define")
		}
		def.allocated[i] = true
	}
	return nil
}

func (def *defBfile) free() {
	for i, lob := range def.lobs {
		if lob == nil {
			continue
		}
		def.lobs[i] = nil
		if !def.allocated[i] {
			continue
		}
		C.OCIDescriptorFree(
			unsafe.Pointer(lob), //void     *descp,
			C.OCI_DTYPE_FILE)    //ub4      type );
	}
	def.arrHlp.close()
}

func (def *defBfile) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	def.free()
	for i := range def.directoryAlias[:cap(def.directoryAlias)] {
		def.directoryAlias[i] = 0
	}
	for i := range def.filename[:cap(def.filename)] {
		def.filename[i] = 0
	}
	def.free()
	if def.lobs != nil {
		C.free(unsafe.Pointer(&def.lobs[0]))
		def.lobs = nil
	}
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxBfile, def)
	return nil
}
