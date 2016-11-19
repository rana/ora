//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"testing"
)

////////////////////////////////////////////////////////////////////////////////
// charB48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_charB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), charB48, t, nil)
}

func TestBindPtr_string_charB48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), charB48, t)
}

func TestBindDefine_OraString_charB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(false), charB48, t, nil)
}

func TestBindSlice_string_charB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), charB48, t, nil)
}

func TestBindSlice_OraString_charB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(false), charB48, t, nil)
}

func TestMultiDefine_charB48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), charB48, t)
}

func TestWorkload_charB48_session(t *testing.T) {
	t.Parallel()
	testWorkload(charB48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// charB48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_charB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), charB48Null, t, nil)
}

func TestBindPtr_string_charB48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), charB48Null, t)
}

func TestBindDefine_OraString_charB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(true), charB48Null, t, nil)
}

func TestBindSlice_string_charB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), charB48Null, t, nil)
}

func TestBindSlice_OraString_charB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(true), charB48Null, t, nil)
}

func TestMultiDefine_charB48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), charB48Null, t)
}

func TestWorkload_charB48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(charB48Null, t)
}

func TestBindDefine_charB48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, charB48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// charC48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_charC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), charC48, t, nil)
}

func TestBindPtr_string_charC48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), charC48, t)
}

func TestBindDefine_OraString_charC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(false), charC48, t, nil)
}

func TestBindSlice_string_charC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), charC48, t, nil)
}

func TestBindSlice_OraString_charC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(false), charC48, t, nil)
}

func TestMultiDefine_charC48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), charC48, t)
}

func TestWorkload_charC48_session(t *testing.T) {
	t.Parallel()
	testWorkload(charC48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// charC48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_charC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), charC48Null, t, nil)
}

func TestBindPtr_string_charC48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), charC48Null, t)
}

func TestBindDefine_OraString_charC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(true), charC48Null, t, nil)
}

func TestBindSlice_string_charC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), charC48Null, t, nil)
}

func TestBindSlice_OraString_charC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(true), charC48Null, t, nil)
}

func TestMultiDefine_charC48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), charC48Null, t)
}

func TestWorkload_charC48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(charC48Null, t)
}

func TestBindDefine_charC48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, charC48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// nchar48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nchar48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), nchar48, t, nil)
}

func TestBindPtr_string_nchar48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), nchar48, t)
}

func TestBindDefine_OraString_nchar48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(false), nchar48, t, nil)
}

func TestBindSlice_string_nchar48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), nchar48, t, nil)
}

func TestBindSlice_OraString_nchar48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(false), nchar48, t, nil)
}

func TestMultiDefine_nchar48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), nchar48, t)
}

func TestWorkload_nchar48_session(t *testing.T) {
	t.Parallel()
	testWorkload(nchar48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// nchar48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string48(), nchar48Null, t, nil)
}

func TestBindPtr_string_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string48(), nchar48Null, t)
}

func TestBindDefine_OraString_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString48(true), nchar48Null, t, nil)
}

func TestBindSlice_string_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice48(), nchar48Null, t, nil)
}

func TestBindSlice_OraString_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice48(true), nchar48Null, t, nil)
}

func TestMultiDefine_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string48(), nchar48Null, t)
}

func TestWorkload_nchar48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(nchar48Null, t)
}

func TestBindDefine_nchar48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, nchar48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// varcharB48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varcharB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varcharB48, t, nil)
}

func TestBindPtr_string_varcharB48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varcharB48, t)
}

func TestBindDefine_OraString_varcharB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), varcharB48, t, nil)
}

func TestBindSlice_string_varcharB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varcharB48, t, nil)
}

func TestBindSlice_OraString_varcharB48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), varcharB48, t, nil)
}

func TestMultiDefine_varcharB48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varcharB48, t)
}

func TestWorkload_varcharB48_session(t *testing.T) {
	t.Parallel()
	testWorkload(varcharB48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// varcharB48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varcharB48Null, t, nil)
}

func TestBindPtr_string_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varcharB48Null, t)
}

func TestBindDefine_OraString_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), varcharB48Null, t, nil)
}

func TestBindSlice_string_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varcharB48Null, t, nil)
}

func TestBindSlice_OraString_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), varcharB48Null, t, nil)
}

func TestMultiDefine_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varcharB48Null, t)
}

func TestWorkload_varcharB48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(varcharB48Null, t)
}

func TestBindDefine_varcharB48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, varcharB48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// varcharC48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varcharC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varcharC48, t, nil)
}

func TestBindPtr_string_varcharC48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varcharC48, t)
}

func TestBindDefine_OraString_varcharC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), varcharC48, t, nil)
}

func TestBindSlice_string_varcharC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varcharC48, t, nil)
}

func TestBindSlice_OraString_varcharC48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), varcharC48, t, nil)
}

func TestMultiDefine_varcharC48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varcharC48, t)
}

func TestWorkload_varcharC48_session(t *testing.T) {
	t.Parallel()
	testWorkload(varcharC48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// varcharC48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varcharC48Null, t, nil)
}

func TestBindPtr_string_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varcharC48Null, t)
}

func TestBindDefine_OraString_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), varcharC48Null, t, nil)
}

func TestBindSlice_string_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varcharC48Null, t, nil)
}

func TestBindSlice_OraString_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), varcharC48Null, t, nil)
}

func TestMultiDefine_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varcharC48Null, t)
}

func TestWorkload_varcharC48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(varcharC48Null, t)
}

func TestBindDefine_varcharC48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, varcharC48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// varchar2B48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varchar2B48, t, nil)
}

func TestBindPtr_string_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varchar2B48, t)
}

func TestBindDefine_OraString_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), varchar2B48, t, nil)
}

func TestBindSlice_string_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varchar2B48, t, nil)
}

func TestBindSlice_OraString_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), varchar2B48, t, nil)
}

func TestMultiDefine_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varchar2B48, t)
}

func TestWorkload_varchar2B48_session(t *testing.T) {
	t.Parallel()
	testWorkload(varchar2B48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// varchar2B48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varchar2B48Null, t, nil)
}

func TestBindPtr_string_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varchar2B48Null, t)
}

func TestBindDefine_OraString_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), varchar2B48Null, t, nil)
}

func TestBindSlice_string_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varchar2B48Null, t, nil)
}

func TestBindSlice_OraString_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), varchar2B48Null, t, nil)
}

func TestMultiDefine_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varchar2B48Null, t)
}

func TestWorkload_varchar2B48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(varchar2B48Null, t)
}

func TestBindDefine_varchar2B48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, varchar2B48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// varchar2C48
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varchar2C48, t, nil)
}

func TestBindPtr_string_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varchar2C48, t)
}

func TestBindDefine_OraString_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), varchar2C48, t, nil)
}

func TestBindSlice_string_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varchar2C48, t, nil)
}

func TestBindSlice_OraString_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), varchar2C48, t, nil)
}

func TestMultiDefine_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varchar2C48, t)
}

func TestWorkload_varchar2C48_session(t *testing.T) {
	t.Parallel()
	testWorkload(varchar2C48, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// varchar2C48Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), varchar2C48Null, t, nil)
}

func TestBindPtr_string_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), varchar2C48Null, t)
}

func TestBindDefine_OraString_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), varchar2C48Null, t, nil)
}

func TestBindSlice_string_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), varchar2C48Null, t, nil)
}

func TestBindSlice_OraString_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), varchar2C48Null, t, nil)
}

func TestMultiDefine_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), varchar2C48Null, t)
}

func TestWorkload_varchar2C48Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(varchar2C48Null, t)
}

func TestBindDefine_varchar2C48Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, varchar2C48Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// nvarchar248
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), nvarchar248, t, nil)
}

func TestBindPtr_string_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), nvarchar248, t)
}

func TestBindDefine_OraString_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), nvarchar248, t, nil)
}

func TestBindSlice_string_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), nvarchar248, t, nil)
}

func TestBindSlice_OraString_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), nvarchar248, t, nil)
}

func TestMultiDefine_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), nvarchar248, t)
}

func TestWorkload_nvarchar248_session(t *testing.T) {
	t.Parallel()
	testWorkload(nvarchar248, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// nvarchar248Null
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), nvarchar248Null, t, nil)
}

func TestBindPtr_string_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testBindPtr(gen_string(), nvarchar248Null, t)
}

func TestBindDefine_OraString_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), nvarchar248Null, t, nil)
}

func TestBindSlice_string_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), nvarchar248Null, t, nil)
}

func TestBindSlice_OraString_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), nvarchar248Null, t, nil)
}

func TestMultiDefine_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), nvarchar248Null, t)
}

func TestWorkload_nvarchar248Null_session(t *testing.T) {
	t.Parallel()
	testWorkload(nvarchar248Null, t)
}

func TestBindDefine_nvarchar248Null_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, nvarchar248Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// long
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_long_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), long, t, nil)
}

//func TestBindPtr_string_long_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), long, t)
//}

func TestBindDefine_OraString_long_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(false), long, t, nil)
}

func TestBindSlice_string_long_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), long, t, nil)
}

func TestBindSlice_OraString_long_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(false), long, t, nil)
}

func TestMultiDefine_long_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), long, t)
}

//func TestWorkload_long_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(long, t)
//}

//////////////////////////////////////////////////////////////////////////////////
//// longNull
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_longNull_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_string(), longNull, t, nil)
}

//func TestBindPtr_string_longNull_session(t *testing.T) {
//	//// ORA-22816: unsupported feature with RETURNING clause
//	//testBindPtr(gen_string(), longNull, t)
//}

func TestBindDefine_OraString_longNull_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraString(true), longNull, t, nil)
}

func TestBindSlice_string_longNull_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_stringSlice(), longNull, t, nil)
}

func TestBindSlice_OraString_longNull_session(t *testing.T) {
	t.Parallel()
	testBindDefine(gen_OraStringSlice(true), longNull, t, nil)
}

func TestMultiDefine_longNull_session(t *testing.T) {
	t.Parallel()
	testMultiDefine(gen_string(), longNull, t)
}

//func TestWorkload_longNull_session(t *testing.T) {
//	//// ORA-01754: a table may contain only one column of type LONG
//	//testWorkload(longNull, t)
//}

func TestBindDefine_longNull_nil_session(t *testing.T) {
	t.Parallel()
	testBindDefine(nil, longNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// clob
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_clob_session(t *testing.T) {
	testBindDefine(gen_string(), clob, t, nil)
}

func TestBindPtr_string_clob_session(t *testing.T) {
	testBindPtr(gen_string(), clob, t)
}

func TestBindDefine_OraString_clob_session(t *testing.T) {
	testBindDefine(gen_OraString(false), clob, t, nil)
}

func TestBindSlice_string_clob_session(t *testing.T) {
	testBindDefine(gen_stringSlice(), clob, t, nil)
}

func TestBindSlice_OraString_clob_session(t *testing.T) {
	testBindDefine(gen_OraStringSlice(false), clob, t, nil)
}

func TestMultiDefine_clob_session(t *testing.T) {
	testMultiDefine(gen_string(), clob, t)
}

func TestWorkload_clob_session(t *testing.T) {
	testWorkload(clob, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// clobNull
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_clobNull_session(t *testing.T) {
	testBindDefine(gen_string(), clobNull, t, nil)
}

func TestBindPtr_string_clobNull_session(t *testing.T) {
	testBindPtr(gen_string(), clobNull, t)
}

func TestBindDefine_OraString_clobNull_session(t *testing.T) {
	testBindDefine(gen_OraString(true), clobNull, t, nil)
}

func TestBindSlice_string_clobNull_session(t *testing.T) {
	testBindDefine(gen_stringSlice(), clobNull, t, nil)
}

func TestBindSlice_OraString_clobNull_session(t *testing.T) {
	testBindDefine(gen_OraStringSlice(true), clobNull, t, nil)
}

func TestMultiDefine_clobNull_session(t *testing.T) {
	testMultiDefine(gen_string(), clobNull, t)
}

func TestWorkload_clobNull_session(t *testing.T) {
	testWorkload(clobNull, t)
}

func TestBindDefine_clobNull_nil_session(t *testing.T) {
	testBindDefine(nil, clobNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// nclob
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nclob_session(t *testing.T) {
	testBindDefine(gen_string(), nclob, t, nil)
}

func TestBindPtr_string_nclob_session(t *testing.T) {
	testBindPtr(gen_string(), nclob, t)
}

func TestBindDefine_OraString_nclob_session(t *testing.T) {
	testBindDefine(gen_OraString(false), nclob, t, nil)
}

func TestBindSlice_string_nclob_session(t *testing.T) {
	testBindDefine(gen_stringSlice(), nclob, t, nil)
}

func TestBindSlice_OraString_nclob_session(t *testing.T) {
	testBindDefine(gen_OraStringSlice(false), nclob, t, nil)
}

func TestMultiDefine_nclob_session(t *testing.T) {
	testMultiDefine(gen_string(), nclob, t)
}

func TestWorkload_nclob_session(t *testing.T) {
	testWorkload(nclob, t)
}

//////////////////////////////////////////////////////////////////////////////////
//// nclobNull
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_string_nclobNull_session(t *testing.T) {
	testBindDefine(gen_string(), nclobNull, t, nil)
}

func TestBindPtr_string_nclobNull_session(t *testing.T) {
	testBindPtr(gen_string(), nclobNull, t)
}

func TestBindDefine_OraString_nclobNull_session(t *testing.T) {
	testBindDefine(gen_OraString(true), nclobNull, t, nil)
}

func TestBindSlice_string_nclobNull_session(t *testing.T) {
	testBindDefine(gen_stringSlice(), nclobNull, t, nil)
}

func TestBindSlice_OraString_nclobNull_session(t *testing.T) {
	testBindDefine(gen_OraStringSlice(true), nclobNull, t, nil)
}

func TestMultiDefine_nclobNull_session(t *testing.T) {
	testMultiDefine(gen_string(), nclobNull, t)
}

func TestWorkload_nclobNull_session(t *testing.T) {
	testWorkload(nclobNull, t)
}

func TestBindDefine_nclobNull_nil_session(t *testing.T) {
	testBindDefine(nil, nclobNull, t, nil)
}
