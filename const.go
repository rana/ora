// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// ColumnGoType defines the Go type returned from a sql select column.
type GoColumnType uint

// go column types
const (
	// D defines a sql select column based on its default mapping.
	D GoColumnType = iota
	// I64 defines a sql select column as a Go int64.
	I64
	// I32 defines a sql select column as a Go int32.
	I32
	// I16 defines a sql select column as a Go int16.
	I16
	// I8 defines a sql select column as a Go int8.
	I8
	// U64 defines a sql select column as a Go uint64.
	U64
	// U32 defines a sql select column as a Go uint32.
	U32
	// U16 defines a sql select column as a Go uint16.
	U16
	// U8 defines a sql select column as a Go uint8.
	U8
	// F64 defines a sql select column as a Go float64.
	F64
	// F32 defines a sql select column as a Go float32.
	F32
	// OraI64 defines a sql select column as a nullable Go ora.Int64.
	OraI64
	// OraI32 defines a sql select column as a nullable Go ora.Int32.
	OraI32
	// OraI16 defines a sql select column as a nullable Go ora.Int16.
	OraI16
	// OraI8 defines a sql select column as a nullable Go ora.Int8.
	OraI8
	// OraU64 defines a sql select column as a nullable Go ora.Uint64.
	OraU64
	// OraU32 defines a sql select column as a nullable Go ora.Uint32.
	OraU32
	// OraU16 defines a sql select column as a nullable Go ora.Uint16.
	OraU16
	// OraU8 defines a sql select column as a nullable Go ora.Uint8.
	OraU8
	// OraF64 defines a sql select column as a nullable Go ora.Float64.
	OraF64
	// OraF32 defines a sql select column as a nullable Go ora.Float32.
	OraF32
	// T defines a sql select column as a Go time.Time.
	T
	// OraT defines a sql select column as a nullable Go ora.Time.
	OraT
	// S defines a sql select column as a Go string.
	S
	// OraS defines a sql select column as a nullable Go ora.String.
	OraS
	// B defines a sql select column as a Go bool.
	B
	// OraB defines a sql select column as a nullable Go ora.Bool.
	OraB
	// Bin defines a sql select column or bind parmeter as a Go byte slice.
	Bin
	// OraBin defines a sql select column as a nullable Go ora.Binary.
	OraBin
)

// bind pool indexes
const (
	bndIdxInt64 int = iota
	bndIdxInt32
	bndIdxInt16
	bndIdxInt8
	bndIdxUint64
	bndIdxUint32
	bndIdxUint16
	bndIdxUint8
	bndIdxFloat64
	bndIdxFloat32

	bndIdxInt64Ptr
	bndIdxInt32Ptr
	bndIdxInt16Ptr
	bndIdxInt8Ptr
	bndIdxUint64Ptr
	bndIdxUint32Ptr
	bndIdxUint16Ptr
	bndIdxUint8Ptr
	bndIdxFloat64Ptr
	bndIdxFloat32Ptr

	bndIdxInt64Slice
	bndIdxInt32Slice
	bndIdxInt16Slice
	bndIdxInt8Slice
	bndIdxUint64Slice
	bndIdxUint32Slice
	bndIdxUint16Slice
	bndIdxUint8Slice
	bndIdxFloat64Slice
	bndIdxFloat32Slice

	bndIdxTime
	bndIdxTimePtr
	bndIdxTimeSlice

	bndIdxDate
	bndIdxDatePtr
	bndIdxDateSlice

	bndIdxString
	bndIdxStringPtr
	bndIdxStringSlice

	bndIdxBool
	bndIdxBoolPtr
	bndIdxBoolSlice

	bndIdxBin
	bndIdxBinSlice
	bndIdxLob
	bndIdxLobPtr
	bndIdxLobSlice

	bndIdxIntervalYM
	bndIdxIntervalYMSlice
	bndIdxIntervalDS
	bndIdxIntervalDSSlice

	bndIdxBfile
	bndIdxRset
	bndIdxNil
)

// define pool indexes
const (
	defIdxInt64 int = iota
	defIdxInt32
	defIdxInt16
	defIdxInt8
	defIdxUint64
	defIdxUint32
	defIdxUint16
	defIdxUint8
	defIdxFloat64
	defIdxFloat32

	defIdxTime
	defIdxDate
	defIdxString
	defIdxBool

	defIdxLob
	defIdxRaw
	defIdxLongRaw

	defIdxIntervalYM
	defIdxIntervalDS
	defIdxBfile
	defIdxRowid
)
