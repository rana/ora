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
	"fmt"
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
	def.ociLobLocator = nil
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
	if r != C.OCI_SUCCESS {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	prefetchLength := C.boolean(C.TRUE)
	return def.rset.stmt.ses.srv.env.setAttr(unsafe.Pointer(def.ocidef), C.OCI_HTYPE_DEFINE, unsafe.Pointer(&prefetchLength), 0, C.OCI_ATTR_LOBPREFETCH_LENGTH)
}

func (def *defLob) Bytes() (value []byte, err error) {
	// Open the lob to obtain length; round-trip to database
	Log.Infof("Bytes OCILobOpen %p", def.ociLobLocator)
	r := C.OCILobOpen(
		def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		def.ociLobLocator,                //OCILobLocator      *locp,
		C.OCI_LOB_READONLY)               //ub1              mode );
	if r == C.OCI_ERROR {
		return nil, def.rset.stmt.ses.srv.env.ociError()
	}
	defer func() {
		Log.Infof("Bytes OCILobClose %p", def.ociLobLocator)
		if closeErr := lobClose(def.rset.stmt.ses.srv, def.ociLobLocator); closeErr != nil && err == nil {
			err = closeErr
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
func (def *defLob) String() (value string, err error) {
	var bytes []byte
	bytes, err = def.Bytes()
	value = string(bytes)
	return value, err
}

// Reader returns an io.Reader for the underlying LOB.
// Also dissociates this def from the LOB!
func (def *defLob) Reader() (io.Reader, error) {
	// Open the lob to obtain length; round-trip to database
	Log.Infof("Reader OCILobOpen %p", def.ociLobLocator)
	r := C.OCILobOpen(
		def.rset.stmt.ses.srv.ocisvcctx,  //OCISvcCtx          *svchp,
		def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
		def.ociLobLocator,                //OCILobLocator      *locp,
		C.OCI_LOB_READONLY)               //ub1              mode );
	if r != C.OCI_SUCCESS {
		return nil, def.rset.stmt.ses.srv.env.ociError()
	}
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

	lr := &lobReader{
		srv:           def.rset.stmt.ses.srv,
		ociLobLocator: def.ociLobLocator,
		charsetForm:   def.charsetForm,
		piece:         C.OCI_FIRST_PIECE,
		Length:        lobLength,
	}
	def.ociLobLocator = nil
	return lr, nil
}

func (def *defLob) value() (value interface{}, err error) {
	if def.gct == Bin {
		if def.null > -1 {
			return def.Reader()
		}
		return value, err
	}
	binValue := Lob{IsNull: def.null < 0}
	if !binValue.IsNull {
		binValue.Reader, err = def.Reader()
	}
	return binValue, err
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
	if def.ociLobLocator == nil { // dissociated or already freed
		return
	}
	C.OCIDescriptorFree(
		unsafe.Pointer(def.ociLobLocator), //void     *descp,
		C.OCI_DTYPE_LOB)                   //ub4      type );
	def.ociLobLocator = nil
}

func (def *defLob) close() (err error) {
	Log.Infof("defLob close %p", def.ociLobLocator)
	lob := def.ociLobLocator
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociLobLocator = nil
	rset.putDef(defIdxLob, def)

	if lob == nil {
		return nil
	}
	return lobClose(rset.stmt.ses.srv, lob)
}

type lobReader struct {
	srv           *Srv
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	piece         C.ub1
	off           C.oraub8
	interrupted   bool
	Length        C.oraub8
}

func lobClose(srv *Srv, lob *C.OCILobLocator) error {
	if lob == nil {
		return nil
	}
	r := C.OCILobClose(
		srv.ocisvcctx,  //OCISvcCtx          *svchp,
		srv.env.ocierr, //OCIError           *errhp,
		lob,            //OCILobLocator      *locp,
	)
	C.OCIDescriptorFree(unsafe.Pointer(lob), //void     *descp,
		C.OCI_DTYPE_LOB) //ub4      type );
	if r == C.OCI_ERROR {
		return srv.env.ociError()
	}
	return nil
}

func (lr *lobReader) Close() error {
	if lr.ociLobLocator == nil {
		return nil
	}
	lob, srv := lr.ociLobLocator, lr.srv
	lr.ociLobLocator, lr.srv = nil, nil
	if lr.interrupted {
		srv.Break()
	}
	Log.Infof("lobReader OCILobClose %p", lr.ociLobLocator)
	return lobClose(srv, lob)
}

func (lr *lobReader) Read(p []byte) (n int, err error) {
	if lr.ociLobLocator == nil {
		return 0, io.EOF
	}
	defer func() {
		if err != nil {
			lr.Close()
		}
	}()

	var byte_amtp C.oraub8 // zero
	Log.Infof("LobRead2 piece=%d off=%d amt=%d", lr.piece, lr.off, len(p))
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
	Log.Infof("LobRead2 returned %d amt=%d", r, byte_amtp)
	switch r {
	case C.OCI_ERROR:
		lr.interrupted = true
		return 0, lr.srv.env.ociError()
	case C.OCI_NO_DATA:
		return int(byte_amtp), io.EOF
	case C.OCI_INVALID_HANDLE:
		return 0, fmt.Errorf("Invalid handle %v", lr.ociLobLocator)
	}
	// byte_amtp represents the amount copied into buffer by oci
	if byte_amtp != 0 {
		lr.off += byte_amtp
		if lr.off == lr.Length {
			return int(byte_amtp), io.EOF
		}
		if lr.piece == C.OCI_FIRST_PIECE {
			lr.piece = C.OCI_NEXT_PIECE
		}
	}
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
		Log.Infof("WriteTo LobRead2 off=%d amt=%d", lr.off, len(buf))
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
		Log.Infof("WriteTo LobRead2 returned %d amt=%d piece=%d", r, byte_amtp, lr.piece)
		switch r {
		case C.OCI_SUCCESS:
		case C.OCI_NO_DATA:
			break
		default:
			return 0, lr.srv.env.ociError()
		}
		// byte_amtp represents the amount copied into buffer by oci
		lr.off += byte_amtp

		if byte_amtp != 0 {
			if k, err = w.Write(buf[:int(byte_amtp)]); err != nil {
				return n, err
			}
			n += int64(k)
			if lr.off == lr.Length {
				break
			}
		}
		if lr.piece == C.OCI_FIRST_PIECE {
			lr.piece = C.OCI_NEXT_PIECE
		}
	}
	return n, nil
}
