// +build go1.8

// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"context"
	"database/sql/driver"
	"errors"

	"golang.org/x/sync/errgroup"
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
		if v.Name != "" {
			return nil, errors.New("named values are not supported!")
		}
		params[n] = v.Value
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var res DrvExecResult
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		var err error
		res.rowsAffected, res.lastInsertId, err = ds.stmt.exeC(ctx, params, false)
		if err != nil {
			return errE(err)
		}
		return nil
	})
	if err := ctx.Err(); err != nil {
		if isCanceled(err) {
			ds.stmt.ses.Break()
		}
		return nil, err
	}
	if err := grp.Wait(); err != nil {
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
		if v.Name != "" {
			return nil, errors.New("named values are not supported!")
		}
		params[n] = v.Value
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var rset *Rset
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		var err error
		rset, err = ds.stmt.qryC(ctx, params)
		if err != nil {
			return errE(err)
		}
		return nil
	})
	if err := ctx.Err(); err != nil {
		if isCanceled(err) {
			ds.stmt.ses.Break()
		}
		return nil, err
	}
	err := grp.Wait()
	return &DrvQueryResult{rset: rset}, err
}
