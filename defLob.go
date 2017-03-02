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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"unicode/utf16"
	"unsafe"
)

const lobChunkSize = (1 << 20) // 1Mb

var lobChunkPool = sync.Pool{
	New: func() interface{} {
		var b [lobChunkSize]byte
		return b
	},
}

type defLob struct {
	ociDef
	gct  GoColumnType
	sqlt C.ub2
	lobs []*C.OCILobLocator
}

func (def *defLob) define(position int, sqlt C.ub2, gct GoColumnType, rset *Rset) error {
	def.rset = rset
	def.gct = gct
	def.sqlt = sqlt
	if def.lobs != nil {
		C.free(unsafe.Pointer(&def.lobs[0]))
	}
	def.lobs = (*((*[MaxFetchLen]*C.OCILobLocator)(C.malloc(C.size_t(rset.fetchLen) * C.sof_LobLocatorp))))[:rset.fetchLen]
	if err := def.ociDef.defineByPos(position, unsafe.Pointer(&def.lobs[0]), int(C.sof_LobLocatorp), int(sqlt)); err != nil {
		return err
	}
	prefetchLength := C.boolean(C.TRUE)
	return def.rset.stmt.ses.srv.env.setAttr(unsafe.Pointer(def.ocidef), C.OCI_HTYPE_DEFINE,
		unsafe.Pointer(&prefetchLength), 0, C.OCI_ATTR_LOBPREFETCH_LENGTH)
}

func (def *defLob) Bytes(offset int) (value []byte, err error) {
	r := def.Reader(offset)
	defer r.Close()
	lr := r.(*lobReader)

	arr := lobChunkPool.Get().([lobChunkSize]byte)
	defer lobChunkPool.Put(arr)
	var buf bytes.Buffer

	n, err := r.Read(arr[:])
	lr.ses.logF(_drv.Cfg().Log.Ses.Prep, "Bytes-1(%p) amt=%d err=%v\n", lr, n, err)
	length := lr.Length
	if length == 0 {
		if err == io.EOF {
			err = nil
		}
		if n == 0 {
			return nil, err
		}
		buf.Grow(n)
		buf.Write(arr[:n])
		return buf.Bytes(), err
	}
	if def.sqlt == C.SQLT_CLOB {
		length *= 4
	}
	buf.Grow(int(length))
	buf.Write(arr[:n])
	_, err = io.Copy(&buf, r)
	if err == io.EOF {
		err = nil
	}
	return buf.Bytes(), err
}

func (def *defLob) String(offset int) (value string, err error) {
	var bytes []byte
	bytes, err = def.Bytes(offset)
	value = string(bytes)
	return value, err
}

// Reader returns an io.Reader for the underlying LOB.
// Also dissociates this def from the LOB!
func (def *defLob) Reader(offset int) io.ReadCloser {
	lr := &lobReader{
		ses:           def.rset.stmt.ses,
		ociLobLocator: def.lobs[offset],
		piece:         C.OCI_FIRST_PIECE,
	}
	//fmt.Printf("%p.Reader(%d): %p\n", def, offset, lr)
	return lr
}

func (def *defLob) value(offset int) (result interface{}, err error) {
	//lob := def.ociLobLocator
	//Log.Infof("value %p null=%d", lob, def.null)
	isNull := def.nullInds[offset] <= -1

	//defer func() { fmt.Printf("%d gct=%v null=%v =>%#v\n", offset, def.gct, isNull, result) }()
	switch def.gct {
	case Bin:
		if isNull {
			return nil, nil
		}
		return def.Bytes(offset)
	case OraBin:
		if isNull {
			return Raw{IsNull: true}, nil
		}
		b, err := def.Bytes(offset)
		return Raw{Value: b}, err

	case S:
		if isNull {
			return "", nil
		}
		//fmt.Printf("offset=%d\n", offset)
		b, err := def.Bytes(offset)
		return string(b), err
	case OraS:
		if isNull {
			return String{IsNull: true}, nil
		}
		//fmt.Printf("offset=%d\n", offset)
		b, err := def.Bytes(offset)
		return String{Value: string(b)}, err

	default: // D or L
		if isNull {
			return (*Lob)(nil), nil
		}
		r := def.Reader(offset)
		return &Lob{Reader: r}, nil
	}
}

func (def *defLob) alloc() error {
	// Allocate lob locator handle
	// For a LOB define, the buffer pointer must be a pointer to a LOB locator of type OCILobLocator, allocated by the OCIDescriptorAlloc() call.
	// OCI_DTYPE_LOB is for a BLOB or CLOB
	for i := range def.lobs {
		r := C.OCIDescriptorAlloc(
			unsafe.Pointer(def.rset.stmt.ses.srv.env.ocienv), //CONST dvoid   *parenth,
			(*unsafe.Pointer)(unsafe.Pointer(&def.lobs[i])),  //dvoid         **descpp,
			C.OCI_DTYPE_LOB,                                  //ub4           type,
			0,                                                //size_t        xtramem_sz,
			nil)                                              //dvoid         **usrmempp);
		if r == C.OCI_ERROR {
			return def.rset.stmt.ses.srv.env.ociError("LOB OCIDescriptorAlloc")
		} else if r == C.OCI_INVALID_HANDLE {
			return errNew("unable to allocate oci lob handle during define")
		}
	}
	return nil
}

func (def *defLob) free() {
	for i, lob := range def.lobs {
		if lob == nil {
			continue
		}
		def.lobs[i] = nil
		lobClose(def.rset.stmt.ses, lob)
		C.OCIDescriptorFree(
			unsafe.Pointer(lob), //void     *descp,
			C.OCI_DTYPE_LOB)     //ub4      type );
	}
}

func (def *defLob) close() (err error) {
	//Log.Infof("defLob close %p", def.ociLobLocator)
	def.free()
	if def.lobs != nil {
		C.free(unsafe.Pointer(unsafe.Pointer(&def.lobs[0])))
		def.lobs = nil
	}
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.arrHlp.close()
	rset.putDef(defIdxLob, def)

	return nil
}

var _ = io.Reader((*lobReader)(nil))
var _ = io.WriterTo((*lobReader)(nil))

type lobReader struct {
	ses           *Ses
	ociLobLocator *C.OCILobLocator
	piece, csfrm  C.ub1
	csid          C.ub2
	off           C.oraub8
	interrupted   bool
	opened        bool

	// Length is the underlying LOB's length.
	// It is 0 before the first Read call!
	Length C.oraub8
}

// Close the LOB reader.
func (lr *lobReader) Close() error {
	if lr == nil || lr.ociLobLocator == nil {
		return nil
	}
	lob, ses := lr.ociLobLocator, lr.ses
	lr.ociLobLocator, lr.ses = nil, nil
	if lr.interrupted {
		ses.log(_drv.Cfg().Log.Ses.Close, "lobClose interrupted")
		ses.Break()
	}
	//Log.Infof("lobReader OCILobClose %p", lr.ociLobLocator)
	return lobClose(ses, lob)
}

// Read into p, the next chunk.
// Will open the LOB at the first call.
func (lr *lobReader) Read(p []byte) (n int, err error) {
	if lr == nil || lr.ociLobLocator == nil {
		return 0, io.EOF
	}
	if !lr.opened {
		lr.opened = true
		// Open the lob to obtain length; round-trip to database
		//Log.Infof("Reader OCILobOpen %p", def.ociLobLocator)
		//fmt.Printf("lobOpen(%p loc=%p)\n", lr, lr.ociLobLocator)
		if lr.Length, lr.csid, lr.csfrm, err = lobOpen(lr.ses, lr.ociLobLocator, C.OCI_LOB_READONLY); err != nil {
			return 0, err
		}
	}

	if lr.Length == 0 || lr.off >= lr.Length {
		return 0, io.EOF
	}
	defer func() {
		if err != nil {
			lr.Close()
		}
	}()

	var byteAmt C.oraub8 // zero
	lr.ses.logF(_drv.Cfg().Log.Ses.Close, "OCILobRead2(%p) piece=%d off=%d amt=%d length=%d\n", lr.ociLobLocator, lr.piece, lr.off, len(p), lr.Length)
	r := C.OCILobRead2(
		lr.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lr.ses.srv.env.ocierr, //OCIError           *errhp,
		lr.ociLobLocator,      //OCILobLocator      *locp,
		&byteAmt,              //oraub8             *byteAmtp,
		nil,                   //oraub8             *char_amtp,
		lr.off+1,              //oraub8             offset, offset is 1-based
		unsafe.Pointer(&p[0]), //void               *bufp,
		C.oraub8(len(p)),      //oraub8             bufl,
		lr.piece,              //ub1                piece,
		nil,                   //void               *ctxp,
		nil,                   //OCICallbackLobRead2 (cbfp)
		lr.csid,               //C.ub2(0),              //ub2                csid,
		lr.csfrm,              //lr.charsetForm,                          //ub1                csfrm );
	)
	//Log.Infof("LobRead2 returned %d amt=%d", r, byteAmt)
	err = nil
	switch r {
	case C.OCI_ERROR:
		lr.interrupted = true
		err = lr.ses.srv.env.ociError("OCILobRead2")
	case C.OCI_NO_DATA:
		err = io.EOF
	case C.OCI_INVALID_HANDLE:
		err = fmt.Errorf("Invalid handle %v", lr.ociLobLocator)
	}
	lr.ses.logF(_drv.Cfg().Log.Ses.Close, "OCILobRead2(%p) off=%d amt=%d csid=%d csfrm=%d err=%v\n", lr.ociLobLocator, lr.off, byteAmt, lr.csid, lr.csfrm, err)
	// byteAmt represents the amount copied into buffer by oci
	if byteAmt != 0 {
		lr.off += byteAmt
		if lr.piece == C.OCI_FIRST_PIECE {
			lr.piece = C.OCI_NEXT_PIECE
		}

		if lr.csid == 2000 && byteAmt > 1 {
			// UTF-16
			u16 := make([]uint16, int(byteAmt/2))
			binary.Read(bytes.NewReader(p[:int(byteAmt)]), binary.BigEndian, &u16)
			return copy(p[:], string(utf16.Decode(u16))), err
		}
	}
	return int(byteAmt), err
}

// WriteTo writes all data from the LOB into the given Writer.
func (lr *lobReader) WriteTo(w io.Writer) (n int64, err error) {
	defer func() {
		if closeErr := lr.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	arr := lobChunkPool.Get().([lobChunkSize]byte)
	defer lobChunkPool.Put(arr)

	for {
		k, err := lr.Read(arr[:])
		if k > 0 {
			if _, err := w.Write(arr[:k]); err != nil {
				return n, err
			}
		}
		n += int64(k)
		if err != nil {
			return n, err
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
	size          C.oraub8
	csid          C.ub2
	csfrm         C.ub1
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
	if lrw.csid == 0 {
		lrw.csid = C.ub2(atomic.LoadUint32(&csIDAl32UTF8))
	}
	byteAmt := C.oraub8(len(p))
	//Log.Infof("LobRead2 off=%d amt=%d", off, len(p))
	r := C.OCILobRead2(
		lrw.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lrw.ses.srv.env.ocierr, //OCIError           *errhp,
		lrw.ociLobLocator,      //OCILobLocator      *locp,
		&byteAmt,               //oraub8             *byteAmtp,
		nil,                    //oraub8             *char_amtp,
		C.oraub8(off)+1,        //oraub8             offset, offset is 1-based
		unsafe.Pointer(&p[0]),  //void               *bufp,
		C.oraub8(len(p)),       //oraub8             bufl,
		C.OCI_ONE_PIECE,        //ub1                piece,
		nil,                    //void               *ctxp,
		nil,                    //OCICallbackLobRead2 (cbfp)
		lrw.csid,               //ub2                csid,
		lrw.csfrm,              //ub1                csfrm );
	)
	//Log.Infof("LobRead2 returned %d amt=%d", r, byteAmt)
	switch r {
	case C.OCI_ERROR:
		return 0, lrw.ses.srv.env.ociError()
	case C.OCI_NO_DATA:
		return int(byteAmt), io.EOF
	case C.OCI_INVALID_HANDLE:
		return 0, fmt.Errorf("Invalid handle %v", lrw.ociLobLocator)
	}
	return int(byteAmt), nil
}

// WriteAt writes data in p into the LOB, starting at off.
func (lrw *lobReadWriter) WriteAt(p []byte, off int64) (n int, err error) {
	//Log.Infof("LobWrite2 off=%d len=%d", off, n)
	byteAmt := C.oraub8(len(p))
	// Write to Oracle
	//fmt.Println("OCILobWrite2")
	if C.OCILobWrite2(
		lrw.ses.ocisvcctx,      //OCISvcCtx          *svchp,
		lrw.ses.srv.env.ocierr, //OCIError           *errhp,
		lrw.ociLobLocator,      //OCILobLocator      *locp,
		&byteAmt,               //oraub8          *byteAmtp,
		nil,                    //oraub8          *char_amtp,
		C.oraub8(off)+1,        //oraub8          offset, starting position is 1
		unsafe.Pointer(&p[0]),  //void            *bufp,
		C.oraub8(len(p)),
		C.OCI_ONE_PIECE, //ub1             piece,
		nil,             //void            *ctxp,
		nil,             //OCICallbackLobWrite2 (cbfp)
		C.ub2(atomic.LoadUint32(&csIDAl32UTF8)), //ub2             csid,
		C.SQLCS_IMPLICIT,                        //ub1             csfrm );
	//fmt.Printf("C.OCI_NEED_DATA %v, C.OCI_SUCCESS %v\n", C.OCI_NEED_DATA, C.OCI_SUCCESS)
	) == C.OCI_ERROR {
		return 0, lrw.ses.srv.env.ociError()
	}
	//fmt.Printf("r %v, current %v, buffer %v\n", r, current, buffer)
	if C.oraub8(off)+byteAmt > lrw.size {
		lrw.size = C.oraub8(off) + byteAmt
	}
	return int(byteAmt), nil
}

func lobOpen(ses *Ses, lob *C.OCILobLocator, mode C.ub1) (
	length C.oraub8, csid C.ub2, csfrm C.ub1, err error,
) {
	ses.RLock()
	ocisvcctx := ses.ocisvcctx
	env := ses.srv.env
	ses.RUnlock()

	ses.log(_drv.Cfg().Log.Ses.Prep, "lobOpen")
	// reopen
	_ = C.OCILobClose(
		ocisvcctx,  //OCISvcCtx          *svchp,
		env.ocierr, //OCIError           *errhp,
		lob,        //OCILobLocator      *locp,
	)
	//Log.Infof("OCILobOpen %p\n%s", lob, getStack(1))
	r := C.OCILobOpen(
		ses.ocisvcctx, //OCISvcCtx          *svchp,
		env.ocierr,    //OCIError           *errhp,
		lob,           //OCILobLocator      *locp,
		mode,          //ub1              mode );
	)
	//Log.Infof("OCILobOpen %p returned %d", lob, r)
	if r != C.OCI_SUCCESS {
		lobClose(ses, lob)
		return 0, csid, csfrm, ses.srv.env.ociError("OCILobOpen")
	}
	// get the length of the lob
	// For character LOBs, it is the number of characters; for binary LOBs and BFILEs,
	// it is the number of bytes in the LOB.<Paste>
	if r = C.OCILobGetLength2(
		ses.ocisvcctx, //OCISvcCtx          *svchp,
		env.ocierr,    //OCIError           *errhp,
		lob,           //OCILobLocator      *locp,
		&length,       //oraub8 *lenp)
	); r == C.OCI_ERROR {
		lobClose(ses, lob)
		return length, csid, csfrm, env.ociError("OCILobGetLength2")
	}

	if r = C.OCILobCharSetId(
		env.ocienv,
		env.ocierr,
		lob,
		&csid,
	); r == C.OCI_ERROR {
		lobClose(ses, lob)
		return length, csid, csfrm, env.ociError("OCILobCharSetId")
	}
	if csid == 0 {
		csid = C.ub2(atomic.LoadUint32(&csIDAl32UTF8))
	}
	if r = C.OCILobCharSetForm(
		env.ocienv,
		env.ocierr,
		lob,
		&csfrm,
	); r == C.OCI_ERROR {
		lobClose(ses, lob)
		return length, csid, csfrm, env.ociError("OCILobCharSetForm")
	}
	return length, csid, csfrm, nil
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
