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
	"os"
	"strings"
	"sync"
	"unsafe"
)

type defString struct {
	ociDef
	sync.RWMutex
	buf               []byte
	isNullable, rTrim bool
	columnSize        int
}

type defNumString struct {
	defString
}

func (def *defNumString) define(position int, isNullable bool, rset *Rset) error {
	return def.defString.define(position, 40, isNullable, false, rset)
}

func (def *defString) define(position int, columnSize int, isNullable, rTrim bool, rset *Rset) error {
	def.Lock()
	defer def.Unlock()
	def.rset = rset
	def.isNullable, def.rTrim = isNullable, rTrim
	//Log.Infof("defString position=%d columnSize=%d", position, columnSize)
	n := columnSize
	// AL32UTF8: one db "char" can be 4 bytes on wire, esp. if the database's
	// character set is not AL32UTF8 (e.g. some 8bit fixed width charset), and
	// the column is VARCHAR2 with byte semantics.
	//
	// For example when the db's charset is EE8ISO8859P2, then a VARCHAR2(1) can
	// contain an "ű", which is 2 bytes AL32UTF8.
	rset.stmt.RLock()
	rset.stmt.ses.RLock()
	isUTF8 := rset.stmt.ses.srv.IsUTF8()
	rset.stmt.ses.RUnlock()
	rset.stmt.RUnlock()
	if !isUTF8 {
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
		//def.buf = make([]byte, n)
		def.buf = bytesPool.Get(n)
	} else {
		def.buf = def.buf[:n]
	}

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.buf[0]), def.columnSize, C.SQLT_CHR)
}

func (def *defString) value(offset int) (value interface{}, err error) {
	def.RLock()
	defer def.RUnlock()
	if offset < 0 || offset >= len(def.nullInds) {
		fmt.Fprintf(os.Stderr, "offset=%d nullInds=%d\n", offset, len(def.nullInds))
	}
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return String{IsNull: true}, nil
		}
		return "", nil
	}
	var s string
	//def.rset.logF(_drv.Cfg().Log.Stmt.Bind,
	//	"%p offset=%d alen=%v, colSize=%d, buf=%v",
	//	def, offset, def.alen, def.columnSize, def.buf[offset*def.columnSize:offset*def.columnSize+int(def.alen[offset])])
	if def.alen[offset] > 0 {
		off := offset * def.columnSize
		s = string(def.buf[off : off+int(def.alen[offset])])
		if def.rTrim {
			s = strings.TrimRight(s, " ")
		}
	}
	if def.isNullable {
		return String{Value: s}, nil
	}
	return s, nil
}

func (def *defString) alloc() error {
	return nil
}

func (def *defString) free() {
	def.Lock()
	def.arrHlp.close()
	if def.buf != nil {
		bytesPool.Put(def.buf)
		def.buf = nil
	}
	def.Unlock()
}

func (def *defString) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	def.free()
	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	rset.putDef(defIdxString, def)
	return nil
}
