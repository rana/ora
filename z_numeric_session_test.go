//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v3"
)

//// numeric
//numberP38S0Identity oracleColumnType = "number(38,0) generated always as identity (start with 1 increment by 1)"
//numberP38S0         oracleColumnType = "number(38,0) not null"
//numberP38S0Null     oracleColumnType = "number(38,0) null"
//numberP16S15        oracleColumnType = "number(16,15) not null"
//numberP16S15Null    oracleColumnType = "number(16,15) null"
//binaryDouble        oracleColumnType = "binary_double not null"
//binaryDoubleNull    oracleColumnType = "binary_double null"
//binaryFloat         oracleColumnType = "binary_float not null"
//binaryFloatNull     oracleColumnType = "binary_float null"
//floatP126           oracleColumnType = "float(126) not null"
//floatP126Null       oracleColumnType = "float(126) null"

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE numberP38S0
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int64(), numberP38S0, t, nil)
}

func TestBindDefine_int32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int32(), numberP38S0, t, nil)
}

func TestBindDefine_int16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int16(), numberP38S0, t, nil)
}

func TestBindDefine_int8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int8(), numberP38S0, t, nil)
}

func TestBindDefine_uint64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint64(), numberP38S0, t, nil)
}

func TestBindDefine_uint32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint32(), numberP38S0, t, nil)
}

func TestBindDefine_uint16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint16(), numberP38S0, t, nil)
}

func TestBindDefine_uint8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint8(), numberP38S0, t, nil)
}

func TestBindDefine_float64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_float64Trunc(), numberP38S0, t, nil)
}

func TestBindDefine_float32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_float32Trunc(), numberP38S0, t, nil)
}
func TestBindDefine_NumString_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_NumStringTrunc(), numberP38S0, t, nil)
}

func TestBindDefine_OraInt64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), numberP38S0, t, nil)
}

func TestBindDefine_OraInt32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), numberP38S0, t, nil)
}

func TestBindDefine_OraInt16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), numberP38S0, t, nil)
}

func TestBindDefine_OraInt8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), numberP38S0, t, nil)
}

func TestBindDefine_OraUint64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), numberP38S0, t, nil)
}

func TestBindDefine_OraUint32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), numberP38S0, t, nil)
}

func TestBindDefine_OraUint16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), numberP38S0, t, nil)
}

func TestBindDefine_OraUint8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), numberP38S0, t, nil)
}

func TestBindDefine_OraFloat64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Trunc(false), numberP38S0, t, nil)
}

func TestBindDefine_OraFloat32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Trunc(false), numberP38S0, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR numberP38S0
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_int64(), numberP38S0, t)
}

func TestBindPtr_int32_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_int32(), numberP38S0, t)
}

func TestBindPtr_int16_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_int16(), numberP38S0, t)
}

func TestBindPtr_int8_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_int8(), numberP38S0, t)
}

func TestBindPtr_uint64_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_uint64(), numberP38S0, t)
}

func TestBindPtr_uint32_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_uint32(), numberP38S0, t)
}

func TestBindPtr_uint16_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_uint16(), numberP38S0, t)
}

func TestBindPtr_uint8_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_uint8(), numberP38S0, t)
}

func TestBindPtr_float64_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_float64Trunc(), numberP38S0, t)
}

func TestBindPtr_float32_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_float32Trunc(), numberP38S0, t)
}
func TestBindPtr_NumString_numberP38S0_session(t *testing.T) {
	testBindPtr(gen_NumStringTrunc(), numberP38S0, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE numberP38S0
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), numberP38S0, t, nil)
}

func TestBindSlice_int32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), numberP38S0, t, nil)
}

func TestBindSlice_int16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), numberP38S0, t, nil)
}

func TestBindSlice_int8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), numberP38S0, t, nil)
}

func TestBindSlice_uint64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), numberP38S0, t, nil)
}

func TestBindSlice_uint32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), numberP38S0, t, nil)
}

func TestBindSlice_uint16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), numberP38S0, t, nil)
}

func TestBindSlice_uint8_numberP38S0_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), numberP38S0, t, sc)
}

func TestBindSlice_float64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_float64TruncSlice(), numberP38S0, t, nil)
}

func TestBindSlice_float32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_float32TruncSlice(), numberP38S0, t, nil)
}
func TestBindSlice_NumString_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_NumStringTruncSlice(), numberP38S0, t, nil)
}

func TestBindSlice_OraUint64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint64Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraUint32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint32Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraUint16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint16Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraUint8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraUint8Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraInt64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraInt32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraInt16_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraInt8_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraFloat64_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraFloat64TruncSlice(false), numberP38S0, t, nil)
}

func TestBindSlice_OraFloat32_numberP38S0_session(t *testing.T) {
	testBindDefine(gen_OraFloat32TruncSlice(false), numberP38S0, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC numberP38S0
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_numberP38S0_session(t *testing.T) {
	testMultiDefine(gen_int64(), numberP38S0, t)
}

func TestWorkload_numberP38S0_session(t *testing.T) {
	testWorkload(numberP38S0, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE numberP38S0Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int64(), numberP38S0Null, t, nil)
}

func TestBindDefine_int32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int32(), numberP38S0Null, t, nil)
}

func TestBindDefine_int16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int16(), numberP38S0Null, t, nil)
}

func TestBindDefine_int8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int8(), numberP38S0Null, t, nil)
}

func TestBindDefine_uint64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint64(), numberP38S0Null, t, nil)
}

func TestBindDefine_uint32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint32(), numberP38S0Null, t, nil)
}

func TestBindDefine_uint16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint16(), numberP38S0Null, t, nil)
}

func TestBindDefine_uint8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint8(), numberP38S0Null, t, nil)
}

func TestBindDefine_float64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_float64Trunc(), numberP38S0Null, t, nil)
}

func TestBindDefine_float32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_float32Trunc(), numberP38S0Null, t, nil)
}
func TestBindDefine_NumString_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_NumStringTrunc(), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraFloat64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Trunc(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraFloat32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Trunc(false), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt64_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt32_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt16_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraInt8_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint64_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint32_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint16_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraUint8_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraFloat64_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Trunc(true), numberP38S0Null, t, nil)
}

func TestBindDefine_OraFloat32_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Trunc(true), numberP38S0Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR numberP38S0Null
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_int64(), numberP38S0Null, t)
}

func TestBindPtr_int32_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_int32(), numberP38S0Null, t)
}

func TestBindPtr_int16_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_int16(), numberP38S0Null, t)
}

func TestBindPtr_int8_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_int8(), numberP38S0Null, t)
}

func TestBindPtr_uint64_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_uint64(), numberP38S0Null, t)
}

func TestBindPtr_uint32_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_uint32(), numberP38S0Null, t)
}

func TestBindPtr_uint16_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_uint16(), numberP38S0Null, t)
}

func TestBindPtr_uint8_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_uint8(), numberP38S0Null, t)
}

func TestBindPtr_float64_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_float64Trunc(), numberP38S0Null, t)
}

func TestBindPtr_float32_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_float32Trunc(), numberP38S0Null, t)
}
func TestBindPtr_NumString_numberP38S0Null_session(t *testing.T) {
	testBindPtr(gen_NumStringTrunc(), numberP38S0Null, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE numberP38S0Null
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_int32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_int16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_int8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_uint64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_uint32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_uint16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), numberP38S0Null, t, nil)
}

func TestBindSlice_uint8_numberP38S0Null_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), numberP38S0Null, t, sc)
}

func TestBindSlice_float64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_float64TruncSlice(), numberP38S0Null, t, nil)
}

func TestBindSlice_float32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_float32TruncSlice(), numberP38S0Null, t, nil)
}
func TestBindSlice_NumString_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_NumStringTruncSlice(), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt16_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt8_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraFloat64_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64TruncSlice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraFloat32_numberP38S0Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32TruncSlice(false), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt64_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(true), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt32_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(true), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt16_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(true), numberP38S0Null, t, nil)
}

func TestBindSlice_OraInt8_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(true), numberP38S0Null, t, nil)
}

func TestBindSlice_OraFloat64_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64TruncSlice(true), numberP38S0Null, t, nil)
}

func TestBindSlice_OraFloat32_numberP38S0Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32TruncSlice(true), numberP38S0Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC numberP38S0Null
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_numberP38S0Null_session(t *testing.T) {
	testMultiDefine(gen_int64(), numberP38S0Null, t)
}

func TestWorkload_numberP38S0Null_session(t *testing.T) {
	testWorkload(numberP38S0Null, t)
}

func TestBindDefine_numberP38S0Null_nil_session(t *testing.T) {
	testBindDefine(nil, numberP38S0Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE numberP16S15
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int64(), numberP16S15, t, nil)
}

func TestBindDefine_int32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int32(), numberP16S15, t, nil)
}

func TestBindDefine_int16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int16(), numberP16S15, t, nil)
}

func TestBindDefine_int8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int8(), numberP16S15, t, nil)
}

func TestBindDefine_uint64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint64(), numberP16S15, t, nil)
}

func TestBindDefine_uint32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint32(), numberP16S15, t, nil)
}

func TestBindDefine_uint16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint16(), numberP16S15, t, nil)
}

func TestBindDefine_uint8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint8(), numberP16S15, t, nil)
}

func TestBindDefine_float64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_float64(), numberP16S15, t, nil)
}

func TestBindDefine_float32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_float32(), numberP16S15, t, nil)
}
func TestBindDefine_NumString_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_NumString(), numberP16S15, t, nil)
}

func TestBindDefine_OraInt64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), numberP16S15, t, nil)
}

func TestBindDefine_OraInt32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), numberP16S15, t, nil)
}

func TestBindDefine_OraInt16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), numberP16S15, t, nil)
}

func TestBindDefine_OraInt8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), numberP16S15, t, nil)
}

func TestBindDefine_OraUint64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), numberP16S15, t, nil)
}

func TestBindDefine_OraUint32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), numberP16S15, t, nil)
}

func TestBindDefine_OraUint16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), numberP16S15, t, nil)
}

func TestBindDefine_OraUint8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), numberP16S15, t, nil)
}

func TestBindDefine_OraFloat64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), numberP16S15, t, nil)
}

func TestBindDefine_OraFloat32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), numberP16S15, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR numberP16S15
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_int64(), numberP16S15, t)
}

func TestBindPtr_int32_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_int32(), numberP16S15, t)
}

func TestBindPtr_int16_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_int16(), numberP16S15, t)
}

func TestBindPtr_int8_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_int8(), numberP16S15, t)
}

func TestBindPtr_uint64_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_uint64(), numberP16S15, t)
}

func TestBindPtr_uint32_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_uint32(), numberP16S15, t)
}

func TestBindPtr_uint16_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_uint16(), numberP16S15, t)
}

func TestBindPtr_uint8_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_uint8(), numberP16S15, t)
}

func TestBindPtr_float64_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_float64(), numberP16S15, t)
}

func TestBindPtr_float32_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_float32(), numberP16S15, t)
}
func TestBindPtr_NumString_numberP16S15_session(t *testing.T) {
	testBindPtr(gen_NumString(), numberP16S15, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE numberP16S15
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_numberP16S15_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_int64Slice(), numberP16S15, t, nil)
}

func TestBindSlice_int32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), numberP16S15, t, nil)
}

func TestBindSlice_int16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), numberP16S15, t, nil)
}

func TestBindSlice_int8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), numberP16S15, t, nil)
}

func TestBindSlice_uint64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), numberP16S15, t, nil)
}

func TestBindSlice_uint32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), numberP16S15, t, nil)
}

func TestBindSlice_uint16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), numberP16S15, t, nil)
}

func TestBindSlice_uint8_numberP16S15_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), numberP16S15, t, sc)
}

func TestBindSlice_float64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), numberP16S15, t, nil)
}

func TestBindSlice_float32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), numberP16S15, t, nil)
}
func TestBindSlice_NumString_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), numberP16S15, t, nil)
}

func TestBindSlice_OraInt64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), numberP16S15, t, nil)
}

func TestBindSlice_OraInt32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), numberP16S15, t, nil)
}

func TestBindSlice_OraInt16_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), numberP16S15, t, nil)
}

func TestBindSlice_OraInt8_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), numberP16S15, t, nil)
}

func TestBindSlice_OraFloat64_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), numberP16S15, t, nil)
}

func TestBindSlice_OraFloat32_numberP16S15_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), numberP16S15, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC numberP16S15
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_numberP16S15_session(t *testing.T) {
	testMultiDefine(gen_int64(), numberP16S15, t)
}

func TestWorkload_numberP16S15_session(t *testing.T) {
	testWorkload(numberP16S15, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE numberP16S15Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int64(), numberP16S15Null, t, nil)
}

func TestBindDefine_int32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int32(), numberP16S15Null, t, nil)
}

func TestBindDefine_int16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int16(), numberP16S15Null, t, nil)
}

func TestBindDefine_int8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int8(), numberP16S15Null, t, nil)
}

func TestBindDefine_uint64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint64(), numberP16S15Null, t, nil)
}

func TestBindDefine_uint32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint32(), numberP16S15Null, t, nil)
}

func TestBindDefine_uint16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint16(), numberP16S15Null, t, nil)
}

func TestBindDefine_uint8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint8(), numberP16S15Null, t, nil)
}

func TestBindDefine_float64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_float64(), numberP16S15Null, t, nil)
}

func TestBindDefine_float32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_float32(), numberP16S15Null, t, nil)
}

func TestBindDefine_NumString_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_NumString(), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraFloat64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraFloat32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt64_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt32_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt16_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraInt8_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint64_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint32_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint16_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraUint8_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraFloat64_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(true), numberP16S15Null, t, nil)
}

func TestBindDefine_OraFloat32_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(true), numberP16S15Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR numberP16S15Null
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_int64(), numberP16S15Null, t)
}

func TestBindPtr_int32_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_int32(), numberP16S15Null, t)
}

func TestBindPtr_int16_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_int16(), numberP16S15Null, t)
}

func TestBindPtr_int8_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_int8(), numberP16S15Null, t)
}

func TestBindPtr_uint64_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_uint64(), numberP16S15Null, t)
}

func TestBindPtr_uint32_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_uint32(), numberP16S15Null, t)
}

func TestBindPtr_uint16_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_uint16(), numberP16S15Null, t)
}

func TestBindPtr_uint8_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_uint8(), numberP16S15Null, t)
}

func TestBindPtr_float64_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_float64(), numberP16S15Null, t)
}

func TestBindPtr_float32_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_float32(), numberP16S15Null, t)
}

func TestBindPtr_NumString_numberP16S15Null_session(t *testing.T) {
	testBindPtr(gen_NumString(), numberP16S15Null, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE numberP16S15Null
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_int32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_int16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_int8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_uint64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_uint32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_uint16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_uint8_numberP16S15Null_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), numberP16S15Null, t, sc)
}

func TestBindSlice_float64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_float32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), numberP16S15Null, t, nil)
}

func TestBindSlice_NumString_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt16_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt8_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraFloat64_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraFloat32_numberP16S15Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt64_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(true), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt32_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(true), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt16_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(true), numberP16S15Null, t, nil)
}

func TestBindSlice_OraInt8_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(true), numberP16S15Null, t, nil)
}

func TestBindSlice_OraFloat64_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(true), numberP16S15Null, t, nil)
}

func TestBindSlice_OraFloat32_numberP16S15Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(true), numberP16S15Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC numberP16S15Null
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_numberP16S15Null_session(t *testing.T) {
	testMultiDefine(gen_int64(), numberP16S15Null, t)
}

func TestWorkload_numberP16S15Null_session(t *testing.T) {
	testWorkload(numberP16S15Null, t)
}

func TestBindDefine_numberP16S15Null_nil_session(t *testing.T) {
	testBindDefine(nil, numberP16S15Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE binaryDouble
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int64(), binaryDouble, t, nil)
}

func TestBindDefine_int32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int32(), binaryDouble, t, nil)
}

func TestBindDefine_int16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int16(), binaryDouble, t, nil)
}

func TestBindDefine_int8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int8(), binaryDouble, t, nil)
}

func TestBindDefine_uint64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint64(), binaryDouble, t, nil)
}

func TestBindDefine_uint32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint32(), binaryDouble, t, nil)
}

func TestBindDefine_uint16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint16(), binaryDouble, t, nil)
}

func TestBindDefine_uint8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint8(), binaryDouble, t, nil)
}

func TestBindDefine_float64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_float64(), binaryDouble, t, nil)
}

func TestBindDefine_float32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_float32(), binaryDouble, t, nil)
}

func TestBindDefine_NumString_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_NumString(), binaryDouble, t, nil)
}

func TestBindDefine_OraInt64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), binaryDouble, t, nil)
}

func TestBindDefine_OraInt32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), binaryDouble, t, nil)
}

func TestBindDefine_OraInt16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), binaryDouble, t, nil)
}

func TestBindDefine_OraInt8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), binaryDouble, t, nil)
}

func TestBindDefine_OraUint64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), binaryDouble, t, nil)
}

func TestBindDefine_OraUint32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), binaryDouble, t, nil)
}

func TestBindDefine_OraUint16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), binaryDouble, t, nil)
}

func TestBindDefine_OraUint8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), binaryDouble, t, nil)
}

func TestBindDefine_OraFloat64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), binaryDouble, t, nil)
}

func TestBindDefine_OraFloat32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), binaryDouble, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR binaryDouble
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_int64(), binaryDouble, t)
}

func TestBindPtr_int32_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_int32(), binaryDouble, t)
}

func TestBindPtr_int16_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_int16(), binaryDouble, t)
}

func TestBindPtr_int8_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_int8(), binaryDouble, t)
}

func TestBindPtr_uint64_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_uint64(), binaryDouble, t)
}

func TestBindPtr_uint32_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_uint32(), binaryDouble, t)
}

func TestBindPtr_uint16_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_uint16(), binaryDouble, t)
}

func TestBindPtr_uint8_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_uint8(), binaryDouble, t)
}

func TestBindPtr_float64_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_float64(), binaryDouble, t)
}

func TestBindPtr_float32_binaryDouble_session(t *testing.T) {
	testBindPtr(gen_float32(), binaryDouble, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE binaryDouble
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), binaryDouble, t, nil)
}

func TestBindSlice_int32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), binaryDouble, t, nil)
}

func TestBindSlice_int16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), binaryDouble, t, nil)
}

func TestBindSlice_int8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), binaryDouble, t, nil)
}

func TestBindSlice_uint64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), binaryDouble, t, nil)
}

func TestBindSlice_uint32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), binaryDouble, t, nil)
}

func TestBindSlice_uint16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), binaryDouble, t, nil)
}

func TestBindSlice_uint8_binaryDouble_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), binaryDouble, t, sc)
}

func TestBindSlice_float64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), binaryDouble, t, nil)
}

func TestBindSlice_float32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), binaryDouble, t, nil)
}

func TestBindSlice_NumString_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), binaryDouble, t, nil)
}

func TestBindSlice_OraInt64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), binaryDouble, t, nil)
}

func TestBindSlice_OraInt32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), binaryDouble, t, nil)
}

func TestBindSlice_OraInt16_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), binaryDouble, t, nil)
}

func TestBindSlice_OraInt8_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), binaryDouble, t, nil)
}

func TestBindSlice_OraFloat64_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), binaryDouble, t, nil)
}

func TestBindSlice_OraFloat32_binaryDouble_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), binaryDouble, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC binaryDouble
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_binaryDouble_session(t *testing.T) {
	testMultiDefine(gen_int64(), binaryDouble, t)
}

func TestWorkload_binaryDouble_session(t *testing.T) {
	testWorkload(binaryDouble, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE binaryDoubleNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int64(), binaryDoubleNull, t, nil)
}

func TestBindDefine_int32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int32(), binaryDoubleNull, t, nil)
}

func TestBindDefine_int16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int16(), binaryDoubleNull, t, nil)
}

func TestBindDefine_int8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int8(), binaryDoubleNull, t, nil)
}

func TestBindDefine_uint64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint64(), binaryDoubleNull, t, nil)
}

func TestBindDefine_uint32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint32(), binaryDoubleNull, t, nil)
}

func TestBindDefine_uint16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint16(), binaryDoubleNull, t, nil)
}

func TestBindDefine_uint8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint8(), binaryDoubleNull, t, nil)
}

func TestBindDefine_float64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_float64(), binaryDoubleNull, t, nil)
}

func TestBindDefine_float32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_float32(), binaryDoubleNull, t, nil)
}

func TestBindDefine_NumString_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_NumString(), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraFloat64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraFloat32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt64_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt32_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt16_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraInt8_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint64_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint32_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint16_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraUint8_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraFloat64_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(true), binaryDoubleNull, t, nil)
}

func TestBindDefine_OraFloat32_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(true), binaryDoubleNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR binaryDoubleNull
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_int64(), binaryDoubleNull, t)
}

func TestBindPtr_int32_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_int32(), binaryDoubleNull, t)
}

func TestBindPtr_int16_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_int16(), binaryDoubleNull, t)
}

func TestBindPtr_int8_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_int8(), binaryDoubleNull, t)
}

func TestBindPtr_uint64_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_uint64(), binaryDoubleNull, t)
}

func TestBindPtr_uint32_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_uint32(), binaryDoubleNull, t)
}

func TestBindPtr_uint16_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_uint16(), binaryDoubleNull, t)
}

func TestBindPtr_uint8_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_uint8(), binaryDoubleNull, t)
}

func TestBindPtr_float64_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_float64(), binaryDoubleNull, t)
}

func TestBindPtr_float32_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_float32(), binaryDoubleNull, t)
}

func TestBindPtr_NumString_binaryDoubleNull_session(t *testing.T) {
	testBindPtr(gen_NumString(), binaryDoubleNull, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE binaryDoubleNull
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_int32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_int16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_int8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_uint64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_uint32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_uint16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_uint8_binaryDoubleNull_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), binaryDoubleNull, t, sc)
}

func TestBindSlice_float64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_float32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_NumString_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt16_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt8_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraFloat64_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraFloat32_binaryDoubleNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt64_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(true), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt32_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(true), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt16_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(true), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraInt8_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(true), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraFloat64_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(true), binaryDoubleNull, t, nil)
}

func TestBindSlice_OraFloat32_binaryDoubleNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(true), binaryDoubleNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC binaryDoubleNull
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_binaryDoubleNull_session(t *testing.T) {
	testMultiDefine(gen_int64(), binaryDoubleNull, t)
}

func TestWorkload_binaryDoubleNull_session(t *testing.T) {
	testWorkload(binaryDoubleNull, t)
}

func TestBindDefine_binaryDoubleNull_nil_session(t *testing.T) {
	testBindDefine(nil, binaryDoubleNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE binaryFloat
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int64(), binaryFloat, t, nil)
}

func TestBindDefine_int32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int32(), binaryFloat, t, nil)
}

func TestBindDefine_int16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int16(), binaryFloat, t, nil)
}

func TestBindDefine_int8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int8(), binaryFloat, t, nil)
}

func TestBindDefine_uint64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint64(), binaryFloat, t, nil)
}

func TestBindDefine_uint32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint32(), binaryFloat, t, nil)
}

func TestBindDefine_uint16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint16(), binaryFloat, t, nil)
}

func TestBindDefine_uint8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint8(), binaryFloat, t, nil)
}

func TestBindDefine_float64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_float64(), binaryFloat, t, nil)
}

func TestBindDefine_float32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_float32(), binaryFloat, t, nil)
}

func TestBindDefine_NumString_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_NumString(), binaryFloat, t, nil)
}

func TestBindDefine_OraInt64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), binaryFloat, t, nil)
}

func TestBindDefine_OraInt32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), binaryFloat, t, nil)
}

func TestBindDefine_OraInt16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), binaryFloat, t, nil)
}

func TestBindDefine_OraInt8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), binaryFloat, t, nil)
}

func TestBindDefine_OraUint64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), binaryFloat, t, nil)
}

func TestBindDefine_OraUint32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), binaryFloat, t, nil)
}

func TestBindDefine_OraUint16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), binaryFloat, t, nil)
}

func TestBindDefine_OraUint8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), binaryFloat, t, nil)
}

func TestBindDefine_OraFloat64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), binaryFloat, t, nil)
}

func TestBindDefine_OraFloat32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), binaryFloat, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR binaryFloat
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_int64(), binaryFloat, t)
}

func TestBindPtr_int32_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_int32(), binaryFloat, t)
}

func TestBindPtr_int16_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_int16(), binaryFloat, t)
}

func TestBindPtr_int8_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_int8(), binaryFloat, t)
}

func TestBindPtr_uint64_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_uint64(), binaryFloat, t)
}

func TestBindPtr_uint32_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_uint32(), binaryFloat, t)
}

func TestBindPtr_uint16_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_uint16(), binaryFloat, t)
}

func TestBindPtr_uint8_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_uint8(), binaryFloat, t)
}

func TestBindPtr_float64_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_float64(), binaryFloat, t)
}

func TestBindPtr_float32_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_float32(), binaryFloat, t)
}

func TestBindPtr_NumString_binaryFloat_session(t *testing.T) {
	testBindPtr(gen_NumString(), binaryFloat, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE binaryFloat
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), binaryFloat, t, nil)
}

func TestBindSlice_int32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), binaryFloat, t, nil)
}

func TestBindSlice_int16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), binaryFloat, t, nil)
}

func TestBindSlice_int8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), binaryFloat, t, nil)
}

func TestBindSlice_uint64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), binaryFloat, t, nil)
}

func TestBindSlice_uint32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), binaryFloat, t, nil)
}

func TestBindSlice_uint16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), binaryFloat, t, nil)
}

func TestBindSlice_uint8_binaryFloat_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), binaryFloat, t, sc)
}

func TestBindSlice_float64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), binaryFloat, t, nil)
}

func TestBindSlice_float32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), binaryFloat, t, nil)
}

func TestBindSlice_NumString_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), binaryFloat, t, nil)
}

func TestBindSlice_OraInt64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), binaryFloat, t, nil)
}

func TestBindSlice_OraInt32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), binaryFloat, t, nil)
}

func TestBindSlice_OraInt16_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), binaryFloat, t, nil)
}

func TestBindSlice_OraInt8_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), binaryFloat, t, nil)
}

func TestBindSlice_OraFloat64_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), binaryFloat, t, nil)
}

func TestBindSlice_OraFloat32_binaryFloat_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), binaryFloat, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC binaryFloat
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_binaryFloat_session(t *testing.T) {
	testMultiDefine(gen_int64(), binaryFloat, t)
}

func TestWorkload_binaryFloat_session(t *testing.T) {
	testWorkload(binaryFloat, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE binaryFloatNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int64(), binaryFloatNull, t, nil)
}

func TestBindDefine_int32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int32(), binaryFloatNull, t, nil)
}

func TestBindDefine_int16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int16(), binaryFloatNull, t, nil)
}

func TestBindDefine_int8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int8(), binaryFloatNull, t, nil)
}

func TestBindDefine_uint64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint64(), binaryFloatNull, t, nil)
}

func TestBindDefine_uint32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint32(), binaryFloatNull, t, nil)
}

func TestBindDefine_uint16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint16(), binaryFloatNull, t, nil)
}

func TestBindDefine_uint8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint8(), binaryFloatNull, t, nil)
}

func TestBindDefine_float64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_float64(), binaryFloatNull, t, nil)
}

func TestBindDefine_float32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_float32(), binaryFloatNull, t, nil)
}

func TestBindDefine_NumString_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_NumString(), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraFloat64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraFloat32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt64_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt32_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt16_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraInt8_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint64_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint32_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint16_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraUint8_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraFloat64_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(true), binaryFloatNull, t, nil)
}

func TestBindDefine_OraFloat32_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(true), binaryFloatNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR binaryFloatNull
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_int64(), binaryFloatNull, t)
}

func TestBindPtr_int32_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_int32(), binaryFloatNull, t)
}

func TestBindPtr_int16_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_int16(), binaryFloatNull, t)
}

func TestBindPtr_int8_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_int8(), binaryFloatNull, t)
}

func TestBindPtr_uint64_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_uint64(), binaryFloatNull, t)
}

func TestBindPtr_uint32_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_uint32(), binaryFloatNull, t)
}

func TestBindPtr_uint16_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_uint16(), binaryFloatNull, t)
}

func TestBindPtr_uint8_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_uint8(), binaryFloatNull, t)
}

func TestBindPtr_float64_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_float64(), binaryFloatNull, t)
}

func TestBindPtr_float32_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_float32(), binaryFloatNull, t)
}

func TestBindPtr_NumString_binaryFloatNull_session(t *testing.T) {
	testBindPtr(gen_NumString(), binaryFloatNull, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE binaryFloatNull
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_int32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_int16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_int8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_uint64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_uint32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_uint16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_uint8_binaryFloatNull_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), binaryFloatNull, t, sc)
}

func TestBindSlice_float64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_float32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), binaryFloatNull, t, nil)
}

func TestBindSlice_NumString_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt16_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt8_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraFloat64_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraFloat32_binaryFloatNull_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt64_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(true), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt32_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(true), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt16_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(true), binaryFloatNull, t, nil)
}

func TestBindSlice_OraInt8_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(true), binaryFloatNull, t, nil)
}

func TestBindSlice_OraFloat64_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(true), binaryFloatNull, t, nil)
}

func TestBindSlice_OraFloat32_binaryFloatNull_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(true), binaryFloatNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC binaryFloatNull
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_binaryFloatNull_session(t *testing.T) {
	testMultiDefine(gen_int64(), binaryFloatNull, t)
}

func TestWorkload_binaryFloatNull_session(t *testing.T) {
	testWorkload(binaryFloatNull, t)
}

func TestBindDefine_binaryFloatNull_nil_session(t *testing.T) {
	testBindDefine(nil, binaryFloatNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE floatP126
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_floatP126_session(t *testing.T) {
	testBindDefine(gen_int64(), floatP126, t, nil)
}

func TestBindDefine_int32_floatP126_session(t *testing.T) {
	testBindDefine(gen_int32(), floatP126, t, nil)
}

func TestBindDefine_int16_floatP126_session(t *testing.T) {
	testBindDefine(gen_int16(), floatP126, t, nil)
}

func TestBindDefine_int8_floatP126_session(t *testing.T) {
	testBindDefine(gen_int8(), floatP126, t, nil)
}

func TestBindDefine_uint64_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint64(), floatP126, t, nil)
}

func TestBindDefine_uint32_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint32(), floatP126, t, nil)
}

func TestBindDefine_uint16_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint16(), floatP126, t, nil)
}

func TestBindDefine_uint8_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint8(), floatP126, t, nil)
}

func TestBindDefine_float64_floatP126_session(t *testing.T) {
	testBindDefine(gen_float64(), floatP126, t, nil)
}

func TestBindDefine_float32_floatP126_session(t *testing.T) {
	testBindDefine(gen_float32(), floatP126, t, nil)
}

func TestBindDefine_NumString_floatP126_session(t *testing.T) {
	testBindDefine(gen_NumString(), floatP126, t, nil)
}

func TestBindDefine_OraInt64_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), floatP126, t, nil)
}

func TestBindDefine_OraInt32_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), floatP126, t, nil)
}

func TestBindDefine_OraInt16_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), floatP126, t, nil)
}

func TestBindDefine_OraInt8_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), floatP126, t, nil)
}

func TestBindDefine_OraUint64_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), floatP126, t, nil)
}

func TestBindDefine_OraUint32_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), floatP126, t, nil)
}

func TestBindDefine_OraUint16_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), floatP126, t, nil)
}

func TestBindDefine_OraUint8_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), floatP126, t, nil)
}

func TestBindDefine_OraFloat64_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), floatP126, t, nil)
}

func TestBindDefine_OraFloat32_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), floatP126, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR floatP126
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_floatP126_session(t *testing.T) {
	testBindPtr(gen_int64(), floatP126, t)
}

func TestBindPtr_int32_floatP126_session(t *testing.T) {
	testBindPtr(gen_int32(), floatP126, t)
}

func TestBindPtr_int16_floatP126_session(t *testing.T) {
	testBindPtr(gen_int16(), floatP126, t)
}

func TestBindPtr_int8_floatP126_session(t *testing.T) {
	testBindPtr(gen_int8(), floatP126, t)
}

func TestBindPtr_uint64_floatP126_session(t *testing.T) {
	testBindPtr(gen_uint64(), floatP126, t)
}

func TestBindPtr_uint32_floatP126_session(t *testing.T) {
	testBindPtr(gen_uint32(), floatP126, t)
}

func TestBindPtr_uint16_floatP126_session(t *testing.T) {
	testBindPtr(gen_uint16(), floatP126, t)
}

func TestBindPtr_uint8_floatP126_session(t *testing.T) {
	testBindPtr(gen_uint8(), floatP126, t)
}

func TestBindPtr_float64_floatP126_session(t *testing.T) {
	testBindPtr(gen_float64(), floatP126, t)
}

func TestBindPtr_float32_floatP126_session(t *testing.T) {
	testBindPtr(gen_float32(), floatP126, t)
}

func TestBindPtr_NumString_floatP126_session(t *testing.T) {
	testBindPtr(gen_NumString(), floatP126, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE floatP126
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_floatP126_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), floatP126, t, nil)
}

func TestBindSlice_int32_floatP126_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), floatP126, t, nil)
}

func TestBindSlice_int16_floatP126_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), floatP126, t, nil)
}

func TestBindSlice_int8_floatP126_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), floatP126, t, nil)
}

func TestBindSlice_uint64_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), floatP126, t, nil)
}

func TestBindSlice_uint32_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), floatP126, t, nil)
}

func TestBindSlice_uint16_floatP126_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), floatP126, t, nil)
}

func TestBindSlice_uint8_floatP126_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), floatP126, t, sc)
}

func TestBindSlice_float64_floatP126_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), floatP126, t, nil)
}

func TestBindSlice_float32_floatP126_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), floatP126, t, nil)
}

func TestBindSlice_NumString_floatP126_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), floatP126, t, nil)
}

func TestBindSlice_OraInt64_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), floatP126, t, nil)
}

func TestBindSlice_OraInt32_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), floatP126, t, nil)
}

func TestBindSlice_OraInt16_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), floatP126, t, nil)
}

func TestBindSlice_OraInt8_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), floatP126, t, nil)
}

func TestBindSlice_OraFloat64_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), floatP126, t, nil)
}

func TestBindSlice_OraFloat32_floatP126_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), floatP126, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC floatP126
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_floatP126_session(t *testing.T) {
	testMultiDefine(gen_int64(), floatP126, t)
}

func TestWorkload_floatP126_session(t *testing.T) {
	testWorkload(floatP126, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND DEFINE VALUE floatP126Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_int64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int64(), floatP126Null, t, nil)
}

func TestBindDefine_int32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int32(), floatP126Null, t, nil)
}

func TestBindDefine_int16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int16(), floatP126Null, t, nil)
}

func TestBindDefine_int8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int8(), floatP126Null, t, nil)
}

func TestBindDefine_uint64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint64(), floatP126Null, t, nil)
}

func TestBindDefine_uint32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint32(), floatP126Null, t, nil)
}

func TestBindDefine_uint16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint16(), floatP126Null, t, nil)
}

func TestBindDefine_uint8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint8(), floatP126Null, t, nil)
}

func TestBindDefine_float64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_float64(), floatP126Null, t, nil)
}

func TestBindDefine_float32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_float32(), floatP126Null, t, nil)
}

func TestBindDefine_NumString_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_NumString(), floatP126Null, t, nil)
}

func TestBindDefine_OraInt64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(false), floatP126Null, t, nil)
}

func TestBindDefine_OraInt32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(false), floatP126Null, t, nil)
}

func TestBindDefine_OraInt16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(false), floatP126Null, t, nil)
}

func TestBindDefine_OraInt8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(false), floatP126Null, t, nil)
}

func TestBindDefine_OraUint64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(false), floatP126Null, t, nil)
}

func TestBindDefine_OraUint32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(false), floatP126Null, t, nil)
}

func TestBindDefine_OraUint16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(false), floatP126Null, t, nil)
}

func TestBindDefine_OraUint8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(false), floatP126Null, t, nil)
}

func TestBindDefine_OraFloat64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(false), floatP126Null, t, nil)
}

func TestBindDefine_OraFloat32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(false), floatP126Null, t, nil)
}

func TestBindDefine_OraInt64_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64(true), floatP126Null, t, nil)
}

func TestBindDefine_OraInt32_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32(true), floatP126Null, t, nil)
}

func TestBindDefine_OraInt16_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16(true), floatP126Null, t, nil)
}

func TestBindDefine_OraInt8_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8(true), floatP126Null, t, nil)
}

func TestBindDefine_OraUint64_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint64(true), floatP126Null, t, nil)
}

func TestBindDefine_OraUint32_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint32(true), floatP126Null, t, nil)
}

func TestBindDefine_OraUint16_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint16(true), floatP126Null, t, nil)
}

func TestBindDefine_OraUint8_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraUint8(true), floatP126Null, t, nil)
}

func TestBindDefine_OraFloat64_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64(true), floatP126Null, t, nil)
}

func TestBindDefine_OraFloat32_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32(true), floatP126Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// BIND PTR floatP126Null
////////////////////////////////////////////////////////////////////////////////
func TestBindPtr_int64_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_int64(), floatP126Null, t)
}

func TestBindPtr_int32_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_int32(), floatP126Null, t)
}

func TestBindPtr_int16_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_int16(), floatP126Null, t)
}

func TestBindPtr_int8_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_int8(), floatP126Null, t)
}

func TestBindPtr_uint64_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_uint64(), floatP126Null, t)
}

func TestBindPtr_uint32_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_uint32(), floatP126Null, t)
}

func TestBindPtr_uint16_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_uint16(), floatP126Null, t)
}

func TestBindPtr_uint8_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_uint8(), floatP126Null, t)
}

func TestBindPtr_float64_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_float64(), floatP126Null, t)
}

func TestBindPtr_float32_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_float32(), floatP126Null, t)
}

func TestBindPtr_NumString_floatP126Null_session(t *testing.T) {
	testBindPtr(gen_NumString(), floatP126Null, t)
}

////////////////////////////////////////////////////////////////////////////////
// BIND SLICE floatP126Null
////////////////////////////////////////////////////////////////////////////////

func TestBindSlice_int64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int64Slice(), floatP126Null, t, nil)
}

func TestBindSlice_int32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int32Slice(), floatP126Null, t, nil)
}

func TestBindSlice_int16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int16Slice(), floatP126Null, t, nil)
}

func TestBindSlice_int8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_int8Slice(), floatP126Null, t, nil)
}

func TestBindSlice_uint64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint64Slice(), floatP126Null, t, nil)
}

func TestBindSlice_uint32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint32Slice(), floatP126Null, t, nil)
}

func TestBindSlice_uint16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_uint16Slice(), floatP126Null, t, nil)
}

func TestBindSlice_uint8_floatP126Null_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	sc.SetByteSlice(ora.U8)
	testBindDefine(gen_uint8Slice(), floatP126Null, t, sc)
}

func TestBindSlice_float64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_float64Slice(), floatP126Null, t, nil)
}

func TestBindSlice_float32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_float32Slice(), floatP126Null, t, nil)
}

func TestBindSlice_NumString_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_NumStringSlice(), floatP126Null, t, nil)
}

func TestBindSlice_OraInt64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraInt32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraInt16_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraInt8_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraFloat64_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraFloat32_floatP126Null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(false), floatP126Null, t, nil)
}

func TestBindSlice_OraInt64_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt64Slice(true), floatP126Null, t, nil)
}

func TestBindSlice_OraInt32_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt32Slice(true), floatP126Null, t, nil)
}

func TestBindSlice_OraInt16_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt16Slice(true), floatP126Null, t, nil)
}

func TestBindSlice_OraInt8_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraInt8Slice(true), floatP126Null, t, nil)
}

func TestBindSlice_OraFloat64_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat64Slice(true), floatP126Null, t, nil)
}

func TestBindSlice_OraFloat32_floatP126Null_null_session(t *testing.T) {
	testBindDefine(gen_OraFloat32Slice(true), floatP126Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// MISC floatP126Null
////////////////////////////////////////////////////////////////////////////////

func TestMultiDefine_floatP126Null_session(t *testing.T) {
	testMultiDefine(gen_int64(), floatP126Null, t)
}

func TestWorkload_floatP126Null_session(t *testing.T) {
	testWorkload(floatP126Null, t)
}

func TestBindDefine_floatP126Null_nil_session(t *testing.T) {
	testBindDefine(nil, floatP126Null, t, nil)
}
