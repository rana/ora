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
	strings *[]string
	values  *[]String
	maxLen  int
	isOra   bool
	arrHlp
}

func (bnd *bndStringSlice) bindOra(values *[]String, position namedPos, stmt *Stmt, isAssocArray bool) (uint32, error) {
	L, C := len(*values), cap(*values)
	if bnd.strings == nil {
		s := make([]string, L, C)
		bnd.strings = &s
	} else if cap(*bnd.strings) < C {
		*bnd.strings = make([]string, L, C)
	} else {
		*bnd.strings = (*bnd.strings)[:L]
	}
	if cap(bnd.nullInds) < C {
		bnd.nullInds = make([]C.sb2, L, C)
	} else {
		bnd.nullInds = bnd.nullInds[:L]
	}
	bnd.values = values
	for n, v := range *values {
		if v.IsNull {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.nullInds[n] = 0
			(*bnd.strings)[n] = v.Value
		}
	}
	bnd.isOra = true
	return bnd.bind(bnd.strings, position, stmt, isAssocArray)
}

func (bnd *bndStringSlice) bind(values *[]string, position namedPos, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	if values == nil {
		values = &[]string{}
	}
	L, C := len(*values), cap(*values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		*values = append(*values, "")
	}
	if !bnd.isOra {
		for i := range bnd.nullInds {
			bnd.nullInds[i] = 0
		}
	}
	bnd.strings = values
	bnd.maxLen = stmt.Cfg().stringPtrBufferSize
	for _, str := range *values {
		strLen := len(str)
		if strLen > bnd.maxLen {
			bnd.maxLen = strLen
		}
	}
	if cap(bnd.bytes) < bnd.maxLen*C {
		bnd.bytes = bytesPool.Get(bnd.maxLen * C)[:bnd.maxLen*L]
	} else {
		bnd.bytes = bnd.bytes[:bnd.maxLen*L]
	}
	for m, str := range *values {
		copy(bnd.bytes[m*bnd.maxLen:], str)
		bnd.alen[m] = C.ACTUAL_LENGTH_TYPE(len(str))
	}
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v cap=%d len=%d curlen=%d curlenp=%p maxlen=%d iterations=%d alen=%v isAssoc=%t",
		bnd, position, cap(bnd.bytes), len(bnd.bytes), bnd.curlen, curlenp, bnd.maxLen, iterations, bnd.alen, isAssocArray)
	ph, phLen, phFree := position.CString()
	if ph != nil {
		defer phFree()
	}

	r := C.bindByNameOrPos(
		bnd.stmt.ocistmt,            //OCIStmt      *stmtp,
		&bnd.ocibnd,                 //OCIBind      **bindpp,
		bnd.stmt.ses.srv.env.ocierr, //OCIError     *errhp,
		C.ub4(position.Ordinal),     //ub4          position,
		ph,    //const OraText          *placeholder,
		phLen, //sb4          placeholder_length,
		unsafe.Pointer(&bnd.bytes[0]),    //void         *valuep,
		C.LENGTH_TYPE(bnd.maxLen),        //sb8          value_sz,
		C.SQLT_CHR,                       //ub2          dty,
		unsafe.Pointer(&bnd.nullInds[0]), //void         *indp,
		&bnd.alen[0],                     //ub4          *alenp,
		&bnd.rcode[0],                    //ub2          *rcodep,
		getMaxarrLen(C, isAssocArray),    //ub4          maxarr_len,
		curlenp,       //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
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
		bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind, "isAssoc=false")
		return nil
	}
	n := int(bnd.curlen)
	if bnd.strings == nil {
		s := make([]string, n)
		bnd.strings = &s
	}
	*bnd.strings = (*bnd.strings)[:n]
	bnd.nullInds = bnd.nullInds[:n]
	if bnd.values != nil {
		*bnd.values = (*bnd.values)[:n]
	}
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"StringSlice.setPtr n=%d alen=%v nulls=%v bytes=%s", n, bnd.alen, bnd.nullInds, bnd.bytes)
	for i, length := range bnd.alen[:n] {
		if bnd.nullInds[i] <= C.sb2(-1) {
			if bnd.values != nil {
				(*bnd.values)[i].IsNull = true
			}
			continue
		}
		(*bnd.strings)[i] = string(bnd.bytes[i*bnd.maxLen : i*bnd.maxLen+int(length)])
		bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
			"StringSlice.setPtr[%d]=%s", i, (*bnd.strings)[i])
		if bnd.values != nil {
			(*bnd.values)[i].IsNull = false
			(*bnd.values)[i].Value = (*bnd.strings)[i]
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
	bytesPool.Put(bnd.bytes)
	bnd.bytes = nil
	bnd.strings = nil
	bnd.isOra = false
	bnd.values = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxStringSlice, bnd)
	return nil
}
