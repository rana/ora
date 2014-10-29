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

type bytesSliceBind struct {
	environment    *Environment
	ocibnd         *C.OCIBind
	ocisvcctx      *C.OCISvcCtx
	ociLobLocators []*C.OCILobLocator
	buffer         []byte
}

func (bytesSliceBind *bytesSliceBind) bindOraBytes(values []Bytes, position int, lobBufferSize int, ocisvcctx *C.OCISvcCtx, ocistmt *C.OCIStmt) error {
	bytesValues := make([][]byte, len(values))
	nullInds := make([]C.sb2, len(values))
	for n, _ := range values {
		if values[n].IsNull {
			nullInds[n] = C.sb2(-1)
		} else {
			bytesValues[n] = values[n].Value
		}
	}
	return bytesSliceBind.bindBytes(bytesValues, nullInds, position, lobBufferSize, ocisvcctx, ocistmt)
}

func (bytesSliceBind *bytesSliceBind) bindBytes(values [][]byte, nullInds []C.sb2, position int, lobBufferSize int, ocisvcctx *C.OCISvcCtx, ocistmt *C.OCIStmt) error {
	bytesSliceBind.ocisvcctx = ocisvcctx
	bytesSliceBind.ociLobLocators = make([]*C.OCILobLocator, len(values))
	if nullInds == nil {
		nullInds = make([]C.sb2, len(values))
	}
	alenp := make([]C.ub4, len(values))
	rcodep := make([]C.ub2, len(values))
	if len(bytesSliceBind.buffer) < lobBufferSize {
		bytesSliceBind.buffer = make([]byte, lobBufferSize)
	}

	for n, valueBytes := range values {
		// OCILobWrite2 doesn't support writing zero bytes
		// nor is writing 1 byte and erasing the one byte supported
		// therefore, throw an error
		if len(valueBytes) == 0 && nullInds[n] > -1 {
			return errNew("writing a zero-length BLOB is unsupported")
		}

		// Allocate lob locator handle
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(bytesSliceBind.environment.ocienv),                    //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&bytesSliceBind.ociLobLocators[n])), //dvoid         **descpp,
			C.OCI_DTYPE_LOB, //ub4           type,
			0,               //size_t        xtramem_sz,
			nil)             //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return bytesSliceBind.environment.ociError()
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci lob handle during bind")
		}
		// Create temporary lob
		r = C.OCILobCreateTemporary(
			ocisvcctx,                         //OCISvcCtx          *svchp,
			bytesSliceBind.environment.ocierr, //OCIError           *errhp,
			bytesSliceBind.ociLobLocators[n],  //OCILobLocator      *locp,
			C.OCI_DEFAULT,                     //ub2                csid,
			C.SQLCS_IMPLICIT,                  //ub1                csfrm,
			C.OCI_TEMP_BLOB,                   //ub1                lobtype,
			C.FALSE,                           //boolean            cache,
			C.OCI_DURATION_SESSION)            //OCIDuration        duration);
		if r == C.OCI_ERROR {
			return bytesSliceBind.environment.ociError()
		}

		if nullInds[n] > -1 {
			var currentBytesToWrite int
			var remainingBytesToWrite int = len(valueBytes)
			var readIndex int
			var byte_amtp C.oraub8 /* Setting Amount to 0 streams the data until use specifies OCI_LAST_PIECE */
			var piece C.ub1 = C.OCI_FIRST_PIECE
			var writing bool = true
			for writing {
				// Copy bytes from slice to buffer
				if remainingBytesToWrite < len(bytesSliceBind.buffer) {
					currentBytesToWrite = remainingBytesToWrite
				} else {
					currentBytesToWrite = len(bytesSliceBind.buffer)
				}
				for n := 0; n < currentBytesToWrite; n++ {
					bytesSliceBind.buffer[n] = valueBytes[readIndex]
					readIndex++
				}
				remainingBytesToWrite = len(valueBytes) - readIndex

				// Write to Oracle
				r = C.OCILobWrite2(
					ocisvcctx,                                 //OCISvcCtx          *svchp,
					bytesSliceBind.environment.ocierr,         //OCIError           *errhp,
					bytesSliceBind.ociLobLocators[n],          //OCILobLocator      *locp,
					&byte_amtp,                                //oraub8          *byte_amtp,
					nil,                                       //oraub8          *char_amtp,
					C.oraub8(1),                               //oraub8          offset, starting position is 1
					unsafe.Pointer(&bytesSliceBind.buffer[0]), //void            *bufp,
					C.oraub8(currentBytesToWrite),             //oraub8          buflen,
					piece,            //ub1             piece,
					nil,              //void            *ctxp,
					nil,              //OCICallbackLobWrite2 (cbfp)
					C.ub2(0),         //ub2             csid,
					C.SQLCS_IMPLICIT) //ub1             csfrm );
				//fmt.Printf("r %v, currentBytesToWrite %v, buffer %v\n", r, currentBytesToWrite, buffer)
				//fmt.Printf("C.OCI_NEED_DATA %v, C.OCI_SUCCESS %v\n", C.OCI_NEED_DATA, C.OCI_SUCCESS)
				if r == C.OCI_ERROR {
					return bytesSliceBind.environment.ociError()
				} else {
					// Determine action for next cycle
					if r == C.OCI_NEED_DATA {
						if remainingBytesToWrite > len(bytesSliceBind.buffer) {
							piece = C.OCI_NEXT_PIECE
						} else {
							piece = C.OCI_LAST_PIECE
						}
					} else if r == C.OCI_SUCCESS {
						writing = false
					}
				}
			}
		}

		alenp[n] = C.ub4(unsafe.Sizeof(bytesSliceBind.ociLobLocators[n]))
	}
	r := C.OCIBindByPos2(
		ocistmt, //OCIStmt      *stmtp,
		(**C.OCIBind)(&bytesSliceBind.ocibnd),                  //OCIBind      **bindpp,
		bytesSliceBind.environment.ocierr,                      //OCIError     *errhp,
		C.ub4(position),                                        //ub4          position,
		unsafe.Pointer(&bytesSliceBind.ociLobLocators[0]),      //void         *valuep,
		C.sb8(unsafe.Sizeof(bytesSliceBind.ociLobLocators[0])), //sb8          value_sz,
		C.SQLT_BLOB,                  //ub2          dty,
		unsafe.Pointer(&nullInds[0]), //void         *indp,
		&alenp[0],                    //ub4          *alenp,
		&rcodep[0],                   //ub2          *rcodep,
		0,                            //ub4          maxarr_len,
		nil,                          //ub4          *curelep,
		C.OCI_DEFAULT)                //ub4          mode );
	if r == C.OCI_ERROR {
		return bytesSliceBind.environment.ociError()
	}

	r = C.OCIBindArrayOfStruct(
		bytesSliceBind.ocibnd,
		bytesSliceBind.environment.ocierr,
		C.ub4(unsafe.Sizeof(bytesSliceBind.ociLobLocators[0])), //ub4         pvskip,
		C.ub4(C.sizeof_sb2),                                    //ub4         indskip,
		C.ub4(C.sizeof_ub4),                                    //ub4         alskip,
		C.ub4(C.sizeof_ub2))                                    //ub4         rcskip
	if r == C.OCI_ERROR {
		return bytesSliceBind.environment.ociError()
	}

	return nil
}

func (bytesSliceBind *bytesSliceBind) setPtr() error {
	return nil
}

func (bytesSliceBind *bytesSliceBind) close() {
	defer func() {
		recover()
	}()
	for n := 0; n < len(bytesSliceBind.ociLobLocators); n++ {
		bytesSliceBind.freeLob(n)
		bytesSliceBind.freeDescriptor(n)
	}
	bytesSliceBind.ocibnd = nil
	bytesSliceBind.ocisvcctx = nil
	bytesSliceBind.ociLobLocators = nil
	bytesSliceBind.environment.bytesSliceBindPool.Put(bytesSliceBind)
}

func (bytesSliceBind *bytesSliceBind) freeLob(n int) {
	defer func() {
		recover()
	}()
	// free temporary lob
	C.OCILobFreeTemporary(
		bytesSliceBind.ocisvcctx,          //OCISvcCtx          *svchp,
		bytesSliceBind.environment.ocierr, //OCIError           *errhp,
		bytesSliceBind.ociLobLocators[n])  //OCILobLocator      *locp,
}

func (bytesSliceBind *bytesSliceBind) freeDescriptor(n int) {
	defer func() {
		recover()
	}()
	// free lob locator handle
	C.OCIDescriptorFree(
		unsafe.Pointer(bytesSliceBind.ociLobLocators[n]), //void     *descp,
		C.OCI_DTYPE_LOB)                                  //ub4      type );
}
