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
	env           *Environment
	ocidef        *C.OCIDefine
	ocisvcctx     *C.OCISvcCtx
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	sqlt          C.ub2
	isNull        C.sb2
	returnType    GoColumnType
}

func (d *lobDefine) define(sqlt C.ub2, charsetForm C.ub1, columnSize int, position int, returnType GoColumnType, ocisvcctx *C.OCISvcCtx, ocistmt *C.OCIStmt) error {
	d.ocisvcctx = ocisvcctx
	d.sqlt = sqlt
	d.charsetForm = charsetForm
	d.returnType = returnType
	r := C.OCIDefineByPos2(
		ocistmt,                          //OCIStmt     *stmtp,
		&d.ocidef,                        //OCIDefine   **defnpp,
		d.env.ocierr,                     //OCIError    *errhp,
		C.ub4(position),                  //ub4         position,
		unsafe.Pointer(&d.ociLobLocator), //void        *valuep,
		C.sb8(columnSize),                //sb8         value_sz,
		sqlt,                             //ub2         dty,
		unsafe.Pointer(&d.isNull), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}
func (d *lobDefine) Bytes() (value []byte, err error) {
	var lobLength C.oraub8
	// Open the lob to obtain length; round-trip to database
	r := C.OCILobOpen(
		d.ocisvcctx,        //OCISvcCtx          *svchp,
		d.env.ocierr,       //OCIError           *errhp,
		d.ociLobLocator,    //OCILobLocator      *locp,
		C.OCI_LOB_READONLY) //ub1              mode );
	if r == C.OCI_ERROR {
		return nil, d.env.ociError()
	}
	// get the length of the lob
	r = C.OCILobGetLength2(
		d.ocisvcctx,     //OCISvcCtx          *svchp,
		d.env.ocierr,    //OCIError           *errhp,
		d.ociLobLocator, //OCILobLocator      *locp,
		&lobLength)      //oraub8 *lenp)
	if r == C.OCI_ERROR {
		return nil, d.env.ociError()
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
				d.ocisvcctx,                //OCISvcCtx          *svchp,
				d.env.ocierr,               //OCIError           *errhp,
				d.ociLobLocator,            //OCILobLocator      *locp,
				&byte_amtp,                 //oraub8             *byte_amtp,
				nil,                        //oraub8             *char_amtp,
				C.oraub8(1),                //oraub8             offset, offset is 1-based
				unsafe.Pointer(&buffer[0]), //void               *bufp,
				C.oraub8(len(buffer)),      //oraub8             bufl,
				piece,         //ub1                piece,
				nil,           //void               *ctxp,
				nil,           //OCICallbackLobRead2 (cbfp)
				C.ub2(0),      //ub2                csid,
				d.charsetForm) //ub1                csfrm );

			if r == C.OCI_ERROR {
				return nil, d.env.ociError()
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
		d.ocisvcctx,     //OCISvcCtx          *svchp,
		d.env.ocierr,    //OCIError           *errhp,
		d.ociLobLocator) //OCILobLocator      *locp,
	if r == C.OCI_ERROR {
		return nil, d.env.ociError()
	}

	return value, nil

}
func (d *lobDefine) String() (value string, err error) {
	var bytes []byte
	bytes, err = d.Bytes()
	value = string(bytes)
	return value, err
}
func (d *lobDefine) value() (value interface{}, err error) {
	if d.sqlt == C.SQLT_BLOB {
		if d.returnType == Bits {
			if d.isNull > -1 {
				value, err = d.Bytes()
			}
		} else {
			bytesValue := Bytes{IsNull: d.isNull < 0}
			if !bytesValue.IsNull {
				bytesValue.Value, err = d.Bytes()
			}
			value = bytesValue
		}
	} else {
		if d.returnType == S {
			if d.isNull > -1 {
				value, err = d.String()
			}
		} else {
			oraString := String{IsNull: d.isNull < 0}
			if !oraString.IsNull {
				oraString.Value, err = d.String()
			}
			value = oraString
		}
	}
	return value, err
}
func (d *lobDefine) alloc() error {
	// Allocate lob locator handle
	// OCI_DTYPE_LOB is for a BLOB or CLOB
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(d.env.ocienv),                        //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&d.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_LOB,                                     //ub4           type,
		0,                                                   //size_t        xtramem_sz,
		nil)                                                 //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return d.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}
func (d *lobDefine) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(d.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                 //ub4      type );
}
func (d *lobDefine) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.ocisvcctx = nil
	d.isNull = C.sb2(0)
	d.env.lobDefinePool.Put(d)
}
