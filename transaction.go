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
	"container/list"
)

// A transaction associated with an Oracle server.
//
// Implements the driver.Tx interface.
type Transaction struct {
	ses  *Session
	elem *list.Element
}

// Commit commits a transaction.
//
// Commit is a member of the driver.Tx interface.
func (tx *Transaction) Commit() (err error) {
	// Validate that the session is open
	err = tx.ses.checkIsOpen()
	if err != nil {
		return err
	}

	r := C.OCITransCommit(
		tx.ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		tx.ses.srv.env.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)         //ub4          flags );

	tx.ses.txs.Remove(tx.elem)

	if r == C.OCI_ERROR {
		return tx.ses.srv.env.ociError()
	}
	return nil
}

// Rollback rolls back a transaction.
//
// Rollback is a member of the driver.Tx interface.
func (tx *Transaction) Rollback() (err error) {
	// Validate that the session is open
	err = tx.ses.checkIsOpen()
	if err != nil {
		return err
	}

	r := C.OCITransRollback(
		tx.ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		tx.ses.srv.env.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)         //ub4          flags );

	tx.ses.txs.Remove(tx.elem)

	if r == C.OCI_ERROR {
		return tx.ses.srv.env.ociError()
	}
	return nil
}
