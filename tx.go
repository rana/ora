// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
import "C"
import (
	"container/list"
	"fmt"
)

// Tx represents an Oracle transaction associated with a session.
//
// Implements the driver.Tx interface.
type Tx struct {
	id   uint64
	ses  *Ses
	elem *list.Element
}

// checkIsOpen validates that the session is open.
func (tx *Tx) checkIsOpen() error {
	if tx.ses != nil {
		return errNewF("Tx is closed (id %v)", tx.id)
	}
	return nil
}

func (tx *Tx) close() {
	if tx.ses != nil {
		tx.ses.txs.Remove(tx.elem)
		tx.ses = nil
		tx.elem = nil
	}
}

// Commit commits the transaction.
//
// Commit is a member of the driver.Tx interface.
func (tx *Tx) Commit() (err error) {
	if tx.checkIsOpen(); err != nil {
		return err
	}
	defer tx.close()
	r := C.OCITransCommit(
		tx.ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		tx.ses.srv.env.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)         //ub4          flags );
	if r == C.OCI_ERROR {
		return tx.ses.srv.env.ociError()
	}
	return nil
}

// Rollback rolls back a transaction.
//
// Rollback is a member of the driver.Tx interface.
func (tx *Tx) Rollback() (err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if tx.checkIsOpen(); err != nil {
		return err
	}
	defer tx.close()
	r := C.OCITransRollback(
		tx.ses.srv.ocisvcctx,  //OCISvcCtx    *svchp,
		tx.ses.srv.env.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)         //ub4          flags );
	if r == C.OCI_ERROR {
		return tx.ses.srv.env.ociError()
	}
	return nil
}
