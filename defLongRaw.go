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

type defLongRaw struct {
	ociDef
	isNullable bool
	buf        []byte
	bufSize    int
}

func (def *defLongRaw) define(position int, bufSize uint32, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if n := rset.fetchLen * int(bufSize); cap(def.buf) < n {
		//def.buf = make([]byte, n)
		def.buf = bytesPool.Get(n)
	} else {
		def.buf = def.buf[:n]
	}
	def.bufSize = int(bufSize)

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), int(bufSize), C.SQLT_LBI)
}

func (def *defLongRaw) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Raw{IsNull: true}, nil
		}
		return nil, nil
	}
	// Make a slice of length equal to the return length
	result := make([]byte, def.alen[offset])
	// Copy returned data
	copyLength := copy(result, def.buf[offset*def.bufSize:(offset+1)*def.bufSize])
	if C.ACTUAL_LENGTH_TYPE(copyLength) != def.alen[offset] {
		return nil, errNew("unable to copy LONG RAW result data from buffer")
	}

	if def.isNullable {
		return Raw{Value: result}, nil
	}
	return result, nil
}

func (def *defLongRaw) alloc() error { return nil }
func (def *defLongRaw) free() {
	def.arrHlp.close()
	if def.buf != nil {
		bytesPool.Put(def.buf)
		def.buf = nil
	}
}

func (def *defLongRaw) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.free()
	rset.putDef(defIdxLongRaw, def)
	return nil
}
