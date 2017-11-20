// +build go1.8

// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

/*
#include <oci.h>
*/
import "C"

var (
	// Ensure that Con implements the needed ...Context interfaces.
	_ = driver.Conn((*Con)(nil))
	_ = driver.ConnBeginTx((*Con)(nil))
	_ = driver.ConnPrepareContext((*Con)(nil))
	_ = driver.Pinger((*Con)(nil))

	// Ensure that DrvStmt implements the needed ...Context interfaces.
	_ = driver.Stmt((*DrvStmt)(nil))
	_ = driver.StmtQueryContext((*DrvStmt)(nil))
	_ = driver.StmtExecContext((*DrvStmt)(nil))
)

// Prepare readies a sql string for use.
//
// Prepare is a member of the driver.Conn interface.
func (con *Con) Prepare(query string) (driver.Stmt, error) {
	return con.PrepareContext(context.Background(), query)
}

// PrepareContext returns a prepared statement, bound to this connection.
// context is for the preparation of the statement,
// it must not store the context within the statement itself.
func (con *Con) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	con.log(_drv.Cfg().Log.Con.Prepare)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	stmt, err := con.ses.Prep(query)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return &DrvStmt{stmt: stmt}, err
}

// BeginTx starts and returns a new transaction.
// The provided context should be used to roll the transaction back
// if it is cancelled.
//
// If the driver does not support setting the isolation
// level and one is set or if there is a set isolation level
// but the set level is not supported, an error must be returned.
//
// If the read-only value is true to either
// set the read-only transaction property if supported
// or return an error if it is not supported.
func (con *Con) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var flags C.ub4
	if opts.ReadOnly {
		flags |= C.OCI_TRANS_READONLY
	}
	switch level := sql.IsolationLevel(opts.Isolation); level {
	case sql.LevelDefault, sql.LevelReadCommitted:
		// this is the default level
	case sql.LevelSerializable:
		flags |= C.OCI_TRANS_SERIALIZABLE
	default:
		return nil, fmt.Errorf("Isolation level %v not supported.", level)
	}
	con.log(_drv.Cfg().Log.Con.Begin)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	var tx *Tx
	done := make(chan error, 1)
	go func() {
		defer close(done)
		var err error
		tx, err = con.ses.StartTx(TxFlags(uint32(flags)))
		done <- err
	}()
	var err error
	select {
	case err = <-done:
		return tx, err
	case <-ctx.Done():
		// select again to avoid race condition if both are done
		select {
		case err = <-done:
			return tx, err
		default:
			if err = ctx.Err(); isCanceled(err) {
				con.ses.Break()
			}
		}
	}
	return nil, maybeBadConn(err)
}

// vim: set fileencoding=utf-8 noet:
