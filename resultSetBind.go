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
	"unsafe"
)

type resultSetBind struct {
	environment *Environment
	statement   *Statement
	ocibnd      *C.OCIBind
	isNull      C.sb2
	value       *ResultSet
	ocistmt     *C.OCIStmt
}

func (resultSetBind *resultSetBind) bind(value *ResultSet, position int, statement *Statement) error {
	resultSetBind.statement = statement
	resultSetBind.value = value
	// Allocate a statement handle
	ocistmt, err := resultSetBind.environment.allocateOciHandle(C.OCI_HTYPE_STMT)
	resultSetBind.ocistmt = (*C.OCIStmt)(ocistmt)
	if err != nil {
		return err
	}
	r := C.OCIBindByPos2(
		statement.ocistmt,                      //OCIStmt      *stmtp,
		(**C.OCIBind)(&resultSetBind.ocibnd),   //OCIBind      **bindpp,
		resultSetBind.environment.ocierr,       //OCIError     *errhp,
		C.ub4(position),                        //ub4          position,
		unsafe.Pointer(&resultSetBind.ocistmt), //void         *valuep,
		C.sb8(0),                              //sb8          value_sz,
		C.SQLT_RSET,                           //ub2          dty,
		unsafe.Pointer(&resultSetBind.isNull), //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return resultSetBind.environment.ociError()
	}

	return nil
}

func (resultSetBind *resultSetBind) setPtr() error {
	err := resultSetBind.value.open(resultSetBind.statement, resultSetBind.ocistmt)
	resultSetBind.statement.resultSets.PushBack(resultSetBind.value)
	if err == nil {
		// open result set is successful; will be freed by ResultSet
		resultSetBind.ocistmt = nil
	}

	return err
}

func (resultSetBind *resultSetBind) close() {
	defer func() {
		recover()
	}()
	// release ocistmt handle for failed ResultSet binding
	// ResultSet will release handle for successful bind
	if resultSetBind.ocistmt != nil {
		resultSetBind.environment.freeOciHandle(unsafe.Pointer(resultSetBind.ocistmt), C.OCI_HTYPE_STMT)
	}
	resultSetBind.statement = nil
	resultSetBind.ocibnd = nil
	resultSetBind.isNull = C.sb2(0)
	resultSetBind.environment.resultSetBindPool.Put(resultSetBind)
}
