// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// A transaction associated with an Oracle server.
//
// Implements the driver.Tx interface.
type Transaction struct {
	session *Session
}

// Commit commits a transaction.
//
// Commit is a member of the driver.Tx interface.
func (transaction *Transaction) Commit() (err error) {
	return transaction.session.CommitTransaction()
}

// Rollback rolls back a transaction.
//
// Rollback is a member of the driver.Tx interface.
func (transaction *Transaction) Rollback() (err error) {
	return transaction.session.RollbackTransaction()
}
