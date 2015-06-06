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
	"sync"
	"unsafe"
)

const lobChunkSize = 16 << 20 // 16Mb

var lobChunkPool = sync.Pool{
	New: func() interface{} {
		var b [lobChunkSize]byte
		return b
	},
}

type defLob struct {
	rset          *Rset
	ocidef        *C.OCIDefine
	ociLobLocator *C.OCILobLocator
	null          C.sb2
	gct           GoColumnType
	sqlt          C.ub2
	charsetForm   C.ub1
}

func (def *defLob) define(position int, charsetForm C.ub1, sqlt C.ub2, gct GoColumnType, rset *Rset) error {
	def.rset = rset
	def.gct = gct
	def.sqlt = sqlt
	def.charsetForm = charsetForm
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                                //OCIStmt     *stmtp,
		&def.ocidef,                                     //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,                //OCIError    *errhp,
		C.ub4(position),                                 //ub4         position,
		unsafe.Pointer(&def.ociLobLocator),              //void        *valuep,
		C.LENGTH_TYPE(unsafe.Sizeof(def.ociLobLocator)), //sb8         value_sz,
		sqlt, //ub2         dty,
		unsafe.Pointer(&def.null), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}
func (def *defLob) Bytes() (value []byte, err error) {
	// Open the lob to obtain length; round-trip to database
	r := C.OCILobOpen(
		def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		def.ociLobLocator,                //OCILobLocator      *locp,
		C.OCI_LOB_READONLY)               //ub1              mode );
	if r == C.OCI_ERROR {
		return nil, def.rset.stmt.ses.srv.env.ociError()
	}
	defer func() {
		if r := C.OCILobClose(
			def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
			def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			def.ociLobLocator,                //OCILobLocator      *locp,
		); r == C.OCI_ERROR {
			closeErr := def.rset.stmt.ses.srv.env.ociError()
			if err == nil {
				err = closeErr
			}
		}
	}()

	var lobLength C.oraub8
	// get the length of the lob
	r = C.OCILobGetLength2(
		def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		def.ociLobLocator,                //OCILobLocator      *locp,
		&lobLength)                       //oraub8 *lenp)
	if r == C.OCI_ERROR {
		return nil, def.rset.stmt.ses.srv.env.ociError()
	}

	if lobLength == 0 {
		return nil, nil
	}

	// Allocate []byte the length of the lob
	value = make([]byte, int(lobLength))
	for off, byte_amtp := 0, lobLength; byte_amtp > 0; byte_amtp = lobLength - C.oraub8(off) {
		Log.Infof("LobRead2 off=%d amt=%d", off, byte_amtp)
		r = C.OCILobRead2(
			def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
			def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			def.ociLobLocator,                //OCILobLocator      *locp,
			&byte_amtp,                       //oraub8             *byte_amtp,
			nil,                              //oraub8             *char_amtp,
			C.oraub8(off+1),                  //oraub8             offset, offset is 1-based
			unsafe.Pointer(&value[off]),      //void               *bufp,
			C.oraub8(lobChunkSize),           //oraub8             bufl,
			C.OCI_ONE_PIECE,                  //ub1                piece,
			nil,                              //void               *ctxp,
			nil,                              //OCICallbackLobRead2 (cbfp)
			C.ub2(0),                         //ub2                csid,
			def.charsetForm)                  //ub1                csfrm );

		if r == C.OCI_ERROR {
			return nil, def.rset.stmt.ses.srv.env.ociError()
		}
		// byte_amtp represents the amount copied into buffer by oci
		off += int(byte_amtp)
	}

	return value, nil
}

func (def *defLob) Reader() (io.Reader, error) {
	// Open the lob to obtain length; round-trip to database
	r := C.OCILobOpen(
		def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		def.ociLobLocator,                //OCILobLocator      *locp,
		C.OCI_LOB_READONLY)               //ub1              mode );
	if r == C.OCI_ERROR {
		return nil, def.rset.stmt.ses.srv.env.ociError()
	}

	return &lobReader{
		srv:           def.rset.stmt.ses.srv,
		ociLobLocator: def.ociLobLocator,
		charsetForm:   def.charsetForm,
		piece:         C.OCI_FIRST_PIECE,
	}, nil
}

func (def *defLob) String() (value string, err error) {
	var bytes []byte
	bytes, err = def.Bytes()
	value = string(bytes)
	return value, err
}
func (def *defLob) value() (value interface{}, err error) {
	if def.sqlt == C.SQLT_BLOB {
		if def.gct == Bin {
			if def.null > -1 {
				value, err = def.Bytes()
			}
		} else {
			binValue := Raw{IsNull: def.null < 0}
			if !binValue.IsNull {
				binValue.Value, err = def.Bytes()
			}
			value = binValue
		}
	} else {
		if def.gct == S {
			if def.null > -1 {
				value, err = def.String()
			}
		} else {
			oraString := String{IsNull: def.null < 0}
			if !oraString.IsNull {
				oraString.Value, err = def.String()
			}
			value = oraString
		}
	}
	return value, err
}
func (def *defLob) alloc() error {
	// Allocate lob locator handle
	// OCI_DTYPE_LOB is for a BLOB or CLOB
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(&def.ociLobLocator)), //dvoid         **descpp,
		C.OCI_DTYPE_LOB,                                       //ub4           type,
		0,                                                     //size_t        xtramem_sz,
		nil)                                                   //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}

func (def *defLob) free() {
	defer func() {
		recover()
	}()
	C.OCIDescriptorFree(
		unsafe.Pointer(def.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                   //ub4      type );
}

func (def *defLob) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errRecover(value)
		}
	}()

	def.free()
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociLobLocator = nil
	rset.putDef(defIdxLob, def)
	return nil
}

type lobReader struct {
	srv           *Srv
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	piece         C.ub1
	off           C.oraub8
	done          bool
}

func (lr *lobReader) Close() error {
	if lr.ociLobLocator == nil {
		return nil
	}
	lob, srv := lr.ociLobLocator, lr.srv
	lr.ociLobLocator, lr.srv = nil, nil
	if !lr.done {
		srv.Break()
	}
	if C.OCILobClose(
		srv.ocisvcctx,  //OCISvcCtx          *svchp,
		srv.env.ocierr, //OCIError           *errhp,
		lob,            //OCILobLocator      *locp,
	) == C.OCI_ERROR {
		return srv.env.ociError()
	}
	return nil
}

func (lr *lobReader) Read(p []byte) (n int, err error) {
	defer func() {
		if err != nil {
			lr.Close()
		}
	}()

	var byte_amtp C.oraub8 // zero
	Log.Infof("LobRead2 off=%d amt=%d", lr.off, len(p))
	r := C.OCILobRead2(
		lr.srv.ocisvcctx,      //OCISvcCtx          *svchp,
		lr.srv.env.ocierr,     //OCIError           *errhp,
		lr.ociLobLocator,      //OCILobLocator      *locp,
		&byte_amtp,            //oraub8             *byte_amtp,
		nil,                   //oraub8             *char_amtp,
		lr.off+1,              //oraub8             offset, offset is 1-based
		unsafe.Pointer(&p[0]), //void               *bufp,
		C.oraub8(len(p)),      //oraub8             bufl,
		lr.piece,              //ub1                piece,
		nil,                   //void               *ctxp,
		nil,                   //OCICallbackLobRead2 (cbfp)
		C.ub2(0),              //ub2                csid,
		lr.charsetForm,        //ub1                csfrm );
	)
	if r == C.OCI_ERROR {
		lr.done = true
		return 0, lr.srv.env.ociError()
	}
	if r == C.OCI_NO_DATA {
		lr.done = true
		return int(byte_amtp), io.EOF
	}
	// byte_amtp represents the amount copied into buffer by oci
	if lr.piece == C.OCI_FIRST_PIECE {
		lr.piece = C.OCI_NEXT_PIECE
	}
	lr.off += byte_amtp
	return int(byte_amtp), nil
}

func (lr *lobReader) WriteTo(w io.Writer) (n int64, err error) {
	defer func() {
		if closeErr := lr.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var byte_amtp C.oraub8 // zero
	arr := lobChunkPool.Get().([lobChunkSize]byte)
	defer lobChunkPool.Put(arr)
	buf := arr[:]

	var k int
	for {
		Log.Infof("LobRead2 off=%d amt=%d", lr.off, len(buf))
		r := C.OCILobRead2(
			lr.srv.ocisvcctx,        //OCISvcCtx          *svchp,
			lr.srv.env.ocierr,       //OCIError           *errhp,
			lr.ociLobLocator,        //OCILobLocator      *locp,
			&byte_amtp,              //oraub8             *byte_amtp,
			nil,                     //oraub8             *char_amtp,
			lr.off+1,                //oraub8             offset, offset is 1-based
			unsafe.Pointer(&buf[0]), //void               *bufp,
			C.oraub8(len(buf)),      //oraub8             bufl,
			lr.piece,                //ub1                piece,
			nil,                     //void               *ctxp,
			nil,                     //OCICallbackLobRead2 (cbfp)
			C.ub2(0),                //ub2                csid,
			lr.charsetForm,          //ub1                csfrm );
		)
		if r == C.OCI_ERROR {
			return 0, lr.srv.env.ociError()
		}
		if r == C.OCI_NO_DATA {
			return n, nil
		}
		// byte_amtp represents the amount copied into buffer by oci
		if lr.piece == C.OCI_FIRST_PIECE {
			lr.piece = C.OCI_NEXT_PIECE
		}
		lr.off += byte_amtp

		if k, err = w.Write(buf[:int(byte_amtp)]); err != nil {
			return n, err
		}
		n += int64(k)
	}
	lr.done = true
	return n, nil
}
