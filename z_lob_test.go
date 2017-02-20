// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	ora "gopkg.in/rana/ora.v4"
)

func TestLobSelect(t *testing.T) {
	tbl := tableName()
	testDb.Exec("DROP TABLE " + tbl)
	qry := "CREATE TABLE " + tbl + " (name VARCHAR2(10), content BLOB)"
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatalf("%s: %v", qry, err)
	}
	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)
	ora.SetCfg(cfg.SetBlob(ora.Bin))

	testCases := map[string][]byte{
		"7f7f7f": []byte{0x7f, 0x7f, 0x7f},
		"empty":  nil,
	}
	for name, want := range testCases {
		qry = fmt.Sprintf("INSERT INTO %s (name, content) VALUES ('%s', HEXTORAW('%x'))", tbl, name, want)
		if _, err := testDb.Exec(qry); err != nil {
			t.Fatalf("%s: %s: %v", name, qry, err)
		}

		// SELECT into []byte
		rows, err := testDb.Query(fmt.Sprintf("SELECT content FROM %s WHERE name = :1", tbl), name)
		if err != nil {
			t.Errorf("%s: SELECT: %v", name, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var v []byte
			if err = rows.Scan(&v); err != nil {
				t.Errorf("%s: Scan: %v", name, err)
			}
			if len(v) != len(want) {
				t.Errorf("%s: got %v, wanted %v.", name, v, want)
			}
		}
	}

	//enableLogging(t)
	ora.SetCfg(cfg.SetBlob(ora.D))

	// SELECT into io.ReadCloser
	for name, want := range testCases {
		rows, err := testDb.Query(fmt.Sprintf("SELECT content FROM %s WHERE name = :1", tbl), name)
		if err != nil {
			t.Errorf("%s: SELECT: %v", name, err)
			return
		}
		defer rows.Close()
		var buf bytes.Buffer
		for rows.Next() {
			var v interface{}
			if err = rows.Scan(&v); err != nil {
				t.Errorf("%s: Scan: %v", name, err)
			}
			//t.Logf("%s: %#v (%T)", name, v, v)
			rc, ok := v.(io.ReadCloser)
			if !ok {
				if rc == nil && want == nil {
					continue
				}
				t.Fatalf("%s: wanted io.ReadCloser for LOB, got %T - cfg.Rset.SetBlob ineffective?", name, v)
			}
			_, err := io.Copy(&buf, rc)
			rc.Close()
			if err != nil {
				t.Errorf("%s: Read: %v", name, err)
			}
			//t.Logf("%s: n=%d data=%v", name, n, buf.Bytes())
			if !bytes.Equal(want, buf.Bytes()) {
				t.Errorf("%s: got %v, wanted %v.", name, buf.Bytes(), want)
			}
			buf.Reset()
		}
	}
}

func TestLobSelectString(t *testing.T) {
	tbl := tableName()
	testDb.Exec("DROP TABLE " + tbl)
	qry := "CREATE TABLE " + tbl + " (name VARCHAR2(10), content CLOB)"
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatalf("%s: %v", qry, err)
	}
	testCases := map[string]string{"xml": "<xml></xml>", "empty": ""}

	qry = "INSERT INTO " + tbl + " (name, content) VALUES (:1, :2)"
	for name, want := range testCases {
		if _, err := testDb.Exec(qry, name, want); err != nil {
			t.Fatalf("%s: %s: %v", name, qry, err)
		}
	}

	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)

	ora.SetCfg(cfg.SetClob(ora.D))

	for name, want := range testCases {
		rows, err := testDb.Query("SELECT content FROM "+tbl+" WHERE name = :1", name)
		if err != nil {
			t.Errorf("%s: SELECT: %v", name, err)
			return
		}
		defer rows.Close()
		var buf bytes.Buffer
		for rows.Next() {
			var v interface{}
			if err = rows.Scan(&v); err != nil {
				t.Errorf("%s: Scan: %v", name, err)
			}
			//t.Logf("%#v (%T)", v, v)
			rc, ok := v.(io.ReadCloser)
			if ok {
				_, err = io.Copy(&buf, rc)
				rc.Close()
			} else if !(v == nil && want == "") {
				buf.WriteString(v.(string))
			}
			if err != nil {
				t.Errorf("%s: Read: %v", name, err)
			}
			//t.Logf("n=%d data=%v", n, buf.Bytes())
			if want != buf.String() {
				t.Errorf("%s: got %q, wanted %q.", name, buf.String(), want)
			}
			buf.Reset()
		}
	}

	// SELECT into string
	ora.SetCfg(cfg.SetClob(ora.S))

	for name, want := range testCases {
		rows, err := testDb.Query("SELECT content FROM "+tbl+" WHERE name = :1", name)
		if err != nil {
			t.Errorf("%s: SELECT: %v", name, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var v string
			if err = rows.Scan(&v); err != nil {
				t.Errorf("%s: Scan: %v", name, err)
			}
			t.Logf("%s: read %q", name, v)
			if v != want {
				t.Errorf("%s: got %q, want %q.", name, v, want)
			}
		}
	}
}

func TestLOBRead(t *testing.T) {
	if _, err := testDb.Exec(`CREATE OR REPLACE
PROCEDURE test_get_json(p_clob OUT CLOB, p_text in VARCHAR2) IS
BEGIN
  DBMS_LOB.createtemporary(p_clob, TRUE);
  IF p_text IS NULL THEN
    RETURN;
  END IF;
  DBMS_LOB.writeappend(p_clob, LENGTH(p_text), p_text);
END test_get_json;`,
	); err != nil {
		t.Skipf("create function: %v", err)
	}
	//enableLogging(t)
	stmt, err := testSes.Prep("CALL test_get_json(:1, :2)", ora.OraBin, ora.S)
	if err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]string{
		"empty": "",
		"json":  `{"message":"this is a json object"}`,
	} {
		lob := ora.Lob{C: true}
		if _, err := stmt.Exe(&lob, want); err != nil {
			if strings.Contains(err.Error(), "ORA-06575:") {
				ce, err2 := ora.GetCompileErrors(testSes, false)
				t.Fatalf("%s: %v\n%v (%v)", name, err, ce, err2)
			}
			t.Fatal(name, err)
		}
		b, err := ioutil.ReadAll(lob)
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
		t.Logf("%s: got %s", name, b)
		if string(b) != want {
			t.Errorf("%s: got %q, wanted %q.", name, b, want)
		}
	}
}
