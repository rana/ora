// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	ora "gopkg.in/rana/ora.v4"
)

func TestLOBCloseStatement(t *testing.T) {
	qry := `DECLARE
  cur1 SYS_REFCURSOR;
  cur2 SYS_REFCURSOR;
  cur3 SYS_REFCURSOR;
  cur4 SYS_REFCURSOR;
BEGIN
  OPEN cur1 FOR
    SELECT '1', SYSDATE dt FROM DUAL
	UNION ALL
	SELECT object_name, SYSDATE FROM all_objects;
  :1 := cur1;
  OPEN cur2 FOR
    SELECT '1', SYSDATE dt FROM DUAL
	UNION ALL
	SELECT object_name, SYSDATE FROM all_objects;
  :2 := cur2;
  OPEN cur3 FOR
    SELECT '1', SYSDATE dt FROM DUAL
	UNION ALL
	SELECT object_name, SYSDATE FROM all_objects;
  :3 := cur3;
  OPEN cur4 FOR
    SELECT '2', SYSDATE dt FROM DUAL
	UNION ALL
	SELECT object_name, SYSDATE FROM all_objects;
  :4 := cur4;
END;`

	testSes := getSes(t)
	defer testSes.Close()

	for _, doNext := range []bool{true, false} {
		stmt, err := testSes.Prep(qry)
		if err != nil {
			t.Error(qry, err)
		}

		cur := make([]*ora.Rset, 4)
		for i := range cur {
			cur[i] = &ora.Rset{}
		}
		if _, err = stmt.Exe(cur[0], cur[1], cur[2], cur[3]); err != nil {
			t.Error(err)
		}
		if !doNext {
			stmt.Close()
			continue
		}
		for _, c := range cur {
			row := c.NextRow()
			if row != nil {
				fmt.Println(row)
			}
		}
		stmt.Close()
	}
}

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
	testSes := getSes(t)
	defer testSes.Close()

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
			t.Fatalf("%s: %v", name, err)
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
	for nm, want := range testCases {
		if _, err := testDb.Exec(qry, nm, want); err != nil {
			t.Fatal(nm, qry, err)
		}
	}

	qry = "SELECT * FROM " + tbl + " WHERE emnekode = :1"

	testSes := getSes(t)
	defer testSes.Close()

	enableLogging(t)
	// LOB
	{
		type EmneInfo struct {
			InstitusjonsNr     ora.Int64
			EmneKode           ora.String
			VersjonsKode       ora.String
			InfoTypeKode       ora.String
			SprakKode          ora.String
			TerminKodeFra      ora.String
			ArstallFra         ora.Int64
			TerminKodeTil      ora.String
			ArstallTil         ora.Int64
			InfoTekst          *ora.Lob
			InfoTekstOriginal  *ora.Lob
			InstitusjonsNrEier ora.Int64
		}
		stmt, err := testSes.Prep(qry,
			ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
			ora.OraI64, ora.OraS, ora.OraI64, ora.L, ora.L, ora.OraI64)
		if err != nil {
			t.Fatal(qry, err)
		}
		defer stmt.Close()

		for nm, _ := range testCases {
			rst, err := stmt.Qry(nm)
			if err != nil {
				t.Fatal(nm, qry, err)
			}

			results := make([]string, 0, 1)
			for rst.Next() {
				info := EmneInfo{
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
				}
				b, err := json.Marshal(info)
				if err != nil {
					t.Fatal(nm, info, err)
				}
				results = append(results, string(b))
			}
			if err := rst.Err(); err != nil {
				t.Fatal(err)
			}
			if len(results) == 0 {
				t.Fatal(nm, "no rows found!")
			}
			t.Log(nm, results)
		}
	}

	// string
	{
		type EmneInfo struct {
			InstitusjonsNr     ora.Int64
			EmneKode           ora.String
			VersjonsKode       ora.String
			InfoTypeKode       ora.String
			SprakKode          ora.String
			TerminKodeFra      ora.String
			ArstallFra         ora.Int64
			TerminKodeTil      ora.String
			ArstallTil         ora.Int64
			InfoTekst          ora.String
			InfoTekstOriginal  ora.String
			InstitusjonsNrEier ora.Int64
		}
		stmt, err := testSes.Prep(qry,
			ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
			ora.OraI64, ora.OraS, ora.OraI64, ora.OraS, ora.OraS, ora.OraI64)
		if err != nil {
			t.Fatal(qry, err)
		}
		defer stmt.Close()

		for nm, want := range testCases {
			rst, err := stmt.Qry(nm)
			if err != nil {
				t.Fatal(nm, qry, err)
			}

			results := make([]EmneInfo, 0, 1)
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
					InfoTekst:          rst.Row[9].(ora.String),
					InfoTekstOriginal:  rst.Row[10].(ora.String),
					InstitusjonsNrEier: rst.Row[11].(ora.Int64),
				})
				got := results[len(results)-1].InfoTekst
				if d := stringEqualNonUnicode(got.Value, want); d != "" {
					t.Errorf("%s: got %q, wanted %q (diff: %v).", nm, got, want, d)
				}
			}
			if err := rst.Err(); err != nil {
				t.Fatal(err)
			}
			if len(results) == 0 {
				t.Fatal(nm, "no rows found!")
			}
			b, err := json.Marshal(results)
			t.Log(nm, string(b), err)
			//t.Logf("%s: %#v", nm, results)
		}
	}
}

func TestLobIssue159Stress(t *testing.T) {
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

	for i := 0; i < 100; i++ {
		nm := fmt.Sprintf("RND-%04d", i)
		testCases[nm] = strings.Repeat(base64.URLEncoding.EncodeToString([]byte(nm)), i+1)
	}

	qry = `INSERT INTO ` + tbl + `
  (INSTITUSJONSNR, EMNEKODE, VERSJONSKODE, INFOTYPEKODE, SPRAKKODE, TERMINKODE_FRA, ARSTALL_FRA, TERMINKODE_TIL, ARSTALL_TIL, INFOTEKST, INFOTEKST_ORIGINAL, INSTITUSJONSNR_EIER)
  VALUES
  (1, :1, 'ver', 'infokode', 'sprakkode', 'term', 2, 'min', 3, :2, :3, 4)`
	for nm, want := range testCases {
		if _, err := testDb.Exec(qry, nm, want, reverseString(want)); err != nil {
			t.Fatal(nm, qry, err)
		}
	}

	testDb.Exec("CREATE UNIQUE INDEX K_" + tbl + " ON " + tbl + "(emnekode)")

	type EmneInfo struct {
		InstitusjonsNr     ora.Int64
		EmneKode           ora.String
		VersjonsKode       ora.String
		InfoTypeKode       ora.String
		SprakKode          ora.String
		TerminKodeFra      ora.String
		ArstallFra         ora.Int64
		TerminKodeTil      ora.String
		ArstallTil         ora.Int64
		InfoTekst          *ora.Lob
		InfoTekstOriginal  *ora.Lob
		InstitusjonsNrEier ora.Int64
	}

	//enableLogging(t)

	qry = "SELECT * FROM " + tbl + " WHERE emnekode = :1"
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU()+2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			t.Log("START", i)
			defer t.Log("END", i)

			ses, err := testSesPool.Get()
			if err != nil {
				t.Fatal(i, err)
			}
			defer ses.Close()
			stmt, err := ses.Prep(qry,
				ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
				ora.OraI64, ora.OraS, ora.OraI64, ora.L, ora.L, ora.OraI64)
			if err != nil {
				t.Fatal(qry, err)
			}
			defer stmt.Close()

			results := make([]string, 0, 1)
			for j := 0; j < 10; j++ {
				for nm, _ := range testCases {
					rst, err := stmt.Qry(nm)
					if err != nil {
						t.Fatal(fmt.Sprintf("%d.%s", i, nm), qry, err)
					}
					nm := fmt.Sprintf("%d:%d.%s", i, j, nm)

					results = results[:0]
					for rst.Next() {
						info := EmneInfo{
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
						}
						b, err := json.Marshal(info)
						//t.Log(nm, "info:", string(b))
						if err != nil {
							rst.Exhaust()
							t.Fatalf("%s: %#v: %v", nm, info, err)
						}
						results = append(results, string(b))
					}
					if err := rst.Err(); err != nil {
						t.Fatal(nm, "rst.Err:", err)
					}
					if len(results) == 0 {
						t.Fatal(nm, "no rows found!")
					}
				}
				//t.Log(nm, results)
			}
		}(i)
	}
	wg.Wait()

}

func TestLobIssue159(t *testing.T) {
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
  (1, :1, 'ver', 'infokode', 'sprakkode', 'term', 2, 'min', 3, :2, :3, 4)`
	for nm, want := range testCases {
		if _, err := testDb.Exec(qry, nm, want, reverseString(want)); err != nil {
			t.Fatal(nm, qry, err)
		}
	}

	testSes := getSes(t)
	defer testSes.Close()

	qry = "SELECT * FROM " + tbl + " WHERE emnekode = :1"

	// string
	{
		type EmneInfo struct {
			InstitusjonsNr     ora.Int64
			EmneKode           ora.String
			VersjonsKode       ora.String
			InfoTypeKode       ora.String
			SprakKode          ora.String
			TerminKodeFra      ora.String
			ArstallFra         ora.Int64
			TerminKodeTil      ora.String
			ArstallTil         ora.Int64
			InfoTekst          string
			InfoTekstOriginal  string
			InstitusjonsNrEier ora.Int64
		}
		stmt, err := testSes.Prep(qry,
			ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
			ora.OraI64, ora.OraS, ora.OraI64, ora.S, ora.S, ora.OraI64)
		if err != nil {
			t.Fatal(qry, err)
		}
		defer stmt.Close()

		for nm, want := range testCases {
			tnaw := reverseString(want)
			rst, err := stmt.Qry(nm)
			if err != nil {
				t.Fatal(nm, qry, err)
			}

			results := make([]EmneInfo, 0, 1)
			for rst.Next() {
				info := EmneInfo{
					InstitusjonsNr:     rst.Row[0].(ora.Int64),
					EmneKode:           rst.Row[1].(ora.String),
					VersjonsKode:       rst.Row[2].(ora.String),
					InfoTypeKode:       rst.Row[3].(ora.String),
					SprakKode:          rst.Row[4].(ora.String),
					TerminKodeFra:      rst.Row[5].(ora.String),
					ArstallFra:         rst.Row[6].(ora.Int64),
					TerminKodeTil:      rst.Row[7].(ora.String),
					ArstallTil:         rst.Row[8].(ora.Int64),
					InfoTekst:          rst.Row[9].(string),
					InfoTekstOriginal:  rst.Row[10].(string),
					InstitusjonsNrEier: rst.Row[11].(ora.Int64),
				}
				results = append(results, info)
				got := info.InfoTekst
				if d := stringEqualNonUnicode(got, want); d != "" {
					t.Errorf("%s: got %q, wanted %q (diff: %v).", nm, got, want, d)
				}
				tog := info.InfoTekstOriginal
				if d := stringEqualNonUnicode(tog, tnaw); d != "" {
					t.Errorf("%s: tog %q, tnawed %q (diff: %v).", nm, tog, tnaw, d)
				}
			}
			if err := rst.Err(); err != nil {
				t.Fatal(err)
			}
			b, err := json.Marshal(results)
			if len(results) == 0 {
				t.Fatal(nm, "no rows found!")
			}
			t.Log(nm, "results:", string(b), "error:", err)
			//t.Logf("%s: %#v", nm, results)
		}
	}

	// LOB
	{
		type EmneInfo struct {
			InstitusjonsNr     ora.Int64
			EmneKode           ora.String
			VersjonsKode       ora.String
			InfoTypeKode       ora.String
			SprakKode          ora.String
			TerminKodeFra      ora.String
			ArstallFra         ora.Int64
			TerminKodeTil      ora.String
			ArstallTil         ora.Int64
			InfoTekst          *ora.Lob
			InfoTekstOriginal  *ora.Lob
			InstitusjonsNrEier ora.Int64
		}
		stmt, err := testSes.Prep(qry,
			ora.OraI64, ora.OraS, ora.OraS, ora.OraS, ora.OraS, ora.OraS,
			ora.OraI64, ora.OraS, ora.OraI64, ora.L, ora.L, ora.OraI64)
		if err != nil {
			t.Fatal(qry, err)
		}
		defer stmt.Close()

		for nm, _ := range testCases {
			rst, err := stmt.Qry(nm)
			if err != nil {
				t.Fatal(nm, qry, err)
			}

			results := make([]string, 0, 1)
			for rst.Next() {
				info := EmneInfo{
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
				}
				b, err := json.Marshal(info)
				t.Log("info:", string(b))
				if err != nil {
					t.Fatal(nm, info, err)
				}
				results = append(results, string(b))
			}
			if err := rst.Err(); err != nil {
				t.Fatal(err)
			}
			if len(results) == 0 {
				t.Fatal(nm, "no rows found!")
			}
			t.Log(nm, results)
		}
	}

}

func TestLobIssue191(t *testing.T) {
	enableLogging(t)
	testSes := getSes(t)
	defer testSes.Close()

	l := &ora.Lob{}
	stmt, err := testSes.Prep(`begin :1 := null; end;`)
	if err != nil {
		t.Fatal("2 - ", err)
	}
	n, err := stmt.Exe(l)

	if err != nil {
		t.Fatal("3 - ", err)
	}
	bb1, err := l.Bytes()
	if err != nil {
		t.Fatal("4 - ", err)
	}
	t.Log("Result - ", n, string(bb1))
}

func stringEqualNonUnicode(a, b string) string {
	if a == b {
		return ""
	}
	aRunes, bRunes := []rune(a), []rune(b)
	for i, r := range aRunes {
		if r == '?' {
			bRunes[i] = '?'
		}
	}
	if string(aRunes) == string(bRunes) {
		return ""
	}
	for i, r := range aRunes {
		if len(bRunes) == i {
			return fmt.Sprintf("extra: %q", string(aRunes[i:]))
		}
		if r != bRunes[i] {
			k := i
			j := i + 5
			if j > len(aRunes) {
				j = len(aRunes)
			}
			if j > len(bRunes) {
				j = len(bRunes)
			}

			i -= 5
			if i < 0 {
				i = 0
			}
			return fmt.Sprintf("@%d %q != %q", k, string(aRunes[i:j]), string(bRunes[i:j]))
		}
	}
	return ""
}

func reverseString(s string) string {
	runes := []rune(s)
	j := len(runes) - 1
	for i := 0; i < j; i++ {
		runes[i], runes[j] = runes[j], runes[i]
		j--
	}
	return string(runes)
}

func TestLobIssue237(t *testing.T) {
	t.Parallel()
	//defer tl.enableLogging(t)()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ms, err := newMetricSet(ctx, testDb)
	if err != nil {
		t.Fatal(err)
	}
	defer ms.Close()

	for i := 0; i < 100; i++ {
		if err := ctx.Err(); err != nil {
			break
		}
		events, err := ms.Fetch(ctx)
		t.Log("events:", len(events))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func newMetricSet(ctx context.Context, db *sql.DB) (*metricSet, error) {
	qry := "select /* metricset: sqlstats */ inst_id, sql_fulltext, last_active_time from gv$sqlstats WHERE ROWNUM < 11"
	stmt, err := db.PrepareContext(ctx, qry)
	if err != nil {
		return nil, err
	}

	return &metricSet{
		stmt: stmt,
	}, nil
}

type metricSet struct {
	stmt *sql.Stmt
}

func (m *metricSet) Close() error {
	st := m.stmt
	m.stmt = nil
	if st == nil {
		return nil
	}
	return st.Close()
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *metricSet) Fetch(ctx context.Context) ([]event, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := m.stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []event
	var buf bytes.Buffer
	for rows.Next() {
		var e event
		var lob *ora.Lob
		if err := rows.Scan(&e.ID, &lob, &e.LastActive); err != nil {
			return events, err
		}
		buf.Reset()
		if _, err := io.Copy(&buf, lob); err != nil {
			return events, err
		}
		e.Text = buf.String()
		events = append(events, e)
	}

	return events, nil
}

type event struct {
	ID         int64
	Text       string
	LastActive time.Time
}
