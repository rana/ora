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

const maxStringLength = 32767

type bndStringPtr struct {
	stmt        *Stmt
	ocibnd      *C.OCIBind
	value       *string
	valueIsNull *bool
	buf         []byte
	alen        [1]C.ACTUAL_LENGTH_TYPE
	nullp
}

func (bnd *bndStringPtr) bind(value *string, valueIsNull *bool, position namedPos, stringPtrBufferSize int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	bnd.valueIsNull = valueIsNull
	if stringPtrBufferSize < 2 {
		stringPtrBufferSize = 2
	} else if stringPtrBufferSize%2 == 1 {
		stringPtrBufferSize++
	}
	L, C := len(bnd.buf), cap(bnd.buf)
	if C < stringPtrBufferSize {
		C = stringPtrBufferSize
	}
	if value != nil {
		lv := len(*value)
		if lv > maxStringLength {
			lv = maxStringLength
			*value = (*value)[:lv]
		}
		if lv > C {
			L, C = lv, lv
		}
	}
	if C%2 == 1 {
		C++
	}
	if cap(bnd.buf) < C {
		//bnd.buf = make([]byte, L, C)
		bnd.buf = bytesPool.Get(C)[:L]
	}
	bnd.nullp.Set(value == nil)
	if value == nil {
		bnd.alen[0] = 0
		bnd.buf = bnd.buf[:2]
	} else {
		if len(*value) == 0 {
			bnd.buf = bnd.buf[:2] // to be able to address bnd.buf[0]
			bnd.buf[0], bnd.buf[1] = 0, 0
		} else {
			L = len(*value)
			if L < 2 {
				L = 2
			} else if L%2 != 0 {
				L++
			}
			bnd.buf = bnd.buf[:L]
			bnd.buf[L-1] = 0
			copy(bnd.buf, []byte(*value))
		}
		bnd.alen[0] = C.ACTUAL_LENGTH_TYPE(len(*value))
	}
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"%p pos=%v cap=%d len=%d alen=%d bufSize=%d", bnd, position, cap(bnd.buf), len(bnd.buf), bnd.alen[0], stringPtrBufferSize)
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
		unsafe.Pointer(&bnd.buf[0]),         //void         *valuep,
		C.LENGTH_TYPE(cap(bnd.buf)),         //sb8          value_sz,
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

func (bnd *bndStringPtr) setPtr() error {
	if bnd.valueIsNull != nil {
		*bnd.valueIsNull = bnd.nullp.IsNull()
	}
	if bnd.value == nil {
		return nil
	}
	bnd.stmt.logF(_drv.Cfg().Log.Stmt.Bind,
		"StringPtr.setPtr isNull=%t alen=%d", bnd.nullp.IsNull(), bnd.alen[0])

	if !bnd.nullp.IsNull() {
		*bnd.value = string(bnd.buf[:bnd.alen[0]])
	} else {
		*bnd.value = ""
	}

	return nil
}

func (bnd *bndStringPtr) close() (err error) {
	/*
		defer func() {
			if value := recover(); value != nil {
				err = errR(value)
			}
		}()
	*/
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.valueIsNull = nil
	bnd.alen[0] = 0
	bytesPool.Put(bnd.buf)
	bnd.buf = nil
	bnd.nullp.Free()
	stmt.putBnd(bndIdxStringPtr, bnd)
	return nil
}
