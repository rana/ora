// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	_ "net/http/pprof"

	"github.com/pkg/errors"

	"gopkg.in/rana/ora.v4"
	"gopkg.in/rana/ora.v4/tstlg"
)

type oracleColumnType string

const (
	// numeric
	numberP38S0Identity oracleColumnType = "number(38,0) generated always as identity (start with 1 increment by 1)"
	numberP38S0         oracleColumnType = "number(38,0) not null"
	numberP38S0Null     oracleColumnType = "number(38,0) null"
	numberP16S15        oracleColumnType = "number(16,15) not null"
	numberP16S15Null    oracleColumnType = "number(16,15) null"
	binaryDouble        oracleColumnType = "binary_double not null"
	binaryDoubleNull    oracleColumnType = "binary_double null"
	binaryFloat         oracleColumnType = "binary_float not null"
	binaryFloatNull     oracleColumnType = "binary_float null"
	floatP126           oracleColumnType = "float(126) not null"
	floatP126Null       oracleColumnType = "float(126) null"
	// time
	dateNotNull        oracleColumnType = "date not null"
	dateNull           oracleColumnType = "date null"
	timestampP9        oracleColumnType = "timestamp(9) not null"
	timestampP9Null    oracleColumnType = "timestamp(9) null"
	timestampTzP9      oracleColumnType = "timestamp(9) with time zone not null"
	timestampTzP9Null  oracleColumnType = "timestamp(9) with time zone null"
	timestampLtzP9     oracleColumnType = "timestamp(9) with local time zone not null"
	timestampLtzP9Null oracleColumnType = "timestamp(9) with local time zone null"
	// interval
	intervalYM     oracleColumnType = "interval year to month not null"
	intervalYMNull oracleColumnType = "interval year to month null"
	intervalDS     oracleColumnType = "interval day to second(9) not null"
	intervalDSNull oracleColumnType = "interval day to second(9) null"
	// string
	charB48         oracleColumnType = "char(48 byte) not null"
	charB48Null     oracleColumnType = "char(48 byte) null"
	charC48         oracleColumnType = "char(48 char) not null"
	charC48Null     oracleColumnType = "char(48 char) null"
	nchar48         oracleColumnType = "nchar(48) not null"
	nchar48Null     oracleColumnType = "nchar(48) null"
	varcharB48      oracleColumnType = "varchar(48 byte) not null"
	varcharB48Null  oracleColumnType = "varchar(48 byte) null"
	varcharC48      oracleColumnType = "varchar(48 char) not null"
	varcharC48Null  oracleColumnType = "varchar(48 char) null"
	varchar2B48     oracleColumnType = "varchar2(48 byte) not null"
	varchar2B48Null oracleColumnType = "varchar2(48 byte) null"
	varchar2C48     oracleColumnType = "varchar2(48 char) not null"
	varchar2C48Null oracleColumnType = "varchar2(48 char) null"
	nvarchar248     oracleColumnType = "nvarchar2(48) not null"
	nvarchar248Null oracleColumnType = "nvarchar2(48) null"
	long            oracleColumnType = "long not null"
	longNull        oracleColumnType = "long null"
	clob            oracleColumnType = "clob not null"
	clobNull        oracleColumnType = "clob null"
	nclob           oracleColumnType = "nclob not null"
	nclobNull       oracleColumnType = "nclob null"
	// string or bool
	charB1     oracleColumnType = "char(1 byte) not null"
	charB1Null oracleColumnType = "char(1 byte) null"
	charC1     oracleColumnType = "char(1 char) not null"
	charC1Null oracleColumnType = "char(1 char) null"
	// bytes
	longRaw     oracleColumnType = "long raw not null"
	longRawNull oracleColumnType = "long raw null"
	raw2000     oracleColumnType = "raw(2000) not null"
	raw2000Null oracleColumnType = "raw(2000) null"
	blob        oracleColumnType = "blob not null"
	blobNull    oracleColumnType = "blob null"
	// bfile
	bfile     oracleColumnType = "bfile not null"
	bfileNull oracleColumnType = "bfile null"
)

var _T_colType = map[string]oracleColumnType{
	"charB1":     charB1,
	"charB1Null": charB1Null,
	"charC1":     charC1,
	"charC1Null": charC1Null,

	"longRaw":     longRaw,
	"longRawNull": longRawNull,
	"raw2000":     raw2000,
	"raw2000Null": raw2000Null,
	"blob":        blob,
	"blobNull":    blobNull,

	"intervalYM":     intervalYM,
	"intervalYMNull": intervalYMNull,
	"intervalDS":     intervalDS,
	"intervalDSNull": intervalDSNull,

	"numberP38S0":      numberP38S0,
	"numberP38S0Null":  numberP38S0Null,
	"numberP16S15":     numberP16S15,
	"numberP16S15Null": numberP16S15Null,
	"binaryDouble":     binaryDouble,
	"binaryDoubleNull": binaryDoubleNull,
	"binaryFloat":      binaryFloat,
	"binaryFloatNull":  binaryFloatNull,
	"floatP126":        floatP126,
	"floatP126Null":    floatP126Null,

	"charB48":         charB48,
	"charB48Null":     charB48Null,
	"charC48":         charC48,
	"charC48Null":     charC48Null,
	"nchar48":         nchar48,
	"nchar48Null":     nchar48Null,
	"varcharB48":      varcharB48,
	"varcharB48Null":  varcharB48Null,
	"varcharC48":      varcharC48,
	"varcharC48Null":  varcharC48Null,
	"varchar2B48":     varchar2B48,
	"varchar2B48Null": varchar2B48Null,
	"varchar2C48":     varchar2C48,
	"varchar2C48Null": varchar2C48Null,
	"nvarchar248":     nvarchar248,
	"nvarchar248Null": nvarchar248Null,

	"long":      long,
	"longNull":  longNull,
	"clob":      clob,
	"clobNull":  clobNull,
	"nclob":     nclob,
	"nclobNull": nclobNull,

	"date":               dateNotNull,
	"dateNull":           dateNull,
	"time":               dateNotNull,
	"timeNull":           dateNull,
	"timestampP9":        timestampP9,
	"timestampP9Null":    timestampP9Null,
	"timestampTzP9":      timestampTzP9,
	"timestampTzP9Null":  timestampTzP9Null,
	"timestampLtzP9":     timestampLtzP9,
	"timestampLtzP9Null": timestampLtzP9Null,
}

var testSrvCfg ora.SrvCfg
var testSesCfg ora.SesCfg
var testConStr string
var testDbsessiontimezone *time.Location
var testTableID uint32
var testWorkloadColumnCount int
var testDb *sql.DB
var testSesPool *ora.Pool

var tableNamePrefix = fmt.Sprintf("test_%d_", os.Getpid())

func init() {
	testSrvCfg = ora.SrvCfg{
		Dblink:  os.Getenv("GO_ORA_DRV_TEST_DB"),
		StmtCfg: ora.NewStmtCfg(),
	}
	if testSrvCfg.IsZero() {
		panic("testSrvCfg is Zero")
	}
	testSesCfg = ora.SesCfg{
		Username: os.Getenv("GO_ORA_DRV_TEST_USERNAME"),
		Password: os.Getenv("GO_ORA_DRV_TEST_PASSWORD"),
		StmtCfg:  testSrvCfg.StmtCfg,
	}
	if testSesCfg.IsZero() {
		panic("testSesCfg is Zero")
	}
	testConStr = fmt.Sprintf("%v/%v@%v", testSesCfg.Username, testSesCfg.Password, testSrvCfg.Dblink)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_DB = '%v'\n", testSrvCfg.Dblink)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_USERNAME = '%v'\n", testSesCfg.Username)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_PASSWORD = '%v'\n", testSesCfg.Password)

	testWorkloadColumnCount = 20
	var err error

	// setup test environment, server and session
	testEnv, err := ora.OpenEnv()
	if err != nil {
		panic(fmt.Sprintf("initError: %v", err))
	}

	testSesPool = testEnv.NewPool(testSrvCfg, testSesCfg, 4)
	testSes, err := testSesPool.Get()
	if err != nil {
		panic(fmt.Sprintf("initError: %v", err))
	}
	defer testSes.Close()

	//ora.SetCfg(func() StmtCfg { cfg = ora.Cfg(); cfg.RTrimChar = false; return cfg }())

	// load session time zone
	testDbsessiontimezone, err = loadDbtimezone()
	if err != nil {
		panic("Error loading session time zone from database: " + err.Error())
	}
	fmt.Printf("Read session time zone from database: %s\n", testDbsessiontimezone)

	// drop all tables from previous test run
	fmt.Println("Dropping previous tables...")
	stmt, err := testSes.Prep(`
BEGIN
	FOR c IN (SELECT table_name FROM user_tables WHERE TABLE_NAME LIKE UPPER('` + tableNamePrefix + `')||'%') LOOP
		EXECUTE IMMEDIATE ('DROP TABLE ' || c.table_name || ' CASCADE CONSTRAINTS');
	END LOOP;
END;`)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	defer stmt.Close()
	start := time.Now()
	_, err = stmt.Exe()
	if err != nil {
		fmt.Println("initError: ", err)
	}
	fmt.Printf("Tables dropped (%s).\n", time.Since(start))

	// setup test db

	start = time.Now()
	testDb, err = sql.Open(ora.Name, testConStr)
	if err != nil {
		fmt.Println("initError: ", err)
	} else {
		fmt.Printf("Connected to %q (%s).\n", testConStr, time.Since(start))
	}

	if os.Getenv("BLOCKPROFILE") == "1" {
		go func() {
			addr := "localhost:8642"
			fmt.Println("blockprofile: go tool pprof http://" + addr + "/debug/pprof/block")
			runtime.SetBlockProfileRate(1)
			fmt.Println(http.ListenAndServe(addr, nil))
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = testDb.PingContext(ctx)
	cancel()
	if err != nil {
		panic(err)
	}

	if err := testSes.Ping(); err != nil {
		panic(err)
	}
}

func getSes(t testing.TB) *ora.Ses {
	testSes, err := testSesPool.Get()
	if err == nil {
		return testSes
	}
	t.Fatal(err)
	return nil
}

func enableLogging(t *testing.T) {
	if t == nil {
		return
	}
	cfg := ora.Cfg()
	cfg.Log.Logger = tstlg.New(t)
	ora.SetCfg(cfg)
}

func testIterations() int {
	if testing.Short() {
		return 1
	} else {
		return 1
	}
}

func testBindDefine(expected interface{}, oct oracleColumnType, t *testing.T, c ora.StmtCfg, goColumnTypes ...ora.GoColumnType) {
	var gct ora.GoColumnType
	if len(goColumnTypes) > 0 {
		gct = goColumnTypes[0]
	} else {
		gct = goColumnTypeFromValue(expected)
	}
	t.Logf("testBindDefine gct (%v, %v)", gct, ora.GctName(gct))

	testSes, err := testSesPool.Get()
	if err != nil {
		t.Fatal(err)
	}
	defer testSes.Close()

	tableName, err := createTable(1, oct, testSes)
	testErr(err, t)
	defer dropTable(tableName, testSes, t)

	// insert
	qry := fmt.Sprintf("insert into %v (c1) values (:c1)", tableName)
	insertStmt, err := testSes.Prep(qry)
	testErr(errors.Wrap(err, qry), t)
	defer insertStmt.Close()
	if !c.IsZero() {
		insertStmt.SetCfg(c)
	}
	rowsAffected, err := insertStmt.Exe(expected)
	testErr(errors.Wrapf(err, "%q, %#v", qry, expected), t)
	expLen := length(expected)
	if gct == ora.Bin || gct == ora.OraBin {
		expLen = 1
	}
	if expLen != int(rowsAffected) {
		t.Fatalf("insert rows affected: expected(%v), actual(%v)", expLen, rowsAffected)
	}

	// select
	selectStmt, err := testSes.Prep(fmt.Sprintf("select c1 from %v", tableName), gct)
	testErr(err, t)
	defer selectStmt.Close()
	rset, err := selectStmt.Qry()
	testErr(err, t)
	defer rset.Exhaust()
	// validate
	validate(expected, rset, t)
}

func testBindDefineDB(expected interface{}, t *testing.T, oct oracleColumnType) {
	for n := 0; n < testIterations(); n++ {
		func() {
			tableName := createTableDB(testDb, t, oct)
			defer dropTableDB(testDb, t, tableName)

			// insert
			stmt, err := testDb.Prepare(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
			testErr(err, t)
			defer stmt.Close()
			execResult, err := stmt.Exec(expected)
			testErr(err, t)
			rowsAffected, err := execResult.RowsAffected()
			testErr(err, t)
			if 1 != rowsAffected {
				t.Fatalf("insert rowsAffected: expected(%v), actual(%v)", 1, rowsAffected)
			}

			// query
			rows, err := testDb.Query(fmt.Sprintf("select c1 from %v", tableName))
			testErr(err, t)
			defer rows.Close()
			testErr(err, t)
			if rows == nil {
				t.Fatalf("no rows returned")
			} else {
				var rowCount int
				var goColumnType ora.GoColumnType
				if oct == longRaw || oct == longRawNull || oct == raw2000 || oct == raw2000Null || oct == blob || oct == blobNull {
					goColumnType = ora.Bin
				} else {
					goColumnType = goColumnTypeFromValue(expected)
				}
				for rows.Next() {
					var actual interface{}
					err := rows.Scan(&actual)
					testErr(err, t)
					compare(expected, actual, goColumnType, t)
					rowCount++
				}
				if 1 != rowCount {
					t.Fatalf("query row count: expected(%v), actual(%v)", 1, rowCount)
				}
			}
		}()
	}
}

func testBindPtr(expected interface{}, oct oracleColumnType, t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	t.Logf("expected=%T", expected)
	for n := 0; n < testIterations(); n++ {
		tableName, err := createTable(1, oct, testSes)
		testErr(err, t)
		defer dropTable(tableName, testSes, t)

		// create pointer to receive actual value
		var actual interface{}
		switch expected.(type) {
		case int64:
			var value int64
			actual = &value
		case int32:
			var value int32
			actual = &value
		case int16:
			var value int16
			actual = &value
		case int8:
			var value int8
			actual = &value
		case uint64:
			var value uint64
			actual = &value
		case uint32:
			var value uint32
			actual = &value
		case uint16:
			var value uint16
			actual = &value
		case uint8:
			var value uint8
			actual = &value
		case float64:
			var value float64
			actual = &value
		case float32:
			var value float32
			actual = &value
		case time.Time:
			var value time.Time
			actual = &value
		case string:
			var value string
			actual = &value
		case bool:
			var value bool
			actual = &value
		default:
			t.Fatalf("no value for %T", expected)
		}

		// insert
		qry := "insert into " + tableName + " (c1) values (:1) returning c1 into :2"
		stmt, err := testSes.Prep(qry)
		testErr(err, t)
		defer stmt.Close()
		t.Logf("%q, [%#v %T]", qry, expected, actual)
		rowsAffected, err := stmt.Exe(expected, actual)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// validate
		t.Logf("actual=%T (%v)", actual, actual)
		compare2(expected, actual, t)
	}
}

func testMultiDefine(expected interface{}, oct oracleColumnType, t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	for n := 0; n < testIterations(); n++ {
		tableName, err := createTable(1, oct, testSes)
		testErr(err, t)
		defer dropTable(tableName, testSes, t)

		// insert
		insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
		testErr(err, t)
		defer insertStmt.Close()
		rowsAffected, err := insertStmt.Exe(expected)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// select
		var qry string
		var params []ora.GoColumnType
		if isNumeric(expected) {
			qry = fmt.Sprintf("select c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1 from %v", tableName)
			params = append(params, ora.I64, ora.I32, ora.I16, ora.I8, ora.U64, ora.U32, ora.U16, ora.U8, ora.F64, ora.F32, ora.OraI64, ora.OraI32, ora.OraI16, ora.OraI8, ora.OraU64, ora.OraU32, ora.OraU16, ora.OraU8, ora.OraF64, ora.OraF32)
		} else if isTime(expected) {
			qry = fmt.Sprintf("select c1, c1 from %v", tableName)
			params = append(params, ora.T, ora.OraT)
		} else if isString(expected) {
			qry = fmt.Sprintf("select c1 from %v", tableName)
			params = append(params, ora.S)
		} else if isBool(expected) {
			qry = fmt.Sprintf("select c1, c1 from %v", tableName)
			params = append(params, ora.B, ora.OraB)
		} else if isBytes(expected) {
			// one LOB cannot be opened twice in the same transaction (c1, c1 not works here)
			col := ora.Bin
			if n%2 == 1 {
				col = ora.OraBin
			}
			qry = fmt.Sprintf("select c1 from %v", tableName)
			params = append(params, col)
		}
		selectStmt, err := testSes.Prep(qry, params...)
		testErr(err, t)
		defer selectStmt.Close()
		rset, err := selectStmt.Qry()
		testErr(err, t)
		defer rset.Exhaust()

		// validate
		hasRow := rset.Next()
		testErr(errors.Wrapf(rset.Err(), "%q %v", qry, params), t)
		if !hasRow {
			t.Fatalf("no row returned")
		} else if len(rset.Row) != len(selectStmt.Gcts()) {
			t.Fatalf("select column count: expected(%v), actual(%v)", len(selectStmt.Gcts()), len(rset.Row))
		} else {
			for n, goColumnType := range selectStmt.Gcts() {
				if isNumeric(expected) {
					compare(castInt(expected, goColumnType), rset.Row[n], goColumnType, t)
				}
				switch goColumnType {
				case ora.T:
					compare_time(expected, rset.Row[n], t)
				case ora.OraT:
					if value, ok := rset.Row[n].(ora.Time); ok {
						compare_time(expected, value.Value, t)
						//} else if value, ok := rset.Row[n].(ora.Date); ok {
						//    compare_time(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value (got %v, expected %v). (%T, %v)", rset.Row[n], expected, rset.Row[n], rset.Row[n])
					}
				case ora.S:
					compare_string(expected, rset.Row[n], t)
				case ora.OraS:
					value, ok := rset.Row[n].(ora.String)
					if ok {
						compare_string(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value. (%T, %v)", rset.Row[n], rset.Row[n])
					}
				case ora.B, ora.OraB:
					compare_bool(expected, rset.Row[n], t)
				case ora.Bin, ora.OraBin:
					compare_bytes(expected, rset.Row[n], t)
				}
			}
		}
	}
}

// Workload tests proper functioning of bind struct re-use and define struct re-use.
// Bind structs and define structs are stored in pools on a per-type basis.
// This function also tests that bind/defines are cleared and properly reused.
// Insert and query are also exercised.
// Creating multiple columns of the same type will instantiate multiple bind/define structs.
// Running the insert and query multiple times will ensure reuse of those structs.
// Slice binding may have fewer columns tested due to OCI memory contraints.
func testWorkload(oct oracleColumnType, t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	for i := 0; i < testIterations(); i++ {
		currentMultiple := testWorkloadColumnCount
		for m := 0; m < 3 && currentMultiple > 0; m++ {
			tableName, err := createTable(currentMultiple, oct, testSes)
			testErr(err, t)
			defer dropTable(tableName, testSes, t)

			// build insert statement and values
			var sql bytes.Buffer
			sql.WriteString(fmt.Sprintf("insert into %v (", tableName))
			for c := 1; c <= currentMultiple; c++ {
				if c > 1 {
					sql.WriteString(", ")
				}
				sql.WriteString(fmt.Sprintf("c%v", c))
			}
			sql.WriteString(") values (")
			expected := make([]interface{}, currentMultiple)
			gcts := make([]ora.GoColumnType, currentMultiple)
			for c := 0; c < currentMultiple; c++ {
				switch oct {
				case numberP38S0, numberP38S0Null:
					expected[c] = gen_int64()
					gcts[c] = ora.I64
				case numberP16S15, numberP16S15Null, binaryDouble, binaryDoubleNull, floatP126, floatP126Null:
					expected[c] = gen_float64()
					gcts[c] = ora.F64
				case binaryFloat, binaryFloatNull:
					expected[c] = gen_float32()
					gcts[c] = ora.F32
				case dateNotNull, dateNull:
					expected[c] = gen_date()
					gcts[c] = ora.T
				case timestampP9, timestampP9Null, timestampTzP9, timestampTzP9Null, timestampLtzP9, timestampLtzP9Null:
					expected[c] = gen_time()
					gcts[c] = ora.T
				case charB48, charB48Null, charC48, charC48Null, nchar48, nchar48Null:
					expected[c] = gen_string48()
					gcts[c] = ora.S
				case varcharB48, varcharB48Null, varcharC48, varcharC48Null, varchar2B48, varchar2B48Null, varchar2C48, varchar2C48Null, nvarchar248, nvarchar248Null, long, longNull, clob, clobNull, nclob, nclobNull:
					expected[c] = gen_string()
					gcts[c] = ora.S
				case charB1, charB1Null, charC1, charC1Null:
					if gct := ora.Cfg().Char1(); gct == ora.B || gct == ora.OraB {
						expected[c] = gen_boolTrue()
						gcts[c] = ora.B
					} else {
						expected[c] = gen_string()[:1]
						gcts[c] = ora.S
					}
				case blob, blobNull, longRaw, longRawNull:
					expected[c] = gen_bytes(9)
					gcts[c] = ora.Bin
				case raw2000, raw2000Null:
					expected[c] = gen_bytes(2000)
					gcts[c] = ora.Bin
				}
				if c > 0 {
					sql.WriteString(", ")
				}
				sql.WriteString(fmt.Sprintf(":c%v", c+1))
			}
			sql.WriteString(")")

			// insert values
			//fmt.Println(sql.String())
			_, err = testSes.PrepAndExe(sql.String(), expected...)
			testErr(err, t)
			//			insertStmt, err := testCon.Prepare(sql.String())
			//			testErr(err, t)
			//			_, err = insertStmt.Exec(expected)
			//			testErr(err, t)
			//			insertStmt.Close()

			// fetch values and compare
			sql.Reset()
			sql.WriteString(fmt.Sprintf("select * from %v", tableName))
			fetchStmt, err := testSes.Prep(sql.String())
			testErr(err, t)
			fetchStmt.SetGcts(gcts)
			rset, err := fetchStmt.Qry()
			testErr(err, t)
			defer rset.Exhaust()
			for rset.Next() {
				if currentMultiple != len(rset.Row) {
					t.Fatalf("select column count: expected(%v), actual(%v)", currentMultiple, len(rset.Row))
				} else {
					for n := 0; n < currentMultiple; n++ {
						expectedElem := elemAt(expected, n)
						compare(expectedElem, rset.Row[n], gcts[n], t)
					}
				}
			}
			testErr(errors.Wrap(rset.Err(), sql.String()), t)
			fetchStmt.Close()

			// Reduce the multiple by half
			currentMultiple = currentMultiple / 2
		}
	}
}

func loadDbtimezone() (*time.Location, error) {
	testSes, err := testSesPool.Get()
	if err != nil {
		return nil, err
	}
	defer testSes.Close()

	return testSes.Timezone()
}

func validate(expected interface{}, rset *ora.Rset, t *testing.T) {
	row := rset.NextRow()
	//t.Logf("Row=%v (%v) Index=%d", rset.Row, row, rset.Index)
	if 1 != len(rset.Row) {
		t.Fatalf("column count: expected(%v), actual(%v)", 1, len(rset.Row))
	}

	switch expected.(type) {
	case int64:
		compare_int64(expected, row[0], t)
	case int32:
		compare_int32(expected, row[0], t)
	case int16:
		compare_int16(expected, row[0], t)
	case int8:
		compare_int8(expected, row[0], t)
	case uint64:
		compare_uint64(expected, row[0], t)
	case uint32:
		compare_uint32(expected, row[0], t)
	case uint16:
		compare_uint16(expected, row[0], t)
	case uint8:
		compare_uint8(expected, row[0], t)
	case float64:
		compare_float64(expected, row[0], t)
	case float32:
		compare_float32(expected, row[0], t)
	case ora.Int64:
		compare_OraInt64(expected, row[0], t)
	case ora.Int32:
		compare_OraInt32(expected, row[0], t)
	case ora.Int16:
		compare_OraInt16(expected, row[0], t)
	case ora.Int8:
		compare_OraInt8(expected, row[0], t)
	case ora.Uint64:
		compare_OraUint64(expected, row[0], t)
	case ora.Uint32:
		compare_OraUint32(expected, row[0], t)
	case ora.Uint16:
		compare_OraUint16(expected, row[0], t)
	case ora.Uint8:
		compare_OraUint8(expected, row[0], t)
	case ora.Float64:
		compare_OraFloat64(expected, row[0], t)
	case ora.Float32:
		compare_OraFloat32(expected, row[0], t)

	case ora.IntervalYM:
		compare_OraIntervalYM(expected, row[0], t)
	case ora.IntervalDS:
		compare_OraIntervalDS(expected, row[0], t)

	case ora.Bfile:
		compare_OraBfile(expected, row[0], t)

	case []int64:
		for {
			//t.Logf("Row=%v Index=%d", rset.Row, rset.Index)
			expectedElem := elemAt(expected, rset.Len()-1)
			compare_int64(expectedElem, rset.Row[0], t)
			if !rset.Next() {
				break
			}
		}

	case []ora.IntervalYM:
		for {
			expectedElem := elemAt(expected, rset.Len()-1)
			compare_OraIntervalYM(expectedElem, rset.Row[0], t)
			if !rset.Next() {
				break
			}
		}
	case []ora.IntervalDS:
		for {
			expectedElem := elemAt(expected, rset.Len()-1)
			compare_OraIntervalDS(expectedElem, rset.Row[0], t)
			if !rset.Next() {
				break
			}
		}
	}
	testErr(rset.Err(), t)
}

func compare2(expected interface{}, actual interface{}, t *testing.T) {
	switch expected.(type) {
	case int64:
		compare_int64(expected, actual, t)
	case int32:
		compare_int32(expected, actual, t)
	case int16:
		compare_int16(expected, actual, t)
	case int8:
		compare_int8(expected, actual, t)
	case uint64:
		compare_uint64(expected, actual, t)
	case uint32:
		compare_uint32(expected, actual, t)
	case uint16:
		compare_uint16(expected, actual, t)
	case uint8:
		compare_uint8(expected, actual, t)
	case float64:
		compare_float64(expected, actual, t)
	case float32:
		compare_float32(expected, actual, t)
	case ora.Int64:
		compare_OraInt64(expected, actual, t)
	case ora.Int32:
		t.Logf("actual=%T (%#v)", actual, actual)
		compare_OraInt32(expected, actual, t)
	case ora.Int16:
		compare_OraInt16(expected, actual, t)
	case ora.Int8:
		compare_OraInt8(expected, actual, t)
	case ora.Uint64:
		compare_OraUint64(expected, actual, t)
	case ora.Uint32:
		compare_OraUint32(expected, actual, t)
	case ora.Uint16:
		compare_OraUint16(expected, actual, t)
	case ora.Uint8:
		compare_OraUint8(expected, actual, t)
	case ora.Float64:
		compare_OraFloat64(expected, actual, t)
	case ora.Float32:
		compare_OraFloat32(expected, actual, t)
	case ora.IntervalYM:
		compare_OraIntervalYM(expected, actual, t)
	case ora.IntervalDS:
		compare_OraIntervalDS(expected, actual, t)
	}
}

func createTable(multiple int, oct oracleColumnType, ses *ora.Ses) (string, error) {
	tableName := fmt.Sprintf("%v_%v", tableName(), multiple)
	qry := createTableSql(tableName, multiple, oct)
	stmt, err := ses.Prep(qry)
	if err != nil {
		return "", errors.Wrap(err, qry)
	}
	defer stmt.Close()
	_, err = stmt.Exe()
	return tableName, errors.Wrap(err, qry)
}

func dropTable(tableName string, ses *ora.Ses, t testing.TB) {
	qry := fmt.Sprintf("drop table %v", tableName)
	stmt, err := ses.Prep(qry)
	if err != nil {
		t.Log(err)
		return
	}
	//testErr(errors.Wrap(err, qry), t)
	defer stmt.Close()
	if _, err = stmt.Exe(); err != nil {
		t.Log(err)
	}
	//testErr(errors.Wrap(err, qry), t)
}

func createTableDB(db *sql.DB, t *testing.T, octs ...oracleColumnType) string {
	tableName := tableName()
	qry := createTableSql(tableName, 1, octs...)
	stmt, err := db.Prepare(qry)
	testErr(errors.Wrap(err, qry), t)
	defer stmt.Close()
	_, err = stmt.Exec()
	testErr(errors.Wrap(err, qry), t)
	return tableName
}

func dropTableDB(db *sql.DB, t *testing.T, tableName string) {
	qry := fmt.Sprintf("drop table %v", tableName)
	stmt, err := db.Prepare(qry)
	testErr(errors.Wrap(err, qry), t)
	defer stmt.Close()
	_, err = stmt.Exec()
	testErr(errors.Wrap(err, qry), t)
}

func createTableSql(tableName string, multiple int, columns ...oracleColumnType) string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("create table %v (", tableName))
	for m := 1; m <= multiple; m++ {
		for n, column := range columns {
			position := (n + 1) * m
			if position > 1 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("c%v %v", position, column))
		}
	}
	b.WriteString(")")
	return b.String()
}

func tableName() string {
	nm := tableNamePrefix + strconv.FormatUint(uint64(atomic.AddUint32(&testTableID, 1)), 10)
	return nm
}

func testErr(err error, t testing.TB, expectedErrs ...error) {
	if err == nil {
		return
	}
	for _, expectedErr := range expectedErrs {
		if expectedErr == err { // skip it
			return
		}
	}
	done := make(chan struct{})
	msg := fmt.Sprintf("%+v: %s", err, getStack(2))
	if true {
		go func() {
			select {
			case <-time.After(300 * time.Second):
				fmt.Printf("\n\nPRINT TIMEOUT\n%s", msg)
			case <-done:
			}
		}()
	}
	t.Fatal(msg)
	close(done)
	if strings.Contains(err.Error(), "ORA-01000:") {
		os.Exit(1)
	}
}

func goColumnTypeFromValue(value interface{}) ora.GoColumnType {
	switch value.(type) {
	case int64, []int64:
		return ora.I64
	case int32, []int32:
		return ora.I32
	case int16, []int16:
		return ora.I16
	case int8, []int8:
		return ora.I8
	case uint64, []uint64:
		return ora.U64
	case uint32, []uint32:
		return ora.U32
	case uint16, []uint16:
		return ora.U16
	case uint8, []uint8:
		return ora.U8
	case float64, []float64:
		return ora.F64
	case float32, []float32:
		return ora.F32
	case ora.Int64, []ora.Int64:
		return ora.OraI64
	case ora.Int32, []ora.Int32:
		return ora.OraI32
	case ora.Int16, []ora.Int16:
		return ora.OraI16
	case ora.Int8, []ora.Int8:
		return ora.OraI8
	case ora.Uint64, []ora.Uint64:
		return ora.OraU64
	case ora.Uint32, []ora.Uint32:
		return ora.OraU32
	case ora.Uint16, []ora.Uint16:
		return ora.OraU16
	case ora.Uint8, []ora.Uint8:
		return ora.OraU8
	case ora.Float64, []ora.Float64:
		return ora.OraF64
	case ora.Float32, []ora.Float32:
		return ora.OraF32
	case ora.Num, []ora.Num:
		return ora.N
	case ora.OraNum, []ora.OraNum:
		return ora.OraN
	case time.Time, []time.Time:
		return ora.T
	case ora.Time, []ora.Time:
		return ora.OraT
	case string, []string:
		return ora.S
	case ora.String, []ora.String:
		return ora.OraS
	case bool, []bool:
		return ora.B
	case ora.Bool, []ora.Bool:
		return ora.OraB
	case ora.Raw:
		return ora.OraBin
	}
	return ora.D
}

func isNumeric(value interface{}) bool {
	switch value.(type) {
	case int8, int16, int32, int64, int, ora.Int8, ora.Int16, ora.Int32, ora.Int64:
		return true
	case uint, uint8, uint16, uint32, uint64, ora.Uint8, ora.Uint16, ora.Uint32, ora.Uint64:
		return true
	case float32, float64, ora.Float32, ora.Float64, ora.Num, ora.OraNum:
		return true
	}
	return false
}

func isTime(value interface{}) bool {
	if _, ok := value.(time.Time); ok {
		return true
	}
	if _, ok := value.(ora.Time); ok {
		return true
	}
	return false
}

func isString(value interface{}) bool {
	if _, ok := value.(string); ok {
		return true
	}
	if _, ok := value.(ora.String); ok {
		return true
	}
	return false
}

func isBool(value interface{}) bool {
	if _, ok := value.(bool); ok {
		return true
	}
	if _, ok := value.(ora.Bool); ok {
		return true
	}
	return false
}

func isBytes(value interface{}) bool {
	if _, ok := value.([]byte); ok {
		return true
	}
	if _, ok := value.(ora.Raw); ok {
		return true
	}
	return false
}

func castInt(v interface{}, goColumnType ora.GoColumnType) interface{} {
	value := reflect.ValueOf(v)
	switch goColumnType {
	case ora.I64:
		return value.Int()
	case ora.I32:
		return int32(value.Int())
	case ora.I16:
		return int16(value.Int())
	case ora.I8:
		return int8(value.Int())
	case ora.U64:
		return uint64(value.Int())
	case ora.U32:
		return uint32(value.Int())
	case ora.U16:
		return uint16(value.Int())
	case ora.U8:
		return uint8(value.Int())
	case ora.F64:
		return float64(value.Int())
	case ora.F32:
		return float32(value.Int())
	case ora.OraI64:
		return ora.Int64{Value: value.Int()}
	case ora.OraI32:
		return ora.Int32{Value: int32(value.Int())}
	case ora.OraI16:
		return ora.Int16{Value: int16(value.Int())}
	case ora.OraI8:
		return ora.Int8{Value: int8(value.Int())}
	case ora.OraU64:
		return ora.Uint64{Value: uint64(value.Int())}
	case ora.OraU32:
		return ora.Uint32{Value: uint32(value.Int())}
	case ora.OraU16:
		return ora.Uint16{Value: uint16(value.Int())}
	case ora.OraU8:
		return ora.Uint8{Value: uint8(value.Int())}
	case ora.OraF64:
		return ora.Float64{Value: float64(value.Int())}
	case ora.OraF32:
		return ora.Float32{Value: float32(value.Int())}
	}
	return nil
}

func length(v interface{}) int {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Slice {
		return value.Len()
	}
	return 1
}

func elemAt(v interface{}, i int) interface{} {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Slice {
		return value.Index(i).Interface()
	}
	return nil
}

func compare(expected interface{}, actual interface{}, goColumnType ora.GoColumnType, t *testing.T) {
	switch goColumnType {
	case ora.I64:
		compare_int64(expected, actual, t)
	case ora.I32:
		compare_int32(expected, actual, t)
	case ora.I16:
		compare_int16(expected, actual, t)
	case ora.I8:
		compare_int8(expected, actual, t)
	case ora.U64:
		compare_uint64(expected, actual, t)
	case ora.U32:
		compare_uint32(expected, actual, t)
	case ora.U16:
		compare_uint16(expected, actual, t)
	case ora.U8:
		compare_uint8(expected, actual, t)
	case ora.F64:
		compare_float64(expected, actual, t)
	case ora.F32:
		compare_float32(expected, actual, t)
	case ora.OraI64:
		compare_OraInt64(expected, actual, t)
	case ora.OraI32:
		compare_OraInt32(expected, actual, t)
	case ora.OraI16:
		compare_OraInt16(expected, actual, t)
	case ora.OraI8:
		compare_OraInt8(expected, actual, t)
	case ora.OraU64:
		compare_OraUint64(expected, actual, t)
	case ora.OraU32:
		compare_OraUint32(expected, actual, t)
	case ora.OraU16:
		compare_OraUint16(expected, actual, t)
	case ora.OraU8:
		compare_OraUint8(expected, actual, t)
	case ora.OraF64:
		compare_OraFloat64(expected, actual, t)
	case ora.OraF32:
		compare_OraFloat32(expected, actual, t)
	case ora.T:
		compare_time(expected, actual, t)
	case ora.OraT:
		compare_OraTime(expected, actual, t)
	case ora.S:
		compare_string(expected, actual, t)
	case ora.OraS:
		compare_OraString(expected, actual, t)
	case ora.B:
		compare_bool(expected, actual, t)
	case ora.OraB:
		compare_OraBool(expected, actual, t)
	case ora.Bin:
		compare_bytes(expected, actual, t)
	case ora.OraBin:
		compare_Bytes(expected, actual, t)
	default:
		compare_nil(expected, actual, t)
	}
}

func compare_int64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(int64)
	if !eOk {
		ePtr, ePtrOk := expected.(*int64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to int64 or *int64. (%T, %v)", expected, expected)
		}
	}
	var a int64
	switch x := actual.(type) {
	case int64:
		a = x
	case *int64:
		a = *x
	case string:
		var err error
		if a, err = strconv.ParseInt(x, 10, 64); err != nil {
			t.Error(err)
		}
	case ora.OCINum:
		var err error
		if a, err = strconv.ParseInt(x.String(), 10, 64); err != nil {
			t.Error(err)
		}
	default:
		t.Fatalf("Unable to cast actual value to int64 or *int64. (%T, %v)", actual, actual)
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_int32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(int32)
	a, aOk := actual.(int32)
	if !eOk {
		ePtr, ePtrOk := expected.(*int32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to int32 or *int32. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int32 or *int32. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_int16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(int16)
	a, aOk := actual.(int16)
	if !eOk {
		ePtr, ePtrOk := expected.(*int16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to int16 or *int16. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int16 or *int16. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_int8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(int8)
	a, aOk := actual.(int8)
	if !eOk {
		ePtr, ePtrOk := expected.(*int8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to int8 or *int8. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int8 or *int8. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_uint64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(uint64)
	a, aOk := actual.(uint64)
	if !eOk {
		ePtr, ePtrOk := expected.(*uint64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to uint64 or *uint64. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint64 or *uint64. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_uint32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(uint32)
	a, aOk := actual.(uint32)
	if !eOk {
		ePtr, ePtrOk := expected.(*uint32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to uint32 or *uint32. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint32 or *uint32. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_uint16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(uint16)
	a, aOk := actual.(uint16)
	if !eOk {
		ePtr, ePtrOk := expected.(*uint16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to uint16 or *uint16. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint16 or *uint16. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_uint8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(uint8)
	a, aOk := actual.(uint8)
	if !eOk {
		ePtr, ePtrOk := expected.(*uint8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to uint8 or *uint8. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint8 or *uint8. (%T, %v)", actual, actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_float64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(float64)
	if !eOk {
		ePtr, ePtrOk := expected.(*float64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to float64 or *float64. (%T, %v)", expected, expected)
		}
	}
	var a float64
	switch x := actual.(type) {
	case float64:
		a = x
	case *float64:
		a = *x
	case float32:
		a = float64(x)
	case *float32:
		a = float64(*x)
	case string:
		var err error
		if a, err = strconv.ParseFloat(x, 64); err != nil {
			t.Error(err)
		}
	case ora.OCINum:
		var err error
		if a, err = strconv.ParseFloat(x.String(), 64); err != nil {
			t.Error(err)
		}
	default:
		t.Fatalf("Unable to cast actual value to float64 or *float64. (%T, %v)", actual, actual)
	}
	if !isFloat64Close(e, a, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_float32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(float32)
	if !eOk {
		ePtr, ePtrOk := expected.(*float32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to float32 or *float32. (%T, %v)", expected, expected)
		}
	}
	var a float32
	switch x := actual.(type) {
	case float32:
		a = x
	case *float32:
		a = *x
	case string:
		f, err := strconv.ParseFloat(x, 32)
		if err != nil {
			t.Error(err)
		}
		a = float32(f)
	case ora.OCINum:
		f, err := strconv.ParseFloat(x.String(), 32)
		if err != nil {
			t.Error(err)
		}
		a = float32(f)
	default:
		t.Fatalf("Unable to cast actual value to float64 or *float64. (%T, %v)", actual, actual)
	}
	if !isFloat32Close(e, a, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Int64)
	a, aOk := actual.(ora.Int64)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Int64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Int64 or *ora.Int64. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int64 or *ora.Int64. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Int32)
	a, aOk := actual.(ora.Int32)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Int32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Int32 or *ora.Int32. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int32 or *ora.Int32. (%T, %v)\n%s", actual, actual, getStack(1))
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Int16)
	a, aOk := actual.(ora.Int16)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Int16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Int16 or *ora.Int16. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int16 or *ora.Int16. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Int8)
	a, aOk := actual.(ora.Int8)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Int8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Int8 or *ora.Int8. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int8 or *ora.Int8. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Uint64)
	a, aOk := actual.(ora.Uint64)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Uint64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Uint64 or *ora.Uint64. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint64 or *ora.Uint64. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Uint32)
	a, aOk := actual.(ora.Uint32)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Uint32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Uint32 or *ora.Uint32. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint32 or *ora.Uint32. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Uint16)
	a, aOk := actual.(ora.Uint16)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Uint16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Uint16 or *ora.Uint16. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint16 or *ora.Uint16. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Uint8)
	a, aOk := actual.(ora.Uint8)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Uint8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Uint8 or *ora.Uint8. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint8 or *ora.Uint8. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraFloat64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Float64)
	a, aOk := actual.(ora.Float64)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Float64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Float64 or *ora.Float64. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Float64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Float64 or *ora.Float64. (%T, %v)", actual, actual)
		}
	}
	if e.IsNull != a.IsNull && !isFloat64Close(e.Value, a.Value, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraFloat32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Float32)
	a, aOk := actual.(ora.Float32)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Float32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Float32 or *ora.Float32. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Float32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Float32 or *ora.Float32. (%T, %v)", actual, actual)
		}
	}
	if e.IsNull != a.IsNull && !isFloat32Close(e.Value, a.Value, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_time(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(time.Time)
	a, aOk := actual.(time.Time)
	if !eOk {
		ePtr, ePtrOk := expected.(*time.Time)
		if ePtrOk {
			e = *ePtr
		} else {
			eOra, eOraOk := expected.(ora.Time)
			if eOraOk {
				e = eOra.Value
			} else {
				t.Fatalf("Unable to cast expected value to time.Time, *time.Time, ora.Time. (%T, %v)", expected, expected)
			}
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*time.Time)
		if aPtrOk {
			a = *aPtr
		} else {
			aOra, aOraOk := actual.(ora.Time)
			if aOraOk {
				a = aOra.Value
			} else {
				t.Fatalf("Unable to cast actual value to time.Time, *time.Time, ora.Time. (%T, %v)", actual, actual)
			}
		}
	}
	if !isTimeEqual(e, a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraTime(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Time)
	a, aOk := actual.(ora.Time)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Time)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Time or *ora.Time. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Time)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Time or *ora.Time. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_string(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(string)
	if !eOk {
		ePtr, ePtrOk := expected.(*string)
		if ePtrOk {
			e = *ePtr
		} else {
			eOra, eOraOk := expected.(ora.String)
			if eOraOk {
				e = eOra.Value
			} else {
				t.Fatalf("Unable to cast expected value to string, *string, ora.String. (%T, %v)", expected, expected)
			}
		}
	}
	var a string
	switch x := actual.(type) {
	case string:
		a = x
	case *string:
		a = *x
	case ora.String:
		a = x.Value
	case *ora.Lob:
		b, err := ioutil.ReadAll(x)
		if err != nil {
			t.Errorf("read %v: %v", x, err)
		}
		x.Close()
		a = string(b)
	default:
		t.Fatalf("Unable to cast actual value to string, *string, ora.String. (%T, %v)", actual, actual)
	}
	if e != a {
		t.Fatalf("expected(%q), actual(%q)\n%s", e, a, getStack(2))
	}
}

func compare_OraString(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.String)
	a, aOk := actual.(ora.String)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.String)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.String or *ora.String. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.String)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.String or *ora.String. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_bool(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(bool)
	a, aOk := actual.(bool)
	if !eOk {
		ePtr, ePtrOk := expected.(*bool)
		if ePtrOk {
			e = *ePtr
		} else {
			eOra, eOraOk := expected.(ora.Bool)
			if eOraOk {
				e = eOra.Value
			} else {
				t.Fatalf("Unable to cast expected value to bool, *bool, ora.Bool. (%T, %v)", expected, expected)
			}
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*bool)
		if aPtrOk {
			a = *aPtr
		} else {
			aOra, aOraOk := actual.(ora.Bool)
			if aOraOk {
				a = aOra.Value
			} else {
				t.Fatalf("Unable to cast actual value to bool, *bool, ora.Bool. (%T, %v): %s", actual, actual, getStack(2))
			}
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraBool(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Bool)
	a, aOk := actual.(ora.Bool)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.Bool)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.Bool or *ora.Bool. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Bool)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Bool or *ora.Bool. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_bytes(expected driver.Value, actual driver.Value, t *testing.T) {
	e, eOk := expected.([]byte)
	if !eOk {
		eOra, eOraOk := expected.(ora.Raw)
		if eOraOk {
			e = eOra.Value
		} else {
			t.Fatalf("Unable to cast expected value to []byte or ora.Raw. (%T, %v)", expected, expected)
		}
	}
	var a []byte
	switch x := actual.(type) {
	case []byte:
		a = x
	case ora.Raw:
		a = x.Value

	case ora.Lob:
		//t.Logf("Lob=%v", x)
		if x.Reader != nil {
			var err error
			a, err = ioutil.ReadAll(x.Reader)
			if err != nil {
				t.Errorf("error reading %v (%T): %v", x, x, err)
			}
		}
		x.Close()
	case io.ReadCloser:
		//t.Logf("ReadCloser=%v", x)
		var err error
		a, err = ioutil.ReadAll(x)
		x.Close()
		if err != nil {
			t.Errorf("error reading %v (%T): %v", x, x, err)
		}
	case io.WriterTo:
		//t.Logf("WriterTo=%v", x)
		var buf bytes.Buffer
		_, err := x.WriteTo(&buf)
		if c, ok := x.(io.Closer); ok {
			c.Close()
		}
		if err != nil {
			t.Errorf("error writing from %v (%T): %v", x, x, err)
		}
		a = buf.Bytes()
	default:
		t.Fatalf("Unable to cast actual value to []byte or ora.Raw. (%T, %v)\n%s", actual, actual, getStack(2))
	}
	if !areBytesEqual(e, a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_Bytes(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Raw)
	a, aOk := actual.(ora.Raw)
	if !eOk {
		t.Fatalf("Unable to cast expected value to ora.Raw. (%T, %v)", expected, expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to ora.Raw. (%T, %v)", actual, actual)
	} else if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraIntervalYM(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.IntervalYM)
	a, aOk := actual.(ora.IntervalYM)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.IntervalYM)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.IntervalYM or *ora.IntervalYM. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.IntervalYM)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.IntervalYM or *ora.IntervalYM. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraIntervalDS(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.IntervalDS)
	a, aOk := actual.(ora.IntervalDS)
	if !eOk {
		ePtr, ePtrOk := expected.(*ora.IntervalDS)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to ora.IntervalDS or *ora.IntervalDS. (%T, %v)", expected, expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.IntervalDS)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.IntervalDS or *ora.IntervalDS. (%T, %v)", actual, actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraBfile(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(ora.Bfile)
	a, aOk := actual.(ora.Bfile)
	if !eOk {
		t.Fatalf("Unable to cast expected value to ora.Bfile. (%T, %v)", expected, expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to ora.Bfile. (%T, %v)", actual, actual)
	} else if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_nil(expected interface{}, actual interface{}, t *testing.T) {
	if expected != nil {
		t.Fatalf("Expected value is not nil. (%T, %v)", expected, expected)
	}
	if actual != nil {
		t.Fatalf("Actual value is not nil. (%T, %v)", actual, actual)
	}
}

func isFloat32Close(x float32, y float32, t *testing.T) bool {
	if x == y {
		return true
	} else {
		xx, err := strconv.ParseFloat(fmt.Sprintf("%.5f", x), 32)
		if err != nil {
			t.Fatalf("Unable to parse float. (%v)", x)
		}
		yy, err := strconv.ParseFloat(fmt.Sprintf("%.5f", y), 32)
		if err != nil {
			t.Fatalf("Unable to parse float. (%v)", y)
		}
		return xx == yy
	}
}

func isFloat64Close(x float64, y float64, t *testing.T) bool {
	if x == y {
		return true
	} else {
		// use scale of 6 to support oracle binaryFloat margin of error
		xx, err := strconv.ParseFloat(fmt.Sprintf("%.6f", x), 64)
		if err != nil {
			t.Fatalf("Unable to parse float. (%v)", x)
		}
		yy, err := strconv.ParseFloat(fmt.Sprintf("%.6f", y), 64)
		if err != nil {
			t.Fatalf("Unable to parse float. (%v)", y)
		}
		return xx == yy
	}
}

func isTimeEqual(x time.Time, y time.Time) bool {
	_, eZoneOffset := x.Zone()
	_, aZoneOffset := y.Zone()
	return x.Year() == y.Year() &&
		x.Month() == y.Month() &&
		x.Day() == y.Day() &&
		x.Hour() == y.Hour() &&
		x.Minute() == y.Minute() &&
		x.Second() == y.Second() &&
		x.Nanosecond() == y.Nanosecond() &&
		eZoneOffset == aZoneOffset
}

func areBytesEqual(x []byte, y []byte) bool {
	return bytes.Equal(x, y)
}

func gen_int64() int64 {
	return int64(9)
}

func gen_int32() int32 {
	return int32(9)
}

func gen_int16() int16 {
	return int16(9)
}

func gen_int8() int8 {
	return int8(9)
}

func gen_uint64() uint64 {
	return uint64(9)
}

func gen_uint32() uint32 {
	return uint32(9)
}

func gen_uint16() uint16 {
	return uint16(9)
}

func gen_uint8() uint8 {
	return uint8(9)
}

func gen_float64() float64 {
	return float64(6.28318) //53071795) //86)
}

func gen_float64Trunc() float64 {
	return float64(6)
}

func gen_float32() float32 {
	return float32(6.28318)
}

func gen_float32Trunc() float32 {
	return float32(6)
}
func gen_NumString() ora.Num {
	return "6.28318" //53071795) //86)
}
func gen_NumStringTrunc() ora.Num {
	return "6" //53071795) //86)
}

func gen_OraInt64(isNull bool) ora.Int64 {
	return ora.Int64{Value: gen_int64(), IsNull: isNull}
}

func gen_OraInt32(isNull bool) ora.Int32 {
	return ora.Int32{Value: gen_int32(), IsNull: isNull}
}

func gen_OraInt16(isNull bool) ora.Int16 {
	return ora.Int16{Value: gen_int16(), IsNull: isNull}
}

func gen_OraInt8(isNull bool) ora.Int8 {
	return ora.Int8{Value: gen_int8(), IsNull: isNull}
}

func gen_OraUint64(isNull bool) ora.Uint64 {
	return ora.Uint64{Value: gen_uint64(), IsNull: isNull}
}

func gen_OraUint32(isNull bool) ora.Uint32 {
	return ora.Uint32{Value: gen_uint32(), IsNull: isNull}
}

func gen_OraUint16(isNull bool) ora.Uint16 {
	return ora.Uint16{Value: gen_uint16(), IsNull: isNull}
}

func gen_OraUint8(isNull bool) ora.Uint8 {
	return ora.Uint8{Value: gen_uint8(), IsNull: isNull}
}

func gen_OraFloat64Trunc(isNull bool) ora.Float64 {
	return ora.Float64{Value: gen_float64Trunc(), IsNull: isNull}
}

func gen_OraFloat32Trunc(isNull bool) ora.Float32 {
	return ora.Float32{Value: gen_float32Trunc(), IsNull: isNull}
}

func gen_int64Slice() []int64 {
	expected := make([]int64, 5)
	expected[0] = -9
	expected[1] = -1
	expected[2] = 0
	expected[3] = 1
	expected[4] = 9
	return expected
}

func gen_int32Slice() []int32 {
	expected := make([]int32, 5)
	expected[0] = -9
	expected[1] = -1
	expected[2] = 0
	expected[3] = 1
	expected[4] = 9
	return expected
}

func gen_int16Slice() []int16 {
	expected := make([]int16, 5)
	expected[0] = -9
	expected[1] = -1
	expected[2] = 0
	expected[3] = 1
	expected[4] = 9
	return expected
}

func gen_int8Slice() []int8 {
	expected := make([]int8, 5)
	expected[0] = -9
	expected[1] = -1
	expected[2] = 0
	expected[3] = 1
	expected[4] = 9
	return expected
}

func gen_uint64Slice() []uint64 {
	expected := make([]uint64, 5)
	expected[0] = 0
	expected[1] = 3
	expected[2] = 5
	expected[3] = 7
	expected[4] = 9
	return expected
}

func gen_uint32Slice() []uint32 {
	expected := make([]uint32, 5)
	expected[0] = 0
	expected[1] = 3
	expected[2] = 5
	expected[3] = 7
	expected[4] = 9
	return expected
}

func gen_uint16Slice() []uint16 {
	expected := make([]uint16, 5)
	expected[0] = 0
	expected[1] = 3
	expected[2] = 5
	expected[3] = 7
	expected[4] = 9
	return expected
}

func gen_uint8Slice() []uint8 {
	expected := make([]uint8, 5)
	expected[0] = 0
	expected[1] = 3
	expected[2] = 5
	expected[3] = 7
	expected[4] = 9
	return expected
}

func gen_float64TruncSlice() []float64 {
	expected := make([]float64, 5)
	expected[0] = -6
	expected[1] = -3
	expected[2] = 0
	expected[3] = 3
	expected[4] = 6
	return expected
}

func gen_float32TruncSlice() []float32 {
	expected := make([]float32, 5)
	expected[0] = -6
	expected[1] = -3
	expected[2] = 0
	expected[3] = 3
	expected[4] = 6
	return expected
}

func gen_NumStringTruncSlice() []ora.Num {
	expected := make([]ora.Num, 5)
	expected[0] = "-6"
	expected[1] = "-3"
	expected[2] = "0"
	expected[3] = "3"
	expected[4] = "6"
	return expected
}

func gen_OraInt64Slice(isNull bool) []ora.Int64 {
	expected := make([]ora.Int64, 5)
	expected[0] = ora.Int64{Value: -9}
	expected[1] = ora.Int64{Value: -1}
	expected[2] = ora.Int64{IsNull: isNull}
	expected[3] = ora.Int64{Value: 1}
	expected[4] = ora.Int64{Value: 9}
	return expected
}

func gen_OraInt32Slice(isNull bool) []ora.Int32 {
	expected := make([]ora.Int32, 5)
	expected[0] = ora.Int32{Value: -9}
	expected[1] = ora.Int32{Value: -1}
	expected[2] = ora.Int32{IsNull: isNull}
	expected[3] = ora.Int32{Value: 1}
	expected[4] = ora.Int32{Value: 9}
	return expected
}

func gen_OraInt16Slice(isNull bool) []ora.Int16 {
	expected := make([]ora.Int16, 5)
	expected[0] = ora.Int16{Value: -9}
	expected[1] = ora.Int16{Value: -1}
	expected[2] = ora.Int16{IsNull: isNull}
	expected[3] = ora.Int16{Value: 1}
	expected[4] = ora.Int16{Value: 9}
	return expected
}

func gen_OraInt8Slice(isNull bool) []ora.Int8 {
	expected := make([]ora.Int8, 5)
	expected[0] = ora.Int8{Value: -9}
	expected[1] = ora.Int8{Value: -1}
	expected[2] = ora.Int8{IsNull: isNull}
	expected[3] = ora.Int8{Value: 1}
	expected[4] = ora.Int8{Value: 9}
	return expected
}

func gen_OraUint64Slice(isNull bool) []ora.Uint64 {
	expected := make([]ora.Uint64, 5)
	expected[0] = ora.Uint64{Value: 0}
	expected[1] = ora.Uint64{Value: 3}
	expected[2] = ora.Uint64{IsNull: isNull}
	expected[3] = ora.Uint64{Value: 7}
	expected[4] = ora.Uint64{Value: 9}
	return expected
}

func gen_OraUint32Slice(isNull bool) []ora.Uint32 {
	expected := make([]ora.Uint32, 5)
	expected[0] = ora.Uint32{Value: 0}
	expected[1] = ora.Uint32{Value: 3}
	expected[2] = ora.Uint32{IsNull: isNull}
	expected[3] = ora.Uint32{Value: 7}
	expected[4] = ora.Uint32{Value: 9}
	return expected
}

func gen_OraUint16Slice(isNull bool) []ora.Uint16 {
	expected := make([]ora.Uint16, 5)
	expected[0] = ora.Uint16{Value: 0}
	expected[1] = ora.Uint16{Value: 3}
	expected[2] = ora.Uint16{IsNull: isNull}
	expected[3] = ora.Uint16{Value: 7}
	expected[4] = ora.Uint16{Value: 9}
	return expected
}

func gen_OraUint8Slice(isNull bool) []ora.Uint8 {
	expected := make([]ora.Uint8, 5)
	expected[0] = ora.Uint8{Value: 0}
	expected[1] = ora.Uint8{Value: 3}
	expected[2] = ora.Uint8{IsNull: isNull}
	expected[3] = ora.Uint8{Value: 7}
	expected[4] = ora.Uint8{Value: 9}
	return expected
}

func gen_OraFloat64TruncSlice(isNull bool) []ora.Float64 {
	expected := make([]ora.Float64, 5)
	expected[0] = ora.Float64{Value: -float64(6)}
	expected[1] = ora.Float64{Value: -float64(3)}
	expected[2] = ora.Float64{IsNull: isNull}
	expected[3] = ora.Float64{Value: float64(3)}
	expected[4] = ora.Float64{Value: float64(6)}
	return expected
}

func gen_OraFloat32TruncSlice(isNull bool) []ora.Float32 {
	expected := make([]ora.Float32, 5)
	expected[0] = ora.Float32{Value: -float32(6)}
	expected[1] = ora.Float32{Value: -float32(3)}
	expected[2] = ora.Float32{IsNull: isNull}
	expected[3] = ora.Float32{Value: float32(3)}
	expected[4] = ora.Float32{Value: float32(6)}
	return expected
}

func gen_date() time.Time {
	return time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)
}

func gen_OraDate(isNull bool) ora.Time {
	return ora.Time{Value: gen_date(), IsNull: isNull}
}

func gen_dateSlice() []time.Time {
	expected := make([]time.Time, 5)
	expected[0] = time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)
	expected[1] = time.Date(2001, 2, 3, 4, 5, 6, 0, testDbsessiontimezone)
	expected[2] = time.Date(2002, 3, 4, 5, 6, 7, 0, testDbsessiontimezone)
	expected[3] = time.Date(2003, 4, 5, 6, 7, 8, 0, testDbsessiontimezone)
	expected[4] = time.Date(2004, 5, 6, 7, 8, 9, 0, testDbsessiontimezone)
	return expected
}

func gen_OraDateSlice(isNull bool) []ora.Time {
	expected := make([]ora.Time, 5)
	expected[0] = ora.Time{Value: time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)}
	expected[1] = ora.Time{Value: time.Date(2001, 2, 3, 4, 5, 6, 0, testDbsessiontimezone)}
	expected[2] = ora.Time{Value: time.Date(2002, 3, 4, 5, 6, 7, 0, testDbsessiontimezone), IsNull: isNull}
	expected[3] = ora.Time{Value: time.Date(2003, 4, 5, 6, 7, 8, 0, testDbsessiontimezone)}
	expected[4] = ora.Time{Value: time.Date(2004, 5, 6, 7, 8, 9, 0, testDbsessiontimezone)}
	return expected
}

func gen_time() time.Time {
	return time.Date(2000, 1, 2, 3, 4, 5, 6, testDbsessiontimezone)
}

func gen_OraTime(isNull bool) ora.Time {
	return ora.Time{Value: gen_time(), IsNull: isNull}
}

func gen_timeSlice() []time.Time {
	expected := make([]time.Time, 5)
	expected[0] = time.Date(2000, 1, 2, 3, 4, 5, 6, testDbsessiontimezone)
	expected[1] = time.Date(2001, 2, 3, 4, 5, 6, 7, testDbsessiontimezone)
	expected[2] = time.Date(2002, 3, 4, 5, 6, 7, 8, testDbsessiontimezone)
	expected[3] = time.Date(2003, 4, 5, 6, 7, 8, 9, testDbsessiontimezone)
	expected[4] = time.Date(2004, 5, 6, 7, 8, 9, 10, testDbsessiontimezone)
	return expected
}

func gen_OraTimeSlice(isNull bool) []ora.Time {
	expected := make([]ora.Time, 5)
	expected[0] = ora.Time{Value: time.Date(2000, 1, 2, 3, 4, 5, 6, testDbsessiontimezone)}
	expected[1] = ora.Time{Value: time.Date(2001, 2, 3, 4, 5, 6, 7, testDbsessiontimezone)}
	expected[2] = ora.Time{Value: time.Date(2002, 3, 4, 5, 6, 7, 8, testDbsessiontimezone), IsNull: isNull}
	expected[3] = ora.Time{Value: time.Date(2003, 4, 5, 6, 7, 8, 9, testDbsessiontimezone)}
	expected[4] = ora.Time{Value: time.Date(2004, 5, 6, 7, 8, 9, 10, testDbsessiontimezone)}
	return expected
}

func gen_string() string {
	return "Sentence with no space at the end."
}
func gen_string48() string {
	return rpad48(gen_string())
}
func rpad48(s string) string {
	if ora.Cfg().RTrimChar {
		return s
	}
	return rpad(s, 48, " ")
}
func rpad(s string, length int, padding string) string {
	n := length - len(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(padding, n/len(padding)+1)[:n]
}

func gen_OraString(isNull bool) ora.String {
	return ora.String{Value: gen_string(), IsNull: isNull}
}
func gen_OraString48(isNull bool) ora.String {
	return ora.String{Value: gen_string48(), IsNull: isNull}
}

// important to test strings of non-equal length
func gen_stringSlice() interface{} {
	expected := make([]string, 5)
	expected[0] = "Go is expressive, concise, clean, and efficient."
	expected[1] = "Its concurrency mechanisms make it easy to"
	expected[2] = "Go compiles quickly to machine code yet has"
	expected[3] = "It's a fast, statically typed, compiled"
	expected[4] = "One of Go's key design goals is code"
	return expected
}
func gen_stringSlice48() interface{} {
	expected := gen_stringSlice().([]string)
	for i, s := range expected {
		expected[i] = rpad48(s)
	}
	return expected
}

func gen_OraStringSlice(isNull bool) interface{} {
	expected := make([]ora.String, 5)
	expected[0] = ora.String{Value: "Go is expressive, concise, clean, and efficient."}
	expected[1] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	expected[2] = ora.String{Value: "Go compiles quickly to machine code yet has", IsNull: isNull}
	expected[3] = ora.String{Value: "It's a fast, statically typed, compiled"}
	expected[4] = ora.String{Value: "One of Go's key design goals is code"}
	return expected
}

func gen_OraStringSlice48(isNull bool) interface{} {
	expected := gen_OraStringSlice(isNull).([]ora.String)
	for i, s := range expected {
		expected[i].Value = rpad48(s.Value)
	}
	return expected
}

func gen_boolFalse() bool {
	return false
}
func gen_boolTrue() bool {
	return true
}

func gen_OraBoolFalse(isNull bool) ora.Bool {
	return ora.Bool{Value: gen_boolFalse(), IsNull: isNull}
}

func gen_OraBoolTrue(isNull bool) ora.Bool {
	return ora.Bool{Value: gen_boolTrue(), IsNull: isNull}
}

func gen_boolSlice() interface{} {
	expected := make([]bool, 5)
	expected[0] = false
	expected[1] = true
	expected[2] = false
	expected[3] = false
	expected[4] = true
	return expected
}

func gen_OraBoolSlice(isNull bool) interface{} {
	expected := make([]ora.Bool, 5)
	expected[0] = ora.Bool{Value: true}
	expected[1] = ora.Bool{Value: false}
	expected[2] = ora.Bool{Value: false, IsNull: isNull}
	expected[3] = ora.Bool{Value: false}
	expected[4] = ora.Bool{Value: true}
	return expected
}

var (
	_gen_bytes    []byte
	_gen_bytes_mu sync.Mutex
)

func gen_bytes(length int) []byte {
	_gen_bytes_mu.Lock()
	defer _gen_bytes_mu.Unlock()
	if len(_gen_bytes) >= length {
		return _gen_bytes[:length:length]
	}
	values := make([]byte, length-len(_gen_bytes))
	rand.Read(values)
	_gen_bytes = append(_gen_bytes, values...)
	return _gen_bytes[:length:length]
}

func gen_OraBytes(length int, isNull bool) ora.Raw {
	return ora.Raw{Value: gen_bytes(length), IsNull: isNull}
}

func gen_OraBytesLob(length int, isNull bool) ora.Lob {
	if isNull {
		return ora.Lob{}
	}
	return ora.Lob{Reader: bytes.NewReader(gen_bytes(length))}
}

func gen_bytesSlice(length int) [][]byte {
	values := make([][]byte, 5)
	values[0] = gen_bytes(length)
	values[1] = gen_bytes(length)
	values[2] = gen_bytes(length)
	values[3] = gen_bytes(length)
	values[4] = gen_bytes(length)

	return values
}

func gen_OraBytesSlice(length int, isNull bool) []ora.Raw {
	values := make([]ora.Raw, 5)
	values[0] = ora.Raw{Value: gen_bytes(2000)}
	values[1] = ora.Raw{Value: gen_bytes(2000)}
	values[2] = ora.Raw{Value: gen_bytes(2000), IsNull: isNull}
	values[3] = ora.Raw{Value: gen_bytes(2000)}
	values[4] = ora.Raw{Value: gen_bytes(2000)}

	return values
}

func gen_OraIntervalYMSlice(isNull bool) []ora.IntervalYM {
	expected := make([]ora.IntervalYM, 5)
	expected[0] = ora.IntervalYM{Year: 1, Month: 1}
	expected[1] = ora.IntervalYM{Year: 99, Month: 9}
	expected[2] = ora.IntervalYM{IsNull: isNull}
	expected[3] = ora.IntervalYM{Year: -1, Month: -1}
	expected[4] = ora.IntervalYM{Year: -99, Month: -9}
	return expected
}

func gen_OraIntervalDSSlice(isNull bool) []ora.IntervalDS {
	expected := make([]ora.IntervalDS, 5)
	expected[0] = ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	expected[1] = ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}
	expected[2] = ora.IntervalDS{IsNull: isNull}
	expected[3] = ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}
	expected[4] = ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}
	return expected
}

func gen_OraBfile(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "TEMP_DIR", Filename: "test.txt"}
}

/*
func gen_OraBfileEmpty(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "", Filename: ""}
}

func gen_OraBfileEmptyDir(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "", Filename: "test.txt"}
}

func gen_OraBfileEmptyFilename(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "TEMP_DIR", Filename: ""}
}
*/

func getStack(stripHeadCalls int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	i := bytes.IndexByte(buf, '\n')
	if i < 0 {
		return string(buf)
	}
	var prefix string
	if bytes.Contains(buf[:i], []byte("goroutine")) {
		prefix, buf = string(buf[:i+1]), buf[i+1:]
	}
Loop:
	for stripHeadCalls > 0 {
		stripHeadCalls--
		for i := 0; i < 2; i++ {
			if j := bytes.IndexByte(buf, '\n'); j < 0 {
				break Loop
			} else {
				buf = buf[j+1:]
			}
		}
	}
	return prefix + string(buf)
}

func TestFils(t *testing.T) {
	// {'default': None, 'autoincrement': True, 'type': NUMBER(precision=5, scale=0, asdecimal=False), 'name': u'leg', 'nullable': False},
	// {'default': None, 'autoincrement': True, 'type': NUMBER(precision=6, scale=0, asdecimal=False), 'name': u'site', 'nullable': False}
	// {'default': None, 'autoincrement': True, 'type': VARCHAR(length=1), 'name': u'hole', 'nullable': False}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(precision=5, scale=0, asdecimal=False), 'name': u'core', 'nullable': False}
	// {'default': None, 'autoincrement': True, 'type': VARCHAR(length=1), 'name': u'core_type', 'nullable': False}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(precision=2, scale=0, asdecimal=False), 'name': u'section_number', 'nullable': False}
	// {'default': None, 'autoincrement': True, 'type': VARCHAR(length=2), 'name': u'section_type', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'top_cm', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'bot_cm', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'depth_mbsf', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'inor_c_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'caco3_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'tot_c_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'org_c_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'nit_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'name': u'sul_wt_pct', 'nullable': True}
	// {'default': None, 'autoincrement': True, 'type': NUMBER(asdecimal=False), 'nam e': u'h_wt_pct', 'nullable': True
	// https://gist.github.com/fils/ffb99e48bc3e994d54f1

	tableName := tableName()
	testDb.Exec(`DROP TABLE ` + tableName)
	if _, err := testDb.Exec(`CREATE TABLE ` + tableName + ` (
		leg NUMBER(5),
		site NUMBER(6),
		hole VARCHAR2(1),
		core NUMBER(5),
		core_type VARCHAR2(1),
		section_number NUMBER(2),
		section_Type VARCHAR2(2) NULL,
		top_cm NUMBER(6,3) NULL,
		bot_cm NUMBER(6,3) NULL,
		depth_mbsf NUMBER NULL,
		inor_c_wt_pct NUMBER NULL,
		caco3_wt_pct NUMBER NULL,
		tot_c_wt_pct NUMBER NULL,
		org_c_wt_pct NUMBER NULL,
		nit_wt_pct NUMBER NULL,
		sul_wt_pct NUMBER NULL,
		h_wt_pct NUMBER(6,3) NULL
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := testDb.Exec(`INSERT INTO ` + tableName + ` (
		leg, site, hole, core, core_type, section_number,
		section_type, top_cm, bot_cm, depth_mbsf,
		inor_c_wt_pct, caco3_wt_pct, tot_c_wt_pct,
		org_c_wt_pct, nit_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (207, 1259, 'C', 3, 'B', 4, '@', 5.2, NULL, 7.6, 8., 9., 10., 11., NULL , 13., 14.)`,
	); err != nil {
		t.Fatal(err)
	}

	//enableLogging(t)

	if _, err := testDb.Exec(`INSERT INTO ` + tableName + ` (
		leg, site, hole, core, core_type, section_number,
		section_type, top_cm, bot_cm, depth_mbsf,
		inor_c_wt_pct, caco3_wt_pct, tot_c_wt_pct,
		org_c_wt_pct, nit_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (171, 1049, 'B', 3, 'B', 4.2, '@', NULL, 6.12, 7.12, 8, 9.99, NULL, 11., NULL , 0.8, 0.42)`,
	); err != nil {
		t.Fatal(err)
	}

	qry := `SELECT
	   leg, site, hole, core, core_type
	 , section_number, section_type
	 , top_cm, bot_cm
	  , depth_mbsf
	 , inor_c_wt_pct
	 , caco3_wt_pct
	 , tot_c_wt_pct
	 , org_c_wt_pct
	 , nit_wt_pct
	 , sul_wt_pct
	 , h_wt_pct
	FROM
	   ` + tableName + `
	WHERE
	       leg = 171
	    AND site = 1049
	   AND hole = 'B'
	ORDER BY leg, site, hole, core, section_number, top_cm
`

	rows, err := testDb.Query(qry)
	if err != nil {
		t.Errorf(`Error with "%s": %s`, qry, err)
		return
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		i++
		var (
			Leg            int
			Site           int
			Hole           string
			Core           int
			Core_type      string
			Section_number int
			Section_type   string
			Top_cm         sql.NullFloat64
			Bot_cm         sql.NullFloat64
			Depth_mbsf     sql.NullFloat64
			Inor_c_wt_pct  sql.NullFloat64
			Caco3_wt_pct   sql.NullFloat64
			Tot_c_wt_pct   sql.NullFloat64
			Org_c_wt_pct   sql.NullFloat64
			Nit_wt_pct     sql.NullFloat64
			Sul_wt_pct     sql.NullFloat64
			H_wt_pct       sql.NullFloat64
		)

		if err := rows.Scan(&Leg, &Site, &Hole, &Core, &Core_type, &Section_number, &Section_type, &Top_cm, &Bot_cm, &Depth_mbsf, &Inor_c_wt_pct, &Caco3_wt_pct, &Tot_c_wt_pct, &Org_c_wt_pct, &Nit_wt_pct, &Sul_wt_pct, &H_wt_pct); err != nil {
			t.Fatalf("scan %d. record: %v", i, err)
		}

		//t.Logf("Results: %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v", Leg, Site, Hole, Core, Core_type, Section_number, Section_type, Top_cm, Bot_cm, Depth_mbsf, Inor_c_wt_pct, Caco3_wt_pct, Tot_c_wt_pct, Org_c_wt_pct, Nit_wt_pct, Sul_wt_pct, H_wt_pct)

	}
	if err := rows.Err(); err != nil {
		t.Error(err)
	}
}

func TestFilsIssue36(t *testing.T) {
	tableName := tableName()
	testDb.Exec(`DROP TABLE ` + tableName)
	testDb.Exec(`DROP VIEW ` + tableName + `_v`)

	checkErr := func(err error, qry string) {
		if err == nil {
			return
		}
		errS := err.Error()
		if qry != "" {
			err = errors.Wrap(err, qry)
		}
		if strings.Contains(errS, "ORA-01031:") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	qry := `CREATE TABLE ` + tableName + ` (
		leg NUMBER(5),
		site NUMBER(6),
		hole VARCHAR2(1),
		core NUMBER(5),
		core_type VARCHAR2(1),
		section_number NUMBER(2),
		section_Type VARCHAR2(2) NULL,
		top_cm NUMBER(6,3) NULL,
		bot_cm NUMBER(6,3) NULL,
		depth_mbsf NUMBER NULL,
		inor_c_wt_pct NUMBER NULL,
		caco3_wt_pct NUMBER NULL,
		tot_c_wt_pct NUMBER NULL,
		org_c_wt_pct NUMBER NULL,
		nit_wt_pct NUMBER NULL,
		sul_wt_pct NUMBER NULL,
		h_wt_pct NUMBER(6,3) NULL
	)`
	if _, err := testDb.Exec(qry); err != nil {
		checkErr(err, qry)
	}

	qry = `INSERT INTO ` + tableName + ` (
		leg, site, hole, core, core_type, section_number,
		section_type, top_cm, bot_cm, depth_mbsf,
		inor_c_wt_pct, caco3_wt_pct, tot_c_wt_pct,
		org_c_wt_pct, nit_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (207, 1259, 'C', 3, 'B', 4, '@', 5.2, NULL, 7.6, 8., 9., 10., 11., NULL , 13., 14.)`
	if _, err := testDb.Exec(qry); err != nil {
		checkErr(err, qry)
	}

	qry = `INSERT INTO ` + tableName + ` (
		leg, site, hole, core, core_type, section_number,
		section_type, top_cm, bot_cm, depth_mbsf,
		inor_c_wt_pct, caco3_wt_pct, tot_c_wt_pct,
		org_c_wt_pct, nit_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (171, 1049, 'B', 3, 'B', 4.2, '@', NULL, 6.12, 7.12, 8, 9.99, NULL, 11., NULL , 0.8, 0.42)`
	if _, err := testDb.Exec(qry); err != nil {
		checkErr(err, qry)
	}

	qry = `CREATE VIEW ` + tableName + `_v AS SELECT * FROM ` + tableName + ``
	if _, err := testDb.Exec(qry); err != nil {
		checkErr(err, qry)
	}

	ocdHole := tableName + "_h"
	ocdSection := tableName + "_s"
	ocdSample := tableName + "_sm"
	ocdChemCarbSample := tableName + "_ccs"
	ocdChemCarbAnalysis := tableName + "_cca"
	testDb.Exec(`DROP TABLE ` + ocdHole)
	testDb.Exec(`DROP TABLE ` + ocdSection)
	testDb.Exec(`DROP TABLE ` + ocdSample)
	testDb.Exec(`DROP TABLE ` + ocdChemCarbSample)
	testDb.Exec(`DROP TABLE ` + ocdChemCarbAnalysis)

	if _, err := testDb.Exec(`CREATE TABLE ` + ocdHole + ` (
 LEG    NUMBER(5) NOT NULL,
 SITE   NUMBER(6) NOT NULL,
 HOLE   VARCHAR2(1) NOT NULL
)`); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec(`DROP TABLE ` + ocdHole)

	if _, err := testDb.Exec(`CREATE TABLE ` + ocdSection + ` (
 SECTION_ID       NUMBER(7) NOT NULL,
 SECTION_NUMBER   NUMBER(2) NOT NULL,
 SECTION_TYPE     VARCHAR2(2),
 LEG              NUMBER(5) NOT NULL,
 SITE             NUMBER(6) NOT NULL,
 HOLE             VARCHAR2(1) NOT NULL,
 CORE             NUMBER(5) NOT NULL,
 CORE_TYPE        VARCHAR2(1) NOT NULL
)`); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec(`DROP TABLE ` + ocdSection)

	if _, err := testDb.Exec(`CREATE TABLE ` + ocdSample + ` (
 SAMPLE_ID               NUMBER(9) NOT NULL,
 LOCATION                VARCHAR2(3) NOT NULL,
 SAM_SECTION_ID          NUMBER(7),
 TOP_INTERVAL            NUMBER(6,3),
 BOTTOM_INTERVAL         NUMBER(6,3)
)`); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec(`DROP TABLE ` + ocdSample)

	if _, err := testDb.Exec(`CREATE TABLE ` + ocdChemCarbSample + ` (
 RUN_ID                  NUMBER(9) NOT NULL,
 SAMPLE_ID               NUMBER(9) NOT NULL,
 LOCATION                VARCHAR2(3) NOT NULL
)`); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec(`DROP TABLE ` + ocdChemCarbSample)

	if _, err := testDb.Exec(`CREATE TABLE ` + ocdChemCarbAnalysis + ` (
 RUN_ID                  NUMBER(9) NOT NULL,
 ANALYSIS_CODE           VARCHAR2(15) NOT NULL,
 METHOD_CODE             VARCHAR2(10) NOT NULL,
 ANALYSIS_RESULT         NUMBER(15,5)
)`); err != nil {
		t.Fatal(err)
	}
	defer testDb.Exec(`DROP TABLE ` + ocdChemCarbAnalysis)

	// create the views

	testDb.Exec(`DROP PUBLIC SYNONYM ` + ocdChemCarbSample)
	testDb.Exec(`DROP VIEW ` + ocdChemCarbSample + `_v`)

	for i, line := range []struct {
		id               int
		analysis, method string
		result           float64
	}{
		{42285, "CaCO3", "C", 1},
		{42285, "INOR_C", "C", 0.12},
		{42290, "CaCO3", "C", 0.45},
		{42290, "H", "CNS", 0.62},
		{42290, "HI", "RE", 187},
		{42290, "INOR_C", "C", 0.054},
		{42290, "NIT", "CNS", 0},
		{42290, "OI", "RE", 56},
		{42290, "ORG_C", "CNS", 0.6},
		{42290, "PC", "RE", 0.07},
		{42290, "PI", "RE", 0.1},
		{42290, "S1", "RE", 0.09},
		{42290, "S2", "RE", 0.77},
		{42290, "S3", "RE", 0.23},
		{42290, "SUL", "CNS", 0.06},
		{42290, "TMX", "RE", 460},
		{42290, "TOC", "RE", 0.41},
		{42290, "TOT_C", "CNS", 0.65},
		{42295, "CaCO3", "C", 0.45},
		{42295, "INOR_C", "C", 0.054},
		{3295, "CaCO3", "C", 73.91},
		{3295, "H", "CNS", 0.27},
		{3295, "INOR_C", "C", 8.87},
		{3295, "ORG_C", "CNS", 0.11},
		{3295, "TOT_C", "CNS", 8.98},
		{3300, "CaCO3", "C", 76.07},
		{3300, "H", "CNS", 0.24},
		{3300, "INOR_C", "C", 9.13},
		{3300, "ORG_C", "CNS", 0},
		{3300, "TOT_C", "CNS", 9.11},
		{3240, "CaCO3", "C", 70.97},
		{3240, "H", "CNS", 0.15},
		{3240, "INOR_C", "C", 8.52},
		{3240, "ORG_C", "CNS", 0},
		{3240, "TOT_C", "CNS", 8.24},
		{3245, "CaCO3", "C", 82.33},
		{3245, "H", "CNS", 0.09},
		{3245, "INOR_C", "C", 9.88},
		{3245, "ORG_C", "CNS", 0},
		{3245, "TOT_C", "CNS", 9.6},
		{3250, "CaCO3", "C", 29.66},
		{3250, "H", "CNS", 0.5},
		{3250, "INOR_C", "C", 3.56},
		{3250, "ORG_C", "CNS", 0.08},
		{3250, "PI", "RE", 0.25},
		{3250, "S1", "RE", 0.02},
		{3250, "S2", "RE", 0.1},
		{3250, "S3", "RE", 1.92},
		{3250, "TMX", "RE", 413},
		{3250, "TOC", "RE", 0},
		{3250, "TOT_C", "CNS", 3.64},
		{3250, "CaCO3", "C", 29.66},
		{3250, "H", "CNS", 0.5},
		{3250, "INOR_C", "C", 3.56},
		{3250, "ORG_C", "CNS", 0.08},
		{3250, "PI", "RE", 0.25},
		{3250, "S1", "RE", 0.02},
		{3250, "S2", "RE", 0.1},
		{3250, "S3", "RE", 1.92},
		{3250, "TMX", "RE", 413},
		{3250, "TOC", "RE", 0},
		{3250, "TOT_C", "CNS", 3.64},
		{3255, "CaCO3", "C", 63.12},
		{3255, "H", "CNS", 0.24},
		{3255, "HI", "RE", 50},
		{3255, "INOR_C", "C", 7.58},
		{3255, "OI", "RE", 1675},
		{3255, "ORG_C", "CNS", 0.16},
		{3255, "PI", "RE", 0},
		{3255, "S1", "RE", 0},
		{3255, "S2", "RE", 0.04},
		{3255, "S3", "RE", 1.34},
		{3255, "TMX", "RE", 410},
		{3255, "TOC", "RE", 0.08},
		{3255, "TOT_C", "CNS", 7.74},
		{3255, "CaCO3", "C", 63.12},
		{3255, "H", "CNS", 0.24},
		{3255, "HI", "RE", 50},
		{3255, "INOR_C", "C", 7.58},
		{3255, "OI", "RE", 1675},
		{3255, "ORG_C", "CNS", 0.16},
		{3255, "PI", "RE", 0},
		{3255, "S1", "RE", 0},
		{3255, "S2", "RE", 0.04},
		{3255, "S3", "RE", 1.34},
		{3255, "TMX", "RE", 410},
		{3255, "TOC", "RE", 0.08},
		{3255, "TOT_C", "CNS", 7.74},
		{3260, "CaCO3", "C", 55.44},
		{3260, "H", "CNS", 0.46},
		{3260, "HI", "RE", 543},
		{3260, "INOR_C", "C", 6.66},
		{3260, "NIT", "CNS", 0.014},
		{3260, "OI", "RE", 152},
		{3260, "ORG_C", "CNS", 1.68},
		{3260, "PI", "RE", 0.01},
		{3260, "S1", "RE", 0.08},
		{3260, "S2", "RE", 6.6},
		{3260, "S3", "RE", 1.86},
		{3260, "TMX", "RE", 403},
		{3260, "TOC", "RE", 1.22},
		{3260, "TOT_C", "CNS", 8.34},
		{3260, "CaCO3", "C", 55.44},
		{3260, "H", "CNS", 0.46},
		{3260, "HI", "RE", 543},
		{3260, "INOR_C", "C", 6.66},
		{3260, "NIT", "CNS", 0.014},
		{3260, "OI", "RE", 152},
		{3260, "ORG_C", "CNS", 1.68},
		{3260, "PI", "RE", 0.01},
		{3260, "S1", "RE", 0.08},
		{3260, "S2", "RE", 6.6},
		{3260, "S3", "RE", 1.86},
		{3260, "TMX", "RE", 403},
		{3260, "TOC", "RE", 1.22},
		{3260, "TOT_C", "CNS", 8.34},
		{3265, "CaCO3", "C", 51.53},
		{3265, "H", "CNS", 0.64},
		{3265, "INOR_C", "C", 6.19},
		{3265, "NIT", "CNS", 0.04},
		{3265, "ORG_C", "CNS", 2.97},
		{3265, "TOT_C", "CNS", 9.16},
		{3265, "CaCO3", "C", 51.53},
		{3265, "H", "CNS", 0.64},
		{3265, "INOR_C", "C", 6.19},
		{3265, "NIT", "CNS", 0.04},
		{3265, "ORG_C", "CNS", 2.97},
		{3265, "TOT_C", "CNS", 9.16},
		{3270, "CaCO3", "C", 74.48},
		{3270, "H", "CNS", 0.34},
		{3270, "HI", "RE", 605},
		{3270, "INOR_C", "C", 8.94},
		{3270, "NIT", "CNS", 0.012},
		{3270, "OI", "RE", 85},
		{3270, "ORG_C", "CNS", 1.69},
		{3270, "PI", "RE", 0.01},
		{3270, "S1", "RE", 0.1},
		{3270, "S2", "RE", 8.4},
		{3270, "S3", "RE", 1.19},
		{3270, "SUL", "CNS", 0.1},
		{3270, "TMX", "RE", 395},
		{3270, "TOC", "RE", 1.39},
		{3270, "TOT_C", "CNS", 10.63},
		{3270, "CaCO3", "C", 74.48},
		{3270, "H", "CNS", 0.34},
		{3270, "HI", "RE", 605},
		{3270, "INOR_C", "C", 8.94},
		{3270, "NIT", "CNS", 0.012},
		{3270, "OI", "RE", 85},
		{3270, "ORG_C", "CNS", 1.69},
		{3270, "PI", "RE", 0.01},
		{3270, "S1", "RE", 0.1},
		{3270, "S2", "RE", 8.4},
		{3270, "S3", "RE", 1.19},
		{3270, "SUL", "CNS", 0.1},
		{3270, "TMX", "RE", 395},
		{3270, "TOC", "RE", 1.39},
		{3270, "TOT_C", "CNS", 10.63},
		{3275, "CaCO3", "C", 41.65},
		{3275, "H", "CNS", 0.73},
		{3275, "HI", "RE", 485},
		{3275, "INOR_C", "C", 5},
		{3275, "NIT", "CNS", 0.038},
		{3275, "OI", "RE", 79},
		{3275, "ORG_C", "CNS", 3.99},
		{3275, "PI", "RE", 0.01},
		{3275, "S1", "RE", 0.18},
		{3275, "S2", "RE", 15.6},
		{3275, "S3", "RE", 2.55},
		{3275, "TMX", "RE", 406},
		{3275, "TOC", "RE", 3.2},
		{3275, "TOT_C", "CNS", 8.99},
		{3275, "CaCO3", "C", 41.65},
		{3275, "H", "CNS", 0.73},
		{3275, "HI", "RE", 485},
		{3275, "INOR_C", "C", 5},
		{3275, "NIT", "CNS", 0.038},
		{3275, "OI", "RE", 79},
		{3275, "ORG_C", "CNS", 3.99},
		{3275, "PI", "RE", 0.01},
		{3275, "S1", "RE", 0.18},
		{3275, "S2", "RE", 15.6},
		{3275, "S3", "RE", 2.55},
		{3275, "TMX", "RE", 406},
		{3275, "TOC", "RE", 3.2},
		{3275, "TOT_C", "CNS", 8.99},
		{3280, "CaCO3", "C", 49.76},
		{3280, "H", "CNS", 1.54},
		{3280, "HI", "RE", 699},
		{3280, "INOR_C", "C", 5.97},
		{3280, "NIT", "CNS", 0.16},
		{3280, "OI", "RE", 45},
		{3280, "ORG_C", "CNS", 11.45},
		{3280, "PI", "RE", 0.02},
		{3280, "S1", "RE", 1.35},
		{3280, "S2", "RE", 70.9},
		{3280, "S3", "RE", 4.63},
		{3280, "SUL", "CNS", 0.62},
		{3280, "TMX", "RE", 393},
		{3280, "TOC", "RE", 10.14},
		{3280, "TOT_C", "CNS", 17.42},
		{3280, "CaCO3", "C", 49.76},
		{3280, "H", "CNS", 1.54},
		{3280, "HI", "RE", 699},
		{3280, "INOR_C", "C", 5.97},
		{3280, "NIT", "CNS", 0.16},
		{3280, "OI", "RE", 45},
		{3280, "ORG_C", "CNS", 11.45},
		{3280, "PI", "RE", 0.02},
		{3280, "S1", "RE", 1.35},
		{3280, "S2", "RE", 70.9},
		{3280, "S3", "RE", 4.63},
		{3280, "SUL", "CNS", 0.62},
		{3280, "TMX", "RE", 393},
		{3280, "TOC", "RE", 10.14},
		{3280, "TOT_C", "CNS", 17.42},
		{3285, "CaCO3", "C", 51.2},
		{3285, "H", "CNS", 0.85},
		{3285, "INOR_C", "C", 6.15},
		{3285, "NIT", "CNS", 0.085},
		{3285, "ORG_C", "CNS", 5.39},
		{3285, "TOT_C", "CNS", 11.54},
		{3285, "CaCO3", "C", 51.2},
		{3285, "H", "CNS", 0.85},
		{3285, "INOR_C", "C", 6.15},
		{3285, "NIT", "CNS", 0.085},
		{3285, "ORG_C", "CNS", 5.39},
		{3285, "TOT_C", "CNS", 11.54},
		{3290, "CaCO3", "C", 88.42},
		{3290, "H", "CNS", 0.007},
		{3290, "INOR_C", "C", 10.61},
		{3290, "ORG_C", "CNS", 0},
		{3290, "TOT_C", "CNS", 10.43},
		{3290, "CaCO3", "C", 88.42},
		{3290, "H", "CNS", 0.007},
		{3290, "INOR_C", "C", 10.61},
		{3290, "ORG_C", "CNS", 0},
		{3290, "TOT_C", "CNS", 10.43},
		{8171, "HI", "RE", 451},
		{8171, "OI", "RE", 70},
		{8171, "PI", "RE", 0.01},
		{8171, "S1", "RE", 0.26},
		{8171, "S2", "RE", 20.6},
		{8171, "S3", "RE", 3.24},
		{8171, "TMX", "RE", 407},
		{8171, "TOC", "RE", 4.57},
		{8176, "OI", "RE", 1016},
		{8176, "PI", "RE", 0.04},
		{8176, "S1", "RE", 0.03},
		{8176, "S2", "RE", 0.8},
		{8176, "S3", "RE", 0.61},
		{8176, "TMX", "RE", 445},
		{8176, "TOC", "RE", 0.06},
	} {
		qry := `INSERT INTO ` + ocdChemCarbAnalysis + ` (RUN_ID,ANALYSIS_CODE,METHOD_CODE,ANALYSIS_RESULT) VALUES (:1, :2, :3, :4)`
		if _, err := testDb.Exec(qry, line.id, line.analysis, line.method, line.result); err != nil {
			t.Fatalf("%q, line %d: %v", qry, i, err)
		}
	}

	//////

	for i, line := range []struct {
		id, sample int
		location   string
	}{
		{42285, 114942, "SHI"},
		{42290, 114943, "SHI"},
		{42295, 114944, "SHI"},
		{3295, 25277, "SHI"},
		{3300, 25263, "SHI"},
		{3240, 25061, "SHI"},
		{3245, 25063, "SHI"},
		{3250, 25107, "SHI"},
		{3250, 25107, "SHI"},
		{3255, 25106, "SHI"},
		{3255, 25106, "SHI"},
		{3260, 25105, "SHI"},
		{3260, 25105, "SHI"},
		{3265, 25102, "SHI"},
		{3265, 25102, "SHI"},
		{3270, 25104, "SHI"},
		{3270, 25104, "SHI"},
		{3275, 25103, "SHI"},
		{3275, 25103, "SHI"},
		{3280, 25101, "SHI"},
		{3280, 25101, "SHI"},
		{3285, 25100, "SHI"},
		{3285, 25100, "SHI"},
		{3290, 25099, "SHI"},
		{3290, 25099, "SHI"},
		{8171, 227155, "SHI"},
		{8176, 227154, "SHI"},
	} {
		qry := `INSERT INTO ` + ocdChemCarbSample + ` (RUN_ID,SAMPLE_ID,LOCATION) VALUES (:1, :2, :3)`
		if _, err := testDb.Exec(qry, line.id, line.sample, line.location); err != nil {
			t.Fatalf("%q line %d: %v", qry, i, err)
		}
	}

	/////

	qry = `Insert into ` + ocdHole + ` (LEG,SITE,HOLE) values (171,1049,'B')`
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatalf("%q: %v", qry, err)
	}

	for i, line := range []struct {
		id          int
		location    string
		section     int
		top, bottom float64
	}{
		{25099, "SHI", 42830, 0.21, 0.22},
		{25100, "SHI", 42830, 0.19, 0.21},
		{25101, "SHI", 42830, 0.175, 0.19},
		{25102, "SHI", 42830, 0.125, 0.22},
		{25103, "SHI", 42830, 0.16, 0.175},
		{25104, "SHI", 42830, 0.135, 0.16},
		{25105, "SHI", 42830, 0.125, 0.135},
		{25106, "SHI", 42830, 0.085, 0.125},
		{25107, "SHI", 42830, 0.07, 0.085},
		{227154, "SHI", 42830, 0.205, 0.22},
		{227155, "SHI", 42830, 0.19, 0.205},
	} {
		qry := `INSERT INTO ` + ocdSample + ` (SAMPLE_ID,LOCATION,SAM_SECTION_ID,TOP_INTERVAL,BOTTOM_INTERVAL) VALUES (:1, :2, :3, :4, :5)`
		if _, err := testDb.Exec(qry,
			line.id, line.location, line.section, line.top, line.bottom); err != nil {
			t.Fatalf("%q line %d: %v", qry, i, err)
		}
	}

	/////

	for i, line := range []struct {
		id, number int
		typ        string
		leg, site  int
		hole       string
		core       int
		core_typ   string
	}{
		{42730, 1, "S", 171, 1049, "B", 8, "H"},
		{42730, 1, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42740, 3, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42740, 3, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42740, 3, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42745, 4, "S", 171, 1049, "B", 8, "H"},
		{42735, 2, "S", 171, 1049, "B", 8, "H"},
		{42735, 2, "S", 171, 1049, "B", 8, "H"},
		{42820, 1, "S", 171, 1049, "B", 11, "X"},
		{42820, 1, "S", 171, 1049, "B", 11, "X"},
		{42820, 1, "S", 171, 1049, "B", 11, "X"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
		{42830, 3, "C", 171, 1049, "B", 11, "X"},
	} {
		qry := `INSERT INTO ` + ocdSection + ` (SECTION_ID,SECTION_NUMBER,SECTION_TYPE,LEG,SITE,HOLE,CORE,CORE_TYPE) VALUES (:1, :2, :3, :4, :5, :6, :7, :8)`
		if _, err := testDb.Exec(qry,
			line.id, line.number, line.typ, line.leg, line.site, line.hole, line.core, line.core_typ); err != nil {
			t.Fatalf("%q line %d: %v", qry, i, err)
		}
	}

	if _, err := testDb.Exec(`CREATE VIEW ` + ocdChemCarbSample + `_v AS
SELECT
    x.leg, x.site, x.hole
  , x.core, x.core_type
  , x.section_number, x.section_type
  , s.top_interval*100.0 top_cm
  , s.bottom_interval*100.0 bot_cm
  , AVG(DECODE(cca.analysis_code,'INOR_C',cca.analysis_result)) INOR_C_wt_pct
  , AVG(DECODE(cca.analysis_code,'CaCO3', cca.analysis_result)) CaCO3_wt_pct
  , AVG(DECODE(cca.analysis_code,'TOT_C', cca.analysis_result)) TOT_C_wt_pct
  , AVG(DECODE(cca.analysis_code,'ORG_C', cca.analysis_result)) ORG_C_wt_pct
  , AVG(DECODE(cca.analysis_code,'NIT',   cca.analysis_result)) NIT_wt_pct
  , AVG(DECODE(cca.analysis_code,'SUL',   cca.analysis_result)) SUL_wt_pct
  , AVG(DECODE(cca.analysis_code,'H',     cca.analysis_result)) H_wt_pct
FROM
    ` + ocdHole + ` h, ` + ocdSection + ` x, ` + ocdSample + ` s
  , ` + ocdChemCarbSample + ` ccs, ` + ocdChemCarbAnalysis + ` cca
WHERE
        h.leg = x.leg
    AND h.site = x.site
    AND h.hole = x.hole
    AND x.section_id = s.sam_section_id
    AND s.sample_id = ccs.sample_id
    AND s.location = ccs.location
    AND ccs.run_id = cca.run_id
GROUP BY x.leg, x.site, x.hole, x.core, x.core_type, x.section_number, x.section_type, s.top_interval, s.bottom_interval
ORDER BY x.leg, x.site, x.hole, x.core, x.core_type, x.section_number, s.top_interval
`); err != nil {
		t.Fatal(err)
	}
	testDb.Exec(`DROP VIEW ` + ocdChemCarbSample + `_v`)

	qry3 := `SELECT
            x.leg, x.site, x.hole
          , x.core, x.core_type
          , x.section_number, x.section_type
          , s.top_interval*100.0 top_cm
          , s.bottom_interval*100.0 bot_cm
          , AVG(DECODE(cca.analysis_code,'INOR_C',cca.analysis_result)) INOR_C_wt_pct
          , AVG(DECODE(cca.analysis_code,'CaCO3', cca.analysis_result)) CaCO3_wt_pct
          , AVG(DECODE(cca.analysis_code,'TOT_C', cca.analysis_result)) TOT_C_wt_pct
          , AVG(DECODE(cca.analysis_code,'ORG_C', cca.analysis_result)) ORG_C_wt_pct
          , AVG(DECODE(cca.analysis_code,'NIT',   cca.analysis_result)) NIT_wt_pct
          , AVG(DECODE(cca.analysis_code,'SUL',   cca.analysis_result)) SUL_wt_pct
          , AVG(DECODE(cca.analysis_code,'H',     cca.analysis_result)) H_wt_pct
        FROM
            ` + ocdHole + ` h, ` + ocdSection + ` x, ` + ocdSample + ` s
          , ` + ocdChemCarbSample + ` ccs, ` + ocdChemCarbAnalysis + ` cca
        WHERE
                h.leg = x.leg
            AND h.site = x.site
            AND h.hole = x.hole
            AND x.section_id = s.sam_section_id
            AND s.sample_id = ccs.sample_id
            AND s.location = ccs.location
            AND ccs.run_id = cca.run_id
            AND x.leg = 171
            AND x.site = 1049
            AND x.hole = upper('B')
        GROUP BY x.leg, x.site, x.hole, x.core, x.core_type, x.section_number, x.section_type, s.top_interval, s.bottom_interval
        ORDER BY x.leg, x.site, x.hole, x.core, x.core_type, x.section_number, s.top_interval
`

	// copied from github.com/tgulacsi/go/orahlp
	t.Logf("Describe query 3\n")
	desc, err := ora.DescribeQuery(testDb, qry3)
	if err != nil {
		t.Errorf(`Error with : %s`, err)
	}
	for i, d := range desc {
		t.Logf("desc[%d]: %#v", i, d)
	}

	t.Logf("Run query 3\n")

	rows3, err := testDb.Query(qry3)
	if err != nil {
		t.Fatalf(`Error with "%s": %s`, qry3, err)
	}
	defer rows3.Close()

	iii := 0
	for rows3.Next() {
		iii++
		var (
			Leg            int
			Site           int
			Hole           string
			Core           int
			Core_type      string
			Section_number int
			Section_type   string
			Top_cm         sql.NullFloat64
			Bot_cm         sql.NullFloat64
			Inor_c_wt_pct  sql.NullFloat64
			Caco3_wt_pct   sql.NullFloat64
			Tot_c_wt_pct   sql.NullFloat64
			Org_c_wt_pct   sql.NullFloat64
			Nit_wt_pct     sql.NullFloat64
			Sul_wt_pct     sql.NullFloat64
			H_wt_pct       sql.NullFloat64
		)

		if err := rows3.Scan(&Leg, &Site, &Hole, &Core, &Core_type, &Section_number, &Section_type, &Top_cm, &Bot_cm, &Inor_c_wt_pct, &Caco3_wt_pct, &Tot_c_wt_pct, &Org_c_wt_pct, &Nit_wt_pct, &Sul_wt_pct, &H_wt_pct); err != nil {
			t.Fatalf("scan %d. record: %v", iii, err)
		}

		//t.Logf("Results: %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v", Leg, Site, Hole, Core, Core_type, Section_number, Section_type, Top_cm, Bot_cm, Inor_c_wt_pct, Caco3_wt_pct, Tot_c_wt_pct, Org_c_wt_pct, Nit_wt_pct, Sul_wt_pct, H_wt_pct)

	}
	if err := rows3.Err(); err != nil {
		t.Error(err)
	}
}

func TestUnderflow(t *testing.T) {
	tbl := tableName()
	testDb.Exec(`DROP VIEW ` + tbl + `_view`)
	testDb.Exec(`DROP TABLE ` + tbl)
	qry := `CREATE TABLE ` + tbl + ` (
		num NUMBER NULL,
		num_6 NUMBER(6) NULL,
		num_6_3 NUMBER(6,3) NULL,
		num_6_n2 NUMBER(6, -2) NULL,
		flo FLOAT NULL,
		flo_6 FLOAT(6) NULL,
		bflo BINARY_FLOAT NULL,
		bdouble BINARY_DOUBLE NULL,
		int INTEGER NULL
	)`
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatalf("%q: %v", qry, err)
	}

	const colCount = 9

	queries := []string{
		`SELECT * FROM ` + tbl,
		`SELECT * FROM (SELECT * FROM ` + tbl + `)`,
	}
	if _, err := testDb.Exec(`CREATE VIEW ` + tbl + `_view1 AS SELECT * FROM ` + tbl); err != nil {
		t.Logf("cannot create view: %v", err)
	} else {
		queries = append(queries, `SELECT * FROM `+tbl+`_view1`)
		if _, err := testDb.Exec(`CREATE VIEW ` + tbl + `_view2 AS
			SELECT
				TO_NUMBER(num) num,
				TO_NUMBER(num_6) num_6,
				TO_NUMBER(num_6_3) num_6_3,
				TO_NUMBER(num_6_n2) num_6_n2,
				TO_NUMBER(flo) flo,
				TO_NUMBER(flo_6) flo_6,
				TO_NUMBER(bflo) bflo,
				TO_NUMBER(bdouble) bdouble,
				TO_NUMBER(int) int
			FROM ` + tbl); err != nil {
			t.Fatal(err)
		}
		queries = append(queries, `SELECT * FROM `+tbl+`_view2`)
	}
	want := make([]interface{}, colCount)
	got := make([]sql.NullFloat64, colCount)
	gotP := make([]interface{}, len(got))
	for i := range gotP {
		gotP[i] = &got[i]
	}
	ins := `INSERT INTO ` + tbl + ` VALUES (` + strings.Repeat(",%f", colCount)[1:] + `)`

	//enableLogging(t)

	for caseNum, test := range [][colCount]float64{
		{0.99, 8, 4.2, 65400., 0.7, 0.6, 3.14, 2.78, 42},
		{0.89, 8, 4.2, 65400., 0.8, 0.5, 3.14, 2.78, 42},
		{0.79, 8, 4.2, 65400., 0.9, 0.4, 3.14, 2.78, 42},
		{0.69, 8, 4.2, 65400., 0.6, 0.3, 3.14, 2.78, 42},
		{0.59, 8, 4.2, 65400., 0.5, 0.2, 3.14, 2.78, 42},
		{0.49, 8, 4.2, 65400., 0.4, 0.1, 3.14, 2.78, 42},
		{0.39, 8, 4.2, 65400., 0.3, 0.01, 3.14, 2.78, 42},
		{0.29, 8, 4.2, 65400., 0.2, 0.71, 3.14, 2.78, 42},
		{0.19, 8, 4.2, 65400., 0.1, 0.81, 3.14, 2.78, 42},
		{0.09, 8, 4.2, 65400., 0.0, 0.91, 3.14, 2.78, 42},
	} {
		testDb.Exec("TRUNCATE TABLE " + tbl)
		for i, f := range test[:] {
			want[i] = f
		}
		if _, err := testDb.Exec(fmt.Sprintf(ins, want...)); err != nil {
			t.Fatalf("%d. %v", caseNum+1, err)
		}

		for _, qry = range queries {
			rows, err := testDb.Query(qry)
			if err != nil {
				t.Errorf(`%d. Error with %q: %s`, caseNum+1, qry, err)
				return
			}
			defer rows.Close()

			i := 0
			for rows.Next() {
				i++

				if err := rows.Scan(gotP...); err != nil {
					t.Fatalf("%d. %q scan %d. record: %v", caseNum+1, qry, i, err)
				}

				t.Logf("Results: %v", got)

				for j, f := range got {
					if !f.Valid || f.Float64 != test[j] && math.Abs(f.Float64-test[j]) > 0.000001 {
						t.Errorf("%d. %q %d. got %v, awaited %v.", caseNum+1, qry, j+1, f, test[j])
					}
				}
			}
			if err := rows.Err(); err != nil {
				t.Errorf("%d. %q: %v", caseNum+1, qry, err)
			}
			rows.Close()
		}
	}
}

// TestIntFloat: see https://github.com/rana/ora/issues/57#issuecomment-185473949
func TestIntFloat(t *testing.T) {
	tbl := tableName()
	testDb.Exec(`DROP TABLE ` + tbl)
	qry := `CREATE TABLE ` + tbl + ` (
			  NUMBER_SIMPLE  NUMBER,
			  NUMBER_SCALE0  NUMBER(*,0)
			)`
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatal(err)
	}
	qry = "INSERT INTO " + tbl + " (NUMBER_SIMPLE, NUMBER_SCALE0) VALUES (:1, :2)"
	stmt, err := testDb.Prepare(qry)
	if err != nil {
		t.Fatal(err)
	}
	for _, numbers := range [][2]string{
		{"1", "2"},
		{"10", "20"},
		{"100", "200"},
		{"1000", "2000"},
		{"10000", "20000"},
		{"100000", "200000"},
		{"1000000", "2000000"},
		//{"1.5", "2.5"},
	} {
		if _, err := stmt.Exec(numbers[0], numbers[1]); err != nil {
			t.Fatalf("INSERT %#v: %v", numbers, err)
		}
	}
	ora.SetCfg(ora.Cfg().SetFloat(ora.N))
	rows, err := testDb.Query("SELECT * FROM " + tbl)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var ni, ni0 int64
		err := rows.Scan(&ni, &ni0)
		if err != nil {
			if !strings.Contains(err.Error(), "e+") {
				t.Errorf("scan: %v", err)
				continue
			}
			var fi, fi0 float64
			if err := rows.Scan(&fi, &fi0); err == nil {
				//t.Logf("float64 (%v, %v)", fi, fi0)
				continue
			}
			var si, si0 string
			if err := rows.Scan(&si, &si0); err != nil {
				t.Errorf("scan: %v", err)
				continue
			}
			//t.Logf("string (%v, %v)", si, si0)
			continue
		}
		//t.Logf("int64 (%v, %v)", ni, ni0)
	}
	err = rows.Err()
	if err != nil {
		t.Error(err)
	}
}

func TestSetDrvCfg(t *testing.T) {
	qry := "SELECT CAST('1' AS CHAR(1)) FROM DUAL"

	//enableLogging(t)
	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)

	ora.SetCfg(cfg.SetChar1(ora.B))
	if got := ora.Cfg().Char1(); got != ora.B {
		t.Fatalf("SetChar1: got %v, wanted %v", got, ora.B)
	}
	var b bool
	if err := testDb.QueryRow(qry).Scan(&b); err != nil {
		t.Fatalf("%s: %v", qry, err)
	}

	ora.SetCfg(cfg.SetChar1(ora.S))
	if got := ora.Cfg().Char1(); got != ora.S {
		t.Fatalf("SetChar1: got %v, wanted %v", got, ora.S)
	}
	var s string
	if err := testDb.QueryRow(qry).Scan(&s); err != nil {
		t.Fatalf("%s: %v", qry, err)
	}
	t.Logf("1=%v", s)
	if s != "1" {
		t.Errorf("got %q, awaited '1'", s)
	}
}

func TestStringSpaces(t *testing.T) {
	tbl := tableName()
	testDb.Exec("DROP TABLE " + tbl)
	qry := "CREATE TABLE " + tbl + " (text VARCHAR2(1024) NOT NULL)"
	if _, err := testDb.Exec(qry); err != nil {
		t.Fatalf("%s: %v", qry, err)
	}
	insQry := "INSERT INTO " + tbl + " (text) VALUES (:1)"
	texts := []string{"nospace", "onespace ", "twospaces  ", "   "}
	//enableLogging(t)
	for i, text := range texts {
		if _, err := testDb.Exec(insQry, text); err != nil {
			t.Fatalf("%d. insert (%q): %v", i, text, err)
		}
		var got, dump string
		if err := testDb.QueryRow("SELECT text, dump(text) FROM "+tbl+" WHERE text LIKE :1", text[0:]).Scan(&got, &dump); err != nil {
			t.Errorf("%d. select %q: %v", i, text, err)
			continue
		}
		if got != text {
			t.Errorf("%d. got %q (%s), awaited %q.", i, got, dump, text)
		}
	}
}

func TestPLSErr(t *testing.T) {
	//enableLogging(t)
	testSes := getSes(t)
	defer testSes.Close()

	qry := `DECLARE v_db PLS_INTEGER; BEGIN
	  SELECT 1 INTO v_db FROM DUAL WHERE 1 = 0;
	END;`
	var err error
	if _, err = testSes.PrepAndExe(qry); err == nil {
		t.Error("awaited error, got nothing!")
	} else {
		t.Log(err)
	}
	if _, err = testDb.Exec(qry); err == nil {
		t.Error("awaited error, got nothing!")
	} else {
		t.Log(err)
	}
}

func TestGetDriverName(t *testing.T) {
	qry := "SELECT sid, program, module, action, client_info FROM V$SESSION"
	rows, err := testDb.Query(qry)
	if err != nil {
		t.Skipf("%q: %v", qry, err)
	}
	for rows.Next() {
		var sid int64
		var program, module, action, clientInfo string
		if err := rows.Scan(&sid, &program, &module, &action, &clientInfo); err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(program, "ora.v4.test") {
			t.Logf("%d: %s/%s/%s/%s", sid, program, module, action, clientInfo)
		}
	}
}
func TestFloat64Prec(t *testing.T) {
	testSes := getSes(t)
	defer testSes.Close()

	var v0 float64 = 123456789.0123456789
	var v1 float64
	t.Logf("v0 = %.12f = %g", v0, v0)
	_, err := testSes.PrepAndExe("begin :1 := 123456789.0123456789; end;", &v1)
	if err != nil {
		t.Fatalf("1 - %v", err)
	}

	t.Logf("v1 = %.12f = %g", v1, v1)
	t.Logf("v0 internal: %064s\n", strconv.FormatUint((*(*uint64)(unsafe.Pointer(&v0))), 2))
	t.Logf("v1 internal: %064s\n", strconv.FormatUint((*(*uint64)(unsafe.Pointer(&v1))), 2))

	var v3 ora.OraNum
	//enableLogging(t)
	_, err = testSes.PrepAndExe("begin :1 := 123456789.0123456789; end;", &v3)
	if err != nil {
		t.Fatalf("2 - %v", err)
	}
	t.Logf("v3 = %s = %#v", v3, v3)

	var v4 ora.OCINum
	_, err = testSes.PrepAndExe("begin :1 := 123456789.0123456789; end;", &v4)
	if err != nil {
		t.Fatalf("3 - %v", err)
	}
	t.Logf("v4 = %s = %#v", v4, v4)

	var v5 ora.Num
	qry := "begin SELECT 1 INTO :2 FROM DUAL; end;"
	_, err = testSes.PrepAndExe(qry, &v5)
	if err != nil {
		t.Errorf("4 - %q: %v", qry, err)
		return
	}
	t.Logf("v5 = %s = %#v", v5, v5)

	qry = "SELECT 0.123, 0.12300000000000001 FROM DUAL"
	enableLogging(t)
	rset, err := testSes.PrepAndQry(qry)
	if err != nil {
		t.Errorf("6 - %q: %v", qry, err)
	}
	for rset.Next() {
		v6 := rset.Row[0]
		t.Logf("v6.1 = %v = %#v", v6, v6)
		if v6 != float64(0.123) {
			t.Logf("got %#v, wanted 0.123", v6)
		}
		v6 = rset.Row[1]
		t.Logf("v6.2  = %v = %#v", v6, v6)
		if v6 != float64(0.12300000000000001) {
			t.Logf("got %#v, wanted 0.12300000000000001", v6)
		}
	}
	if err = rset.Err(); err != nil {
		t.Error(err)
	}

	v7 := ora.Num("0.1")
	var v8 ora.Num
	_, err = testSes.PrepAndExe("begin SELECT :1 into :2 FROM DUAL; end;", v7, &v8)
	if err != nil {
		t.Logf("7 - %v", err)
		return
	}
	t.Logf("v7 = %s", v8)
	if v8 != v7 {
		t.Errorf("8 - got %q, wanted %q from %#v.", v8, v7, v8)
	}

	oCfg := testSes.Cfg()
	defer testSes.SetCfg(oCfg)
	testSes.SetCfg(oCfg.
		SetNumberInt(ora.OraN).
		SetNumberBigInt(ora.OraN).
		SetNumberFloat(ora.OraN).
		SetNumberBigFloat(ora.OraN).
		SetBinaryDouble(ora.OraN).
		SetBinaryFloat(ora.OraN).
		SetFloat(ora.OraN))
	v := "-23452342342423423423423.12345678901234567"
	rset, err = testSes.PrepAndQry(fmt.Sprintf("select dump(%s), %s from dual", v, v))
	if err != nil {
		t.Fatal(err)
	}
	for rset.Next() {
		t.Log("Original Value - ", v)
		t.Log("Value from     - ", rset.Row[1])
		t.Log("Column type    - ", reflect.TypeOf(rset.Row[1]))
		t.Log("Dump from DB   - ", rset.Row[0])
		t.Log("Dump from ora.OraOCINum -    ", []byte(rset.Row[1].(ora.OraOCINum).Value))
	}
	if err = rset.Err(); err != nil {
		t.Error(err)
	}

}
