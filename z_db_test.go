//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/rana/ora.v4"
)

func Test_open_cursors_db(t *testing.T) {
	//enableLogging(t)
	// This needs "GRANT SELECT ANY DICTIONARY TO test"
	// or at least "GRANT SELECT ON v_$mystat TO test".
	// use 'opened cursors current' statistic#=5 to determine opened cursors on oracle server
	// SELECT A.STATISTIC#, A.NAME, B.VALUE
	// FROM V$STATNAME A, V$MYSTAT B
	// WHERE A.STATISTIC# = B.STATISTIC#
	qry := "SELECT VALUE FROM V$MYSTAT WHERE STATISTIC#=5"
	stmt, err := testDb.Prepare(qry)
	if err != nil {
		t.Fatalf("%q: %v", qry, err)
	}
	var before, after int
	if err = stmt.QueryRow().Scan(&before); err != nil {
		t.Skipf("%q: %v", qry, err)
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

func TestSelectNull_db(t *testing.T) {
	t.Parallel()
	cfg := ora.Cfg()
	cfg.Log.Rset.BeginRow = true
	ora.SetCfg(cfg)
	//enableLogging(t)
	var (
		s   string
		oS  ora.String
		i   int64
		oI  ora.Int64
		tim ora.Time
	)
	for tN, tC := range []struct {
		Field string
		Dest  interface{}
	}{
		{"''", &s},
		{"''", &oS},
		{"NULL + 0", &i},
		{"NULL + 0", &oI},
		{"SYSDATE + NULL", &tim},
	} {
		qry := "SELECT " + tC.Field + " x FROM DUAL"
		rows, err := testDb.Query(qry)
		if err != nil {
			t.Errorf("%d. %s: %v", tN, qry, err)
			return
		}
		for rows.Next() {
			if err = rows.Scan(&tC.Dest); err != nil {
				t.Errorf("%d. Scan: %v", tN, err)
				break
			}
		}
		if rows.Err() != nil {
			t.Errorf("%d. rows: %v", tN, rows.Err())
		}
		rows.Close()
	}
}

func TestSetConnMaxLifetime(t *testing.T) {
	var db *sql.DB
	var err error
	db, err = sql.Open("ora", testConStr)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Second)
	defer db.Close()

	done := make(chan struct{})
	var wg sync.WaitGroup
	dbRoutine := func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			var temp int
			db.QueryRow("SELECT 1 FROM DUAL").Scan(&temp)
			if rand.Int()%10 == 0 {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go dbRoutine()
	}
	time.Sleep(3 * time.Second)
	close(done)
	wg.Wait()
}

func Test_numberP38S0Identity_db(t *testing.T) {
	t.Parallel()
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
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := db.Ping(); err != nil {
			t.Skipf("%q: %v", dsn, err)
		}
	}()

	select {
	case <-time.After(10 * time.Second):
		t.Error("Sysdba test timed out!")
	case <-done:
		return
	}
}

func TestZeroRowsAffected(t *testing.T) {
	t.Parallel()
	tableName := tableName()
	if _, err := testDb.Exec("CREATE TABLE " + tableName + " (id NUMBER(3))"); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec("DROP TABLE " + tableName)
	res, err := testDb.Exec("UPDATE " + tableName + " SET id=1 WHERE 1=0")
	if err != nil {
		t.Fatal(err)
	}
	if ra, err := res.RowsAffected(); err != nil {
		t.Error(err)
	} else if ra != 0 {
		t.Errorf("got %d, wanted 0 rows affected!", ra)
	}
	if _, err := res.LastInsertId(); err == nil {
		t.Error("wanted error for LastInsertId, got nil")
	}
}

func Test_db(t *testing.T) {
	for valName, tc := range map[string]struct {
		gen    func() interface{}
		cTypes []string
	}{
		"int64": {
			gen:    func() interface{} { return gen_int64() },
			cTypes: []string{"numberP38S0", "numberP38S0Null"},
		},
		"float64": {
			gen: func() interface{} { return gen_float64() },
			cTypes: []string{
				"numberP16S15", "numberP16S15Null",
				"binaryDouble", "binaryDoubleNull",
				"binaryFloat", "binaryFloatNull",
				"floatP126", "floatP126Null",
			},
		},

		"date": {
			gen:    func() interface{} { return gen_date() },
			cTypes: []string{"date", "dateNull"},
		},

		"time": {
			gen: func() interface{} { return gen_time() },
			cTypes: []string{
				"timestampP9", "timestampP9Null",
				"timestampTzP9", "timestampTzP9Null",
				"timestampLtzP9", "timestampLtzP9Null",
			},
		},

		"string48": {
			gen: func() interface{} { return gen_string48() },
			cTypes: []string{
				"charB48", "charB48Null",
				"charC48", "charC48Null",
				"nchar48", "nchar48Null",
			},
		},

		"string": {
			gen: func() interface{} { return gen_string() },
			cTypes: []string{
				"varcharB48", "varcharB48Null",
				"varcharC48", "varcharC48Null",
				"varchar2B48", "varchar2B48Null",
				"varchar2C48", "varchar2C48Null",
				"nvarchar248", "nvarchar248Null",
				"long", "longNull",
				"clob", "clobNull",
				"nclob", "nclobNull",
			},
		},

		"bytes9": {
			gen: func() interface{} { return gen_bytes(9) },
			cTypes: []string{
				"longRaw", "longRawNull",
				"raw2000", "raw2000Null",
				"blob", "blobNull",
			},
		},
	} {
		tc := tc
		for _, ctName := range tc.cTypes {
			ct := _T_colType[ctName]
			t.Run(ctName+"_"+valName, func(t *testing.T) {
				t.Parallel()
				if strings.Contains(ctName, "clob") {
					enableLogging(t)
				}
				testBindDefineDB(tc.gen(), t, ct)
			})
		}
	}

	for _, ctName := range []string{
		"charB1", "charB1Null",
		"charC1", "charC1Null",
	} {
		ct := _T_colType[ctName]
		t.Run(ctName+"_bool", func(t *testing.T) {
			cfg := ora.Cfg()
			defer ora.SetCfg(cfg)
			ora.SetCfg(cfg.SetChar1(ora.B))
			//enableLogging(t)
			testBindDefineDB(gen_boolTrue(), t, ct)
		})
	}
}
