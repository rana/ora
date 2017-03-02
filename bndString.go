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

type bndString struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	cString *C.char
	alen    [1]C.ACTUAL_LENGTH_TYPE
	nullp
}

// https://ellebaek.wordpress.com/2011/02/25/oracle-type-code-mappings/

func (bnd *bndString) bind(value string, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.cString = C.CString(value)
	bnd.alen[0] = C.ACTUAL_LENGTH_TYPE(len(value))
	bnd.nullp.Set(value == "")
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v alen=%d",
		bnd, position, bnd.alen[0])

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
		unsafe.Pointer(bnd.cString), //void         *valuep,
		C.LENGTH_TYPE(len(value)),   //sb8          value_sz,
		// http://www.devsuperpage.com/search/Articles.aspx?G=4&ArtID=560386
		// "You may find that trailing spaces are truncated when you use SQLT_CHR or SQLT_STR."
		C.SQLT_CHR,                          //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()), //void         *indp,
		&bnd.alen[0],                        //ub2          *alenp,
		nil,                                 //ub2          *rcodep,
		0,                                   //ub4          maxarr_len,
		nil,                                 //ub4          *curelep,
		C.OCI_DEFAULT)                       //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndString) setPtr() error {
	return nil
}

func (bnd *bndString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()
	if bnd.cString != nil {
		C.free(unsafe.Pointer(bnd.cString))
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.cString = nil
	stmt.putBnd(bndIdxString, bnd)
	return nil
}
