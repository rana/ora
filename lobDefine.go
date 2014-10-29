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

type lobDefine struct {
	environment   *Environment
	ocidef        *C.OCIDefine
	ocisvcctx     *C.OCISvcCtx
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	sqlt          C.ub2
	isNull        C.sb2
	returnType    GoColumnType
}

func (lobDefine *lobDefine) define(sqlt C.ub2, charsetForm C.ub1, columnSize int, position int, returnType GoColumnType, ocisvcctx *C.OCISvcCtx, ocistmt *C.OCIStmt) error {
	lobDefine.ocisvcctx = ocisvcctx
	lobDefine.sqlt = sqlt
	lobDefine.charsetForm = charsetForm
	lobDefine.returnType = returnType
	r := C.OCIDefineByPos2(
		ocistmt,                                  //OCIStmt     *stmtp,
		&lobDefine.ocidef,                        //OCIDefine   **defnpp,
		lobDefine.environment.ocierr,             //OCIError    *errhp,
		C.ub4(position),                          //ub4         position,
		unsafe.Pointer(&lobDefine.ociLobLocator), //void        *valuep,
		C.sb8(columnSize),                        //sb8         value_sz,
		sqlt,                                     //ub2         dty,
		unsafe.Pointer(&lobDefine.isNull), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return lobDefine.environment.ociError()
	}
	return nil
}
func (lobDefine *lobDefine) Bytes() (value []byte, err error) {
	var lobLength C.oraub8
	// Open the lob to obtain length; round-trip to database
	r := C.OCILobOpen(
		lobDefine.ocisvcctx,          //OCISvcCtx          *svchp,
		lobDefine.environment.ocierr, //OCIError           *errhp,
		lobDefine.ociLobLocator,      //OCILobLocator      *locp,
		C.OCI_LOB_READONLY)           //ub1              mode );
	if r == C.OCI_ERROR {
		return nil, lobDefine.environment.ociError()
	}
	// get the length of the lob
	r = C.OCILobGetLength2(
		lobDefine.ocisvcctx,          //OCISvcCtx          *svchp,
		lobDefine.environment.ocierr, //OCIError           *errhp,
		lobDefine.ociLobLocator,      //OCILobLocator      *locp,
		&lobLength)                   //oraub8 *lenp)
	if r == C.OCI_ERROR {
		return nil, lobDefine.environment.ociError()
	}

	if lobLength > 0 {
		// Allocate []byte the length of the lob
		value = make([]byte, int(lobLength))
		// buffer is size of oracle.LobBufferSize
		var buffer [1 << 24]byte
		var writeIndex int
		var byte_amtp C.oraub8 = lobLength
		var piece C.ub1 = C.OCI_FIRST_PIECE
		var loading bool = true
		for loading {
			r = C.OCILobRead2(
				lobDefine.ocisvcctx,          //OCISvcCtx          *svchp,
				lobDefine.environment.ocierr, //OCIError           *errhp,
				lobDefine.ociLobLocator,      //OCILobLocator      *locp,
				&byte_amtp,                   //oraub8             *byte_amtp,
				nil,                          //oraub8             *char_amtp,
				C.oraub8(1),                  //oraub8             offset, offset is 1-based
				unsafe.Pointer(&buffer[0]),   //void               *bufp,
				C.oraub8(len(buffer)),        //oraub8             bufl,
				piece,                 //ub1                piece,
				nil,                   //void               *ctxp,
				nil,                   //OCICallbackLobRead2 (cbfp)
				C.ub2(0),              //ub2                csid,
				lobDefine.charsetForm) //ub1                csfrm );

			if r == C.OCI_ERROR {
				return nil, lobDefine.environment.ociError()
			} else {
				// Write buffer to return slice
				// byte_amtp represents the amount copied into buffer by oci
				for n := 0; n < int(byte_amtp); n++ {
					value[writeIndex] = buffer[n]
					writeIndex++
				}
				// Determine action for next cycle
				if r == C.OCI_NEED_DATA {
					piece = C.OCI_NEXT_PIECE
				} else if r == C.OCI_SUCCESS {
					loading = false
				}
			}
		}
	}

	r = C.OCILobClose(
		lobDefine.ocisvcctx,          //OCISvcCtx          *svchp,
		lobDefine.environment.ocierr, //OCIError           *errhp,
		lobDefine.ociLobLocator)      //OCILobLocator      *locp,
	if r == C.OCI_ERROR {
		return nil, lobDefine.environment.ociError()
	}

	return value, nil

}
func (lobDefine *lobDefine) String() (value string, err error) {
	var bytes []byte
	bytes, err = lobDefine.Bytes()
	value = string(bytes)
	return value, err
}
func (lobDefine *lobDefine) value() (value interface{}, err error) {
	if lobDefine.sqlt == C.SQLT_BLOB {
		if lobDefine.returnType == Bits {
			if lobDefine.isNull > -1 {
				value, err = lobDefine.Bytes()
			}
		} else {
			bytesValue := Bytes{IsNull: lobDefine.isNull < 0}
			if !bytesValue.IsNull {
				bytesValue.Value, err = lobDefine.Bytes()
			}
			value = bytesValue
		}
	} else {
		if lobDefine.returnType == S {
			if lobDefine.isNull > -1 {
				value, err = lobDefine.String()
			}
		} else {
			oraString := String{IsNull: lobDefine.isNull < 0}
			if !oraString.IsNull {
				oraString.Value, err = lobDefine.String()
			}
			value = oraString
		}
	}
	return value, err
}
func (lobDefine *lobDefine) alloc() error {
	// Allocate lob locator handle
	// OCI_DTYPE_LOB is for a BLOB or CLOB
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(lobDefine.environment.ocienv),                //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&lobDefine.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_LOB,                                             //ub4           type,
		0,                                                           //size_t        xtramem_sz,
		nil)                                                         //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return lobDefine.environment.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}
func (lobDefine *lobDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(lobDefine.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                         //ub4      type );
}
func (lobDefine *lobDefine) close() {
	defer func() {
		recover()
	}()
	lobDefine.ocidef = nil
	lobDefine.ocisvcctx = nil
	lobDefine.isNull = C.sb2(0)
	lobDefine.environment.lobDefinePool.Put(lobDefine)
}
