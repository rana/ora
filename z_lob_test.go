// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"encoding/json"
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
	ora.SetCfg(cfg.SetBlob(ora.L))

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
			if rc == nil && want == nil {
				continue
			}
			buf.Reset()
			//fmt.Printf("rc=%#v\n", rc)
			_, err := io.Copy(&buf, rc)
			rc.Close()
			if err != nil {
				t.Errorf("%s: Read: %v", name, err)
			}
			//t.Logf("%s: n=%d data=%v", name, n, buf.Bytes())
			if !bytes.Equal(want, buf.Bytes()) {
				t.Errorf("%s: got %v, wanted %v.", name, buf.Bytes(), want)
			}
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
	testCases := map[string]string{
		"empty": "",
		"xml":   "<xml></xml>",
		"long":  strings.Repeat("0123456789", 100000),
	}

	qry = "INSERT INTO " + tbl + " (name, content) VALUES (:1, :2)"
	for name, want := range testCases {
		if _, err := testDb.Exec(qry, name, want); err != nil {
			t.Fatalf("%s: %s: %v", name, qry, err)
		}
	}

	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)

	ora.SetCfg(cfg.SetClob(ora.L))

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

func TestLobRead(t *testing.T) {
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
		lob := &ora.Lob{C: true}
		if _, err := stmt.Exe(lob, want); err != nil {
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

func TestLobIssue156(t *testing.T) {
	tbl := tableName()
	qry := `CREATE TABLE ` + tbl + `
	(
	"INSTITUSJONSNR" NUMBER(8,0) NOT NULL ENABLE,
	"EMNEKODE" VARCHAR2(12 CHAR) NOT NULL ENABLE,
	"VERSJONSKODE" VARCHAR2(3 CHAR) NOT NULL ENABLE,
	"INFOTYPEKODE" VARCHAR2(10 CHAR) NOT NULL ENABLE,
	"SPRAKKODE" VARCHAR2(10 CHAR) NOT NULL ENABLE,
	"TERMINKODE_FRA" VARCHAR2(4 CHAR) NOT NULL ENABLE,
	"ARSTALL_FRA" NUMBER(4,0) NOT NULL ENABLE,
	"TERMINKODE_TIL" VARCHAR2(4 CHAR),
	"ARSTALL_TIL" NUMBER(4,0),
	"INFOTEKST" CLOB,
	"INFOTEKST_ORIGINAL" CLOB,
	"INSTITUSJONSNR_EIER" NUMBER(8,0) NOT NULL ENABLE
	)`
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatal(qry, err)
	}
	defer testDb.Exec("DROP TABLE " + tbl)
	testCases := map[string]string{
		"empty": "",
		"a": `Pedagogiske metoder:

Veiledet praksis. Veiledning individuelt og i grupper. Refleksjonsgrupper.

Obligatoriske arbeidskrav:

Obligatorisk frammøte tilsvarer 90 % av studietid i praksis.`,
		"b": `Godkjente arbeidskrav.Se undervisningsplan for praksisstudier 3. studieår

Læringsutbytte - Kunnskap:

Læringsutbyttet defineres i forhold til områder for kunnskap, ferdigheter og holdninger - se Undervisningsplan for praksissstudier 3. studieår`,
	}
	qry = `INSERT INTO ` + tbl + `
  (INSTITUSJONSNR, EMNEKODE, VERSJONSKODE, INFOTYPEKODE, SPRAKKODE, TERMINKODE_FRA, ARSTALL_FRA, TERMINKODE_TIL, ARSTALL_TIL, INFOTEKST, INFOTEKST_ORIGINAL, INSTITUSJONSNR_EIER)
  VALUES
  (1, :1, 'ver', 'infokode', 'sprakkode', 'term', 2, 'min', 3, :2, '', 4)`

	type EmneInfo struct {
		InstitusjonsNr     ora.Int64  `json:"institusjonsnr"`
		EmneKode           ora.String `json:"emnekode"`
		VersjonsKode       ora.String `json:"versjonskode"`
		InfoTypeKode       ora.String `json:"versjonskode"`
		SprakKode          ora.String `json:"versjonskode"`
		TerminKodeFra      ora.String `json:"versjonskode"`
		ArstallFra         ora.Int64  `json:"versjonskode"`
		TerminKodeTil      ora.String `json:"versjonskode"`
		ArstallTil         ora.Int64  `json:"versjonskode"`
		InfoTekst          *ora.Lob   `json:"versjonskode"`
		InfoTekstOriginal  *ora.Lob   `json:"versjonskode"`
		InstitusjonsNrEier ora.Int64  `json:"versjonskode"`
	}
	for nm, want := range testCases {
		if _, err := testDb.Exec(qry, nm, want); err != nil {
			t.Fatal(nm, qry, err)
		}

		qry = "SELECT * FROM " + tbl + " WHERE emnekode = :1"
		stmt, err := testSes.Prep(qry,
			ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
			ora.OraI64, ora.OraS, ora.OraI64, ora.L, ora.L, ora.OraI64)
		if err != nil {
			t.Fatal(nm, qry, err)
		}
		rst, err := stmt.Qry(nm)
		if err != nil {
			t.Fatal(nm, qry, err)
		}

		results := make([]EmneInfo, 0)
		for rst.Next() {
			results = append(results, EmneInfo{
				InstitusjonsNr:     rst.Row[0].(ora.Int64),
				EmneKode:           rst.Row[1].(ora.String),
				VersjonsKode:       rst.Row[2].(ora.String),
				InfoTypeKode:       rst.Row[3].(ora.String),
				SprakKode:          rst.Row[4].(ora.String),
				TerminKodeFra:      rst.Row[5].(ora.String),
				ArstallFra:         rst.Row[6].(ora.Int64),
				TerminKodeTil:      rst.Row[7].(ora.String),
				ArstallTil:         rst.Row[8].(ora.Int64),
				InfoTekst:          rst.Row[9].(*ora.Lob),
				InfoTekstOriginal:  rst.Row[10].(*ora.Lob),
				InstitusjonsNrEier: rst.Row[11].(ora.Int64),
			})
		}
		b, err := json.Marshal(results)
		t.Log(nm, b, err)
	}
}
