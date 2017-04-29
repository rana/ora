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
	"time"
	"unsafe"

	"gopkg.in/rana/ora.v4/date"
)

type defDate struct {
	ociDef
	ociDate    []date.Date
	isNullable bool
	timezone   *time.Location
}

func (def *defDate) define(position int, isNullable bool, rset *Rset) error {
	var err error
	if def.timezone, err = rset.stmt.ses.Timezone(); err != nil {
		return err
	}
	def.rset = rset
	def.isNullable = isNullable
	if def.ociDate != nil {
		C.free(unsafe.Pointer(&def.ociDate[0]))
	}
	def.ociDate = (*((*[MaxFetchLen]date.Date)(C.malloc(C.size_t(rset.fetchLen) * 7))))[:rset.fetchLen]

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociDate[0]), 7, C.SQLT_DAT)
}

func (def *defDate) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Time{IsNull: true}, nil
		}
		return nil, nil
	}
	if def.isNullable {
		return Time{Value: def.ociDate[offset].GetIn(def.timezone)}, nil
	}
	return def.ociDate[offset].GetIn(def.timezone), nil
}

func (def *defDate) alloc() error { return nil }
func (def *defDate) free() {
	if def.ociDate != nil {
		C.free(unsafe.Pointer(&def.ociDate[0]))
		def.ociDate = nil
	}
	def.arrHlp.close()
}

func (def *defDate) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.free()
	rset.putDef(defIdxDate, def)
	return nil
}
