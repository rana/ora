// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	//	"fmt"
	"unsafe"
)

type oraInt64Define struct {
	env       *Environment
	ocidef    *C.OCIDefine
	ociNumber C.OCINumber
	isNull    C.sb2
}

func (d *oraInt64Define) define(position int, ocistmt *C.OCIStmt) error {
	r := C.OCIDefineByPos2(
		ocistmt,                      //OCIStmt     *stmtp,
		&d.ocidef,                    //OCIDefine   **defnpp,
		d.env.ocierr,                 //OCIError    *errhp,
		C.ub4(position),              //ub4         position,
		unsafe.Pointer(&d.ociNumber), //void        *valuep,
		C.sb8(C.sizeof_OCINumber),    //sb8         value_sz,
		C.SQLT_VNU,                   //ub2         dty,
		unsafe.Pointer(&d.isNull),    //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return d.env.ociError()
	}
	return nil
}

func (d *oraInt64Define) value() (value interface{}, err error) {
	int64Value := Int64{IsNull: d.isNull < 0}
	if !int64Value.IsNull {
		r := C.OCINumberToInt(
			d.env.ocierr,                      //OCIError              *err,
			&d.ociNumber,                      //const OCINumber       *number,
			C.uword(8),                        //uword                 rsl_length,
			C.OCI_NUMBER_SIGNED,               //uword                 rsl_flag,
			unsafe.Pointer(&int64Value.Value)) //void                  *rsl );
		if r == C.OCI_ERROR {
			err = d.env.ociError()
		}
	}
	value = int64Value
	return value, err
}

func (oraInt64Define *oraInt64Define) alloc() error {
	return nil
}

func (d *oraInt64Define) free() {
	defer func() {
		recover()
	}()
}

func (d *oraInt64Define) close() {
	defer func() {
		recover()
	}()
	d.ocidef = nil
	d.isNull = C.sb2(0)
	d.env.oraInt64DefinePool.Put(d)
}
