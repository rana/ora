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
	"unsafe"
)

type bytesBind struct {
	env           *Environment
	ocibnd        *C.OCIBind
	ocisvcctx     *C.OCISvcCtx
	ociLobLocator *C.OCILobLocator
	buf           []byte
}

func (b *bytesBind) bind(value []byte, position int, lobBufferSize int, ocisvcctx *C.OCISvcCtx, ocistmt *C.OCIStmt) error {
	//fmt.Printf("bytesBind.bind \n")
	// OCILobWrite2 doesn't support writing zero bytes
	// nor is writing 1 byte and erasing the one byte supported
	// therefore, throw an error
	if len(value) == 0 {
		return errNew("writing a zero-length BLOB is unsupported")
	}
	b.ocisvcctx = ocisvcctx
	if len(b.buf) < lobBufferSize {
		b.buf = make([]byte, lobBufferSize)
	}

	// Allocate lob locator handle
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(b.env.ocienv),                        //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&b.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_LOB,                                     //ub4           type,
		0,                                                   //size_t        xtramem_sz,
		nil)                                                 //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during bind")
	}
	// Create temporary lob
	r = C.OCILobCreateTemporary(
		ocisvcctx,              //OCISvcCtx          *svchp,
		b.env.ocierr,           //OCIError           *errhp,
		b.ociLobLocator,        //OCILobLocator      *locp,
		C.OCI_DEFAULT,          //ub2                csid,
		C.SQLCS_IMPLICIT,       //ub1                csfrm,
		C.OCI_TEMP_BLOB,        //ub1                lobtype,
		C.FALSE,                //boolean            cache,
		C.OCI_DURATION_SESSION) //OCIDuration        duration);
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}
	// write bytes to lob locator
	var currentBytesToWrite int
	var remainingBytesToWrite int = len(value)
	var readIndex int
	var byte_amtp C.oraub8 /* Setting Amount to 0 streams the data until use specifies OCI_LAST_PIECE */
	var piece C.ub1 = C.OCI_FIRST_PIECE
	var writing bool = true
	for writing {
		// Copy bytes from slice to buffer
		if remainingBytesToWrite < len(b.buf) {
			currentBytesToWrite = remainingBytesToWrite
		} else {
			currentBytesToWrite = len(b.buf)
		}
		for n := 0; n < currentBytesToWrite; n++ {
			b.buf[n] = value[readIndex]
			readIndex++
		}
		remainingBytesToWrite = len(value) - readIndex

		// Write to Oracle
		r = C.OCILobWrite2(
			ocisvcctx,                     //OCISvcCtx          *svchp,
			b.env.ocierr,                  //OCIError           *errhp,
			b.ociLobLocator,               //OCILobLocator      *locp,
			&byte_amtp,                    //oraub8          *byte_amtp,
			nil,                           //oraub8          *char_amtp,
			C.oraub8(1),                   //oraub8          offset, starting position is 1
			unsafe.Pointer(&b.buf[0]),     //void            *bufp,
			C.oraub8(currentBytesToWrite), //oraub8          buflen,
			piece,            //ub1             piece,
			nil,              //void            *ctxp,
			nil,              //OCICallbackLobWrite2 (cbfp)
			C.ub2(0),         //ub2             csid,
			C.SQLCS_IMPLICIT) //ub1             csfrm );
		//fmt.Printf("r %v, currentBytesToWrite %v, buffer %v\n", r, currentBytesToWrite, buffer)
		//fmt.Printf("C.OCI_NEED_DATA %v, C.OCI_SUCCESS %v\n", C.OCI_NEED_DATA, C.OCI_SUCCESS)
		if r == C.OCI_ERROR {
			return b.env.ociError()
		} else {
			// Determine action for next cycle
			if r == C.OCI_NEED_DATA {
				if remainingBytesToWrite > len(b.buf) {
					piece = C.OCI_NEXT_PIECE
				} else {
					piece = C.OCI_LAST_PIECE
				}
			} else if r == C.OCI_SUCCESS {
				writing = false
			}
		}
	}

	r = C.OCIBindByPos2(
		ocistmt,                               //OCIStmt      *stmtp,
		(**C.OCIBind)(&b.ocibnd),              //OCIBind      **bindpp,
		b.env.ocierr,                          //OCIError     *errhp,
		C.ub4(position),                       //ub4          position,
		unsafe.Pointer(&b.ociLobLocator),      //void         *valuep,
		C.sb8(unsafe.Sizeof(b.ociLobLocator)), //sb8          value_sz,
		C.SQLT_BLOB,                           //ub2          dty,
		nil,                                   //void         *indp,
		nil,                                   //ub2          *alenp,
		nil,                                   //ub2          *rcodep,
		0,                                     //ub4          maxarr_len,
		nil,                                   //ub4          *curelep,
		C.OCI_DEFAULT)                         //ub4          mode );
	if r == C.OCI_ERROR {
		return b.env.ociError()
	}

	return nil
}

func (b *bytesBind) setPtr() error {
	return nil
}

func (b *bytesBind) close() {
	defer func() {
		recover()
	}()
	b.freeLob()
	b.freeDescriptor()
	b.ocibnd = nil
	b.ocisvcctx = nil
	b.ociLobLocator = nil
	b.env.bytesBindPool.Put(b)
}

func (b *bytesBind) freeLob() {
	defer func() {
		recover()
	}()
	// free temporary lob
	C.OCILobFreeTemporary(
		b.ocisvcctx,     //OCISvcCtx          *svchp,
		b.env.ocierr,    //OCIError           *errhp,
		b.ociLobLocator) //OCILobLocator      *locp,
}

func (b *bytesBind) freeDescriptor() {
	defer func() {
		recover()
	}()
	// free lob locator handle
	C.OCIDescriptorFree(
		unsafe.Pointer(b.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                 //ub4      type );
}
