// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// RsetCfg affects the association of Oracle select-list columns to
// Go types.
//
// Though it is unlucky, an empty RsetCfg is unusable!
// Please use NewRsetCfg().
//
// RsetCfg is immutable, so all Set... methods returns a new copy!
type RsetCfg struct {
	numberInt      GoColumnType
	numberBigInt   GoColumnType
	numberFloat    GoColumnType
	numberBigFloat GoColumnType
	binaryDouble   GoColumnType
	binaryFloat    GoColumnType
	float          GoColumnType
	date           GoColumnType
	timestamp      GoColumnType
	timestampTz    GoColumnType
	timestampLtz   GoColumnType
	char1          GoColumnType
	char           GoColumnType
	varchar        GoColumnType
	long           GoColumnType
	clob           GoColumnType
	blob           GoColumnType
	raw            GoColumnType
	longRaw        GoColumnType

	// TrueRune is rune a Go bool true value from SQL select-list character column.
	//
	// The is default is '1'.
	TrueRune rune

	// Err is the error from the last Set... method.
	Err error
}

func (c RsetCfg) IsZero() bool { return c.numberInt == 0 }

// NewRsetCfg returns a RsetCfg with default values.
func NewRsetCfg() RsetCfg {
	var c RsetCfg
	c.numberInt = I64
	c.numberBigInt = N
	c.numberFloat = F64
	c.numberBigFloat = N
	c.binaryDouble = F64
	c.binaryFloat = F32
	c.float = F64
	c.date = T
	c.timestamp = T
	c.timestampTz = T
	c.timestampLtz = T
	c.char1 = S
	c.char = S
	c.varchar = S
	c.long = S
	c.clob = S
	c.blob = Bin
	c.raw = Bin
	c.longRaw = Bin

	c.TrueRune = '1'
	return c
}

// SetNumberInt sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero and precision <= 19.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetNumberInt(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.numberInt = gct
	return c
}

// NumberInt returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero and precision <= 19.
//
// The default is I64.
//
// The database/sql package uses NumberInt.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, NumberInt is used.
func (c RsetCfg) NumberInt() GoColumnType {
	return c.numberInt
}

// SetNumberBigInt sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero and precision unknown or > 19.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetNumberBigInt(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.numberBigInt = gct
	return c
}

// NumberBigInt returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero and precision unknown or > 19.
//
// The default is N.
//
// The database/sql package uses NumberBigInt.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, NumberInt is used.
func (c RsetCfg) NumberBigInt() GoColumnType {
	return c.numberBigInt
}

// SetNumberFloat sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero and precision <= 15.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetNumberFloat(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.numberFloat = gct
	return c
}

// NumberFloat returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero.
//
// The default is F64.
//
// NumberFloat is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, NumberFloat is used.
func (c RsetCfg) NumberFloat() GoColumnType {
	return c.numberFloat
}

// SetNumberBigFloat sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero and precision unkonw or > 15.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetNumberBigFloat(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.numberBigFloat = gct
	return c
}

// NumberBigFloat returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero and precision unknown or > 15.
//
// The default is N.
//
// NumberBugFloat is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, NumberFloat is used.
func (c RsetCfg) NumberBigFloat() GoColumnType {
	return c.numberBigFloat
}

// SetBinaryDouble sets a GoColumnType associated to an Oracle select-list
// BINARY_DOUBLE column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetBinaryDouble(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.binaryDouble = gct
	return c
}

// BinaryDouble returns a GoColumnType associated to an Oracle select-list
// BINARY_DOUBLE column.
//
// The default is F64.
//
// BinaryDouble is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, BinaryDouble is used.
func (c RsetCfg) BinaryDouble() GoColumnType {
	return c.binaryDouble
}

// SetBinaryFloat sets a GoColumnType associated to an Oracle select-list
// BINARY_FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// Num, OraNum.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetBinaryFloat(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.binaryFloat = gct
	return c
}

// BinaryFloat returns a GoColumnType associated to an Oracle select-list
// BINARY_FLOAT column.
//
// The default for the database/sql package is F64.
//
// The default for the ora package is F32.
//
// BinaryFloat is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, BinaryFloat is used.
func (c RsetCfg) BinaryFloat() GoColumnType {
	return c.binaryFloat
}

// SetFloat sets a GoColumnType associated to an Oracle select-list
// FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32,
// N, OraN.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c RsetCfg) SetFloat(gct GoColumnType) RsetCfg {
	if err := checkNumericColumn(gct, ""); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.float = gct
	return c
}

// Float returns a GoColumnType associated to an Oracle select-list
// FLOAT column.
//
// The default is F64.
//
// Float is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Float is used.
func (c RsetCfg) Float() GoColumnType {
	return c.float
}

// SetDate sets a GoColumnType associated to an Oracle select-list
// DATE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c RsetCfg) SetDate(gct GoColumnType) RsetCfg {
	if err := checkTimeColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.date = gct
	return c
}

// Date returns a GoColumnType associated to an Oracle select-list
// DATE column.
//
// The default is T.
//
// Date is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Date is used.
func (c RsetCfg) Date() GoColumnType {
	return c.date
}

// SetTimestamp sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c RsetCfg) SetTimestamp(gct GoColumnType) RsetCfg {
	if err := checkTimeColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		c.timestamp = gct
	}
	return c
}

// Timestamp returns a GoColumnType associated to an Oracle select-list
// TIMESTAMP column.
//
// The default is T.
//
// Timestamp is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Timestamp is used.
func (c RsetCfg) Timestamp() GoColumnType {
	return c.timestamp
}

// SetTimestampTz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c RsetCfg) SetTimestampTz(gct GoColumnType) RsetCfg {
	if err := checkTimeColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.timestampTz = gct
	return c
}

// TimestampTz returns a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH TIME ZONE column.
//
// The default is T.
//
// TimestampTz is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, TimestampTz is used.
func (c RsetCfg) TimestampTz() GoColumnType {
	return c.timestampTz
}

// SetTimestampLtz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH LOCAL TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c RsetCfg) SetTimestampLtz(gct GoColumnType) RsetCfg {
	if err := checkTimeColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.timestampLtz = gct
	return c
}

// TimestampLtz returns a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH LOCAL TIME ZONE column.
//
// The default is T.
//
// TimestampLtz is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, TimestampLtz is used.
func (c RsetCfg) TimestampLtz() GoColumnType {
	return c.timestampLtz
}

// SetChar1 sets a GoColumnType associated to an Oracle select-list
// CHAR column with length 1 and NCHAR column with length 1.
//
// Valid values are B, OraB, S and OraS.
//
// Returns an error if a non-bool or non-string GoColumnType is specified.
func (c RsetCfg) SetChar1(gct GoColumnType) RsetCfg {
	if err := checkBoolOrStringColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.char1 = gct
	return c
}

// Char1 returns a GoColumnType associated to an Oracle select-list
// CHAR column with length 1 and NCHAR column with length 1.
//
// The default is B.
//
// Char1 is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Char1 is used.
func (c RsetCfg) Char1() GoColumnType {
	return c.char1
}

// SetChar sets a GoColumnType associated to an Oracle select-list
// CHAR column and NCHAR column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetChar(gct GoColumnType) RsetCfg {
	if err := checkStringColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.char = gct
	return c
}

// Char returns a GoColumnType associated to an Oracle select-list
// CHAR column and NCHAR column.
//
// The default is S.
//
// Char is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Char is used.
func (c RsetCfg) Char() GoColumnType {
	return c.char
}

// SetVarchar sets a GoColumnType associated to an Oracle select-list
// VARCHAR column, VARCHAR2 column and NVARCHAR2 column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetVarchar(gct GoColumnType) RsetCfg {
	if err := checkStringColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.varchar = gct
	return c
}

// Varchar returns a GoColumnType associated to an Oracle select-list
// VARCHAR column, VARCHAR2 column and NVARCHAR2 column.
//
// The default is S.
//
// Varchar is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Varchar is used.
func (c RsetCfg) Varchar() GoColumnType {
	return c.varchar
}

// SetLong sets a GoColumnType associated to an Oracle select-list
// LONG column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetLong(gct GoColumnType) RsetCfg {
	if err := checkStringColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.long = gct
	return c
}

// Long returns a GoColumnType associated to an Oracle select-list
// LONG column.
//
// The default is S.
//
// Long is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Long is used.
func (c RsetCfg) Long() GoColumnType {
	return c.long
}

// SetClob sets a GoColumnType associated to an Oracle select-list
// CLOB column and NCLOB column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetClob(gct GoColumnType) RsetCfg {
	if gct != D && gct != L {
		if err := checkStringColumn(gct); err != nil {
			if c.Err == nil {
				c.Err = err
			}
			return c
		}
	}
	c.clob = gct
	return c
}

// Clob returns a GoColumnType associated to an Oracle select-list
// CLOB column and NCLOB column.
//
// The default is S.
//
// Clob is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Clob is used.
func (c RsetCfg) Clob() GoColumnType {
	return c.clob
}

// SetBlob sets a GoColumnType associated to an Oracle select-list
// BLOB column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetBlob(gct GoColumnType) RsetCfg {
	if gct != D && gct != L {
		if err := checkBinColumn(gct); err != nil {
			if c.Err == nil {
				c.Err = err
			}
			return c
		}
	}
	c.blob = gct
	return c
}

// Blob returns a GoColumnType associated to an Oracle select-list
// BLOB column.
//
// The default is Bits.
//
// Blob is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Blob is used.
func (c RsetCfg) Blob() GoColumnType {
	return c.blob
}

// SetRaw sets a GoColumnType associated to an Oracle select-list
// RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetRaw(gct GoColumnType) RsetCfg {
	if err := checkBinColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.raw = gct
	return c
}

// Raw returns a GoColumnType associated to an Oracle select-list
// RAW column.
//
// The default is Bits.
//
// Raw is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, Raw is used.
func (c RsetCfg) Raw() GoColumnType {
	return c.raw
}

// SetLongRaw sets a GoColumnType associated to an Oracle select-list
// LONG RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c RsetCfg) SetLongRaw(gct GoColumnType) RsetCfg {
	if err := checkBinColumn(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.longRaw = gct
	return c
}

// LongRaw returns a GoColumnType associated to an Oracle select-list
// LONG RAW column.
//
// The default is Bits.
//
// LongRaw is used by the database/sql package.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, LongRaw is used.
func (c RsetCfg) LongRaw() GoColumnType {
	return c.longRaw
}

// numericColumnType returns the GoColumnType for the NUMBER/INTEGER
// column, based on precision and scale.
//
// See issue #33 and #36 for the reason this became a testable separate function.
func (c RsetCfg) numericColumnType(precision, scale int) (gct GoColumnType) {
	//defer func() {
	//    fmt.Printf("numericColumnType(%d, %d): %s\n", precision, scale, gct)
	//}()

	// If the precision is zero and scale is -127, the it is a NUMBER;
	// if the precision is nonzero and scale is -127, then it is a FLOAT;
	// if the scale is positive, then it is a NUMBER(precision, scale);
	// otherwise, it's an int.
	if precision != 0 {
		if scale == 0 {
			if precision <= 19 {
				return c.numberInt
			}
			return c.numberBigInt
		}
		if precision <= 15 {
			return c.numberFloat
		}
		return c.numberBigFloat
	}
	if scale == -127 {
		return c.float
	}
	return c.numberBigFloat
}
