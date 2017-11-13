//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/rana/ora.v4"
	"gopkg.in/rana/ora.v4/date"
)

func Test_plsarr_num_session(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	for _, qry := range []string{
		`CREATE OR REPLACE PACKAGE TST_ora_plsarr_num AS
  TYPE pls_tab_typ IS TABLE OF NUMBER INDEX BY PLS_INTEGER;
  PROCEDURE slice(p_nums IN pls_tab_typ);
  FUNCTION count_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2;
  FUNCTION count_slice_int(p_nums IN pls_tab_typ) RETURN PLS_INTEGER;
  FUNCTION sum_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2;
  FUNCTION sum_slice_num(p_nums IN pls_tab_typ) RETURN NUMBER;
END TST_ora_plsarr_num;`,
		`CREATE OR REPLACE PACKAGE BODY TST_ora_plsarr_num AS
  PROCEDURE slice(p_nums IN pls_tab_typ) IS
  BEGIN
    NULL;
  END slice;
  FUNCTION count_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2 IS
  BEGIN
    RETURN 'COUNT='||TO_CHAR(count_slice_int(p_nums));
  END count_slice_vc;

  FUNCTION count_slice_int(p_nums IN pls_tab_typ) RETURN PLS_INTEGER IS
  BEGIN
    RETURN p_nums.COUNT;
  END count_slice_int;

  FUNCTION sum_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2 IS
  BEGIN
    RETURN 'SUM='||TO_CHAR(sum_slice_num(p_nums));
  END sum_slice_vc;

  FUNCTION sum_slice_num(p_nums IN pls_tab_typ) RETURN NUMBER IS
    s NUMBER := 0;
	i PLS_INTEGER;
  BEGIN
    i := p_nums.FIRST;
    WHILE i IS NOT NULL LOOP
	  s := s + p_nums(i);
	  i := p_nums.NEXT(i);
    END LOOP;
    RETURN s;
  END sum_slice_num;
END TST_ora_plsarr_num;`,
	} {
		if _, err := testSes.PrepAndExe(qry); err != nil {
			t.Fatal(err)
		}
		checkCompile(t, testSes)
	}

	//enableLogging(t)
	for _, rt := range []reflect.Type{
		reflect.TypeOf(ora.Float64{}),
		reflect.TypeOf(ora.Float32{}),
		reflect.TypeOf(ora.Int32{}),
		reflect.TypeOf(ora.Int64{}),
	} {
		numbersV := reflect.MakeSlice(reflect.SliceOf(rt), 9, 9)
		if strings.HasPrefix(rt.Name(), "Float") {
			for i := 0; i < 9; i++ {
				numbersV.Index(i).FieldByName("Value").SetFloat(float64(i))
			}
		} else {
			for i := 0; i < 9; i++ {
				numbersV.Index(i).FieldByName("Value").SetInt(int64(i))
			}
		}
		prefix := "[]" + rt.Name() + " "
		numbers := numbersV.Interface()
		var (
			retStr string
			retInt int32
			retNum float64
		)
		for i, tc := range []struct {
			qry    string
			params []interface{}
			await  interface{}
		}{
			{"BEGIN TST_ora_plsarr_num.slice(:1); END;", []interface{}{numbers}, nil},
			{"BEGIN :1 := TST_ora_plsarr_num.count_slice_vc(:2); END;", []interface{}{&retStr, numbers}, "COUNT=9"},
			{"BEGIN :1 := TST_ora_plsarr_num.sum_slice_vc(:2); END;", []interface{}{&retStr, numbers}, "SUM=36"},
			{"BEGIN :1 := TST_ora_plsarr_num.count_slice_int(:2); END;", []interface{}{&retInt, numbers}, int32(9)},
			{"BEGIN :1 := TST_ora_plsarr_num.sum_slice_num(:2); END;", []interface{}{&retNum, numbers}, float64(36)},
		} {

			if _, err := testSes.PrepAndExeP(tc.qry, tc.params...); err != nil {
				t.Fatalf(prefix+"%d. %q (%#v): %v", i, tc.qry, tc.params, err)
			}
			if len(tc.params) == 0 {
				continue
			}
			if tc.params[0] == nil {
				t.Errorf(prefix+"%d. got nil", i)
				continue
			}
			got := reflect.Indirect(reflect.ValueOf(tc.params[0])).Interface()
			t.Logf(prefix+"%d: got %#v, awaited %#v", i, got, tc.await)
			if tc.await == nil {
				continue
			}
			if !reflect.DeepEqual(got, tc.await) {
				t.Errorf(prefix+"%d. got %#v, awaited %#v.", i, got, tc.await)
			}
		}
	}
}

func Test_plsarr_dt_session(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Parallel()
	for _, qry := range []string{
		`CREATE OR REPLACE PACKAGE TST_ora_plsarr_dt AS
  TYPE string_tab_typ IS TABLE OF VARCHAR2(1000) INDEX BY PLS_INTEGER;
  TYPE date_tab_typ IS TABLE OF DATE INDEX BY PLS_INTEGER;
  FUNCTION str_slice_concat(p_strings IN string_tab_typ) RETURN VARCHAR2;
  FUNCTION date_slice_concat(p_dates IN date_tab_typ) RETURN VARCHAR2;
END TST_ora_plsarr_dt;`,
		`CREATE OR REPLACE PACKAGE BODY TST_ora_plsarr_dt AS
  FUNCTION str_slice_concat(p_strings IN string_tab_typ) RETURN VARCHAR2 IS
    i PLS_INTEGER;
    s VARCHAR2(32767);
  BEGIN
    i := p_strings.FIRST;
    WHILE i IS NOT NULL LOOP
	  s := s||p_strings(i)||CHR(10);
	  i := p_strings.NEXT(i);
    END LOOP;
	RETURN(s);
  END str_slice_concat;
  FUNCTION date_slice_concat(p_dates IN date_tab_typ) RETURN VARCHAR2 IS
    i PLS_INTEGER;
    s VARCHAR2(32767);
  BEGIN
    i := p_dates.FIRST;
    WHILE i IS NOT NULL LOOP
	  s := s||TO_CHAR(p_dates(i), 'YYYY-MM-DD HH24:MI:SS')||CHR(10);
	  i := p_dates.NEXT(i);
    END LOOP;
	RETURN(s);
  END date_slice_concat;
END TST_ora_plsarr_dt;`,
	} {
		if _, err := testSes.PrepAndExe(qry); err != nil {
			t.Fatal(err)
		}
		checkCompile(t, testSes)
	}

	//enableLogging(t)
	var ret string
	now := time.Now()
	for i, tc := range []struct {
		qry    string
		params []interface{}
		await  string
	}{
		{
			"BEGIN :1 := TST_ora_plsarr_dt.str_slice_concat(:2); END;",
			[]interface{}{&ret, []string{"a", "Bb", "cCc", "dDdD", "árvíztűrő tükörfúrógép"}},
			"a\nBb\ncCc\ndDdD\nárvíztűrő tükörfúrógép\n",
		},
		{
			"BEGIN :1 := TST_ora_plsarr_dt.date_slice_concat(:2); END;",
			[]interface{}{&ret, []ora.Date{{Date: date.FromTime(now)}, {Date: date.FromTime(now.Add(-24 * time.Hour))}}},
			now.Format("2006-01-02 15:04:05") + "\n" + now.Add(-24*time.Hour).Format("2006-01-02 15:04:05") + "\n",
		},
	} {
		if _, err := testSes.PrepAndExeP(tc.qry, tc.params...); err != nil {
			t.Fatalf("%d. %q (%#v): %v", i, tc.qry, tc.params, err)
		}
		if len(tc.params) == 0 {
			continue
		}
		if tc.params[0] == nil {
			t.Errorf("%d. got nil", i)
			continue
		}
		t.Logf("%d: got %#v, awaited %#v", i, ret, tc.await)
		if ret != tc.await {
			t.Errorf("%d. got %#v, awaited %#v.", i, ret, tc.await)
		}

	}
}

func checkCompile(t *testing.T, testSes *ora.Ses) {
	errs, err := ora.GetCompileErrors(testSes, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		errS := make([]string, len(errs))
		for i, ce := range errs {
			errS[i] = ce.Error()
		}
		t.Fatal(errors.New(strings.Join(errS, "\n")))
	}
}

func TestIssue188(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	stmt, err := testSes.Prep(`BEGIN :1(1) := 'test'; END;`)
	if err != nil {
		t.Fatal("1 - ", err)
	}
	defer stmt.Close()

	sret := make([]string, 0, 1000)
	enableLogging(t)
	n, err := stmt.ExeP(&sret)
	if err != nil {
		t.Fatal("2 - ", err)
	}
	t.Log("Result - ", n, len(sret), sret)
	if len(sret) != 1 || sret[0] != "test" {
		t.Errorf("Want \"test\", got %#v", sret)
	}
}
