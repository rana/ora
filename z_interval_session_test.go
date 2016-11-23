//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/rana/ora.v4"
)

//// interval
//intervalYM     oracleColumnType = "interval year to month not null"
//intervalYMNull oracleColumnType = "interval year to month null"
//intervalDS     oracleColumnType = "interval day to second not null"
//intervalDSNull oracleColumnType = "interval day to second null"

////////////////////////////////////////////////////////////////////////////////
// intervalYM
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalYM(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, interval := range []ora.IntervalYM{
		ora.IntervalYM{Year: 1, Month: 1},
		ora.IntervalYM{Year: 99, Month: 9},
		ora.IntervalYM{Year: -1, Month: -1},
		ora.IntervalYM{Year: -99, Month: -9},
	} {
		for _, ctName := range []string{"intervalYM", "intervalYMNull"} {
			t.Run(fmt.Sprintf("%s_%s", ctName, interval), func(t *testing.T) {
				t.Parallel()
				testBindDefine(interval, _T_colType[ctName], t, sc)
			})
		}
	}

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		testBindDefine(ora.IntervalYM{IsNull: true}, intervalYMNull, t, sc)
	})
}

func TestBindDefine_OraIntervalYMSlice(t *testing.T) {
	sc := ora.NewStmtCfg()
	t.Run("notnull", func(t *testing.T) {
		t.Parallel()
		testBindDefine(gen_OraIntervalYMSlice(false), intervalYM, t, sc)
	})
	t.Run("null", func(t *testing.T) {
		t.Parallel()
		testBindDefine(gen_OraIntervalYMSlice(true), intervalYMNull, t, sc)
	})
}

////////////////////////////////////////////////////////////////////////////////
// intervalDS
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalDS(t *testing.T) {
	sc := ora.NewStmtCfg()
	for _, interval := range []ora.IntervalDS{
		ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789},
		ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789},
		ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789},
		ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789},
	} {
		for _, ctName := range []string{"intervalDS", "intervalDSNull"} {
			ctName := ctName
			t.Run(fmt.Sprintf("%s_%s", interval, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(interval, _T_colType[ctName], t, sc)
			})
		}
	}
	t.Run("intervalDSNull_null", func(t *testing.T) {
		t.Parallel()
		testBindDefine(ora.IntervalDS{IsNull: true}, intervalDSNull, t, sc)
	})
}

func TestBindDefine_OraIntervalDSSlice(t *testing.T) {
	sc := ora.NewStmtCfg()
	t.Run("notnull", func(t *testing.T) {
		t.Parallel()
		testBindDefine(gen_OraIntervalDSSlice(false), intervalDS, t, sc)
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		testBindDefine(gen_OraIntervalDSSlice(true), intervalDSNull, t, sc)
	})
}

////////////////////////////////////////////////////////////////////////////////
// ShiftTime
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_OraIntervalYM_ShiftTime_session(t *testing.T) {
	t.Parallel()
	interval := ora.IntervalYM{Year: 1, Month: 1}
	actual := interval.ShiftTime(time.Date(2000, 1, 0, 0, 0, 0, 0, time.Local))
	expected := time.Date(2001, 2, 0, 0, 0, 0, 0, time.Local)
	if !expected.Equal(actual) {
		t.Fatalf("expected(%v), actual(%v)", expected, actual)
	}
}

func TestBindDefine_OraIntervalDS_ShiftTime_session(t *testing.T) {
	t.Parallel()
	interval := ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	actual := interval.ShiftTime(time.Date(2000, time.Month(1), 1, 0, 0, 0, 0, time.Local))
	expected := time.Date(2000, time.Month(1), 2, 1, 1, 1, 123456789, time.Local)
	if !expected.Equal(actual) {
		t.Fatalf("expected(%v), actual(%v)", expected, actual)
	}
}
