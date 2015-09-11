// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/rana/ora.v3"
	"gopkg.in/rana/ora.v3/tstlg"
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
	date               oracleColumnType = "date not null"
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

var testSrvCfg *ora.SrvCfg
var testSesCfg *ora.SesCfg
var testUsername string
var testPassword string
var testConStr string
var testDbsessiontimezone *time.Location
var testTableId int
var testWorkloadColumnCount int
var testEnv *ora.Env
var testSrv *ora.Srv
var testSes *ora.Ses
var testDb *sql.DB

func init() {
	testSrvCfg = ora.NewSrvCfg()
	testSrvCfg.Dblink = os.Getenv("GO_ORA_DRV_TEST_DB")
	testSesCfg = ora.NewSesCfg()
	testSesCfg.Username = os.Getenv("GO_ORA_DRV_TEST_USERNAME")
	testSesCfg.Password = os.Getenv("GO_ORA_DRV_TEST_PASSWORD")
	testConStr = fmt.Sprintf("%v/%v@%v", testSesCfg.Username, testSesCfg.Password, testSrvCfg.Dblink)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_DB = '%v'\n", testSrvCfg.Dblink)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_USERNAME = '%v'\n", testSesCfg.Username)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_PASSWORD = '%v'\n", testSesCfg.Password)

	testWorkloadColumnCount = 20
	var err error

	// setup test environment, server and session
	testEnv, err := ora.OpenEnv(nil)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	testSrv, err = testEnv.OpenSrv(testSrvCfg)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	testSes, err = testSrv.OpenSes(testSesCfg)
	if err != nil {
		fmt.Println("initError: ", err)
	}

	// load session time zone
	testDbsessiontimezone, err = loadDbtimezone()
	if err != nil {
		fmt.Println("Error loading session time zone from database: ", err)
	} else {
		fmt.Println("Read session time zone from database...")
	}

	// drop all tables from previous test run
	fmt.Println("Dropping previous tables...")
	stmt, err := testSes.Prep(`
BEGIN
	FOR c IN (SELECT table_name FROM user_tables) LOOP
		EXECUTE IMMEDIATE ('DROP TABLE ' || c.table_name || ' CASCADE CONSTRAINTS');
	END LOOP;
END;`)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	defer stmt.Close()
	_, err = stmt.Exe()
	if err != nil {
		fmt.Println("initError: ", err)
	}
	fmt.Println("Tables dropped.")

	// setup test db
	ora.Register(nil)
	testDb, err = sql.Open(ora.Name, testConStr)
	if err != nil {
		fmt.Println("initError: ", err)
	}
}

var enableLoggingMu sync.Mutex

func enableLogging(t *testing.T) {
	enableLoggingMu.Lock()
	defer enableLoggingMu.Unlock()
	if t != nil {
		ora.Cfg().Log.Logger = tstlg.New(t)
		return
	}
}

func testIterations() int {
	if testing.Short() {
		return 1
	} else {
		return 1
	}
}

func testBindDefine(expected interface{}, oct oracleColumnType, t *testing.T, c *ora.StmtCfg, goColumnTypes ...ora.GoColumnType) {
	var gct ora.GoColumnType
	if len(goColumnTypes) > 0 {
		gct = goColumnTypes[0]
	} else {
		gct = goColumnTypeFromValue(expected)
	}
	//t.Logf("testBindDefine gct (%v, %v)", gct, ora.GctName(gct))

	tableName, err := createTable(1, oct, testSes)
	testErr(err, t)
	//defer dropTable(tableName, testSes, t)

	// insert
	insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
	if c != nil {
		insertStmt.SetCfg(c)
	}
	defer insertStmt.Close()
	testErr(err, t)
	rowsAffected, err := insertStmt.Exe(expected)
	testErr(err, t)
	expLen := length(expected)
	if gct == ora.Bin || gct == ora.OraBin {
		expLen = 1
	}
	if expLen != int(rowsAffected) {
		t.Fatalf("insert rows affected: expected(%v), actual(%v)", expLen, rowsAffected)
	}

	// select
	selectStmt, err := testSes.Prep(fmt.Sprintf("select c1 from %v", tableName), gct)
	defer selectStmt.Close()
	testErr(err, t)
	rset, err := selectStmt.Qry()
	testErr(err, t)
	// validate
	validate(expected, rset, t)
}

func testBindDefineDB(expected interface{}, t *testing.T, oct oracleColumnType) {
	for n := 0; n < testIterations(); n++ {
		tableName := createTableDB(testDb, t, oct)
		defer dropTableDB(testDb, t, tableName)

		// insert
		stmt, err := testDb.Prepare(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
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
	}
}

func testBindPtr(expected interface{}, oct oracleColumnType, t *testing.T) {
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
		}

		// insert
		stmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (:1) returning c1 into :2", tableName))
		defer stmt.Close()
		testErr(err, t)
		rowsAffected, err := stmt.Exe(expected, actual)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// validate
		compare2(expected, actual, t)
	}
}

func testMultiDefine(expected interface{}, oct oracleColumnType, t *testing.T) {
	for n := 0; n < testIterations(); n++ {
		tableName, err := createTable(1, oct, testSes)
		testErr(err, t)
		defer dropTable(tableName, testSes, t)

		// insert
		insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
		defer insertStmt.Close()
		testErr(err, t)
		rowsAffected, err := insertStmt.Exe(expected)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// select
		var selectStmt *ora.Stmt
		var rset *ora.Rset
		if isNumeric(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1 from %v", tableName), ora.I64, ora.I32, ora.I16, ora.I8, ora.U64, ora.U32, ora.U16, ora.U8, ora.F64, ora.F32, ora.OraI64, ora.OraI32, ora.OraI16, ora.OraI8, ora.OraU64, ora.OraU32, ora.OraU16, ora.OraU8, ora.OraF64, ora.OraF32)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isTime(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), ora.T, ora.OraT)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isString(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName), ora.S)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isBool(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), ora.B, ora.OraB)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isBytes(expected) {
			// one LOB cannot be opened twice in the same transaction (c1, c1 not works here)
			col := ora.Bin
			if n%2 == 1 {
				col = ora.OraBin
			}
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1 from %v", tableName), col)
			defer selectStmt.Close()
			testErr(err, t)
		}
		rset, err = selectStmt.Qry()
		testErr(err, t)

		// validate
		hasRow := rset.Next()
		testErr(rset.Err, t)
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
					value, ok := rset.Row[n].(ora.Time)
					if ok {
						compare_time(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value. (%v, %v)", reflect.TypeOf(rset.Row[n]).Name(), rset.Row[n])
					}
				case ora.S:
					compare_string(expected, rset.Row[n], t)
				case ora.OraS:
					value, ok := rset.Row[n].(ora.String)
					if ok {
						compare_string(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value. (%v, %v)", reflect.TypeOf(rset.Row[n]).Name(), rset.Row[n])
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
				case date, dateNull:
					expected[c] = gen_date()
					gcts[c] = ora.T
				case timestampP9, timestampP9Null, timestampTzP9, timestampTzP9Null, timestampLtzP9, timestampLtzP9Null:
					expected[c] = gen_time()
					gcts[c] = ora.T
				case charB48, charB48Null, charC48, charC48Null, nchar48, nchar48Null, varcharB48, varcharB48Null, varcharC48, varcharC48Null, varchar2B48, varchar2B48Null, varchar2C48, varchar2C48Null, nvarchar248, nvarchar248Null, long, longNull, clob, clobNull, nclob, nclobNull:
					expected[c] = gen_string()
					gcts[c] = ora.S
				case charB1, charB1Null, charC1, charC1Null:
					expected[c] = gen_boolTrue()
					gcts[c] = ora.B
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
			testErr(rset.Err, t)
			fetchStmt.Close()

			// Reduce the multiple by half
			currentMultiple = currentMultiple / 2
		}
	}
}

func loadDbtimezone() (*time.Location, error) {
	stmt, err := testSes.Prep("select tz_offset(sessiontimezone) from dual")
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	rset, err := stmt.Qry()
	if err != nil {
		return nil, err
	}
	hasRow := rset.Next()
	if !hasRow {
		return nil, errors.New("no time zone returned from database")
	}
	if value, ok := rset.Row[0].(string); ok {
		value = strings.Trim(value, " ")
		var sign int
		if strings.HasPrefix(value, "-") {
			sign = -1
			value = strings.Replace(value, "-", "", 1)
		} else {
			sign = 1
		}
		strs := strings.Split(value, ":")
		if strs == nil || len(strs) != 2 {
			return nil, errors.New("unable to parse database timezone offset")
		}
		hourOffset, err := strconv.ParseInt(strs[0], 10, 32)
		if err != nil {
			return nil, err
		}
		minStr := strs[1]
		nullIndex := strings.IndexRune(minStr, '\x00')
		if nullIndex > -1 {
			minStr = minStr[:nullIndex]
		}
		minOffset, err := strconv.ParseInt(minStr, 10, 32)
		if err != nil {
			return nil, err
		}
		offset := sign * ((int(hourOffset) * 3600) + (int(minOffset) * 60))
		return time.FixedZone("SESSIONTIMEZONE", offset), nil
	} else {
		return nil, errors.New("unable to retrieve database timezone")
	}
}

func validate(expected interface{}, rset *ora.Rset, t *testing.T) {
	if 1 != len(rset.Row) {
		t.Fatalf("column count: expected(%v), actual(%v)", 1, len(rset.Row))
	}

	switch expected.(type) {
	case int64:
		row := rset.NextRow()
		compare_int64(expected, row[0], t)
	case int32:
		row := rset.NextRow()
		compare_int32(expected, row[0], t)
	case int16:
		row := rset.NextRow()
		compare_int16(expected, row[0], t)
	case int8:
		row := rset.NextRow()
		compare_int8(expected, row[0], t)
	case uint64:
		row := rset.NextRow()
		compare_uint64(expected, row[0], t)
	case uint32:
		row := rset.NextRow()
		compare_uint32(expected, row[0], t)
	case uint16:
		row := rset.NextRow()
		compare_uint16(expected, row[0], t)
	case uint8:
		row := rset.NextRow()
		compare_uint8(expected, row[0], t)
	case float64:
		row := rset.NextRow()
		compare_float64(expected, row[0], t)
	case float32:
		row := rset.NextRow()
		compare_float32(expected, row[0], t)
	case ora.Int64:
		row := rset.NextRow()
		compare_OraInt64(expected, row[0], t)
	case ora.Int32:
		row := rset.NextRow()
		compare_OraInt32(expected, row[0], t)
	case ora.Int16:
		row := rset.NextRow()
		compare_OraInt16(expected, row[0], t)
	case ora.Int8:
		row := rset.NextRow()
		compare_OraInt8(expected, row[0], t)
	case ora.Uint64:
		row := rset.NextRow()
		compare_OraUint64(expected, row[0], t)
	case ora.Uint32:
		row := rset.NextRow()
		compare_OraUint32(expected, row[0], t)
	case ora.Uint16:
		row := rset.NextRow()
		compare_OraUint16(expected, row[0], t)
	case ora.Uint8:
		row := rset.NextRow()
		compare_OraUint8(expected, row[0], t)
	case ora.Float64:
		row := rset.NextRow()
		compare_OraFloat64(expected, row[0], t)
	case ora.Float32:
		row := rset.NextRow()
		compare_OraFloat32(expected, row[0], t)

	case ora.IntervalYM:
		row := rset.NextRow()
		compare_OraIntervalYM(expected, row[0], t)
	case ora.IntervalDS:
		row := rset.NextRow()
		compare_OraIntervalDS(expected, row[0], t)

	case ora.Bfile:
		row := rset.NextRow()
		compare_OraBfile(expected, row[0], t)

	case []int64:
		for rset.Next() {
			expectedElem := elemAt(expected, rset.Index)
			compare_int64(expectedElem, rset.Row[0], t)
		}

	case []ora.IntervalYM:
		for rset.Next() {
			expectedElem := elemAt(expected, rset.Index)
			compare_OraIntervalYM(expectedElem, rset.Row[0], t)
		}
	case []ora.IntervalDS:
		for rset.Next() {
			expectedElem := elemAt(expected, rset.Index)
			compare_OraIntervalDS(expectedElem, rset.Row[0], t)
		}
	}
	testErr(rset.Err, t)
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
	stmt, err := ses.Prep(createTableSql(tableName, multiple, oct))
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	_, err = stmt.Exe()
	return tableName, err
}

func dropTable(tableName string, ses *ora.Ses, t *testing.T) {
	stmt, err := ses.Prep(fmt.Sprintf("drop table %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exe()
	testErr(err, t)
}

func createTableDB(db *sql.DB, t *testing.T, octs ...oracleColumnType) string {
	tableName := tableName()
	stmt, err := db.Prepare(createTableSql(tableName, 1, octs...))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)
	return tableName
}

func dropTableDB(db *sql.DB, t *testing.T, tableName string) {
	stmt, err := db.Prepare(fmt.Sprintf("drop table %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
	testErr(err, t)
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
	testTableId++
	return "t" + strconv.Itoa(testTableId)
}

func testErr(err error, t *testing.T, expectedErrs ...error) {
	if err != nil {
		if expectedErrs == nil {
			t.Fatalf("%v: %s", err, getStack(1))
		} else {
			var isSkipping bool
			for _, expectedErr := range expectedErrs {
				isSkipping = expectedErr == err
				if isSkipping {
					break
				}
			}
			if !isSkipping {
				t.Fatal(err)
			}
		}
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
	if _, ok := value.(int64); ok {
		return true
	}
	if _, ok := value.(int32); ok {
		return true
	}
	if _, ok := value.(int16); ok {
		return true
	}
	if _, ok := value.(int8); ok {
		return true
	}
	if _, ok := value.(uint64); ok {
		return true
	}
	if _, ok := value.(uint32); ok {
		return true
	}
	if _, ok := value.(uint16); ok {
		return true
	}
	if _, ok := value.(uint8); ok {
		return true
	}
	if _, ok := value.(float64); ok {
		return true
	}
	if _, ok := value.(float32); ok {
		return true
	}
	if _, ok := value.(ora.Int64); ok {
		return true
	}
	if _, ok := value.(ora.Int32); ok {
		return true
	}
	if _, ok := value.(ora.Int16); ok {
		return true
	}
	if _, ok := value.(ora.Int8); ok {
		return true
	}
	if _, ok := value.(ora.Uint64); ok {
		return true
	}
	if _, ok := value.(ora.Uint32); ok {
		return true
	}
	if _, ok := value.(ora.Uint16); ok {
		return true
	}
	if _, ok := value.(ora.Uint8); ok {
		return true
	}
	if _, ok := value.(ora.Float64); ok {
		return true
	}
	if _, ok := value.(ora.Float32); ok {
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

func goColumnTypeFromSlice(value interface{}) ora.GoColumnType {
	if _, ok := value.([]int64); ok {
		return ora.I64
	}
	if _, ok := value.([]int32); ok {
		return ora.I32
	}
	if _, ok := value.([]int16); ok {
		return ora.I16
	}
	if _, ok := value.([]int8); ok {
		return ora.I8
	}
	if _, ok := value.([]uint64); ok {
		return ora.U64
	}
	if _, ok := value.([]uint32); ok {
		return ora.U32
	}
	if _, ok := value.([]uint16); ok {
		return ora.U16
	}
	if _, ok := value.([]uint8); ok {
		return ora.U8
	}
	if _, ok := value.([]float64); ok {
		return ora.F64
	}
	if _, ok := value.([]float32); ok {
		return ora.F32
	}
	if _, ok := value.([]ora.Int64); ok {
		return ora.OraI64
	}
	if _, ok := value.([]ora.Int32); ok {
		return ora.OraI32
	}
	if _, ok := value.([]ora.Int16); ok {
		return ora.OraI16
	}
	if _, ok := value.([]ora.Int8); ok {
		return ora.OraI8
	}
	if _, ok := value.([]ora.Uint64); ok {
		return ora.OraU64
	}
	if _, ok := value.([]ora.Uint32); ok {
		return ora.OraU32
	}
	if _, ok := value.([]ora.Uint16); ok {
		return ora.OraU16
	}
	if _, ok := value.([]ora.Uint8); ok {
		return ora.OraU8
	}
	if _, ok := value.([]ora.Float64); ok {
		return ora.OraF64
	}
	if _, ok := value.([]ora.Float32); ok {
		return ora.OraF32
	}
	if _, ok := value.([]time.Time); ok {
		return ora.T
	}
	if _, ok := value.([]ora.Time); ok {
		return ora.OraT
	}
	if _, ok := value.([]string); ok {
		return ora.S
	}
	if _, ok := value.([]ora.String); ok {
		return ora.OraS
	}
	if _, ok := value.([]bool); ok {
		return ora.B
	}
	if _, ok := value.([]ora.Bool); ok {
		return ora.OraB
	}
	return ora.D
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

func slice(goColumnType ora.GoColumnType, length int) interface{} {
	switch goColumnType {
	case ora.I64:
		return make([]int64, length)
	case ora.I32:
		return make([]int32, length)
	case ora.I16:
		return make([]int16, length)
	case ora.I8:
		return make([]int8, length)
	case ora.U64:
		return make([]uint64, length)
	case ora.U32:
		return make([]uint32, length)
	case ora.U16:
		return make([]uint16, length)
	case ora.U8:
		return make([]uint8, length)
	case ora.F64:
		return make([]float64, length)
	case ora.F32:
		return make([]float32, length)
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

func printValues(v interface{}) {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Slice {
		for n := 0; n < value.Len(); n++ {
			fmt.Printf("%v, ", value.Index(n))
		}
		fmt.Println()
	}
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
	a, aOk := actual.(int64)
	if !eOk {
		ePtr, ePtrOk := expected.(*int64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to int64 or *int64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int64 or *int64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
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
			t.Fatalf("Unable to cast expected value to int32 or *int32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int32 or *int32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to int16 or *int16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int16 or *int16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to int8 or *int8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*int8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to int8 or *int8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to uint64 or *uint64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint64 or *uint64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to uint32 or *uint32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint32 or *uint32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to uint16 or *uint16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint16 or *uint16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to uint8 or *uint8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*uint8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to uint8 or *uint8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_float64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(float64)
	a, aOk := actual.(float64)
	if !eOk {
		ePtr, ePtrOk := expected.(*float64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to float64 or *float64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*float64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to float64 or *float64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !isFloat64Close(e, a, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_float32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(float32)
	a, aOk := actual.(float32)
	if !eOk {
		ePtr, ePtrOk := expected.(*float32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to float32 or *float32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*float32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to float32 or *float32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
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
			t.Fatalf("Unable to cast expected value to ora.Int64 or *ora.Int64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int64 or *ora.Int64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Int32 or *ora.Int32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int32 or *ora.Int32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Int16 or *ora.Int16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int16 or *ora.Int16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Int8 or *ora.Int8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Int8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Int8 or *ora.Int8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Uint64 or *ora.Uint64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint64 or *ora.Uint64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Uint32 or *ora.Uint32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint32 or *ora.Uint32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Uint16 or *ora.Uint16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint16 or *ora.Uint16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Uint8 or *ora.Uint8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Uint8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Uint8 or *ora.Uint8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Float64 or *ora.Float64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Float64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Float64 or *ora.Float64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Float32 or *ora.Float32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Float32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Float32 or *ora.Float32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
				t.Fatalf("Unable to cast expected value to time.Time, *time.Time, ora.Time. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
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
				t.Fatalf("Unable to cast actual value to time.Time, *time.Time, ora.Time. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.Time or *ora.Time. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Time)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Time or *ora.Time. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
				t.Fatalf("Unable to cast expected value to string, *string, ora.String. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
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
	case ora.Lob:
		b, err := ioutil.ReadAll(x)
		if err != nil {
			t.Errorf("read %v: %v", x, err)
		}
		a = string(b)
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
		t.Fatalf("expected(%v), actual(%v)\n%s", e, a, getStack(2))
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
			t.Fatalf("Unable to cast expected value to ora.String or *ora.String. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.String)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.String or *ora.String. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
				t.Fatalf("Unable to cast expected value to bool, *bool, ora.Bool. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
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
				t.Fatalf("Unable to cast actual value to bool, *bool, ora.Bool. (%v, %v): %s", reflect.TypeOf(actual), actual, getStack(2))
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
			t.Fatalf("Unable to cast expected value to ora.Bool or *ora.Bool. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.Bool)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.Bool or *ora.Bool. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to []byte or ora.Raw. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	var a []byte
	switch x := actual.(type) {
	case []byte:
		a = x
	case ora.Raw:
		a = x.Value

	case ora.Lob:
		t.Logf("Lob=%v", x)
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
		t.Fatalf("Unable to cast expected value to ora.Raw. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to ora.Raw. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.IntervalYM or *ora.IntervalYM. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.IntervalYM)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.IntervalYM or *ora.IntervalYM. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			t.Fatalf("Unable to cast expected value to ora.IntervalDS or *ora.IntervalDS. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*ora.IntervalDS)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to ora.IntervalDS or *ora.IntervalDS. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
		t.Fatalf("Unable to cast expected value to ora.Bfile. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to ora.Bfile. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
	} else if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_nil(expected interface{}, actual interface{}, t *testing.T) {
	if expected != nil {
		t.Fatalf("Expected value is not nil. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
	}
	if actual != nil {
		t.Fatalf("Actual value is not nil. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
		//fmt.Printf("isFloat32Close xx, yy: %v, %v\n", xx, yy)
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

func gen_OraFloat64(isNull bool) ora.Float64 {
	return ora.Float64{Value: gen_float64(), IsNull: isNull}
}

func gen_OraFloat64Trunc(isNull bool) ora.Float64 {
	return ora.Float64{Value: gen_float64Trunc(), IsNull: isNull}
}

func gen_OraFloat32(isNull bool) ora.Float32 {
	return ora.Float32{Value: gen_float32(), IsNull: isNull}
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

func gen_float64Slice() []float64 {
	expected := make([]float64, 5)
	expected[0] = -float64(6.28318) //5307179586)
	expected[1] = -float64(3.14159) //2653589793)
	expected[2] = 0
	expected[3] = float64(3.14159) //2653589793)
	expected[4] = float64(6.28318) //5307179586)
	return expected
}

func gen_float64TruncSlice() []float64 {
	expected := make([]float64, 5)
	expected[0] = -float64(6)
	expected[1] = -float64(3)
	expected[2] = 0
	expected[3] = float64(3)
	expected[4] = float64(6)
	return expected
}

func gen_float32Slice() []float32 {
	expected := make([]float32, 5)
	expected[0] = -float32(6.28318)
	expected[1] = -float32(3.14159)
	expected[2] = 0
	expected[3] = float32(3.14159)
	expected[4] = float32(6.28318)
	return expected
}

func gen_float32TruncSlice() []float32 {
	expected := make([]float32, 5)
	expected[0] = -float32(6)
	expected[1] = -float32(3)
	expected[2] = 0
	expected[3] = float32(3)
	expected[4] = float32(6)
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

func gen_OraFloat64Slice(isNull bool) []ora.Float64 {
	expected := make([]ora.Float64, 5)
	expected[0] = ora.Float64{Value: -float64(6.28318)}
	expected[1] = ora.Float64{Value: -float64(3.14159)}
	expected[2] = ora.Float64{IsNull: isNull}
	expected[3] = ora.Float64{Value: float64(3.14159)}
	expected[4] = ora.Float64{Value: float64(6.28318)}
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

func gen_OraFloat32Slice(isNull bool) []ora.Float32 {
	expected := make([]ora.Float32, 5)
	expected[0] = ora.Float32{Value: -float32(6.28318)}
	expected[1] = ora.Float32{Value: -float32(3.14159)}
	expected[2] = ora.Float32{IsNull: isNull}
	expected[3] = ora.Float32{Value: float32(3.14159)}
	expected[4] = ora.Float32{Value: float32(6.28318)}
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
	return "Go is expressive, concise, clean, and efficient."
}

func gen_OraString(isNull bool) ora.String {
	return ora.String{Value: gen_string(), IsNull: isNull}
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

func gen_OraStringSlice(isNull bool) interface{} {
	expected := make([]ora.String, 5)
	expected[0] = ora.String{Value: "Go is expressive, concise, clean, and efficient."}
	expected[1] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	expected[2] = ora.String{Value: "Go compiles quickly to machine code yet has", IsNull: isNull}
	expected[3] = ora.String{Value: "It's a fast, statically typed, compiled"}
	expected[4] = ora.String{Value: "One of Go's key design goals is code"}
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

func gen_OraBfileEmpty(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "", Filename: ""}
}

func gen_OraBfileEmptyDir(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "", Filename: "test.txt"}
}

func gen_OraBfileEmptyFilename(isNull bool) interface{} {
	return ora.Bfile{IsNull: isNull, DirectoryAlias: "TEMP_DIR", Filename: ""}
}

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

	testDb.Exec(`DROP TABLE test_janus`)
	if _, err := testDb.Exec(`CREATE TABLE test_janus (
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
	if _, err := testDb.Exec(`INSERT INTO test_janus (
		leg, site, hole, core, core_type, section_number,
		section_type, top_cm, bot_cm, depth_mbsf,
		inor_c_wt_pct, caco3_wt_pct, tot_c_wt_pct,
		org_c_wt_pct, nit_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (207, 1259, 'C', 3, 'B', 4, '@', 5.2, NULL, 7.6, 8., 9., 10., 11., NULL , 13., 14.)`,
	); err != nil {
		t.Fatal(err)
	}

	enableLogging(t)

	if _, err := testDb.Exec(`INSERT INTO test_janus (
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
	   test_janus
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

		t.Logf("Results: %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v", Leg, Site, Hole, Core, Core_type, Section_number, Section_type, Top_cm, Bot_cm, Depth_mbsf, Inor_c_wt_pct, Caco3_wt_pct, Tot_c_wt_pct, Org_c_wt_pct, Nit_wt_pct, Sul_wt_pct, H_wt_pct)

	}
	if err := rows.Err(); err != nil {
		t.Error(err)
	}
}

func TestUnderflow(t *testing.T) {
	tbl := "test_underflow"
	testDb.Exec(`DROP TABLE ` + tbl)
	if _, err := testDb.Exec(`CREATE TABLE ` + tbl + ` (
		caco3_wt_pct NUMBER NULL,
		sul_wt_pct NUMBER NULL,
		h_wt_pct NUMBER(6,3) NULL
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := testDb.Exec(`INSERT INTO ` + tbl + ` (
		caco3_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (9., 13., 14.)`,
	); err != nil {
		t.Fatal(err)
	}

	enableLogging(t)

	for caseNum, test := range [][3]float64{
		{10., 0.8, 4.2},
		{9.19, 0.8, 0.12},
		{9.99, 0.8, 0.42},
	} {
		testDb.Exec("TRUNCATE TABLE " + tbl)
		if _, err := testDb.Exec(fmt.Sprintf(
			`INSERT INTO `+tbl+` (
		caco3_wt_pct, sul_wt_pct, h_wt_pct)
	VALUES (%f, %f, %f)`, test[0], test[1], test[2]),
		); err != nil {
			t.Fatalf("%d. %v", caseNum, err)
		}

		qry := `SELECT caco3_wt_pct, sul_wt_pct, h_wt_pct
		FROM ` + tbl

		rows, err := testDb.Query(qry)
		if err != nil {
			t.Errorf(`%d. Error with "%s": %s`, caseNum, qry, err)
			return
		}
		defer rows.Close()

		i := 0
		for rows.Next() {
			i++
			got := make([]sql.NullFloat64, len(test))

			if err := rows.Scan(&got[0], &got[1], &got[2]); err != nil {
				t.Fatalf("scan %d. record: %v", i, err)
			}

			t.Logf("Results: %v", got)

			for j, f := range got {
				if !f.Valid || f.Float64 != test[j] {
					t.Errorf("%d. %d. got %v, awaited %v.", caseNum, j, f, test[j])
				}
			}
		}
		if err := rows.Err(); err != nil {
			t.Errorf("%d. %v", caseNum, err)
		}
		rows.Close()
	}
}
