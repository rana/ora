// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// GoColumnType defines the Go type returned from a sql select column.
type GoColumnType uint

// go column types
const (
	// D defines a sql select column based on its default mapping.
	D GoColumnType = iota + 1
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
	// N defines a sql select column as a Go string for number.
	N
	// OraN defines a sql select column as a nullable Go string for number.
	OraN
	// L defins an sql select column as an ora.Lob.
	L
)

func GctName(gct GoColumnType) string {
	switch gct {
	case D:
		return "D"
	case I64:
		return "I64"
	case I32:
		return "I32"
	case I16:
		return "I16"
	case I8:
		return "I8"
	case U64:
		return "U64"
	case U32:
		return "U32"
	case U16:
		return "U16"
	case U8:
		return "U8"
	case F64:
		return "F64"
	case F32:
		return "F32"
	case OraI64:
		return "OraI64"
	case OraI32:
		return "OraI32"
	case OraI16:
		return "OraI16"
	case OraI8:
		return "OraI8"
	case OraU64:
		return "OraU64"
	case OraU32:
		return "OraU32"
	case OraU16:
		return "OraU16"
	case OraU8:
		return "OraU8"
	case OraF64:
		return "OraF64"
	case OraF32:
		return "OraF32"
	case T:
		return "T"
	case OraT:
		return "OraT"
	case S:
		return "S"
	case OraS:
		return "OraS"
	case B:
		return "B"
	case OraB:
		return "OraB"
	case Bin:
		return "Bin"
	case OraBin:
		return "OraBin"
	case N:
		return "N"
	case OraN:
		return "OraN"
	case L:
		return "L"
	}
	return ""
}

func (gct GoColumnType) String() string {
	return GctName(gct)
}

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
	bndIdxNumString
	bndIdxOCINum

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
	bndIdxNumStringPtr
	bndIdxOCINumPtr

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
	bndIdxNumStringSlice
	bndIdxOCINumSlice

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
	defIdxOCINum

	defIdxTime
	defIdxDate
	defIdxString
	defIdxNumString
	defIdxBool

	defIdxLob
	defIdxRaw
	defIdxLongRaw

	defIdxIntervalYM
	defIdxIntervalDS
	defIdxBfile
	defIdxRowid
	defIdxRset
)
