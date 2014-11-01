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
	"bytes"
	"unsafe"
)

type stringSliceBind struct {
	env    *Environment
	ocibnd *C.OCIBind
	bytes  []byte
	buffer bytes.Buffer
}

func (b *stringSliceBind) bindOraStringSlice(values []String, position int, ocistmt *C.OCIStmt) error {
	stringValues := make([]string, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			stringValues[n] = values[n].Value
		}
	}
	return b.bindStringSlice(stringValues, nullInds, position, ocistmt)
}

func (b *stringSliceBind) bindStringSlice(values []string, nullInds []C.sb2, position int, ocistmt *C.OCIStmt) (err error) {
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	var maxLen int
	for _, str := range values {
		strLen := len(str)
		if strLen > maxLen {
			maxLen = strLen
		}
	}
	for n, str := range values {
		_, err = b.buffer.WriteString(str)
		if err != nil {
			return err
		}
		// pad to make equal to max len if necessary
		padLen := maxLen - len(str)
		for n := 0; n < padLen; n++ {
			_, err = b.buffer.WriteRune('0')
			if err != nil {
				return err
			}
		}
		alenp[n] = C.ub4(len(str))
	}
	b.bytes = b.buffer.Bytes()
	r := C.OCIBindByPos2(
		ocistmt,                      //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),     //OCIBind      **bindpp,
		b.env.ocierr,                 //OCIError     *errhp,
		C.ub4(position),              //ub4          position,
		unsafe.Pointer(&b.bytes[0]),  //void         *valuep,
		C.sb8(maxLen),                //sb8          value_sz,
		C.SQLT_CHR,                   //ub2          dty,
		unsafe.Pointer(&nullInds[0]), //void         *indp,
		&alenp[0],                    //ub4          *alenp,
		&rcodep[0],                   //ub2          *rcodep,
		0,                            //ub4          maxarr_len,
		nil,                          //ub4          *curelep,
		C.OCI_DEFAULT)                //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		b.ocibnd,
		b.env.ocierr,
		C.ub4(maxLen),       //ub4         pvskip,
		C.ub4(C.sizeof_sb2), //ub4         indskip,
		C.ub4(C.sizeof_ub4), //ub4         alskip,
		C.ub4(C.sizeof_ub2)) //ub4         rcskip
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	return nil
}

func (b *stringSliceBind) setPtr() error {
	return nil
}

func (b *stringSliceBind) close() {
	defer func() {
		recover()
	}()
	b.ocibnd = nil
	b.bytes = nil
	b.buffer.Reset()
	b.env.stringSliceBindPool.Put(b)
}
