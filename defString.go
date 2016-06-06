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
import "unsafe"

type defString struct {
	ociDef
	buf        []byte
	isNullable bool
	columnSize int
}

func (def *defString) define(position int, columnSize int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	//Log.Infof("defString position=%d columnSize=%d", position, columnSize)
	n := columnSize
	// AL32UTF8: one db "char" can be 4 bytes on wire, esp. if the database's
	// character set is not AL32UTF8 (e.g. some 8bit fixed width charset), and
	// the column is VARCHAR2 with byte semantics.
	//
	// For example when the db's charset is EE8ISO8859P2, then a VARCHAR2(1) can
	// contain an "ű", which is 2 bytes AL32UTF8.
	if !rset.stmt.ses.srv.dbIsUTF8 {
		n *= 2
	}
	if n == 0 {
		n = 2
	}
	if n%2 != 0 {
		n++
	}
	def.columnSize = n
	if n := rset.fetchLen * def.columnSize; cap(def.buf) < n {
		def.buf = make([]byte, n)
	} else {
		def.buf = def.buf[:n]
	}

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), def.columnSize, C.SQLT_CHR)
}

func (def *defString) value(offset int) (value interface{}, err error) {
	if def.isNullable {
		oraStringValue := String{IsNull: def.nullInds[offset] < 0}
		if !oraStringValue.IsNull && def.alen[offset] > 0 {
			off := offset * def.columnSize
			oraStringValue.Value = string(def.buf[off : off+int(def.alen[offset])])
		}
		return oraStringValue, nil
	}
	if def.nullInds[offset] < 0 || def.alen[offset] <= 0 {
		return "", nil
	}
	off := offset * def.columnSize
	return string(def.buf[off : off+int(def.alen[offset])]), nil
}

func (def *defString) alloc() error {
	return nil
}

func (def *defString) free() {
}

func (def *defString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.buf = nil
	def.arrHlp.close()
	rset.putDef(defIdxString, def)
	return nil
}
