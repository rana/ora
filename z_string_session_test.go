//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/rana/ora.v4"
)

var _T_stringGen = map[string](func() interface{}){
	"string48":        func() interface{} { return gen_string48() },
	"OraString48":     func() interface{} { return gen_OraString48(false) },
	"OraString48Null": func() interface{} { return gen_OraString48(true) },
}

var _T_stringCols = []string{
	"charB48", "charB48Null",
	"charC48", "charC48Null",
	"nchar48", "nchar48Null",
	"varcharB48", "varcharB48Null",
	"varcharC48", "varcharC48Null",
	"varchar2B48", "varchar2B48Null",
	"varchar2C48", "varchar2C48Null",
	"nvarchar248", "nvarchar248Null",
}

func TestBindDefine_string(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, ctName := range _T_stringCols {
		for valName, gen := range _T_stringGen {
			t.Run(fmt.Sprintf("%s_%s", valName, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

func TestBindSlice_string(t *testing.T) {
	sc := ora.NewStmtCfg()
	for valName, gen := range map[string](func() interface{}){
		"stringSlice48":        func() interface{} { return gen_stringSlice48() },
		"OraStringSlice48":     func() interface{} { return gen_OraStringSlice48(false) },
		"OraStringSlice48Null": func() interface{} { return gen_OraStringSlice48(true) },
	} {
		for _, ctName := range _T_stringCols {
			t.Run(fmt.Sprintf("%s_%s", valName, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

func TestMultiDefine_string(t *testing.T) {
	for _, ctName := range _T_stringCols {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(gen_string48(), _T_colType[ctName], t)
		})
	}
}

func TestWorkload_charB48_session(t *testing.T) {
	for _, ctName := range _T_stringCols {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testWorkload(_T_colType[ctName], t)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// long
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_long(t *testing.T) {
	sc := ora.NewStmtCfg()
	for valName, gen := range map[string](func() interface{}){
		"string":             func() interface{} { return gen_string() },
		"stringSlice":        func() interface{} { return gen_stringSlice() },
		"OraString":          func() interface{} { return gen_OraString(false) },
		"OraStringSlice":     func() interface{} { return gen_OraString(false) },
		"OraStringNull":      func() interface{} { return gen_OraString(true) },
		"OraStringSliceNull": func() interface{} { return gen_OraString(true) },
	} {
		for _, ctName := range []string{
			"long", "longNull",
			"clob", "clobNull",
			"nclob", "nclobNull",
		} {
			if strings.HasSuffix(valName, "Null") && !strings.HasSuffix(ctName, "Null") {
				continue
			}
			t.Run(valName+"_"+ctName, func(t *testing.T) {
				if !strings.Contains(ctName, "lob") {
					t.Parallel()
				}
				testBindDefine(gen(), _T_colType[ctName], t, sc)
			})
		}
	}
}

//func TestBindPtr_string_long_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), long, t)
//}

func TestMultiDefine_long_session(t *testing.T) {
	for _, ctName := range []string{
		"long", "longNull",
		"clob", "clobNull",
		"nclob", "nclobNull",
	} {
		t.Run(ctName, func(t *testing.T) {
			t.Parallel()
			testMultiDefine(gen_string(), _T_colType[ctName], t)
		})
	}
}

//func TestWorkload_long_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(long, t)
//}

//func TestBindPtr_string_longNull_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), longNull, t)
//}

//func TestWorkload_longNull_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(longNull, t)
//}

func TestStringSlice(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	for _, nls_param := range []string{
		//`NLS_LANGUAGE = 'hungarian_hungary.ee9iso8859p2'`,
		`NLS_DATE_FORMAT = 'YYYY-MM-DD"T"HH24:MI:SS'`,
		`NLS_TIMESTAMP_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF'`,
		`NLS_NUMERIC_CHARACTERS = '.,'`,
	} {
		qry := "ALTER SESSION SET " + nls_param
		if _, err := testSes.PrepAndExe(qry); err != nil {
			t.Fatal(qry, err)
		}
	}
	tbl := tableName()
	qry := "CREATE TABLE " + tbl + ` (
  "BUC_BUG_ID" NUMBER(9,0),
  "BUC_NUMMER" NUMBER(7,0),
  "BUC_DEBITOR_KREDITOR_KENNZ" VARCHAR2(38),
  "BUC_SACHKONTENGRUPPE" VARCHAR2(32),
  "BUC_SKT_CODE" VARCHAR2(25),
  "BUC_KTK_CODE" VARCHAR2(25),
  "BUC_WAE_CODE" VARCHAR2(25),
  "BUC_BETRAG" NUMBER(15,2),
  "BUC_STEUERSCHLUESSEL" VARCHAR2(32),
  "BUC_ZAHLUNGSZIEL" DATE,
  "BUC_TYP" VARCHAR2(19),
  "BUC_MANDANT_PAR_NUMMER" NUMBER(11,0),
  "BUC_MANDANT_PAD_CODE" VARCHAR2(32),
  "BUC_MANDANT_FIBU_KONTONR" NUMBER(16,0),
  "BUC_PARTNER_PAD_CODE" VARCHAR2(32),
  "BUC_PARTNER_NUMMER" NUMBER(11,0),
  "BUC_PARTNER_FIBU_KONTONR" NUMBER(16,0),
  "BUC_VER_NUMMER" NUMBER(11,0),
  "BUC_VVS_NUMMER" NUMBER(9,0),
  "BUC_DEC_NUMMER" NUMBER(9,0),
  "BUC_PRD_NUMMER" VARCHAR2(27),
  "BUC_PBS_NUMMER" NUMBER(9,0),
  "BUC_OBJ_ID" NUMBER(9,0),
  "BUC_OBJ_NUMMER" NUMBER(9,0),
  "BUC_PARTNER_KONTONUMMER" VARCHAR2(36),
  "BUC_PARTNER_BANKLEITZAHL" VARCHAR2(37),
  "BUC_TEXT" VARCHAR2(74),
  "BUC_ERSTELLT_VON" VARCHAR2(29),
  "BUC_ERSTELLUNGSZEITPUNKT" DATE,
  "BUC_GEAENDERT_VON" VARCHAR2(30),
  "BUC_AENDERUNGSZEITPUNKT" DATE,
  "BUC_DPR_ID" NUMBER(9,0),
  "BUC_BETRAG_HW" NUMBER(15,2),
  "BUC_PBT_ID" NUMBER(9,0),
  "BUC_PARTNER_BANKKONTOINHABER" NUMBER(11,0),
  "BUC_PVN_ID" NUMBER(9,0),
  "BUC_KTO_LFDNUMMER" NUMBER(9,0),
  "BUC_BTP_ID" NUMBER(9,0),
  "BUC_MAKLER_ABRECHNUNGS_KNZ" VARCHAR2(38),
  "BUC_SWIFTCODE" VARCHAR2(26),
  "BUC_BANKNAME" VARCHAR2(25),
  "BUC_BANKORT" VARCHAR2(24),
  "BUC_IBAN" VARCHAR2(21),
  "BUC_RV_VERTRAGSNUMMER" NUMBER(12,0),
  "BUC_RV_VERTRAGSVERSION" NUMBER(10,0),
  "BUC_RV_VERTRAGSART" NUMBER(2,0),
  "BUC_HTV_CODE" VARCHAR2(25),
  "BUC_BELEGART_NB" VARCHAR2(27),
  "BUC_VSV_EXTERNER_CODE" VARCHAR2(33),
  "BUC_ORG_NUMMER_VEAB" VARCHAR2(32),
  "BUC_ORG_NUMMER_VTVZ" VARCHAR2(32),
  "BUC_KOSTENSTELLE_VM" VARCHAR2(32),
  "BUC_KIRCHENSTEUERSATZ" NUMBER(5,2),
  "BUC_RELIGIONSGEMEINSCHAFT" VARCHAR2(38)
  , "BUC_COC_SPERR_ID" RAW(16)
  , "BUC_COC_LOESCH_ID" RAW(16)
  )`
	if _, err := testSes.PrepAndExe(qry); err != nil {
		t.Fatal(qry, err)
	}
	defer testSes.PrepAndExe("DROP TABLE " + tbl)
	qry = "INSERT INTO " + tbl + `
	("BUC_BUG_ID", "BUC_NUMMER", "BUC_DEBITOR_KREDITOR_KENNZ", "BUC_SACHKONTENGRUPPE", "BUC_SKT_CODE", "BUC_KTK_CODE", "BUC_WAE_CODE", "BUC_BETRAG", "BUC_STEUERSCHLUESSEL", "BUC_ZAHLUNGSZIEL", "BUC_TYP", "BUC_MANDANT_PAR_NUMMER", "BUC_MANDANT_PAD_CODE", "BUC_MANDANT_FIBU_KONTONR", "BUC_PARTNER_PAD_CODE", "BUC_PARTNER_NUMMER", "BUC_PARTNER_FIBU_KONTONR", "BUC_VER_NUMMER", "BUC_VVS_NUMMER", "BUC_DEC_NUMMER", "BUC_PRD_NUMMER", "BUC_PBS_NUMMER", "BUC_OBJ_ID", "BUC_OBJ_NUMMER", "BUC_PARTNER_KONTONUMMER", "BUC_PARTNER_BANKLEITZAHL", "BUC_TEXT", "BUC_ERSTELLT_VON", "BUC_ERSTELLUNGSZEITPUNKT", "BUC_GEAENDERT_VON", "BUC_AENDERUNGSZEITPUNKT", "BUC_DPR_ID", "BUC_BETRAG_HW", "BUC_PBT_ID", "BUC_PARTNER_BANKKONTOINHABER", "BUC_PVN_ID", "BUC_KTO_LFDNUMMER", "BUC_BTP_ID", "BUC_MAKLER_ABRECHNUNGS_KNZ", "BUC_SWIFTCODE", "BUC_BANKNAME", "BUC_BANKORT", "BUC_IBAN", "BUC_RV_VERTRAGSNUMMER", "BUC_RV_VERTRAGSVERSION", "BUC_RV_VERTRAGSART", "BUC_HTV_CODE", "BUC_BELEGART_NB", "BUC_VSV_EXTERNER_CODE", "BUC_ORG_NUMMER_VEAB", "BUC_ORG_NUMMER_VTVZ", "BUC_KOSTENSTELLE_VM", "BUC_KIRCHENSTEUERSATZ", "BUC_RELIGIONSGEMEINSCHAFT", "BUC_COC_SPERR_ID", "BUC_COC_LOESCH_ID")
	VALUES
	(
	:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12, :13, :14, :15, :16, :17, :18, :19, :20, :21, :22, :23, :24, :25, :26, :27, :28, :29, :30, :31, :32, :33, :34, :35, :36, :37, :38, :39, :40, :41, :42, :43, :44, :45, :46, :47, :48, :49, :50, :51, :52, :53, :54
	, :55, :56
	)`
	stmt, err := testSes.Prep(qry)
	if err != nil {
		t.Fatal(qry, err)
	}
	defer stmt.Close()

	params := []interface{}{
		[]string{"3979", "3979"}, []string{"5", "6"}, []string{"D", "D"}, []string{"03", "03"}, []string{"429999", "429999"}, []string{"P10DLP000000901", "P10DLP000000901"}, []string{"HUF", "HUF"}, []string{"215.00", "171.00"}, []string{"S0", "S0"}, []string{"2007-11-12T00:00:00", "2007-11-12T00:00:00"}, []string{"F", "F"}, []string{"1", "1"}, []string{"MAN", "MAN"}, []string{"0", "0"}, []string{"VN", "VN"}, []string{"870009593", "870009593"}, []string{"870009593", "870009593"}, []string{"90000044121", "90000044132"}, []string{"1", "1"}, []string{"2", "1"}, []string{"01", "03"}, []string{"2", "1"}, []string{"5680656", "0"}, []string{"1", "0"}, []string{"1034827949010019", "1034827949010019"}, []string{"10300002", "10300002"}, []string{"háztartási vagyonrész betöréses lopás", "Magánszemélyek felelősségbiztosítása"}, []string{"BARCSAI", "BARCSAI"}, []string{"2007-11-12T13:04:22", "2007-11-12T13:04:25"}, []string{"", ""}, []string{"", ""}, []string{"6718", "6719"}, []string{"215.00", "171.00"}, []string{"0", "0"}, []string{"870009593", "870009593"}, []string{"0", "0"}, []string{"1", "1"}, []string{"0", "0"}, []string{"", ""}, []string{"", ""}, []string{"", ""}, []string{"", ""}, []string{"", ""}, []string{"0", "0"}, []string{"0", "0"}, []string{"0", "0"}, []string{"A110A110", "A110A110"}, []string{"11", "11"}, []string{"Z0", "Z0"}, []string{"", ""}, []string{"", ""}, []string{"", ""}, []string{"0", "0"}, []string{"", ""},
		[][]uint8{[]uint8(nil), []uint8(nil)}, [][]uint8{[]uint8(nil), []uint8(nil)},
	}

	for i := 0; i < 10; i++ {
		t.Logf("%d.", i)
		if _, err := stmt.Exe(params...); err != nil {
			t.Fatal(i, err)
		}
	}

	qry = "SELECT COUNT(0) FROM " + tbl
	rset, err := testSes.PrepAndQry(qry)
	if err != nil {
		t.Fatal(qry, err)
	}
	rset.Next()
	t.Log(rset.Row[0])

}
