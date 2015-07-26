//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/rana/ora.v2"
)

func Test_plsarr_session(t *testing.T) {
	for _, qry := range []string{
		`CREATE OR REPLACE PACKAGE TST_ora AS
  TYPE pls_tab_typ IS TABLE OF NUMBER INDEX BY PLS_INTEGER;
  PROCEDURE slice(p_nums IN pls_tab_typ);
  FUNCTION count_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2;
  FUNCTION count_slice_int(p_nums IN pls_tab_typ) RETURN PLS_INTEGER;
  FUNCTION sum_slice_vc(p_nums IN pls_tab_typ) RETURN VARCHAR2;
  FUNCTION sum_slice_num(p_nums IN pls_tab_typ) RETURN NUMBER;
END TST_ora;`,
		`CREATE OR REPLACE PACKAGE BODY TST_ora AS
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
END TST_ora;`,
	} {
		if _, err := testSes.PrepAndExe(qry); err != nil {
			t.Fatal(err)
		}
	}

	enableLogging(t)
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
			{"BEGIN TST_ora.slice(:1); END;", []interface{}{numbers}, nil},
			{"BEGIN :1 := TST_ora.count_slice_vc(:2); END;", []interface{}{&retStr, numbers}, "COUNT=9"},
			{"BEGIN :1 := TST_ora.sum_slice_vc(:2); END;", []interface{}{&retStr, numbers}, "SUM=36"},
			{"BEGIN :1 := TST_ora.count_slice_int(:2); END;", []interface{}{&retInt, numbers}, int32(9)},
			{"BEGIN :1 := TST_ora.sum_slice_num(:2); END;", []interface{}{&retNum, numbers}, float64(36)},
		} {

			if _, err := testSes.PrepAndExe(tc.qry, tc.params...); err != nil {
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
