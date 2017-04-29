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
	"unicode/utf8"
	"unsafe"
)

type defBool struct {
	ociDef
	isNullable bool
	columnSize int
	buf        []byte
}

func (def *defBool) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	def.columnSize = columnSize
	if n := rset.fetchLen * columnSize; cap(def.buf) < n {
		//def.buf = make([]byte, n)
		def.buf = bytesPool.Get(n)
	} else {
		def.buf = def.buf[:n]
	}
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), columnSize, C.SQLT_AFC)
}

func (def *defBool) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Bool{IsNull: true}, nil
		}
		return nil, nil
	}
	//Log.Infof("%v.value", def)
	buf := def.buf[offset*def.columnSize : (offset+1)*def.columnSize]
	if def.isNullable {
		r, _ := utf8.DecodeRune(buf)
		return Bool{Value: r == def.rset.stmt.Cfg().TrueRune}, nil
	}
	r, _ := utf8.DecodeRune(buf)
	return r == def.rset.stmt.Cfg().TrueRune, nil
}

func (def *defBool) alloc() error {
	return nil
}

func (def *defBool) free() {
	if def.buf != nil {
		bytesPool.Put(def.buf)
		def.buf = nil
	}
	def.arrHlp.close()
}

func (def *defBool) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxBool, def)
	return nil
}
