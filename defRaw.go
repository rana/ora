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
	"unsafe"
)

type defRaw struct {
	ociDef
	ociRaw     *C.OCIRaw
	isNullable bool
	buf        []byte
	columnSize int
}

func (def *defRaw) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	def.columnSize = columnSize
	if n := rset.fetchLen * columnSize; cap(def.buf) < n {
		//def.buf = make([]byte, n)
		def.buf = bytesPool.Get(n)
	} else {
		def.buf = def.buf[:n]
	}

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), columnSize, C.SQLT_BIN)
}

func (def *defRaw) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Raw{IsNull: true}, nil
		}
		return nil, nil
	}
	n := int(def.alen[offset])
	off := offset * def.columnSize
	if def.isNullable {
		return Raw{Value: def.buf[off : off+n]}, nil
	}
	return def.buf[off : off+n], nil
}

func (def *defRaw) alloc() error {
	return nil
}

func (def *defRaw) free() {
	def.arrHlp.close()
	if def.buf != nil {
		bytesPool.Put(def.buf)
		def.buf = nil
	}
}

func (def *defRaw) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.ociRaw = nil
	def.free()
	rset.putDef(defIdxRaw, def)
	return nil
}
