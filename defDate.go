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
import "gopkg.in/rana/ora.v3/date"

type defDate struct {
	ociDef
	ociDate    []date.Date
	isNullable bool
}

func (def *defDate) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociDate != nil {
		C.free(unsafe.Pointer(&def.ociDate[0]))
	}
	def.ociDate = (*((*[MaxFetchLen]date.Date)(C.malloc(C.size_t(rset.fetchLen) * 7))))[:rset.fetchLen]

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociDate[0]), 7, C.SQLT_DAT)
}

func (def *defDate) value(offset int) (value interface{}, err error) {
	if def.isNullable {
		oraTimeValue := Date{IsNull: def.nullInds[offset] < 0}
		if !oraTimeValue.IsNull {
			oraTimeValue.Value = def.ociDate[offset].Get()
		}
		return oraTimeValue, nil
	}
	return def.ociDate[offset].Get(), nil
}

func (def *defDate) alloc() error { return nil }
func (def *defDate) free()        {}

func (def *defDate) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	if def.ociDate != nil {
		C.free(unsafe.Pointer(&def.ociDate[0]))
		def.ociDate = nil
	}
	def.arrHlp.close()
	rset.putDef(defIdxDate, def)
	return nil
}
