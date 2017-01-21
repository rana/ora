// +build go1.8

// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"context"
	"database/sql/driver"
)

// ExecContext enhances the Stmt interface by providing Exec with context.
// ExecContext must honor the context timeout and return when it is cancelled.
func (ds *DrvStmt) ExecContext(ctx context.Context, values []driver.NamedValue) (driver.Result, error) {
	ds.log(true)
	if err := ds.checkIsOpen(); err != nil {
		return nil, errE(err)
	}
	params := make([]interface{}, len(values))
	for n, v := range values {
		params[n] = v
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var res DrvExecResult
	done := make(chan error)
	go func() {
		defer close(done)
		var err error
		res.rowsAffected, res.lastInsertId, err = ds.stmt.exeC(ctx, params, false)
		if err != nil {
			done <- errE(err)
			return
		}
		done <- nil
	}()
	var err error
	select {
	case <-ctx.Done():
		if err = ctx.Err(); isCanceled(err) {
			ds.stmt.ses.Break()
		}
	case err = <-done:
	}
	if err != nil {
		return nil, err
	}
	if res.rowsAffected == 0 {
		return driver.RowsAffected(0), nil
	}
	return &res, nil
}

// QueryContext enhances the Stmt interface by providing Query with context.
// QueryContext must honor the context timeout and return when it is cancelled.
func (ds *DrvStmt) QueryContext(ctx context.Context, values []driver.NamedValue) (driver.Rows, error) {
	ds.log(true)
	if err := ds.checkIsOpen(); err != nil {
		return nil, errE(err)
	}
	params := make([]interface{}, len(values))
	for n, v := range values {
		params[n] = v
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var rset *Rset
	done := make(chan error)
	go func() {
		defer close(done)
		var err error
		rset, err = ds.stmt.qryC(ctx, params)
		if err != nil {
			done <- errE(err)
			return
		}
		done <- nil
	}()
	var err error
	select {
	case <-ctx.Done():
		if err = ctx.Err(); isCanceled(err) {
			ds.stmt.ses.Break()
		}
		return nil, err
	case err = <-done:
	}
	return &DrvQueryResult{rset: rset}, err
}

// vim: set fileencoding=utf-8 noet:
