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
	"strconv"
	"unsafe"
)

type boolBind struct {
	environment *Environment
	ocibnd      *C.OCIBind
	cstringp    *C.char
}

func (boolBind *boolBind) bind(value bool, position int, config StatementConfig, ocistmt *C.OCIStmt) (err error) {
	var str string
	if value {
		str, err = strconv.Unquote(strconv.QuoteRune(config.TrueRune))
	} else {
		str, err = strconv.Unquote(strconv.QuoteRune(config.FalseRune))
	}
	if err != nil {
		return err
	}
	boolBind.cstringp = C.CString(str)
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&boolBind.ocibnd),   //OCIBind      **bindpp,
		boolBind.environment.ocierr,       //OCIError     *errhp,
		C.ub4(position),                   //ub4          position,
		unsafe.Pointer(boolBind.cstringp), //void         *valuep,
		C.sb8(1),      //sb8          value_sz,
		C.SQLT_CHR,    //ub2          dty,
		nil,           //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return boolBind.environment.ociError()
	}
	return nil
}

func (boolBind *boolBind) setPtr() error {
	return nil
}

func (boolBind *boolBind) close() {
	defer func() {
		recover()
	}()
	// free c string memory
	C.free(unsafe.Pointer(boolBind.cstringp))
	boolBind.ocibnd = nil
	boolBind.cstringp = nil
	boolBind.environment.boolBindPool.Put(boolBind)
}
