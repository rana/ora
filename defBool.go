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
		def.buf = make([]byte, n)
	} else {
		def.buf = def.buf[:n]
	}
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), columnSize, C.SQLT_AFC)
}

func (def *defBool) value(offset int) (value interface{}, err error) {
	//Log.Infof("%v.value", def)
	buf := def.buf[offset*def.columnSize : (offset+1)*def.columnSize]
	if def.isNullable {
		oraBoolValue := Bool{IsNull: def.nullInds[offset] < 0}
		if !oraBoolValue.IsNull {
			r, _ := utf8.DecodeRune(buf)
			oraBoolValue.Value = r == def.rset.stmt.cfg.Rset.TrueRune
		}
		return oraBoolValue, nil
	}
	if def.nullInds[offset] > -1 {
		r, _ := utf8.DecodeRune(buf)
		return r == def.rset.stmt.cfg.Rset.TrueRune, nil
	}
	// NULL is false, too
	return false, nil
}

func (def *defBool) alloc() error {
	return nil
}

func (def *defBool) free() {
}

func (def *defBool) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.arrHlp.close()
	clear(def.buf, 0)
	rset.putDef(defIdxBool, def)
	return nil
}
