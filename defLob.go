// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
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
	rset        *Rset
	ocidef      *C.OCIDefine
	gct         GoColumnType
	sqlt        C.ub2
	charsetForm C.ub1
	lobLocatorp
	nullp
}

func (def *defLob) define(position int, charsetForm C.ub1, sqlt C.ub2, gct GoColumnType, rset *Rset) error {
	def.rset = rset
	def.gct = gct
	def.sqlt = sqlt
	def.charsetForm = charsetForm
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                          //OCIStmt     *stmtp,
		&def.ocidef,                               //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,          //OCIError    *errhp,
		C.ub4(position),                           //ub4         position,
		unsafe.Pointer(def.lobLocatorp.Pointer()), //void        *valuep,
		C.LENGTH_TYPE(def.lobLocatorp.Size()),     //sb8         value_sz,
		sqlt, //ub2         dty,
		unsafe.Pointer(def.nullp.Pointer()), //void        *indp,
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
	//Log.Infof("Bytes OCILobOpen %p", def.ociLobLocator)
	lobLength, err := lobOpen(def.rset.stmt.ses, def.lobLocatorp.Value(), C.OCI_LOB_READONLY)
	if err != nil {
		return nil, err
	}
	defer func() {
		//Log.Infof("Bytes OCILobClose %p", def.ociLobLocator)
		if closeErr := lobClose(def.rset.stmt.ses, def.lobLocatorp.Value()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if lobLength == 0 {
		return nil, nil
	}

	// Allocate []byte the length of the lob
	value = make([]byte, int(lobLength))
	for off, byte_amtp := 0, lobLength; byte_amtp > 0; byte_amtp = lobLength - C.oraub8(off) {
		//Log.Infof("LobRead2 off=%d amt=%d", off, byte_amtp)
		r := C.OCILobRead2(
			def.rset.stmt.ses.ocisvcctx,      //OCISvcCtx          *svchp,
			def.rset.stmt.ses.srv.env.ocierr, //OCIError           *errhp,
			def.lobLocatorp.Value(),          //OCILobLocator      *locp,
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
	//Log.Infof("Reader OCILobOpen %p", def.ociLobLocator)
	lobLength, err := lobOpen(def.rset.stmt.ses, def.lobLocatorp.Value(), C.OCI_LOB_READONLY)
	if err != nil {
		return nil, err
	}

	lr := &lobReader{
		ses:           def.rset.stmt.ses,
		ociLobLocator: def.lobLocatorp.Value(),
		charsetForm:   def.charsetForm,
		piece:         C.OCI_FIRST_PIECE,
		Length:        lobLength,
	}
	return lr, nil
}

func (def *defLob) value() (value interface{}, err error) {
	//lob := def.ociLobLocator
	//Log.Infof("value %p null=%d", lob, def.null)
	if def.gct == Bin {
		if !def.nullp.IsNull() {
			return def.Reader()
		}
		return value, err
	}
	if def.nullp.IsNull() {
		return Lob{}, nil
	}
	var r io.Reader
	r, err = def.Reader()
	binValue := Lob{Reader: r}
	//Log.Infof("value %p returns %#v (%v)", lob, binValue, err)
	return binValue, err
}
func (def *defLob) alloc() error {
	// Allocate lob locator handle
	// OCI_DTYPE_LOB is for a BLOB or CLOB
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv),             //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(def.lobLocatorp.Pointer())), //dvoid         **descpp,
		C.OCI_DTYPE_LOB, //ub4           type,
		0,               //size_t        xtramem_sz,
		nil)             //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci lob handle during define")
	}
	return nil
}

func (def *defLob) free() {
	if def.lobLocatorp.Value() != nil {
		C.OCIDescriptorFree(
			unsafe.Pointer(def.lobLocatorp.Pointer()), //void     *descp,
			C.OCI_DTYPE_LOB)                           //ub4      type );
	}
}

func (def *defLob) close() (err error) {
	//Log.Infof("defLob close %p", def.ociLobLocator)
	lob := def.lobLocatorp.Value()
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.nullp.Free()
	def.lobLocatorp.Free()
	rset.putDef(defIdxLob, def)

	if lob == nil {
		return nil
	}
	return lobClose(rset.stmt.ses, lob)
}

func lobOpen(ses *Ses, lob *C.OCILobLocator, mode C.ub1) (length C.oraub8, err error) {
	//Log.Infof("OCILobOpen %p\n%s", lob, getStack(1))
	r := C.OCILobOpen(
		ses.ocisvcctx,      //OCISvcCtx          *svchp,
		ses.srv.env.ocierr, //OCIError           *errhp,
		lob,                //OCILobLocator      *locp,
		mode)               //ub1              mode );
	//Log.Infof("OCILobOpen %p returned %d", lob, r)
	if r != C.OCI_SUCCESS {
		lobClose(ses, lob)
		return 0, ses.srv.env.ociError()
	}
	// get the length of the lob
	r = C.OCILobGetLength2(
		ses.ocisvcctx,      //OCISvcCtx          *svchp,
		ses.srv.env.ocierr, //OCIError           *errhp,
		lob,                //OCILobLocator      *locp,
		&length)            //oraub8 *lenp)
	if r == C.OCI_ERROR {
		lobClose(ses, lob)
		return length, ses.srv.env.ociError()
	}
	return length, nil
}

func lobClose(ses *Ses, lob *C.OCILobLocator) error {
	if lob == nil {
		return nil
	}
	//Log.Infof("OCILobClose %p\n%s", lob, getStack(1))
	r := C.OCILobClose(
		ses.ocisvcctx,      //OCISvcCtx          *svchp,
		ses.srv.env.ocierr, //OCIError           *errhp,
		lob,                //OCILobLocator      *locp,
	)
	C.OCIDescriptorFree(unsafe.Pointer(lob), //void     *descp,
		C.OCI_DTYPE_LOB) //ub4      type );
	if r == C.OCI_ERROR {
		return ses.srv.env.ociError()
	}
	return nil
}

var _ = io.Reader((*lobReader)(nil))
var _ = io.WriterTo((*lobReader)(nil))

type lobReader struct {
	ses           *Ses
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	piece         C.ub1
	off           C.oraub8
	interrupted   bool
	Length        C.oraub8
}

// Close the LOB reader.
func (lr *lobReader) Close() error {
	if lr.ociLobLocator == nil {
		return nil
	}
	lob, ses := lr.ociLobLocator, lr.ses
	lr.ociLobLocator, lr.ses = nil, nil
	if lr.interrupted {
		ses.Break()
	}
	//Log.Infof("lobReader OCILobClose %p", lr.ociLobLocator)
	return lobClose(ses, lob)
}

// Read into p, the next chunk.
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
	//Log.Infof("LobRead2 piece=%d off=%d amt=%d", lr.piece, lr.off, len(p))
	r := C.OCILobRead2(
		lr.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lr.ses.srv.env.ocierr, //OCIError           *errhp,
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
	//Log.Infof("LobRead2 returned %d amt=%d", r, byte_amtp)
	switch r {
	case C.OCI_ERROR:
		lr.interrupted = true
		return 0, lr.ses.srv.env.ociError()
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

// WriteTo writes all data from the LOB into the given Writer.
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
		//Log.Infof("WriteTo LobRead2 off=%d amt=%d", lr.off, len(buf))
		r := C.OCILobRead2(
			lr.ses.ocisvcctx,        //OCISvcCtx          *svchp,
			lr.ses.srv.env.ocierr,   //OCIError           *errhp,
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
		//Log.Infof("WriteTo LobRead2 returned %d amt=%d piece=%d", r, byte_amtp, lr.piece)
		switch r {
		case C.OCI_SUCCESS:
		case C.OCI_NO_DATA:
			break
		default:
			return 0, lr.ses.srv.env.ociError()
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

// TODO(tgulacsi): find how to return lobReadWriter.

var _ = io.ReaderAt((*lobReadWriter)(nil))
var _ = io.WriterAt((*lobReadWriter)(nil))

type lobReadWriter struct {
	ses           *Ses
	ociLobLocator *C.OCILobLocator
	charsetForm   C.ub1
	size          C.oraub8
}

// Size returns the actual size of the LOB.
func (lrw lobReadWriter) Size() uint64 {
	return uint64(lrw.size)
}

// Close the LOB.
func (lrw *lobReadWriter) Close() error {
	lob := lrw.ociLobLocator
	if lob == nil {
		return nil
	}
	lrw.ociLobLocator = nil
	return lobClose(lrw.ses, lob)
}

// Truncate the lob to the given length.
func (lrw *lobReadWriter) Truncate(length int64) error {
	if C.OCILobTrim2(
		lrw.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lrw.ses.srv.env.ocierr, //OCIError           *errhp,
		lrw.ociLobLocator,      //OCILobLocator      *locp,
		C.oraub8(length),       //oraub8             *newlen)
	) == C.OCI_ERROR {
		return lrw.ses.srv.env.ociError()
	}
	return nil
}

// ReadAt reads into p, starting from off.
func (lrw *lobReadWriter) ReadAt(p []byte, off int64) (n int, err error) {
	byte_amtp := C.oraub8(len(p))
	//Log.Infof("LobRead2 off=%d amt=%d", off, len(p))
	r := C.OCILobRead2(
		lrw.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lrw.ses.srv.env.ocierr, //OCIError           *errhp,
		lrw.ociLobLocator,      //OCILobLocator      *locp,
		&byte_amtp,             //oraub8             *byte_amtp,
		nil,                    //oraub8             *char_amtp,
		C.oraub8(off)+1,        //oraub8             offset, offset is 1-based
		unsafe.Pointer(&p[0]),  //void               *bufp,
		C.oraub8(len(p)),       //oraub8             bufl,
		C.OCI_ONE_PIECE,        //ub1                piece,
		nil,                    //void               *ctxp,
		nil,                    //OCICallbackLobRead2 (cbfp)
		C.ub2(0),               //ub2                csid,
		lrw.charsetForm,        //ub1                csfrm );
	)
	//Log.Infof("LobRead2 returned %d amt=%d", r, byte_amtp)
	switch r {
	case C.OCI_ERROR:
		return 0, lrw.ses.srv.env.ociError()
	case C.OCI_NO_DATA:
		return int(byte_amtp), io.EOF
	case C.OCI_INVALID_HANDLE:
		return 0, fmt.Errorf("Invalid handle %v", lrw.ociLobLocator)
	}
	return int(byte_amtp), nil
}

// WriteAt writes data in p into the LOB, starting at off.
func (lrw *lobReadWriter) WriteAt(p []byte, off int64) (n int, err error) {
	//Log.Infof("LobWrite2 off=%d len=%d", off, n)
	byte_amtp := C.oraub8(len(p))
	// Write to Oracle
	if C.OCILobWrite2(
		lrw.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lrw.ses.srv.env.ocierr, //OCIError           *errhp,
		lrw.ociLobLocator,      //OCILobLocator      *locp,
		&byte_amtp,             //oraub8          *byte_amtp,
		nil,                    //oraub8          *char_amtp,
		C.oraub8(off)+1,        //oraub8          offset, starting position is 1
		unsafe.Pointer(&p[0]),  //void            *bufp,
		C.oraub8(len(p)),
		C.OCI_ONE_PIECE,  //ub1             piece,
		nil,              //void            *ctxp,
		nil,              //OCICallbackLobWrite2 (cbfp)
		C.ub2(0),         //ub2             csid,
		C.SQLCS_IMPLICIT, //ub1             csfrm );
	//fmt.Printf("r %v, current %v, buffer %v\n", r, current, buffer)
	//fmt.Printf("C.OCI_NEED_DATA %v, C.OCI_SUCCESS %v\n", C.OCI_NEED_DATA, C.OCI_SUCCESS)
	) == C.OCI_ERROR {
		return 0, lrw.ses.srv.env.ociError()
	}
	if C.oraub8(off)+byte_amtp > lrw.size {
		lrw.size = C.oraub8(off) + byte_amtp
	}
	return int(byte_amtp), nil
}
