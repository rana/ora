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
	"bytes"
	"unsafe"
)

const rowidLen = 19

type defRowid struct {
	ociDef
	buf []byte
}

func (def *defRowid) define(position int, rset *Rset) error {
	def.rset = rset
	// using a character host variable of width between 19
	// (18 bytes plus the null-terminator) and 4001 as the
	// host bind variable for universal ROWID.
	if n := rset.fetchLen * rowidLen; cap(def.buf) < n {
		//def.buf = make([]byte, n)
		def.buf = bytesPool.Get(n)
	} else {
		def.buf = def.buf[:n]
	}
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), rowidLen, C.SQLT_STR)
}

func (def *defRowid) value(offset int) (value interface{}, err error) {
	n := bytes.Index(def.buf[offset*rowidLen:(offset+1)*rowidLen], []byte{0})
	if n == -1 {
		n = rowidLen
	}
	value = string(def.buf[offset*rowidLen : offset*rowidLen+n])
	return value, err
}

func (def *defRowid) alloc() error { return nil }
func (def *defRowid) free() {
	def.arrHlp.close()
	if def.buf != nil {
		bytesPool.Put(def.buf)
		def.buf = nil
	}
}

func (def *defRowid) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.free()
	rset.putDef(defIdxRowid, def)
	return nil
}
