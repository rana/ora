//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"gopkg.in/rana/ora.v3"
)

func Test_open_cursors_db(t *testing.T) {
	//enableLogging(t)
	// This needs "GRANT SELECT ANY DICTIONARY TO test"
	// or at least "GRANT SELECT ON v_$mystat TO test".
	// use 'opened cursors current' statistic#=5 to determine opened cursors on oracle server
	// SELECT A.STATISTIC#, A.NAME, B.VALUE
	// FROM V$STATNAME A, V$MYSTAT B
	// WHERE A.STATISTIC# = B.STATISTIC#
	stmt, err := testDb.Prepare("SELECT VALUE FROM V$MYSTAT WHERE STATISTIC#=5")
	if err != nil {
		t.Fatal(err)
	}
	var before, after int
	if err = stmt.QueryRow().Scan(&before); err != nil {
		t.Skip(err)
	}
	rounds := 100
	for i := 0; i < rounds; i++ {
		func() {
			stmt, err := testDb.Prepare("SELECT 1 FROM user_objects WHERE ROWNUM < 100")
			if err != nil {
				t.Fatal(err)
			}
			defer stmt.Close()
			rows, err := stmt.Query()
			if err != nil {
				t.Errorf("SELECT: %v", err)
				return
			}
			defer rows.Close()
			j := 0
			for rows.Next() {
				j++
			}
			//t.Logf("%d objects, error=%v", j, rows.Err())
		}()
	}
	if err = stmt.QueryRow().Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after-before >= rounds {
		t.Errorf("before=%d after=%d, awaited less than %d increment!", before, after, rounds)
		return
	}
	t.Logf("before=%d after=%d", before, after)
}

func TestSelectNullString_db(t *testing.T) {
	ora.Cfg().Log.Rset.BeginRow = true
	//enableLogging(t)
	var s string
	rows, err := testDb.Query("SELECT '' x FROM DUAL")
	if err != nil {
		t.Errorf("SELECT '' FROM DUAL: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&s); err != nil {
			t.Errorf("Scan: %v", err)
			break
		}
	}
	if rows.Err() != nil {
		t.Errorf("rows: %v", rows.Err())
	}
}

func Test_numberP38S0Identity_db(t *testing.T) {
	tableName := tableName()
	stmt, err := testDb.Prepare(createTableSql(tableName, 1, numberP38S0Identity, varchar2C48))
	if err == nil {
		defer stmt.Close()
		_, err = stmt.Exec()
	}
	if err != nil {
		t.Skipf("SKIP create table with identity: %v", err)
		return
	}
	defer dropTableDB(testDb, t, tableName)

	stmt, err = testDb.Prepare(fmt.Sprintf("insert into %v (c2) values ('go') returning c1 /*lastInsertId*/ into :c1", tableName))
	defer stmt.Close()

	// pass nil to Exec when using 'returning into' clause with sql.DB
	result, err := stmt.Exec(nil)
	testErr(err, t)
	actual, err := result.LastInsertId()
	testErr(err, t)
	if 1 != actual {
		t.Fatalf("LastInsertId: expected(%v), actual(%v)", 1, actual)
	}
}

func Test_numberP38S0_int64_db(t *testing.T) {
	testBindDefineDB(gen_int64(), t, numberP38S0)
}

func Test_numberP38S0Null_int64_db(t *testing.T) {
	testBindDefineDB(gen_int64(), t, numberP38S0Null)
}

func Test_numberP16S15_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, numberP16S15)
}

func Test_numberP16S15Null_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, numberP16S15Null)
}

func Test_binaryDouble_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, binaryDouble)
}

func Test_binaryDoubleNull_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, binaryDoubleNull)
}

func Test_binaryFloat_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, binaryFloat)
}

func Test_binaryFloatNull_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, binaryFloatNull)
}

func Test_floatP126_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, floatP126)
}

func Test_floatP126Null_float64_db(t *testing.T) {
	testBindDefineDB(gen_float64(), t, floatP126Null)
}

func Test_date_time_db(t *testing.T) {
	testBindDefineDB(gen_date(), t, date)
}

func Test_dateNull_time_db(t *testing.T) {
	testBindDefineDB(gen_date(), t, dateNull)
}

func Test_timestampP9_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampP9)
}

func Test_timestampP9Null_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampP9Null)
}

func Test_timestampTzP9_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampTzP9)
}

func Test_timestampTzP9Null_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampTzP9Null)
}

func Test_timestampLtzP9_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampLtzP9)
}

func Test_timestampLtzP9Null_time_db(t *testing.T) {
	testBindDefineDB(gen_time(), t, timestampLtzP9Null)
}

func Test_charB48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, charB48)
}

func Test_charB48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, charB48Null)
}

func Test_charC48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, charC48)
}

func Test_charC48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, charC48Null)
}

func Test_nchar48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nchar48)
}

func Test_nchar48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nchar48Null)
}

func Test_varcharB48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varcharB48)
}

func Test_varcharB48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varcharB48Null)
}

func Test_varcharC48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varcharC48)
}

func Test_varcharC48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varcharC48Null)
}

func Test_varchar2B48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varchar2B48)
}

func Test_varchar2B48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varchar2B48Null)
}

func Test_varchar2C48_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varchar2C48)
}

func Test_varchar2C48Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, varchar2C48Null)
}

func Test_nvarchar248_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nvarchar248)
}

func Test_nvarchar248Null_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nvarchar248Null)
}

func Test_long_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, long)
}

func Test_longNull_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, longNull)
}

func Test_clob_string_db(t *testing.T) {
	//enableLogging(t)
	testBindDefineDB(gen_string(), t, clob)
}

func Test_clobNull_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, clobNull)
}

func Test_nclob_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nclob)
}

func Test_nclobNull_string_db(t *testing.T) {
	testBindDefineDB(gen_string(), t, nclobNull)
}

func Test_charB1_bool_true_db(t *testing.T) {
	defer setC1Bool()()
	testBindDefineDB(gen_boolTrue(), t, charB1)
}

func Test_charB1Null_bool_true_db(t *testing.T) {
	//enableLogging(t)
	defer setC1Bool()()
	testBindDefineDB(gen_boolTrue(), t, charB1Null)
}

func Test_charC1_bool_true_db(t *testing.T) {
	//enableLogging(t)
	defer setC1Bool()()
	testBindDefineDB(gen_boolTrue(), t, charC1)
}

func Test_charC1Null_bool_true_db(t *testing.T) {
	defer setC1Bool()()
	testBindDefineDB(gen_boolTrue(), t, charC1Null)
}

func Test_longRaw_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(9), t, longRaw)
}

func Test_longRawNull_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(9), t, longRawNull)
}

func Test_raw2000_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(2000), t, raw2000)
}

func Test_raw2000Null_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(2000), t, raw2000Null)
}

func Test_blob_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(9), t, blob)
}

func Test_blobNull_bytes_db(t *testing.T) {
	testBindDefineDB(gen_bytes(9), t, blobNull)
}

func TestSysdba(t *testing.T) {
	u := os.Getenv("GO_ORA_DRV_TEST_SYSDBA_USERNAME")
	p := os.Getenv("GO_ORA_DRV_TEST_SYSDBA_PASSWORD")
	if u == "" {
		u = testSesCfg.Username
		p = testSesCfg.Password
	}
	dsn := fmt.Sprintf("%s/%s@%s AS SYSDBA", u, p, testSrvCfg.Dblink)
	db, err := sql.Open("ora", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skipf("%q: %v", dsn, err)
	}
}
