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
	"fmt"
	"sync"
)

// LogTxCfg represents Tx logging configuration values.
type LogTxCfg struct {
	// Commit determines whether the Tx.Commit method is logged.
	//
	// The default is true.
	Commit bool

	// Rollback determines whether the Tx.Rollback method is logged.
	//
	// The default is true.
	Rollback bool
}

// NewLogTxCfg creates a LogTxCfg with default values.
func NewLogTxCfg() LogTxCfg {
	c := LogTxCfg{}
	c.Commit = true
	c.Rollback = true
	return c
}

// Tx represents an Oracle transaction associated with a session.
//
// Implements the driver.Tx interface.
type Tx struct {
	id  uint64
	ses *Ses
	mu  sync.Mutex
}

// checkIsOpen validates that the session is open.
func (tx *Tx) checkIsOpen() error {
	if tx == nil || tx.ses == nil {
		return er("Tx is closed.")
	}
	return tx.ses.checkClosed()
}

// closeWithRemove releases allocated resources and removes the Tx from the
// Ses.openTxss list.
func (tx *Tx) closeWithRemove() (err error) {
	tx.ses.openTxs.remove(tx)
	return tx.close()
}

// close releases allocated resources.
func (tx *Tx) close() (err error) {
	if tx.ses != nil {
		tx.ses = nil
		_drv.txPool.Put(tx)
	}
	return nil
}

// Commit commits the transaction.
//
// Commit is a member of the driver.Tx interface.
func (tx *Tx) Commit() (err error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx == nil {
		return nil
	}
	tx.log(_drv.cfg.Log.Tx.Commit)
	if err = tx.checkIsOpen(); err != nil {
		return err
	}
	defer tx.closeWithRemove()
	r := C.OCITransCommit(
		tx.ses.ocisvcctx,      //OCISvcCtx    *svchp,
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
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx == nil {
		return nil
	}
	tx.log(_drv.cfg.Log.Tx.Rollback)
	if err = tx.checkIsOpen(); err != nil {
		return err
	}
	if tx.ses == nil || tx.ses.srv == nil {
		return nil
	}
	defer tx.closeWithRemove()
	r := C.OCITransRollback(
		tx.ses.ocisvcctx,      //OCISvcCtx    *svchp,
		tx.ses.srv.env.ocierr, //OCIError     *errhp,
		C.OCI_DEFAULT)         //ub4          flags );
	if r == C.OCI_ERROR {
		return tx.ses.srv.env.ociError()
	}
	return nil
}

// sysName returns a string representing the Tx.
func (tx *Tx) sysName() string {
	if tx == nil {
		return "E_S_S_T_"
	}
	return tx.ses.sysName() + fmt.Sprintf("T%v", tx.id)
}

// log writes a message with an Tx system name and caller info.
func (tx *Tx) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", tx.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", tx.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with an Tx system name and caller info.
func (tx *Tx) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", tx.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", tx.sysName(), callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}
