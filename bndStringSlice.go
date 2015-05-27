// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"bytes"
	"unsafe"
)

type bndStringSlice struct {
	stmt   *Stmt
	ocibnd *C.OCIBind
	bytes  []byte
	buf    bytes.Buffer
}

func (bnd *bndStringSlice) bindOra(values []String, position int, stmt *Stmt) error {
	stringValues := make([]string, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			stringValues[n] = values[n].Value
		}
	}
	return bnd.bind(stringValues, nullInds, position, stmt)
}

func (bnd *bndStringSlice) bind(values []string, nullInds []C.sb2, position int, stmt *Stmt) (err error) {
	bnd.stmt = stmt
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ACTUAL_LENGTH_TYPE, len(values))
	rcodep := make([]C.ub2, len(values))
	var maxLen int
	for _, str := range values {
		strLen := len(str)
		if strLen > maxLen {
			maxLen = strLen
		}
	}
	for n, str := range values {
		_, err = bnd.buf.WriteString(str)
		if err != nil {
			return err
		}
		// pad to make equal to max len if necessary
		padLen := maxLen - len(str)
		for n := 0; n < padLen; n++ {
			_, err = bnd.buf.WriteRune('0')
			if err != nil {
				return err
			}
		}
		alenp[n] = C.ACTUAL_LENGTH_TYPE(len(str))
	}
	bnd.bytes = bnd.buf.Bytes()
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,              //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),    //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,   //OCIError     *errhp,
		C.ub4(position),               //ub4          position,
		unsafe.Pointer(&bnd.bytes[0]), //void         *valuep,
		C.LENGTH_TYPE(maxLen),         //sb8          value_sz,
		C.SQLT_CHR,                    //ub2          dty,
		unsafe.Pointer(&nullInds[0]),  //void         *indp,
		&alenp[0],                     //ub4          *alenp,
		&rcodep[0],                    //ub2          *rcodep,
		0,                             //ub4          maxarr_len,
		nil,                           //ub4          *curelep,
		C.OCI_DEFAULT)                 //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(maxLen),       //ub4         pvskip,
		C.ub4(C.sizeof_sb2), //ub4         indskip,
		C.ub4(C.sizeof_ub4), //ub4         alskip,
		C.ub4(C.sizeof_ub2)) //ub4         rcskip
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndStringSlice) setPtr() error {
	return nil
}

func (bnd *bndStringSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.bytes = nil
	bnd.buf.Reset()
	stmt.putBnd(bndIdxStringSlice, bnd)
	return nil
}
