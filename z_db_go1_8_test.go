// +build go1.8

//Copyright 2016 Tamás Gulácsi. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"context"
	"database/sql"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

func TestNamedArgs(t *testing.T) {
	t.Parallel()
	qry := "SELECT object_name FROM user_objects WHERE object_type = :typ AND ROWNUM < :num AND object_name <> :typ"
	stmt, err := testDb.Prepare(qry)
	if err != nil {
		t.Fatal(errors.Wrap(err, qry))
	}
	defer stmt.Close()
	var s string
	if err := stmt.QueryRow(sql.Named("typ", "TABLE"), sql.Named("num", 2)).Scan(&s); err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}

func TestRapidCancelIssue192(t *testing.T) {
	wait := uint64(500)
	dbQuery := func(db *sql.DB) error {
		w := atomic.LoadUint64(&wait)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(w))
		defer cancel()
		if w > 100 {
			atomic.StoreUint64(&wait, w>>2)
		}

		rows, err := db.QueryContext(ctx, "select table_name from all_tables")
		w = atomic.LoadUint64(&wait)
		if err != nil {
			t.Log(w, err)
			if err == context.DeadlineExceeded && !strings.Contains(err.Error(), "ORA-01013") {
				atomic.StoreUint64(&wait, w+1)
			}
			return err
		}
		return rows.Close()
	}

	breakStuff := func(ctx context.Context, db *sql.DB) error {
		for ctx.Err() == nil {
			if err := dbQuery(db); err != nil && err != context.DeadlineExceeded && !strings.Contains(err.Error(), "ORA-01013") {
				return err
			}
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grp, ctx := errgroup.WithContext(ctx)
	for i := 0; i < 8; i++ {
		grp.Go(func() error { return breakStuff(ctx, testDb) })
	}

	if err := grp.Wait(); err != nil {
		t.Error(err)
	}
}
