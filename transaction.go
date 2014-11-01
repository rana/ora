// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// A transaction associated with an Oracle server.
//
// Implements the driver.Tx interface.
type Transaction struct {
	ses *Session
}

// Commit commits a transaction.
//
// Commit is a member of the driver.Tx interface.
func (tx *Transaction) Commit() (err error) {
	return tx.ses.CommitTransaction()
}

// Rollback rolls back a transaction.
//
// Rollback is a member of the driver.Tx interface.
func (tx *Transaction) Rollback() (err error) {
	return tx.ses.RollbackTransaction()
}
