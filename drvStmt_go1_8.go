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

	done := make(chan struct{})
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			// select again to avoid race condition if both are done
			select {
			case <-done:
			default:
				if isCanceled(ctx.Err()) {
					ds.stmt.RLock()
					ses := ds.stmt.ses
					ds.stmt.RUnlock()
					ses.Break()
				}
			}
		}
	}()

	var err error
	var res DrvExecResult
	res.rowsAffected, res.lastInsertId, err = ds.stmt.exeC(ctx, params, false)
	close(done)

	if err != nil {
		return nil, maybeBadConn(err)
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
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			// select again to avoid race condition if both are done
			select {
			case <-done:
			default:
				if isCanceled(ctx.Err()) {
					ds.stmt.RLock()
					ses := ds.stmt.ses
					ds.stmt.RUnlock()
					ses.Break()
				}
			}
		}
	}()

	rset, err := ds.stmt.qryC(ctx, params)
	close(done)

	if err != nil {
		return nil, maybeBadConn(err)
	}
	return &DrvQueryResult{rset: rset}, nil
}

// vim: set fileencoding=utf-8 noet:
