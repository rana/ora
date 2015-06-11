//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"ora"
	"testing"
	"time"
)

//// interval
//intervalYM     oracleColumnType = "interval year to month not null"
//intervalYMNull oracleColumnType = "interval year to month null"
//intervalDS     oracleColumnType = "interval day to second not null"
//intervalDSNull oracleColumnType = "interval day to second null"

////////////////////////////////////////////////////////////////////////////////
// intervalYM
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalYM_intervalYM_Positive1_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: 1, Month: 1}, intervalYM, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYM_Positive99_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: 99, Month: 9}, intervalYM, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYM_Negative1_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: -1, Month: -1}, intervalYM, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYM_Negative99_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: -99, Month: -9}, intervalYM, t, nil)
}

func TestBindDefine_OraIntervalYMSlice_intervalYM_session(t *testing.T) {
	testBindDefine(gen_OraIntervalYMSlice(false), intervalYM, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// intervalYMNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalYM_intervalYMNull_Positive1_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: 1, Month: 1}, intervalYMNull, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYMNull_Positive99_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: 99, Month: 9}, intervalYMNull, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYMNull_Negative1_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: -1, Month: -1}, intervalYMNull, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYMNull_Negative99_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{Year: -99, Month: -9}, intervalYMNull, t, nil)
}

func TestBindDefine_OraIntervalYM_intervalYMNull_null_session(t *testing.T) {
	testBindDefine(ora.IntervalYM{IsNull: true}, intervalYMNull, t, nil)
}

func TestBindDefine_OraIntervalYMSlice_intervalYMNull_session(t *testing.T) {
	testBindDefine(gen_OraIntervalYMSlice(true), intervalYMNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// intervalDS
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalDS_intervalDS_Positive1_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}, intervalDS, t, nil)
}

func TestBindDefine_OraIntervalDS_intervalDS_Positive59_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}, intervalDS, t, nil)
}

func TestBindDefine_OraIntervalDS_intervalDS_Negative1_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}, intervalDS, t, nil)
}

func TestBindDefine_OraIntervalDS_intervalDS_Negative59_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}, intervalDS, t, nil)
}

func TestBindDefine_OraIntervalDSSlice_intervalDS_session(t *testing.T) {
	testBindDefine(gen_OraIntervalDSSlice(false), intervalDS, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// intervalDSNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalDSNull_intervalDSNull_Positive1_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}, intervalDSNull, t, nil)
}

func TestBindDefine_OraIntervalDSNull_intervalDSNull_Positive59_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}, intervalDSNull, t, nil)
}

func TestBindDefine_OraIntervalDSNull_intervalDSNull_Negative1_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}, intervalDSNull, t, nil)
}

func TestBindDefine_OraIntervalDSNull_intervalDSNull_Negative59_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}, intervalDSNull, t, nil)
}

func TestBindDefine_OraIntervalDS_intervalDSNull_null_session(t *testing.T) {
	testBindDefine(ora.IntervalDS{IsNull: true}, intervalDSNull, t, nil)
}

func TestBindDefine_OraIntervalDSSlice_intervalDSNull_session(t *testing.T) {
	testBindDefine(gen_OraIntervalDSSlice(true), intervalDSNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// ShiftTime
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalYM_ShiftTime_session(t *testing.T) {
	interval := ora.IntervalYM{Year: 1, Month: 1}
	actual := interval.ShiftTime(time.Date(2000, 1, 0, 0, 0, 0, 0, time.Local))
	expected := time.Date(2001, 2, 0, 0, 0, 0, 0, time.Local)
	if !expected.Equal(actual) {
		t.Fatalf("expected(%v), actual(%v)", expected, actual)
	}
}

func TestBindDefine_OraIntervalDS_ShiftTime_session(t *testing.T) {
	interval := ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	actual := interval.ShiftTime(time.Date(2000, time.Month(1), 1, 0, 0, 0, 0, time.Local))
	expected := time.Date(2000, time.Month(1), 2, 1, 1, 1, 123456789, time.Local)
	if !expected.Equal(actual) {
		t.Fatalf("expected(%v), actual(%v)", expected, actual)
	}
}
