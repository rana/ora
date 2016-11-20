//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"fmt"
	"strings"
	"testing"

	ora "gopkg.in/rana/ora.v3"
)

var _T_timeCols = []string{
	"date", "dateNull",
	"time", "timeNull",
	"timestampP9", "timestampP9Null",
	"timestampTzP9", "timestampTzP9Null",
	"timestampLtzP9", "timestampLtzP9Null",
}

func TestBindDefine_time(t *testing.T) {
	sc := ora.NewStmtCfg()
	for valName, gen := range map[string](func() interface{}){
		"date":             func() interface{} { return gen_date() },
		"OraDate":          func() interface{} { return gen_OraDate(false) },
		"OraDateNull":      func() interface{} { return gen_OraDate(true) },
		"dateSlice":        func() interface{} { return gen_dateSlice() },
		"OraDateSlice":     func() interface{} { return gen_OraDateSlice(false) },
		"OraDateSliceNull": func() interface{} { return gen_OraDateSlice(true) },

		"time":             func() interface{} { return gen_time() },
		"OraTime":          func() interface{} { return gen_OraTime(false) },
		"OraTimeNull":      func() interface{} { return gen_OraTime(true) },
		"timeSlice":        func() interface{} { return gen_timeSlice() },
		"OraTimeSlice":     func() interface{} { return gen_OraTimeSlice(false) },
		"OraTimeSliceNull": func() interface{} { return gen_OraTimeSlice(true) },
	} {
		valName := valName
		for _, ctName := range _T_timeCols {
			if strings.HasSuffix(valName, "Null") && !strings.HasSuffix(ctName, "Null") {
				continue
			}
			if strings.HasPrefix(ctName, "time") && !strings.Contains(valName, "ime") {
				continue
			}
			if strings.HasPrefix(ctName, "date") && !strings.Contains(valName, "ate") {
				continue
			}
			t.Run(fmt.Sprintf("%s_%s", valName, ctName), func(t *testing.T) {
				t.Parallel()
				testBindDefine(gen(), _T_colType[ctName], t, sc)
				testBindPtr(gen(), _T_colType[ctName], t)
			})
		}
	}
}

func TestMultiDefine_date_session(t *testing.T) {
	for _, ctName := range []string{"date"} {
		t.Run(ctName, func(t *testing.T) {
			testMultiDefine(gen_date(), _T_colType[ctName], t)
		})
	}
}

func TestWorkload_date_session(t *testing.T) {
	for _, ctName := range _T_timeCols {
		t.Run(ctName, func(t *testing.T) {
			testWorkload(_T_colType[ctName], t)
		})
	}
}
