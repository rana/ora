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
	"strconv"
	"unsafe"
)

type bndBool struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	cString *C.char
}

func (bnd *bndBool) bind(value bool, position namedPos, c StmtCfg, stmt *Stmt) (err error) {
	//Log.Infof("%s.bind(%t, %d)", bnd, value, position)
	bnd.stmt = stmt
	var str string
	if value {
		str, err = strconv.Unquote(strconv.QuoteRune(c.TrueRune))
	} else {
		str, err = strconv.Unquote(strconv.QuoteRune(c.FalseRune))
	}
	if err != nil {
		return err
	}
	bnd.cString = C.CString(str)
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
		C.LENGTH_TYPE(1),            //sb8          value_sz,
		C.SQLT_AFC,                  //ub2          dty,
		nil,                         //void         *indp,
		nil,                         //ub2          *alenp,
		nil,                         //ub2          *rcodep,
		0,                           //ub4          maxarr_len,
		nil,                         //ub4          *curelep,
		C.OCI_DEFAULT)               //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndBool) setPtr() error {
	return nil
}

func (bnd *bndBool) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	C.free(unsafe.Pointer(bnd.cString))
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.cString = nil
	stmt.putBnd(bndIdxBool, bnd)
	return nil
}
