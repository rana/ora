// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import "unsafe"

type defDate struct {
	ociDef
	ociDate    [MaxFetchLen]C.OCIDate
	isNullable bool
}

func (def *defDate) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable

	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociDate[0]), C.sizeof_OCIDate, C.SQLT_ODT)
}

func (def *defDate) value(offset int) (value interface{}, err error) {
	if def.isNullable {
		oraTimeValue := Date{IsNull: def.nullInds[offset] < 0}
		if !oraTimeValue.IsNull {
			oraTimeValue.Value = ociGetDateTime(def.ociDate[offset])
		}
		return oraTimeValue, nil
	}
	return Date{Value: ociGetDateTime(def.ociDate[0])}, nil
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
	def.arrHlp.close()
	rset.putDef(defIdxDate, def)
	return nil
}
