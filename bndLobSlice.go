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
	"io"
	"unsafe"
)

type bndLobSlice struct {
	stmt           *Stmt
	ocibnd         *C.OCIBind
	ociLobLocators []*C.OCILobLocator
	buf            []byte
	readers        []io.Reader
	arrHlp
}

func (bnd *bndLobSlice) bindOra(values []Lob, position namedPos, lobBufferSize int, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	L, C := len(values), cap(values)
	if cap(bnd.readers) < C {
		bnd.readers = make([]io.Reader, L, C)
	} else {
		bnd.readers = bnd.readers[:L]
	}
	if cap(bnd.nullInds) < C {
		bnd.nullInds = make([]C.sb2, L, C)
	} else {
		bnd.nullInds = bnd.nullInds[:L]
	}
	for n := range values {
		if values[n].Reader == nil {
			bnd.nullInds[n] = C.sb2(-1)
		} else {
			bnd.readers[n] = values[n].Reader
		}
	}
	return bnd.bindReaders(bnd.readers, position, lobBufferSize, stmt, isAssocArray)
}

func (bnd *bndLobSlice) bindReaders(values []io.Reader, position namedPos, lobBufferSize int, stmt *Stmt, isAssocArray bool) (iterations uint32, err error) {
	bnd.stmt = stmt
	// ensure we have at least 1 slot in the slice
	L, C := len(values), cap(values)
	iterations, curlenp, needAppend := bnd.ensureBindArrLength(&L, &C, isAssocArray)
	if needAppend {
		values = append(values, nil)
	}
	bnd.readers = values
	if cap(bnd.ociLobLocators) < C {
		bnd.ociLobLocators = make([]*C.OCILobLocator, L, C)
	} else {
		bnd.ociLobLocators = bnd.ociLobLocators[:L]
	}
	if len(bnd.buf) < lobBufferSize {
		//bnd.buf = make([]byte, lobBufferSize)
		bnd.buf = bytesPool.Get(lobBufferSize)
	}
	finishers := make([]func(), len(values))
	defer func() {
		if err == nil {
			return
		}
		for _, finish := range finishers {
			if finish == nil {
				continue
			}
			finish()
		}
	}()

	for i, r := range values {
		bnd.ociLobLocators[i], finishers[i], err = allocTempLob(bnd.stmt)
		if err != nil {
			return iterations, err
		}

		bnd.alen[i] = C.ACTUAL_LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocators[i]))
		if bnd.nullInds[i] <= C.sb2(-1) {
			continue
		}
		if err = writeLob(bnd.ociLobLocators[i], bnd.stmt, r, lobBufferSize); err != nil {
			bnd.stmt.ses.Break()
			return iterations, err
		}
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
		unsafe.Pointer(&bnd.ociLobLocators[0]),              //void         *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(bnd.ociLobLocators[0])), //sb8          value_sz,
		C.SQLT_BLOB,                      //ub2          dty,
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
		C.ub4(unsafe.Sizeof(bnd.ociLobLocators[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                         //ub4         indskip,
		C.ub4(C.sizeof_ub4),                         //ub4         alskip,
		C.ub4(C.sizeof_ub2))                         //ub4         rcskip
	if r == C.OCI_ERROR {
		return iterations, bnd.stmt.ses.srv.env.ociError()
	}

	return iterations, nil
}

func (bnd *bndLobSlice) setPtr() error {
	return nil
}

func (bnd *bndLobSlice) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	for n := 0; n < len(bnd.ociLobLocators); n++ {
		// free temporary lob
		C.OCILobFreeTemporary(
			bnd.stmt.ses.ocisvcctx,      //OCISvcCtx          *svchp,
			bnd.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			bnd.ociLobLocators[n])       //OCILobLocator      *locp,
		// free lob locator handle
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.ociLobLocators[n]), //void     *descp,
			C.OCI_DTYPE_LOB)                       //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.ociLobLocators = nil
	bytesPool.Put(bnd.buf)
	bnd.buf = nil
	bnd.readers = nil
	bnd.arrHlp.close()
	stmt.putBnd(bndIdxBinSlice, bnd)
	return nil
}
