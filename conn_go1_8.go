// +build go1.8

// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"golang.org/x/sync/errgroup"
)

/*
#include <oci.h>
*/
import "C"

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
		return nil, err
	}
	return &DrvStmt{stmt: stmt}, err
}

// BeginContext starts and returns a new transaction.
// The provided context should be used to roll the transaction back
// if it is cancelled.
//
// This must call IsolationFromContext to determine if there is a set
// isolation level. If the driver does not support setting the isolation
// level and one is set or if there is a set isolation level
// but the set level is not supported, an error must be returned.
//
// This must also call ReadOnlyFromContext to determine if the read-only
// value is true to either set the read-only transaction property if supported
// or return an error if it is not supported.
func (con *Con) BeginContext(ctx context.Context) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var flags C.ub4
	if driver.ReadOnlyFromContext(ctx) {
		flags |= C.OCI_TRANS_READONLY
	}
	level, ok := driver.IsolationFromContext(ctx)
	if ok {
		switch sql.IsolationLevel(level) {
		case sql.LevelDefault, sql.LevelReadCommitted:
			// this is the default level
		case sql.LevelSerializable:
			flags |= C.OCI_TRANS_SERIALIZABLE
		default:
			return nil, fmt.Errorf("Isolation level %s not supported.", level)
		}
	}
	con.log(_drv.Cfg().Log.Con.Begin)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	grp, ctx := errgroup.WithContext(ctx)
	var tx *Tx
	grp.Go(func() error {
		var err error
		tx, err = con.ses.StartTx(TxFlags(uint32(flags)))
		return err
	})
	<-ctx.Done()
	if err := ctx.Err(); err != nil {
		if isCanceled(err) {
			con.ses.Break()
		}
		return nil, err
	}
	return tx, grp.Wait()
}
