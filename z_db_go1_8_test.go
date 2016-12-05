// +build go1.8

//Copyright 2016 Tamás Gulácsi. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"database/sql"
	"testing"

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
