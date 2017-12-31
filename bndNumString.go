// Copyright 2016 Tamás Gulácsi. All rights reserved.
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
	"sync"
	"unsafe"
)

type bndNumString struct {
	stmt      *Stmt
	ocibnd    *C.OCIBind
	ociNumber [1]C.OCINumber
	fmtBuf    [numStringLen]byte
}

const (
	numberFmt    = "TM9" // Has some compromise: goes scientific over 64 digits (See https://docs.oracle.com/cd/B10501_01/server.920/a96540/sql_elements4a.htm#34597)
	numberFmtLen = len(numberFmt)
	numberNLS    = "NLS_NUMERIC_CHARACTERS='.,'" // DG
	numberNLSLen = len(numberNLS)
	numStringLen = 64
)

var (
	numberFmtC *C.oratext
	numberNLSC *C.oratext
)

func init() {
	numberFmtC = (*C.oratext)(unsafe.Pointer(C.CString(numberFmt)))
	numberNLSC = (*C.oratext)(unsafe.Pointer(C.CString(numberNLS)))
}

// formatFor returns the number format for the value for OCINumberFromText.
func formatFor(buf []byte, num string) []byte {
	if cap(buf) < len(num) {
		buf = make([]byte, len(num))
	} else {
		buf = buf[:len(num)]
	}
	for i, r := range num {
		if r == '.' {
			buf[i] = 'D'
		} else {
			buf[i] = '9'
		}
	}
	//fmt.Printf("formatFor(%q): %q\n", num, buf)
	return buf
}

var fmtBufPool = sync.Pool{New: func() interface{} { z := make([]byte, numStringLen); return &z }}

func (env *Env) numberFromText(dest *C.OCINumber, value string) error {
	buf := (*(fmtBufPool.Get().(*[]byte)))[:len(value)]
	copy(buf, value)
	fmtBuf := formatFor(*(fmtBufPool.Get().(*[]byte)), value)
	//fmt.Printf("buf=%q fmtBuf=%q\n", buf, fmtBuf)
	r := C.OCINumberFromText(
		env.ocierr,                               //OCIError            *err,
		(*C.oratext)(unsafe.Pointer(&buf[0])),    //CONST OraText      *str,
		C.ub4(len(value)),                        //ub4                str_length,
		(*C.oratext)(unsafe.Pointer(&fmtBuf[0])), //CONST OraText *fmt,
		C.ub4(len(fmtBuf)),                       //ub4                fmt_length,
		numberNLSC,                               //CONST OraText      *nls_params,
		C.ub4(numberNLSLen),                      //ub4                nls_p_length,
		dest,                                     //OCINumber          *number );
	)
	fmtBufPool.Put(&buf)
	fmtBufPool.Put(&fmtBuf)
	if r == C.OCI_ERROR {
		err := env.ociError()
		//fmt.Printf("numberFromText(%q [%d]): %v\n", value, len(value), err)
		return err
	}
	return nil
}

func (bnd *bndNumString) bind(value Num, position namedPos, stmt *Stmt) error {
	bnd.stmt = stmt
	cstr := (*C.oratext)(unsafe.Pointer(C.CString(string(value))))
	defer C.free(unsafe.Pointer(cstr))
	if err := bnd.stmt.ses.srv.env.numberFromText(&bnd.ociNumber[0], string(value)); err != nil {
		return err
	}
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}
	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,
		phLen,
		unsafe.Pointer(&bnd.ociNumber[0]), //void         *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber), //sb8          value_sz,
		C.SQLT_VNU,                        //ub2          dty,
		nil,                               //void         *indp,
		nil,                               //ub2          *alenp,
		nil,                               //ub2          *rcodep,
		0,                                 //ub4          maxarr_len,
		nil,                               //ub4          *curelep,
		C.OCI_DEFAULT)                     //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndNumString) setPtr() error {
	return nil
}

func (bnd *bndNumString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	stmt.putBnd(bndIdxNumString, bnd)
	return nil
}
