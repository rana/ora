// Copyright 2016 Tamás Gulácsi, Valentin Kuznetsov. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/rana/ora.v4"
)

const (
	geoTableName     = "test_geoloc"
	geoTableRowCount = 1000
)

var geoTableOnce sync.Once

func createGeoTable() error {
	tableName := geoTableName
	var cnt int64
	if err := testDb.QueryRow("SELECT COUNT(0) FROM " + tableName).Scan(&cnt); err == nil && cnt == geoTableRowCount {
		return nil
	}
	testDb.Exec("ALTER SESSION SET NLS_NUMERIC_CHARACTERS = '.,'")
	testDb.Exec("DROP TABLE " + tableName)
	if _, err := testDb.Exec(`CREATE TABLE ` + tableName + ` (
		id NUMBER(3) NOT NULL,
	"RECORD_ID" NUMBER(*,0) NOT NULL ENABLE,
	"PERSON_ID" NUMBER(*,0),
	"PERSON_ACCOUNT_ID" NUMBER(*,0),
	"ORGANIZATION_ID" NUMBER(*,0),
	"ORGANIZATION_MEMBERSHIP_ID" NVARCHAR2(45),
	"LOCATION" NVARCHAR2(2000) NOT NULL ENABLE,
	"DEVICE_ID" NVARCHAR2(45),
	"DEVICE_REGISTRATION_ID" NVARCHAR2(500),
	"DEVICE_NAME" NVARCHAR2(45),
	"DEVICE_TYPE" NVARCHAR2(45),
	"DEVICE_OS_NAME" NVARCHAR2(45),
	"DEVICE_TOKEN" NVARCHAR2(45),
	"DEVICE_OTHER_DETAILS" NVARCHAR2(100)
	)`,
	); err != nil {
		return err
	}
	testData := [][]string{
		{"1", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5518407 104.0685472)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"2", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5520498 104.0686355)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"3", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5517747 104.0684895)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"4", "8.64522675633357E16", "8.64522734353613E16", "", "1220457", "POINT(30.55187 104.06856)", "3A9D1838-3B2D-4119-9E07-77C6CDAC53C5", "noUwBnWojdY:APA91bE8aGLEECS9_Q1EKrp8i2B36H1X8GwIj3v58KUcuXglhf0rXJb8Ez5meQ6D5MgTAQghYEe3s9vOntU3pYPQoc6ASNw3QzhzQevAqlMQC2ukUMNyLD8Rve-IA1-6lttsCXYsYIKh", "User3’s iPhone", "iPhone", "iPhone OS", "", "DeviceID:3A9D1838-3B2D-4119-9E07-77C6CDAC53C5, SystemVersion:8.4, LocalizedModel:iPhone"},
		{"5", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5517458 104.0685809)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"6", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.551802 104.0685301)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"7", "8.64522675633357E16", "8.64522734353613E16", "", "1220457", "POINT(30.55187 104.06856)", "3A9D1838-3B2D-4119-9E07-77C6CDAC53C5", "noUwBnWojdY:APA91bE8aGLEECS9_Q1EKrp8i2B36H1X8GwIj3v58KUcuXglhf0rXJb8Ez5meQ6D5MgTAQghYEe3s9vOnt,3pYPQoc6ASNw3QzhzQevAqlMQC2ukUMNyLD8Rve-IA1-6lttsCXYsYIKh", "User3’s iPhone", "iPhone", "iPhone OS", "", "DeviceID:3A9D1838-3B2D-4119-9E07-77C6CDAC53C5, SystemVersion:8.4, LocalizedModel:iPhone"},
		{"8", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.551952 104.0685893)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"9", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5518439 104.0685473)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
		{"10", "8.37064876162908E16", "8.37064898728264E16", "12", "6506", "POINT(30.5518439 104.0685473)", "a71223186cef459b", "", "Samsung SCH-I545", "Mobile", "Android 4.4.2", "", ""},
	}
	dataI := make([][]interface{}, len(testData))
	for i, data := range testData {
		dataI[i] = make([]interface{}, 1, len(data)+1)
		for _, d := range data {
			dataI[i] = append(dataI[i], d)
		}
	}

	stmt, err := testDb.Prepare("INSERT INTO " + tableName + `
  (ID,RECORD_ID,PERSON_ID,PERSON_ACCOUNT_ID,ORGANIZATION_ID,ORGANIZATION_MEMBERSHIP_ID,
   LOCATION,DEVICE_ID,DEVICE_REGISTRATION_ID,DEVICE_NAME,DEVICE_TYPE,
   DEVICE_OS_NAME,DEVICE_TOKEN,DEVICE_OTHER_DETAILS)
   VALUES (:1,:2,:3,:4,:5,
           :6,:7,:8,:9,:10,
		   :11,:12, :13, :14)`)
	if err != nil {
		return err
	}
Loop:
	for rn := 0; rn < geoTableRowCount; {
		for i, data := range dataI {
			data[0] = strconv.Itoa(rn)
			if _, err := stmt.Exec(data...); err != nil {
				return fmt.Errorf("%d. %v\n%q", i, err, data)
			}
			rn++
			if rn == geoTableRowCount {
				break Loop
			}
		}
	}
	return nil
}

func TestSelectOrder(t *testing.T) {
	t.Parallel()
	const limit = 1013
	var cnt int64
	tbl := "user_objects"
	start := time.Now()
	if err := testDb.QueryRow("SELECT count(0) FROM " + tbl).Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s rowcount=%d (%s)", tbl, cnt, time.Since(start))
	if cnt == 0 {
		cnt = 10
		tbl = "(SELECT 1 FROM DUAL " + strings.Repeat("\nUNION ALL SELECT 1 FROM DUAL ", int(cnt)-1) + ")"
	}
	qry := "SELECT ROWNUM FROM " + tbl
	for i := cnt; i < limit; i *= cnt {
		qry += ", " + tbl
	}
	t.Logf("qry=%s", qry)
	rows, err := testDb.Query(qry)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var rn int
		if err = rows.Scan(&rn); err != nil {
			t.Fatal(err)
		}
		i++
		if rn != i {
			t.Errorf("got %d, wanted %d.", rn, i)
		}
		if i > limit {
			break
		}
	}
	for rows.Next() {
	}
}

// go test -c && ./ora.v4.test -test.run=^$ -test.bench=Date -test.cpuprofile=/tmp/cpu.prof && go tool pprof ora.v4.test /tmp/cpu.prof
func BenchmarkSelectDate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; {
		rows, err := testDb.Query("SELECT CAST(TO_DATE('2006-01-02 15:04:05', 'YYYY-MM-DD HH24:MI:SS') AS DATE) dt FROM user_objects, (select 1 from dual)")
		if err != nil {
			b.Fatal(err)
		}
		for rows.Next() && i < b.N {
			var dt time.Time
			if err = rows.Scan(&dt); err != nil {
				rows.Close()
				b.Fatal(err)
			}
			i++
		}
		rows.Close()
	}
}

func BenchmarkSelect(b *testing.B) {
	geoTableOnce.Do(func() {
		if err := createGeoTable(); err != nil {
			b.Fatal(err)
		}
	})
	b.ResetTimer()
	for i := 0; i < b.N; {
		rows, err := testDb.Query("SELECT record_id FROM " + geoTableName)
		if err != nil {
			b.Fatal(err)
		}
		for rows.Next() && i < b.N {
			var id int
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				b.Fatal(err)
			}
			i++
		}
		rows.Close()
	}
}

func BenchmarkPrepare(b *testing.B) {
	rows, err := testDb.Query("SELECT A.object_name from user_objects A")
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
	rows.Close()
}

func BenchmarkIter(b *testing.B) {
	b.StopTimer()
	rows, err := testDb.Query("SELECT A.object_name from user_objects A")
	if err != nil {
		b.Fatal(err)
	}
	defer rows.Close()
	b.StartTimer()
	i := 0
	for rows.Next() && i < b.N {
		i++
	}
	b.SetBytes(int64(i))
}

// BenchmarkMemory usage for querying rows.
//
// go test -c && ./ora.v4.test -test.run=^$ -test.bench=Memory -test.memprofilerate=1 -test.memprofile=/tmp/mem.prof && go tool pprof --alloc_space ora.v4.test /tmp/mem.prof
func TestMemoryNumString(t *testing.T) {
	n := 1000
	//enableLogging(t)
	benchMem(t, n, 1870, `SELECT
		TO_NUMBER('123456789012345678') bn01
		, TO_NUMBER('223456789012345678') bn02
		, TO_NUMBER('323456789012345678') bn03
		, TO_NUMBER('423456789012345678') bn04
		, TO_NUMBER('523456789012345678') bn05
		, TO_NUMBER('623456789012345678') bn06
		, TO_NUMBER('723456789012345678') bn07
		, TO_NUMBER('823456789012345678') bn08
		, TO_NUMBER('923456789012345678') bn09
		, TO_NUMBER('023456789012345678') bn10
	FROM user_OBJECTS B, user_objects A, (SELECT 1 FROM DUAL) AA
	WHERE ROWNUM <= :1`)
}
func TestMemoryNumStringI64(t *testing.T) {
	cfg := ora.Cfg()
	defer ora.SetCfg(cfg)
	ora.SetCfg(cfg.
		SetNumberBigInt(ora.I64).
		SetNumberBigFloat(ora.I64))
	n := 1000
	benchMem(t, n, 1801, `SELECT
		TO_NUMBER('123456789012345678') bn01
		, TO_NUMBER('223456789012345678') bn02
		, TO_NUMBER('323456789012345678') bn03
		, TO_NUMBER('423456789012345678') bn04
		, TO_NUMBER('523456789012345678') bn05
		, TO_NUMBER('623456789012345678') bn06
		, TO_NUMBER('723456789012345678') bn07
		, TO_NUMBER('823456789012345678') bn08
		, TO_NUMBER('923456789012345678') bn09
		, TO_NUMBER('023456789012345678') bn10
	FROM user_OBJECTS B, user_objects A, (select 1 from dual)
	WHERE ROWNUM <= :1`)
}

func TestMemoryString(t *testing.T) {
	n := 1000
	benchMem(t, n, 2014, `SELECT
		'123456789012345678' bs01
		, '223456789012345678' bs02
		, '323456789012345678' bs03
		, '423456789012345678' bs04
		, '523456789012345678' bs05
		, '623456789012345678' bs06
		, '723456789012345678' bs07
		, '823456789012345678' bs08
		, '923456789012345678' bs09
		, '023456789012345678' bs10
	FROM user_OBJECTS B, user_objects A, (select 1 from dual)
	WHERE ROWNUM <= :1`)
}

func benchMem(tb testing.TB, n int, maxBytesPerRun uint64, qry string) {
	columns, err := ora.DescribeQuery(testDb, qry)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Logf("columns: %#v", columns)

	cols := make([]string, len(columns))
	for i, c := range columns {
		cols[i] = c.Name
	}
	args := []interface{}{int64(n)}

	type Record map[string]interface{}

	execute := func(qry string, cols []string, args ...interface{}) []Record {
		var out []Record

		rows, err := testDb.Query(qry, args...)
		if err != nil {
			tb.Fatalf("ERROR: DB.Query, query='%s' args='%v' error=%v", qry, args, err)
		}
		defer rows.Close()

		count := len(cols)
		vals := make([]interface{}, count)
		valPtrs := make([]interface{}, count)
		for i, _ := range cols {
			valPtrs[i] = &vals[i]
		}
		// loop over rows
		for rows.Next() {
			err := rows.Scan(valPtrs...)
			//        err := rows.Scan(vals...)
			if err != nil {
				tb.Fatalf("ERROR: rows.Scan, dest='%v', error=%v", vals, err)
			}
			rec := make(Record)
			length := 0
			for i, _ := range cols {
				rec[cols[i]] = vals[i]
				if s, ok := vals[i].(string); ok {
					length += len(s)
				}
			}
			out = append(out, rec)
			if len(out) == 1 {
				tb.Logf("One record's length: %d", length)
			}
		}
		if err = rows.Err(); err != nil {
			tb.Fatal(err)
		}
		return out
	}

	var ostat, nstat runtime.MemStats
	runtime.ReadMemStats(&ostat)
	results := execute(qry, cols, args...)
	runtime.ReadMemStats(&nstat)
	d := nstat.TotalAlloc - ostat.TotalAlloc
	tb.Logf("nres=%d, allocated %d bytes\n", len(results), d)
	maxBytes := maxBytesPerRun * uint64(n)
	if maxBytes > 0 && d > maxBytes {
		tb.Errorf("nres=%d, allocated %d bytes (max: %d)", len(results), d, maxBytes)
	}
	ostat = nstat
}
