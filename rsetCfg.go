// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// RsetCfg affects the association of Oracle select-list columns to
// Go types.
type RsetCfg struct {
	numberInt    GoColumnType
	numberFloat  GoColumnType
	binaryDouble GoColumnType
	binaryFloat  GoColumnType
	float        GoColumnType
	date         GoColumnType
	timestamp    GoColumnType
	timestampTz  GoColumnType
	timestampLtz GoColumnType
	char1        GoColumnType
	char         GoColumnType
	varchar      GoColumnType
	long         GoColumnType
	clob         GoColumnType
	blob         GoColumnType
	raw          GoColumnType
	longRaw      GoColumnType

	// TrueRune is rune a Go bool true value from SQL select-list character column.
	//
	// The is default is '1'.
	TrueRune rune
}

// NewRsetCfg returns a RsetCfg with default values.
func NewRsetCfg() RsetCfg {
	c := RsetCfg{}
	c.numberInt = I64
	c.numberFloat = F64
	c.binaryDouble = F64
	c.binaryFloat = F32
	c.float = F64
	c.date = T
	c.timestamp = T
	c.timestampTz = T
	c.timestampLtz = T
	c.char1 = B
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
// NUMBER column defined with scale zero.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c *RsetCfg) SetNumberInt(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct, "")
	if err == nil {
		c.numberInt = gct
	}
	return err
}

// NumberInt returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero.
//
// The default is I64.
//
// The database/sql package uses NumberInt.
//
// When using the ora package directly, custom GoColumnType associations may
// be specified to the Ses.Prep method. If no custom GoColumnType association
// is specified, NumberInt is used.
func (c *RsetCfg) NumberInt() GoColumnType {
	return c.numberInt
}

// SetNumberFloat sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c *RsetCfg) SetNumberFloat(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct, "")
	if err == nil {
		c.numberFloat = gct
	}
	return err
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
func (c *RsetCfg) NumberFloat() GoColumnType {
	return c.numberFloat
}

// SetBinaryDouble sets a GoColumnType associated to an Oracle select-list
// BINARY_DOUBLE column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c *RsetCfg) SetBinaryDouble(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct, "")
	if err == nil {
		c.binaryDouble = gct
	}
	return err
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
func (c *RsetCfg) BinaryDouble() GoColumnType {
	return c.binaryDouble
}

// SetBinaryFloat sets a GoColumnType associated to an Oracle select-list
// BINARY_FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c *RsetCfg) SetBinaryFloat(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct, "")
	if err == nil {
		c.binaryFloat = gct
	}
	return err
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
func (c *RsetCfg) BinaryFloat() GoColumnType {
	return c.binaryFloat
}

// SetFloat sets a GoColumnType associated to an Oracle select-list
// FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (c *RsetCfg) SetFloat(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct, "")
	if err == nil {
		c.float = gct
	}
	return err
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
func (c *RsetCfg) Float() GoColumnType {
	return c.float
}

// SetDate sets a GoColumnType associated to an Oracle select-list
// DATE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c *RsetCfg) SetDate(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		c.date = gct
	}
	return err
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
func (c *RsetCfg) Date() GoColumnType {
	return c.date
}

// SetTimestamp sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c *RsetCfg) SetTimestamp(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		c.timestamp = gct
	}
	return err
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
func (c *RsetCfg) Timestamp() GoColumnType {
	return c.timestamp
}

// SetTimestampTz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c *RsetCfg) SetTimestampTz(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		c.timestampTz = gct
	}
	return err
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
func (c *RsetCfg) TimestampTz() GoColumnType {
	return c.timestampTz
}

// SetTimestampLtz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH LOCAL TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (c *RsetCfg) SetTimestampLtz(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		c.timestampLtz = gct
	}
	return err
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
func (c *RsetCfg) TimestampLtz() GoColumnType {
	return c.timestampLtz
}

// SetChar1 sets a GoColumnType associated to an Oracle select-list
// CHAR column with length 1 and NCHAR column with length 1.
//
// Valid values are B, OraB, S and OraS.
//
// Returns an error if a non-bool or non-string GoColumnType is specified.
func (c *RsetCfg) SetChar1(gct GoColumnType) (err error) {
	err = checkBoolOrStringColumn(gct)
	if err == nil {
		c.char1 = gct
	}
	return err
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
func (c *RsetCfg) Char1() GoColumnType {
	return c.char1
}

// SetChar sets a GoColumnType associated to an Oracle select-list
// CHAR column and NCHAR column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetChar(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		c.char = gct
	}
	return err
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
func (c *RsetCfg) Char() GoColumnType {
	return c.char
}

// SetVarchar sets a GoColumnType associated to an Oracle select-list
// VARCHAR column, VARCHAR2 column and NVARCHAR2 column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetVarchar(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		c.varchar = gct
	}
	return err
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
func (c *RsetCfg) Varchar() GoColumnType {
	return c.varchar
}

// SetLong sets a GoColumnType associated to an Oracle select-list
// LONG column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetLong(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		c.long = gct
	}
	return err
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
func (c *RsetCfg) Long() GoColumnType {
	return c.long
}

// SetClob sets a GoColumnType associated to an Oracle select-list
// CLOB column and NCLOB column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetClob(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		c.clob = gct
	}
	return err
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
func (c *RsetCfg) Clob() GoColumnType {
	return c.clob
}

// SetBlob sets a GoColumnType associated to an Oracle select-list
// BLOB column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetBlob(gct GoColumnType) (err error) {
	err = checkBinColumn(gct)
	if err == nil {
		c.blob = gct
	}
	return err
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
func (c *RsetCfg) Blob() GoColumnType {
	return c.blob
}

// SetRaw sets a GoColumnType associated to an Oracle select-list
// RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetRaw(gct GoColumnType) (err error) {
	err = checkBinColumn(gct)
	if err == nil {
		c.raw = gct
	}
	return err
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
func (c *RsetCfg) Raw() GoColumnType {
	return c.raw
}

// SetLongRaw sets a GoColumnType associated to an Oracle select-list
// LONG RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (c *RsetCfg) SetLongRaw(gct GoColumnType) (err error) {
	err = checkBinColumn(gct)
	if err == nil {
		c.longRaw = gct
	}
	return err
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
func (c *RsetCfg) LongRaw() GoColumnType {
	return c.longRaw
}
