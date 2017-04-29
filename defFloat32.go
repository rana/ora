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

// defFloat32.go is generated from defFloat32.go!

type defFloat32 struct {
	ociDef
	ociNumber  []C.OCINumber
	isNullable bool
}

func (def *defFloat32) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
	}
	def.ociNumber = (*((*[MaxFetchLen]C.OCINumber)(C.malloc(C.size_t(rset.fetchLen) * C.sizeof_OCINumber))))[:rset.fetchLen]
	return def.ociDef.defineByPos(position, unsafe.Pointer(&def.ociNumber[0]), C.sizeof_OCINumber, C.SQLT_VNU)
}

func (def *defFloat32) value(offset int) (value interface{}, err error) {
	if def.nullInds[offset] < 0 {
		if def.isNullable {
			return Float32{IsNull: true}, nil
		}
		return nil, nil
	}
	var float32Value float32
	on := def.ociNumber[offset]
	r := C.OCINumberToReal(
		def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
		&on,                           //const OCINumber     *number,
		byteWidth32,                   //uword               rsl_length,
		unsafe.Pointer(&float32Value)) //void                *rsl );
	if r == C.OCI_ERROR {
		err = def.rset.stmt.ses.srv.env.ociError()
	}
	//fmt.Printf("%d. %#v = %#v\n", offset, on, float32Value)
	if def.isNullable {
		return Float32{Value: float32Value}, err
	}
	return float32Value, err
}

func (def *defFloat32) alloc() error {
	return nil
}

func (def *defFloat32) free() {
	def.arrHlp.close()
}

func (def *defFloat32) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	if def.ociNumber != nil {
		C.free(unsafe.Pointer(&def.ociNumber[0]))
		def.ociNumber = nil
	}
	rset.putDef(defIdxFloat32, def)
	return nil
}
