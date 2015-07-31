// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type bndStringSlice struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	bytes   []byte
	strings []string
	values  []String
	maxLen  int
	arrHlp
}

func (bnd *bndStringSlice) bindOra(values []String, position int, stmt *Stmt) (uint32, error) {
	if cap(bnd.strings) < cap(values) {
		bnd.strings = make([]string, len(values), cap(values))
	} else {
		bnd.strings = bnd.strings[:len(values)]
	}
	if cap(bnd.nullInds) < cap(values) {
		bnd.nullInds = make([]C.sb2, len(values), cap(values))
	} else {
		bnd.nullInds = bnd.nullInds[:len(values)]
	}
	for n, _ := range values {
		if values[n].IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = 0
			bnd.strings[n] = values[n].Value
		}
	}
	return bnd.bind(bnd.strings, position, stmt)
}

func (bnd *bndStringSlice) bind(values []string, position int, stmt *Stmt) (iterations uint32, err error) {
	bnd.stmt = stmt
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, stmt.stmtType)
	if needAppend {
		values = append(values, "")
	}
	bnd.strings = values
	bnd.maxLen = stmt.cfg.stringPtrBufferSize
	for _, str := range values {
		strLen := len(str)
		if strLen > bnd.maxLen {
			bnd.maxLen = strLen
		}
	}
	if cap(bnd.bytes) < bnd.maxLen*C {
		bnd.bytes = make([]byte, bnd.maxLen*L, bnd.maxLen*C)
	} else {
		bnd.bytes = bnd.bytes[:bnd.maxLen*L]
	}
	for m, str := range values {
		copy(bnd.bytes[m*bnd.maxLen:], []byte(str))
		bnd.alen[m] = C.ACTUAL_LENGTH_TYPE(len(str))
	}
	bnd.stmt.logF(_drv.cfg.Log.Stmt.Bind,
		"%p pos=%d cap=%d len=%d curlen=%d curlenp=%p maxlen=%d iterations=%d alen=%v",
		bnd, position, cap(bnd.bytes), len(bnd.bytes), bnd.curlen, curlenp, bnd.maxLen, iterations, bnd.alen)
	r := C.OCIBINDBYPOS(
		bnd.stmt.ocistmt,                 //OCIStmt      *stmtp,
		(**C.OCIBind)(&bnd.ocibnd),       //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr,      //OCIError     *errhp,
		C.ub4(position),                  //ub4          position,
		unsafe.Pointer(&bnd.bytes[0]),    //void         *valuep,
		C.LENGTH_TYPE(bnd.maxLen),        //sb8          value_sz,
		C.SQLT_CHR,                       //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]), //void         *indp,
		&bnd.alen[0],                     //ub4          *alenp,
		&bnd.rcode[0],                    //ub2          *rcodep,
		C.ACTUAL_LENGTH_TYPE(C),          //ub4          maxarr_len,
		curlenp,                          //ub4          *curelep,
		C.OCI_DEFAULT)                    //ub4          mode );
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	r = C.OCIBindArrayOfStruct(
		bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,
		C.ub4(bnd.maxLen),                  //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                //ub4         indskip,
		C.ub4(C.sizeof_ACTUAL_LENGTH_TYPE), //ub4         alskip,
		C.ub4(C.sizeof_ub2))                //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}
	return iterations, nil
}

func (bnd *bndStringSlice) setPtr() error {
	if !bnd.IsAssocArr() {
		return nil
	}
	n := int(bnd.curlen)
	bnd.strings = bnd.strings[:n]
	for i, length := range bnd.alen[:n] {
		if bnd.nullInds[i] > C.sb2(-1) {
			bnd.strings[i] = string(bnd.bytes[i*bnd.maxLen : i*bnd.maxLen+int(length)])
			if bnd.values != nil {
				bnd.values[i].IsNull = false
				bnd.values[i].Value = bnd.strings[i]
			}
		} else if bnd.values != nil {
			bnd.values[i].IsNull = true
		}
	}
	return nil
}

func (bnd *bndStringSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.values = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxStringSlice, bnd)
	return nil
}
