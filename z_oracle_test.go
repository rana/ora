// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
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

var testServerName string
var testUsername string
var testPassword string
var testConStr string
var testDbsessiontimezone *time.Location
var testTableId int
var testWorkloadColumnCount int
var testEnv *Env
var testSrv *Srv
var testSes *Ses
var testDb *sql.DB
var testCon driver.Conn

func init() {
	testWorkloadColumnCount = 20
	testServerName = os.Getenv("GO_ORA_DRV_TEST_DB")
	testUsername = os.Getenv("GO_ORA_DRV_TEST_USERNAME")
	testPassword = os.Getenv("GO_ORA_DRV_TEST_PASSWORD")
	testConStr = fmt.Sprintf("%v/%v@%v", testUsername, testPassword, testServerName)

	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_DB = '%v'\n", testServerName)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_USERNAME = '%v'\n", testUsername)
	fmt.Printf("Read environment variable GO_ORA_DRV_TEST_PASSWORD = '%v'\n", testPassword)

	var err error

	// setup test environment, server and session
	testEnv, err := GetDrv().OpenEnv()
	if err != nil {
		fmt.Println("initError: ", err)
	}
	testSrv, err = testEnv.OpenSrv(testServerName)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	testSes, err = testSrv.OpenSes(testUsername, testPassword)
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
	var buf bytes.Buffer
	buf.WriteString("BEGIN ")
	buf.WriteString("FOR c IN (SELECT table_name FROM user_tables) LOOP ")
	buf.WriteString("EXECUTE IMMEDIATE ('DROP TABLE ' || c.table_name || ' CASCADE CONSTRAINTS'); ")
	buf.WriteString("END LOOP; ")
	buf.WriteString("END;")
	stmt, err := testSes.Prep(buf.String())
	if err != nil {
		fmt.Println("initError: ", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println("initError: ", err)
	}

	// setup test db
	testDb, err = sql.Open(Name, testConStr)
	if err != nil {
		fmt.Println("initError: ", err)
	}
	testCon, err = GetDrv().Open(testConStr)
	if err != nil {
		fmt.Println("initError: ", err)
	}
}

func testIterations() int {
	if testing.Short() {
		return 1
	} else {
		return 1
	}
}

func testBindDefine(expected interface{}, oct oracleColumnType, t *testing.T, c *StmtConfig, goColumnTypes ...GoColumnType) {
	var gct GoColumnType
	if len(goColumnTypes) > 0 {
		gct = goColumnTypes[0]
	} else {
		gct = goColumnTypeFromValue(expected)
	}
	//fmt.Printf("testBindDefine (%v)\n", gctName(gct))

	for n := 0; n < testIterations(); n++ {
		tableName, err := createTable(1, oct, testSes)
		testErr(err, t)
		defer dropTable(tableName, testSes, t)

		// insert
		insertStmt, err := testSes.Prep(fmt.Sprintf("insert into %v (c1) values (:c1)", tableName))
		if c != nil {
			insertStmt.Config = *c
		}
		defer insertStmt.Close()
		testErr(err, t)
		rowsAffected, err := insertStmt.Exec(expected)
		testErr(err, t)
		expLen := length(expected)
		if gct == Bin || gct == OraBin {
			expLen = 1
		}
		if expLen != int(rowsAffected) {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", expLen, rowsAffected)
		}

		// select
		selectStmt, err := testSes.Prep(fmt.Sprintf("select c1 from %v", tableName), gct)
		defer selectStmt.Close()
		testErr(err, t)
		rset, err := selectStmt.Query()
		testErr(err, t)

		// validate
		validate(expected, rset, t)
	}
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
			var goColumnType GoColumnType
			if oct == longRaw || oct == longRawNull || oct == raw2000 || oct == raw2000Null || oct == blob || oct == blobNull {
				goColumnType = Bin
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
		rowsAffected, err := stmt.Exec(expected, actual)
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
		rowsAffected, err := insertStmt.Exec(expected)
		testErr(err, t)
		if rowsAffected != 1 {
			t.Fatalf("insert rows affected: expected(%v), actual(%v)", 1, rowsAffected)
		}

		// select
		var selectStmt *Stmt
		var rset *Rset
		if isNumeric(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, c1 from %v", tableName), I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isTime(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), T, OraT)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isString(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), S, S)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isBool(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), B, OraB)
			defer selectStmt.Close()
			testErr(err, t)
		} else if isBytes(expected) {
			selectStmt, err = testSes.Prep(fmt.Sprintf("select c1, c1 from %v", tableName), Bin, OraBin)
			defer selectStmt.Close()
			testErr(err, t)
		}
		rset, err = selectStmt.Query()
		testErr(err, t)

		// validate
		hasRow := rset.Next()
		testErr(rset.Err, t)
		if !hasRow {
			t.Fatalf("no row returned")
		} else if len(rset.Row) != len(selectStmt.gcts) {
			t.Fatalf("select column count: expected(%v), actual(%v)", len(selectStmt.gcts), len(rset.Row))
		} else {
			for n, goColumnType := range selectStmt.gcts {
				if isNumeric(expected) {
					compare(castInt(expected, goColumnType), rset.Row[n], goColumnType, t)
				}
				switch goColumnType {
				case T:
					compare_time(expected, rset.Row[n], t)
				case OraT:
					value, ok := rset.Row[n].(Time)
					if ok {
						compare_time(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value. (%v, %v)", reflect.TypeOf(rset.Row[n]).Name(), rset.Row[n])
					}
				case S:
					compare_string(expected, rset.Row[n], t)
				case OraS:
					value, ok := rset.Row[n].(String)
					if ok {
						compare_string(expected, value.Value, t)
					} else {
						t.Fatalf("Unpexected rset.Row[n] value. (%v, %v)", reflect.TypeOf(rset.Row[n]).Name(), rset.Row[n])
					}
				case B, OraB:
					compare_bool(expected, rset.Row[n], t)
				case Bin, OraBin:
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
			expected := make([]driver.Value, currentMultiple)
			gcts := make([]GoColumnType, currentMultiple)
			for c := 0; c < currentMultiple; c++ {
				switch oct {
				case numberP38S0, numberP38S0Null, numberP16S15, numberP16S15Null, binaryDouble, binaryDoubleNull, binaryFloat, binaryFloatNull, floatP126, floatP126Null:
					expected[c] = gen_int64()
					gcts[c] = I64
				case date, dateNull:
					expected[c] = gen_date()
					gcts[c] = T
				case timestampP9, timestampP9Null, timestampTzP9, timestampTzP9Null, timestampLtzP9, timestampLtzP9Null:
					expected[c] = gen_time()
					gcts[c] = T
				case charB48, charB48Null, charC48, charC48Null, nchar48, nchar48Null, varcharB48, varcharB48Null, varcharC48, varcharC48Null, varchar2B48, varchar2B48Null, varchar2C48, varchar2C48Null, nvarchar248, nvarchar248Null, long, longNull, clob, clobNull, nclob, nclobNull:
					expected[c] = gen_string()
					gcts[c] = S
				case charB1, charB1Null, charC1, charC1Null:
					expected[c] = gen_boolTrue()
					gcts[c] = B
				case blob, blobNull, longRaw, longRawNull:
					expected[c] = gen_bytes(9)
					gcts[c] = Bin
				case raw2000, raw2000Null:
					expected[c] = gen_bytes(2000)
					gcts[c] = Bin
				}
				if c > 0 {
					sql.WriteString(", ")
				}
				sql.WriteString(fmt.Sprintf(":c%v", c+1))
			}
			sql.WriteString(")")

			// insert values
			//fmt.Println(sql.String())
			insertStmt, err := testCon.Prepare(sql.String())
			testErr(err, t)
			_, err = insertStmt.Exec(expected)
			testErr(err, t)
			insertStmt.Close()

			// fetch values and compare
			sql.Reset()
			sql.WriteString(fmt.Sprintf("select * from %v", tableName))
			fetchStmt, err := testSes.Prep(sql.String())
			testErr(err, t)
			fetchStmt.gcts = gcts
			rset, err := fetchStmt.Query()
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
	rset, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	hasRow := rset.Next()
	if !hasRow {
		return nil, errNew("no time zone returned from database")
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
			return nil, errNew("unable to parse database timezone offset")
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
		return nil, errNew("unable to retrieve database timezone")
	}
}

func validate(expected interface{}, rset *Rset, t *testing.T) {
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
	case Int64:
		row := rset.NextRow()
		compare_OraInt64(expected, row[0], t)
	case Int32:
		row := rset.NextRow()
		compare_OraInt32(expected, row[0], t)
	case Int16:
		row := rset.NextRow()
		compare_OraInt16(expected, row[0], t)
	case Int8:
		row := rset.NextRow()
		compare_OraInt8(expected, row[0], t)
	case Uint64:
		row := rset.NextRow()
		compare_OraUint64(expected, row[0], t)
	case Uint32:
		row := rset.NextRow()
		compare_OraUint32(expected, row[0], t)
	case Uint16:
		row := rset.NextRow()
		compare_OraUint16(expected, row[0], t)
	case Uint8:
		row := rset.NextRow()
		compare_OraUint8(expected, row[0], t)
	case Float64:
		row := rset.NextRow()
		compare_OraFloat64(expected, row[0], t)
	case Float32:
		row := rset.NextRow()
		compare_OraFloat32(expected, row[0], t)

	case IntervalYM:
		row := rset.NextRow()
		compare_OraIntervalYM(expected, row[0], t)
	case IntervalDS:
		row := rset.NextRow()
		compare_OraIntervalDS(expected, row[0], t)

	case []int64:
		for rset.Next() {
			expectedElem := elemAt(expected, rset.Index)
			compare_int64(expectedElem, rset.Row[0], t)
		}

	case []IntervalYM:
		for rset.Next() {
			expectedElem := elemAt(expected, rset.Index)
			compare_OraIntervalYM(expectedElem, rset.Row[0], t)
		}
	case []IntervalDS:
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
	case Int64:
		compare_OraInt64(expected, actual, t)
	case Int32:
		compare_OraInt32(expected, actual, t)
	case Int16:
		compare_OraInt16(expected, actual, t)
	case Int8:
		compare_OraInt8(expected, actual, t)
	case Uint64:
		compare_OraUint64(expected, actual, t)
	case Uint32:
		compare_OraUint32(expected, actual, t)
	case Uint16:
		compare_OraUint16(expected, actual, t)
	case Uint8:
		compare_OraUint8(expected, actual, t)
	case Float64:
		compare_OraFloat64(expected, actual, t)
	case Float32:
		compare_OraFloat32(expected, actual, t)
	case IntervalYM:
		compare_OraIntervalYM(expected, actual, t)
	case IntervalDS:
		compare_OraIntervalDS(expected, actual, t)
	}
}

func createTable(multiple int, oct oracleColumnType, ses *Ses) (string, error) {
	tableName := fmt.Sprintf("%v_%v", tableName(), multiple)
	stmt, err := ses.Prep(createTableSql(tableName, multiple, oct))
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	return tableName, err
}

func dropTable(tableName string, ses *Ses, t *testing.T) {
	stmt, err := ses.Prep(fmt.Sprintf("drop table %v", tableName))
	defer stmt.Close()
	testErr(err, t)
	_, err = stmt.Exec()
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
			t.Fatal(err)
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

func goColumnTypeFromValue(value interface{}) GoColumnType {
	switch value.(type) {
	case int64, []int64:
		return I64
	case int32, []int32:
		return I32
	case int16, []int16:
		return I16
	case int8, []int8:
		return I8
	case uint64, []uint64:
		return U64
	case uint32, []uint32:
		return U32
	case uint16, []uint16:
		return U16
	case uint8, []uint8:
		return U8
	case float64, []float64:
		return F64
	case float32, []float32:
		return F32
	case Int64, []Int64:
		return OraI64
	case Int32, []Int32:
		return OraI32
	case Int16, []Int16:
		return OraI16
	case Int8, []Int8:
		return OraI8
	case Uint64, []Uint64:
		return OraU64
	case Uint32, []Uint32:
		return OraU32
	case Uint16, []Uint16:
		return OraU16
	case Uint8, []Uint8:
		return OraU8
	case Float64, []Float64:
		return OraF64
	case Float32, []Float32:
		return OraF32
	case time.Time, []time.Time:
		return T
	case Time, []Time:
		return OraT
	case string, []string:
		return S
	case String, []String:
		return OraS
	case bool, []bool:
		return B
	case Bool, []Bool:
		return OraB
	case Binary:
		return OraBin
	}

	return D
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
	if _, ok := value.(Int64); ok {
		return true
	}
	if _, ok := value.(Int32); ok {
		return true
	}
	if _, ok := value.(Int16); ok {
		return true
	}
	if _, ok := value.(Int8); ok {
		return true
	}
	if _, ok := value.(Uint64); ok {
		return true
	}
	if _, ok := value.(Uint32); ok {
		return true
	}
	if _, ok := value.(Uint16); ok {
		return true
	}
	if _, ok := value.(Uint8); ok {
		return true
	}
	if _, ok := value.(Float64); ok {
		return true
	}
	if _, ok := value.(Float32); ok {
		return true
	}
	return false
}

func isTime(value interface{}) bool {
	if _, ok := value.(time.Time); ok {
		return true
	}
	if _, ok := value.(Time); ok {
		return true
	}
	return false
}

func isString(value interface{}) bool {
	if _, ok := value.(string); ok {
		return true
	}
	if _, ok := value.(String); ok {
		return true
	}
	return false
}

func isBool(value interface{}) bool {
	if _, ok := value.(bool); ok {
		return true
	}
	if _, ok := value.(Bool); ok {
		return true
	}
	return false
}

func isBytes(value interface{}) bool {
	if _, ok := value.([]byte); ok {
		return true
	}
	if _, ok := value.(Binary); ok {
		return true
	}
	return false
}

func goColumnTypeFromSlice(value interface{}) GoColumnType {
	if _, ok := value.([]int64); ok {
		return I64
	}
	if _, ok := value.([]int32); ok {
		return I32
	}
	if _, ok := value.([]int16); ok {
		return I16
	}
	if _, ok := value.([]int8); ok {
		return I8
	}
	if _, ok := value.([]uint64); ok {
		return U64
	}
	if _, ok := value.([]uint32); ok {
		return U32
	}
	if _, ok := value.([]uint16); ok {
		return U16
	}
	if _, ok := value.([]uint8); ok {
		return U8
	}
	if _, ok := value.([]float64); ok {
		return F64
	}
	if _, ok := value.([]float32); ok {
		return F32
	}
	if _, ok := value.([]Int64); ok {
		return OraI64
	}
	if _, ok := value.([]Int32); ok {
		return OraI32
	}
	if _, ok := value.([]Int16); ok {
		return OraI16
	}
	if _, ok := value.([]Int8); ok {
		return OraI8
	}
	if _, ok := value.([]Uint64); ok {
		return OraU64
	}
	if _, ok := value.([]Uint32); ok {
		return OraU32
	}
	if _, ok := value.([]Uint16); ok {
		return OraU16
	}
	if _, ok := value.([]Uint8); ok {
		return OraU8
	}
	if _, ok := value.([]Float64); ok {
		return OraF64
	}
	if _, ok := value.([]Float32); ok {
		return OraF32
	}
	if _, ok := value.([]time.Time); ok {
		return T
	}
	if _, ok := value.([]Time); ok {
		return OraT
	}
	if _, ok := value.([]string); ok {
		return S
	}
	if _, ok := value.([]String); ok {
		return OraS
	}
	if _, ok := value.([]bool); ok {
		return B
	}
	if _, ok := value.([]Bool); ok {
		return OraB
	}

	return D
}

func castInt(v interface{}, goColumnType GoColumnType) interface{} {
	value := reflect.ValueOf(v)
	switch goColumnType {
	case I64:
		return value.Int()
	case I32:
		return int32(value.Int())
	case I16:
		return int16(value.Int())
	case I8:
		return int8(value.Int())
	case U64:
		return uint64(value.Int())
	case U32:
		return uint32(value.Int())
	case U16:
		return uint16(value.Int())
	case U8:
		return uint8(value.Int())
	case F64:
		return float64(value.Int())
	case F32:
		return float32(value.Int())
	case OraI64:
		return Int64{Value: value.Int()}
	case OraI32:
		return Int32{Value: int32(value.Int())}
	case OraI16:
		return Int16{Value: int16(value.Int())}
	case OraI8:
		return Int8{Value: int8(value.Int())}
	case OraU64:
		return Uint64{Value: uint64(value.Int())}
	case OraU32:
		return Uint32{Value: uint32(value.Int())}
	case OraU16:
		return Uint16{Value: uint16(value.Int())}
	case OraU8:
		return Uint8{Value: uint8(value.Int())}
	case OraF64:
		return Float64{Value: float64(value.Int())}
	case OraF32:
		return Float32{Value: float32(value.Int())}
	}
	return nil
}

func slice(goColumnType GoColumnType, length int) interface{} {
	switch goColumnType {
	case I64:
		return make([]int64, length)
	case I32:
		return make([]int32, length)
	case I16:
		return make([]int16, length)
	case I8:
		return make([]int8, length)
	case U64:
		return make([]uint64, length)
	case U32:
		return make([]uint32, length)
	case U16:
		return make([]uint16, length)
	case U8:
		return make([]uint8, length)
	case F64:
		return make([]float64, length)
	case F32:
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

func compare(expected interface{}, actual interface{}, goColumnType GoColumnType, t *testing.T) {
	switch goColumnType {
	case I64:
		compare_int64(expected, actual, t)
	case I32:
		compare_int32(expected, actual, t)
	case I16:
		compare_int16(expected, actual, t)
	case I8:
		compare_int8(expected, actual, t)
	case U64:
		compare_uint64(expected, actual, t)
	case U32:
		compare_uint32(expected, actual, t)
	case U16:
		compare_uint16(expected, actual, t)
	case U8:
		compare_uint8(expected, actual, t)
	case F64:
		compare_float64(expected, actual, t)
	case F32:
		compare_float32(expected, actual, t)
	case OraI64:
		compare_OraInt64(expected, actual, t)
	case OraI32:
		compare_OraInt32(expected, actual, t)
	case OraI16:
		compare_OraInt16(expected, actual, t)
	case OraI8:
		compare_OraInt8(expected, actual, t)
	case OraU64:
		compare_OraUint64(expected, actual, t)
	case OraU32:
		compare_OraUint32(expected, actual, t)
	case OraU16:
		compare_OraUint16(expected, actual, t)
	case OraU8:
		compare_OraUint8(expected, actual, t)
	case OraF64:
		compare_OraFloat64(expected, actual, t)
	case OraF32:
		compare_OraFloat32(expected, actual, t)
	case T:
		compare_time(expected, actual, t)
	case OraT:
		compare_OraTime(expected, actual, t)
	case S:
		compare_string(expected, actual, t)
	case OraS:
		compare_OraString(expected, actual, t)
	case B:
		compare_bool(expected, actual, t)
	case OraB:
		compare_OraBool(expected, actual, t)
	case Bin:
		compare_bytes(expected, actual, t)
	case OraBin:
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
	e, eOk := expected.(Int64)
	a, aOk := actual.(Int64)
	if !eOk {
		ePtr, ePtrOk := expected.(*Int64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Int64 or *Int64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Int64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Int64 or *Int64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Int32)
	a, aOk := actual.(Int32)
	if !eOk {
		ePtr, ePtrOk := expected.(*Int32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Int32 or *Int32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Int32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Int32 or *Int32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Int16)
	a, aOk := actual.(Int16)
	if !eOk {
		ePtr, ePtrOk := expected.(*Int16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Int16 or *Int16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Int16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Int16 or *Int16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraInt8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Int8)
	a, aOk := actual.(Int8)
	if !eOk {
		ePtr, ePtrOk := expected.(*Int8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Int8 or *Int8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Int8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Int8 or *Int8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Uint64)
	a, aOk := actual.(Uint64)
	if !eOk {
		ePtr, ePtrOk := expected.(*Uint64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Uint64 or *Uint64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Uint64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Uint64 or *Uint64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Uint32)
	a, aOk := actual.(Uint32)
	if !eOk {
		ePtr, ePtrOk := expected.(*Uint32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Uint32 or *Uint32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Uint32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Uint32 or *Uint32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint16(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Uint16)
	a, aOk := actual.(Uint16)
	if !eOk {
		ePtr, ePtrOk := expected.(*Uint16)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Uint16 or *Uint16. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Uint16)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Uint16 or *Uint16. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraUint8(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Uint8)
	a, aOk := actual.(Uint8)
	if !eOk {
		ePtr, ePtrOk := expected.(*Uint8)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Uint8 or *Uint8. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Uint8)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Uint8 or *Uint8. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraFloat64(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Float64)
	a, aOk := actual.(Float64)
	if !eOk {
		ePtr, ePtrOk := expected.(*Float64)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Float64 or *Float64. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Float64)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Float64 or *Float64. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if e.IsNull != a.IsNull && !isFloat64Close(e.Value, a.Value, t) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraFloat32(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Float32)
	a, aOk := actual.(Float32)
	if !eOk {
		ePtr, ePtrOk := expected.(*Float32)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Float32 or *Float32. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Float32)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Float32 or *Float32. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			eOra, eOraOk := expected.(Time)
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
			aOra, aOraOk := actual.(Time)
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
	e, eOk := expected.(Time)
	a, aOk := actual.(Time)
	if !eOk {
		ePtr, ePtrOk := expected.(*Time)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Time or *Time. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Time)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Time or *Time. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_string(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(string)
	a, aOk := actual.(string)
	if !eOk {
		ePtr, ePtrOk := expected.(*string)
		if ePtrOk {
			e = *ePtr
		} else {
			eOra, eOraOk := expected.(String)
			if eOraOk {
				e = eOra.Value
			} else {
				t.Fatalf("Unable to cast expected value to string, *string, ora.String. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
			}
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*string)
		if aPtrOk {
			a = *aPtr
		} else {
			aOra, aOraOk := actual.(String)
			if aOraOk {
				a = aOra.Value
			} else {
				t.Fatalf("Unable to cast actual value to string, *string, ora.String. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
			}
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraString(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(String)
	a, aOk := actual.(String)
	if !eOk {
		ePtr, ePtrOk := expected.(*String)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to String or *String. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*String)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to String or *String. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
			eOra, eOraOk := expected.(Bool)
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
			aOra, aOraOk := actual.(Bool)
			if aOraOk {
				a = aOra.Value
			} else {
				t.Fatalf("Unable to cast actual value to bool, *bool, ora.Bool. (%v, %v)", reflect.TypeOf(actual), actual)
			}
		}
	}
	if e != a {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraBool(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Bool)
	a, aOk := actual.(Bool)
	if !eOk {
		ePtr, ePtrOk := expected.(*Bool)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to Bool or *Bool. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*Bool)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to Bool or *Bool. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_bytes(expected driver.Value, actual driver.Value, t *testing.T) {
	e, eOk := expected.([]byte)
	a, aOk := actual.([]byte)
	if !eOk {
		eOra, eOraOk := expected.(Binary)
		if eOraOk {
			e = eOra.Value
		} else {
			t.Fatalf("Unable to cast expected value to []byte or ora.Binary. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	} else if !aOk {
		aOra, aOraOk := actual.(Binary)
		if aOraOk {
			a = aOra.Value
		} else {
			t.Fatalf("Unable to cast actual value to []byte or ora.Binary. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	} else if !areBytesEqual(e, a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_Bytes(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Binary)
	a, aOk := actual.(Binary)
	if !eOk {
		t.Fatalf("Unable to cast expected value to ora.Binary. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to ora.Binary. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
	} else if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraIntervalYM(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(IntervalYM)
	a, aOk := actual.(IntervalYM)
	if !eOk {
		ePtr, ePtrOk := expected.(*IntervalYM)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to IntervalYM or *IntervalYM. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*IntervalYM)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to IntervalYM or *IntervalYM. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraIntervalDS(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(IntervalDS)
	a, aOk := actual.(IntervalDS)
	if !eOk {
		ePtr, ePtrOk := expected.(*IntervalDS)
		if ePtrOk {
			e = *ePtr
		} else {
			t.Fatalf("Unable to cast expected value to IntervalDS or *IntervalDS. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
		}
	}
	if !aOk {
		aPtr, aPtrOk := actual.(*IntervalDS)
		if aPtrOk {
			a = *aPtr
		} else {
			t.Fatalf("Unable to cast actual value to IntervalDS or *IntervalDS. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
		}
	}
	if !e.Equals(a) {
		t.Fatalf("expected(%v), actual(%v)", e, a)
	}
}

func compare_OraBfile(expected interface{}, actual interface{}, t *testing.T) {
	e, eOk := expected.(Bfile)
	a, aOk := actual.(Bfile)
	if !eOk {
		t.Fatalf("Unable to cast expected value to Bfile. (%v, %v)", reflect.TypeOf(expected).Name(), expected)
	} else if !aOk {
		t.Fatalf("Unable to cast actual value to Bfile. (%v, %v)", reflect.TypeOf(actual).Name(), actual)
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
	if len(x) != len(y) {
		return false
	} else {
		for n := 0; n < len(x); n++ {
			if x[n] != y[n] {
				return false
			}
		}
	}
	return true
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

func gen_OraInt64(isNull bool) Int64 {
	return Int64{Value: gen_int64(), IsNull: isNull}
}

func gen_OraInt32(isNull bool) Int32 {
	return Int32{Value: gen_int32(), IsNull: isNull}
}

func gen_OraInt16(isNull bool) Int16 {
	return Int16{Value: gen_int16(), IsNull: isNull}
}

func gen_OraInt8(isNull bool) Int8 {
	return Int8{Value: gen_int8(), IsNull: isNull}
}

func gen_OraUint64(isNull bool) Uint64 {
	return Uint64{Value: gen_uint64(), IsNull: isNull}
}

func gen_OraUint32(isNull bool) Uint32 {
	return Uint32{Value: gen_uint32(), IsNull: isNull}
}

func gen_OraUint16(isNull bool) Uint16 {
	return Uint16{Value: gen_uint16(), IsNull: isNull}
}

func gen_OraUint8(isNull bool) Uint8 {
	return Uint8{Value: gen_uint8(), IsNull: isNull}
}

func gen_OraFloat64(isNull bool) Float64 {
	return Float64{Value: gen_float64(), IsNull: isNull}
}

func gen_OraFloat64Trunc(isNull bool) Float64 {
	return Float64{Value: gen_float64Trunc(), IsNull: isNull}
}

func gen_OraFloat32(isNull bool) Float32 {
	return Float32{Value: gen_float32(), IsNull: isNull}
}

func gen_OraFloat32Trunc(isNull bool) Float32 {
	return Float32{Value: gen_float32Trunc(), IsNull: isNull}
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

func gen_OraInt64Slice(isNull bool) []Int64 {
	expected := make([]Int64, 5)
	expected[0] = Int64{Value: -9}
	expected[1] = Int64{Value: -1}
	expected[2] = Int64{IsNull: isNull}
	expected[3] = Int64{Value: 1}
	expected[4] = Int64{Value: 9}
	return expected
}

func gen_OraInt32Slice(isNull bool) []Int32 {
	expected := make([]Int32, 5)
	expected[0] = Int32{Value: -9}
	expected[1] = Int32{Value: -1}
	expected[2] = Int32{IsNull: isNull}
	expected[3] = Int32{Value: 1}
	expected[4] = Int32{Value: 9}
	return expected
}

func gen_OraInt16Slice(isNull bool) []Int16 {
	expected := make([]Int16, 5)
	expected[0] = Int16{Value: -9}
	expected[1] = Int16{Value: -1}
	expected[2] = Int16{IsNull: isNull}
	expected[3] = Int16{Value: 1}
	expected[4] = Int16{Value: 9}
	return expected
}

func gen_OraInt8Slice(isNull bool) []Int8 {
	expected := make([]Int8, 5)
	expected[0] = Int8{Value: -9}
	expected[1] = Int8{Value: -1}
	expected[2] = Int8{IsNull: isNull}
	expected[3] = Int8{Value: 1}
	expected[4] = Int8{Value: 9}
	return expected
}

func gen_OraUint64Slice(isNull bool) []Uint64 {
	expected := make([]Uint64, 5)
	expected[0] = Uint64{Value: 0}
	expected[1] = Uint64{Value: 3}
	expected[2] = Uint64{IsNull: isNull}
	expected[3] = Uint64{Value: 7}
	expected[4] = Uint64{Value: 9}
	return expected
}

func gen_OraUint32Slice(isNull bool) []Uint32 {
	expected := make([]Uint32, 5)
	expected[0] = Uint32{Value: 0}
	expected[1] = Uint32{Value: 3}
	expected[2] = Uint32{IsNull: isNull}
	expected[3] = Uint32{Value: 7}
	expected[4] = Uint32{Value: 9}
	return expected
}

func gen_OraUint16Slice(isNull bool) []Uint16 {
	expected := make([]Uint16, 5)
	expected[0] = Uint16{Value: 0}
	expected[1] = Uint16{Value: 3}
	expected[2] = Uint16{IsNull: isNull}
	expected[3] = Uint16{Value: 7}
	expected[4] = Uint16{Value: 9}
	return expected
}

func gen_OraUint8Slice(isNull bool) []Uint8 {
	expected := make([]Uint8, 5)
	expected[0] = Uint8{Value: 0}
	expected[1] = Uint8{Value: 3}
	expected[2] = Uint8{IsNull: isNull}
	expected[3] = Uint8{Value: 7}
	expected[4] = Uint8{Value: 9}
	return expected
}

func gen_OraFloat64Slice(isNull bool) []Float64 {
	expected := make([]Float64, 5)
	expected[0] = Float64{Value: -float64(6.28318)}
	expected[1] = Float64{Value: -float64(3.14159)}
	expected[2] = Float64{IsNull: isNull}
	expected[3] = Float64{Value: float64(3.14159)}
	expected[4] = Float64{Value: float64(6.28318)}
	return expected
}

func gen_OraFloat64TruncSlice(isNull bool) []Float64 {
	expected := make([]Float64, 5)
	expected[0] = Float64{Value: -float64(6)}
	expected[1] = Float64{Value: -float64(3)}
	expected[2] = Float64{IsNull: isNull}
	expected[3] = Float64{Value: float64(3)}
	expected[4] = Float64{Value: float64(6)}
	return expected
}

func gen_OraFloat32Slice(isNull bool) []Float32 {
	expected := make([]Float32, 5)
	expected[0] = Float32{Value: -float32(6.28318)}
	expected[1] = Float32{Value: -float32(3.14159)}
	expected[2] = Float32{IsNull: isNull}
	expected[3] = Float32{Value: float32(3.14159)}
	expected[4] = Float32{Value: float32(6.28318)}
	return expected
}

func gen_OraFloat32TruncSlice(isNull bool) []Float32 {
	expected := make([]Float32, 5)
	expected[0] = Float32{Value: -float32(6)}
	expected[1] = Float32{Value: -float32(3)}
	expected[2] = Float32{IsNull: isNull}
	expected[3] = Float32{Value: float32(3)}
	expected[4] = Float32{Value: float32(6)}
	return expected
}

func gen_date() time.Time {
	return time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)
}

func gen_OraDate(isNull bool) Time {
	return Time{Value: gen_date(), IsNull: isNull}
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

func gen_OraDateSlice(isNull bool) []Time {
	expected := make([]Time, 5)
	expected[0] = Time{Value: time.Date(2000, 1, 2, 3, 4, 5, 0, testDbsessiontimezone)}
	expected[1] = Time{Value: time.Date(2001, 2, 3, 4, 5, 6, 0, testDbsessiontimezone)}
	expected[2] = Time{Value: time.Date(2002, 3, 4, 5, 6, 7, 0, testDbsessiontimezone), IsNull: isNull}
	expected[3] = Time{Value: time.Date(2003, 4, 5, 6, 7, 8, 0, testDbsessiontimezone)}
	expected[4] = Time{Value: time.Date(2004, 5, 6, 7, 8, 9, 0, testDbsessiontimezone)}
	return expected
}

func gen_time() time.Time {
	return time.Date(2000, 1, 2, 3, 4, 5, 6, testDbsessiontimezone)
}

func gen_OraTime(isNull bool) Time {
	return Time{Value: gen_time(), IsNull: isNull}
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

func gen_OraTimeSlice(isNull bool) []Time {
	expected := make([]Time, 5)
	expected[0] = Time{Value: time.Date(2000, 1, 2, 3, 4, 5, 6, testDbsessiontimezone)}
	expected[1] = Time{Value: time.Date(2001, 2, 3, 4, 5, 6, 7, testDbsessiontimezone)}
	expected[2] = Time{Value: time.Date(2002, 3, 4, 5, 6, 7, 8, testDbsessiontimezone), IsNull: isNull}
	expected[3] = Time{Value: time.Date(2003, 4, 5, 6, 7, 8, 9, testDbsessiontimezone)}
	expected[4] = Time{Value: time.Date(2004, 5, 6, 7, 8, 9, 10, testDbsessiontimezone)}
	return expected
}

func gen_string() string {
	return "Go is expressive, concise, clean, and efficient."
}

func gen_OraString(isNull bool) String {
	return String{Value: gen_string(), IsNull: isNull}
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
	expected := make([]String, 5)
	expected[0] = String{Value: "Go is expressive, concise, clean, and efficient."}
	expected[1] = String{Value: "Its concurrency mechanisms make it easy to"}
	expected[2] = String{Value: "Go compiles quickly to machine code yet has", IsNull: isNull}
	expected[3] = String{Value: "It's a fast, statically typed, compiled"}
	expected[4] = String{Value: "One of Go's key design goals is code"}
	return expected
}

func gen_boolFalse() bool {
	return false
}
func gen_boolTrue() bool {
	return true
}

func gen_OraBoolFalse(isNull bool) Bool {
	return Bool{Value: gen_boolFalse(), IsNull: isNull}
}

func gen_OraBoolTrue(isNull bool) Bool {
	return Bool{Value: gen_boolTrue(), IsNull: isNull}
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
	expected := make([]Bool, 5)
	expected[0] = Bool{Value: true}
	expected[1] = Bool{Value: false}
	expected[2] = Bool{Value: false, IsNull: isNull}
	expected[3] = Bool{Value: false}
	expected[4] = Bool{Value: true}
	return expected
}

func gen_bytes(length int) []byte {
	values := make([]byte, length)
	rand.Read(values)
	return values
}

func gen_OraBytes(length int, isNull bool) Binary {
	return Binary{Value: gen_bytes(length), IsNull: isNull}
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

func gen_OraBytesSlice(length int, isNull bool) []Binary {
	values := make([]Binary, 5)
	values[0] = Binary{Value: gen_bytes(2000)}
	values[1] = Binary{Value: gen_bytes(2000)}
	values[2] = Binary{Value: gen_bytes(2000), IsNull: isNull}
	values[3] = Binary{Value: gen_bytes(2000)}
	values[4] = Binary{Value: gen_bytes(2000)}

	return values
}

func gen_OraIntervalYMSlice(isNull bool) []IntervalYM {
	expected := make([]IntervalYM, 5)
	expected[0] = IntervalYM{Year: 1, Month: 1}
	expected[1] = IntervalYM{Year: 99, Month: 9}
	expected[2] = IntervalYM{IsNull: isNull}
	expected[3] = IntervalYM{Year: -1, Month: -1}
	expected[4] = IntervalYM{Year: -99, Month: -9}
	return expected
}

func gen_OraIntervalDSSlice(isNull bool) []IntervalDS {
	expected := make([]IntervalDS, 5)
	expected[0] = IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	expected[1] = IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}
	expected[2] = IntervalDS{IsNull: isNull}
	expected[3] = IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}
	expected[4] = IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}
	return expected
}

func gen_OraBfile(isNull bool) interface{} {
	return Bfile{IsNull: isNull, DirectoryAlias: "TEMP_DIR", Filename: "test.txt"}
}

func gen_OraBfileEmpty(isNull bool) interface{} {
	return Bfile{IsNull: isNull, DirectoryAlias: "", Filename: ""}
}

func gen_OraBfileEmptyDir(isNull bool) interface{} {
	return Bfile{IsNull: isNull, DirectoryAlias: "", Filename: "test.txt"}
}

func gen_OraBfileEmptyFilename(isNull bool) interface{} {
	return Bfile{IsNull: isNull, DirectoryAlias: "TEMP_DIR", Filename: ""}
}
