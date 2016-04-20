//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"testing"
)

////////////////////////////////////////////////////////////////////////////////
// date
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_d_date_session(t *testing.T) {
	testBindDefine(gen_date(), date, t, nil)
}

func TestBindPtr_time_d_date_session(t *testing.T) {
	testBindPtr(gen_date(), date, t)
}

func TestBindDefine_OraTime_d_date_session(t *testing.T) {
	testBindDefine(gen_OraDate(false), date, t, nil)
}

func TestBindSlice_time_d_date_session(t *testing.T) {
	testBindDefine(gen_dateSlice(), date, t, nil)
}

func TestBindSlice_OraTime_d_date_session(t *testing.T) {
	testBindDefine(gen_OraDateSlice(false), date, t, nil)
}

func TestMultiDefine_date_session(t *testing.T) {
	testMultiDefine(gen_date(), date, t)
}

func TestWorkload_date_session(t *testing.T) {
	testWorkload(date, t)
}

////////////////////////////////////////////////////////////////////////////////
// dateNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_d_dateNull_session(t *testing.T) {
	testBindDefine(gen_date(), dateNull, t, nil)
}

func TestBindPtr_time_d_dateNull_session(t *testing.T) {
	testBindPtr(gen_date(), dateNull, t)
}

func TestBindDefine_OraTime_d_dateNull_session(t *testing.T) {
	testBindDefine(gen_OraDate(true), dateNull, t, nil)
}

func TestBindSlice_time_d_dateNull_session(t *testing.T) {
	testBindDefine(gen_dateSlice(), dateNull, t, nil)
}

func TestBindSlice_OraTime_d_dateNull_session(t *testing.T) {
	testBindDefine(gen_OraDateSlice(true), dateNull, t, nil)
}

func TestMultiDefine_dateNull_session(t *testing.T) {
	testMultiDefine(gen_date(), dateNull, t)
}

func TestWorkload_dateNull_session(t *testing.T) {
	testWorkload(dateNull, t)
}

func TestBindDefine_dateNull_nil_session(t *testing.T) {
	testBindDefine(nil, dateNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// timestampP9
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampP9_session(t *testing.T) {
	testBindDefine(gen_time(), timestampP9, t, nil)
}

func TestBindPtr_time_timestampP9_session(t *testing.T) {
	testBindPtr(gen_time(), timestampP9, t)
}

func TestBindDefine_OraTime_timestampP9_session(t *testing.T) {
	testBindDefine(gen_OraTime(false), timestampP9, t, nil)
}

func TestBindSlice_time_timestampP9_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampP9, t, nil)
}

func TestBindSlice_OraTime_timestampP9_session(t *testing.T) {
	testBindDefine(gen_OraTimeSlice(false), timestampP9, t, nil)
}

func TestMultiDefine_timestampP9_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampP9, t)
}

func TestWorkload_timestampP9_session(t *testing.T) {
	testWorkload(timestampP9, t)
}

////////////////////////////////////////////////////////////////////////////////
// timestampP9Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampP9Null_session(t *testing.T) {
	testBindDefine(gen_time(), timestampP9Null, t, nil)
}

func TestBindPtr_time_timestampP9Null_session(t *testing.T) {
	testBindPtr(gen_time(), timestampP9Null, t)
}

func TestBindDefine_OraTime_timestampP9Null_session(t *testing.T) {
	testBindDefine(gen_OraTime(true), timestampP9Null, t, nil)
}

func TestBindSlice_time_timestampP9Null_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampP9Null, t, nil)
}

func TestBindSlice_OraTime_timestampP9Null_session(t *testing.T) {
	testBindDefine(gen_OraTimeSlice(true), timestampP9Null, t, nil)
}

func TestMultiDefine_timestampP9Null_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampP9Null, t)
}

func TestWorkload_timestampP9Null_session(t *testing.T) {
	testWorkload(timestampP9Null, t)
}

func TestBindDefine_timestampP9Null_nil_session(t *testing.T) {
	testBindDefine(nil, timestampP9Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// timestampTzP9
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampTzP9_session(t *testing.T) {
	testBindDefine(gen_time(), timestampTzP9, t, nil)
}

func TestBindPtr_time_timestampTzP9_session(t *testing.T) {
	testBindPtr(gen_time(), timestampTzP9, t)
}

func TestBindDefine_OraTime_timestampTzP9_session(t *testing.T) {
	testBindDefine(gen_OraTime(false), timestampTzP9, t, nil)
}

func TestBindSlice_time_timestampTzP9_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampTzP9, t, nil)
}

func TestBindSlice_OraTime_timestampTzP9_session(t *testing.T) {
	testBindDefine(gen_OraTimeSlice(false), timestampTzP9, t, nil)
}

func TestMultiDefine_timestampTzP9_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampTzP9, t)
}

func TestWorkload_timestampTzP9_session(t *testing.T) {
	testWorkload(timestampTzP9, t)
}

////////////////////////////////////////////////////////////////////////////////
// timestampTzP9Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampTzP9Null_session(t *testing.T) {
	testBindDefine(gen_time(), timestampTzP9Null, t, nil)
}

func TestBindPtr_time_timestampTzP9Null_session(t *testing.T) {
	testBindPtr(gen_time(), timestampTzP9Null, t)
}

func TestBindDefine_OraTime_timestampTzP9Null_session(t *testing.T) {
	testBindDefine(gen_OraTime(true), timestampTzP9Null, t, nil)
}

func TestBindSlice_time_timestampTzP9Null_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampTzP9Null, t, nil)
}

func TestBindSlice_OraTime_timestampTzP9Null_session(t *testing.T) {
	testBindDefine(gen_OraTimeSlice(true), timestampTzP9Null, t, nil)
}

func TestMultiDefine_timestampTzP9Null_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampTzP9Null, t)
}

func TestWorkload_timestampTzP9Null_session(t *testing.T) {
	testWorkload(timestampTzP9Null, t)
}

func TestBindDefine_timestampTzP9Null_nil_session(t *testing.T) {
	testBindDefine(nil, timestampTzP9Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// timestampLtzP9
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampLtzP9_session(t *testing.T) {
	testBindDefine(gen_time(), timestampLtzP9, t, nil)
}

func TestBindPtr_time_timestampLtzP9_session(t *testing.T) {
	testBindPtr(gen_time(), timestampLtzP9, t)
}

func TestBindDefine_OraTime_timestampLtzP9_session(t *testing.T) {
	testBindDefine(gen_OraTime(false), timestampLtzP9, t, nil)
}

func TestBindSlice_time_timestampLtzP9_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampLtzP9, t, nil)
}

func TestBindSlice_OraTime_timestampLtzP9_session(t *testing.T) {
	testBindDefine(gen_OraTimeSlice(false), timestampLtzP9, t, nil)
}

func TestMultiDefine_timestampLtzP9_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampLtzP9, t)
}

func TestWorkload_timestampLtzP9_session(t *testing.T) {
	testWorkload(timestampLtzP9, t)
}

////////////////////////////////////////////////////////////////////////////////
// timestampLtzP9Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_time_timestampLtzP9Null_session(t *testing.T) {
	testBindDefine(gen_time(), timestampLtzP9Null, t, nil)
}

func TestBindPtr_time_timestampLtzP9Null_session(t *testing.T) {
	testBindPtr(gen_time(), timestampLtzP9Null, t)
}

func TestBindDefine_OraTime_timestampLtzP9Null_session(t *testing.T) {
	testBindDefine(gen_OraTime(true), timestampLtzP9Null, t, nil)
}

func TestBindSlice_time_timestampLtzP9Null_session(t *testing.T) {
	testBindDefine(gen_timeSlice(), timestampLtzP9Null, t, nil)
}

func TestBindSlice_OraTime_timestampLtzP9Null_session(t *testing.T) {
	enableLogging(t)
	testBindDefine(gen_OraTimeSlice(true), timestampLtzP9Null, t, nil)
}

func TestMultiDefine_timestampLtzP9Null_session(t *testing.T) {
	testMultiDefine(gen_time(), timestampLtzP9Null, t)
}

func TestWorkload_timestampLtzP9Null_session(t *testing.T) {
	testWorkload(timestampLtzP9Null, t)
}

func TestBindDefine_timestampLtzP9Null_nil_session(t *testing.T) {
	testBindDefine(nil, timestampLtzP9Null, t, nil)
}
