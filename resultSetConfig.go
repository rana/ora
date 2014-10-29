// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// ResultSetConfig affects the association of Oracle select-list columns to
// Go types.
type ResultSetConfig struct {
	numberScaless GoColumnType
	numberScaled  GoColumnType
	binaryDouble  GoColumnType
	binaryFloat   GoColumnType
	float         GoColumnType
	date          GoColumnType
	timestamp     GoColumnType
	timestampTz   GoColumnType
	timestampLtz  GoColumnType
	char1         GoColumnType
	char          GoColumnType
	varchar       GoColumnType
	long          GoColumnType
	clob          GoColumnType
	blob          GoColumnType
	raw           GoColumnType
	longRaw       GoColumnType

	// TrueRune is rune a Go bool true value from SQL select-list character column.
	//
	// The is default is '1'.
	TrueRune rune
}

// NewResultSetConfig returns a ResultSetConfig with default values.
func NewResultSetConfig() ResultSetConfig {
	var rsc ResultSetConfig
	rsc.Reset()
	return rsc
}

// Reset sets driver-defined values to all fields.
func (resultSetConfig *ResultSetConfig) Reset() {
	resultSetConfig.TrueRune = '1'

	resultSetConfig.numberScaless = I64
	resultSetConfig.numberScaled = F64
	resultSetConfig.binaryDouble = F64
	resultSetConfig.binaryFloat = F32
	resultSetConfig.float = F64
	resultSetConfig.date = T
	resultSetConfig.timestamp = T
	resultSetConfig.timestampTz = T
	resultSetConfig.timestampLtz = T
	resultSetConfig.char1 = B
	resultSetConfig.char = S
	resultSetConfig.varchar = S
	resultSetConfig.long = S
	resultSetConfig.clob = S
	resultSetConfig.blob = Bits
	resultSetConfig.raw = Bits
	resultSetConfig.longRaw = Bits
}

// SetNumberScaless sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetNumberScaless(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct)
	if err == nil {
		resultSetConfig.numberScaless = gct
	}
	return err
}

// NumberScaless returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with scale zero.
//
// The default is I64.
//
// The database/sql package uses NumberScaless.
//
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, NumberScaless is used.
func (resultSetConfig *ResultSetConfig) NumberScaless() GoColumnType {
	return resultSetConfig.numberScaless
}

// SetNumberScaled sets a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetNumberScaled(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct)
	if err == nil {
		resultSetConfig.numberScaled = gct
	}
	return err
}

// NumberScaled returns a GoColumnType associated to an Oracle select-list
// NUMBER column defined with a scale greater than zero.
//
// The default is F64.
//
// NumberScaled is used by the database/sql package.
//
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, NumberScaled is used.
func (resultSetConfig *ResultSetConfig) NumberScaled() GoColumnType {
	return resultSetConfig.numberScaled
}

// SetBinaryDouble sets a GoColumnType associated to an Oracle select-list
// BINARY_DOUBLE column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetBinaryDouble(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct)
	if err == nil {
		resultSetConfig.binaryDouble = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, BinaryDouble is used.
func (resultSetConfig *ResultSetConfig) BinaryDouble() GoColumnType {
	return resultSetConfig.binaryDouble
}

// SetBinaryFloat sets a GoColumnType associated to an Oracle select-list
// BINARY_FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetBinaryFloat(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct)
	if err == nil {
		resultSetConfig.binaryFloat = gct
	}
	return err
}

// BinaryFloat returns a GoColumnType associated to an Oracle select-list
// BINARY_FLOAT column.
//
// The default for the database/sql package is F64.
//
// The default for the oracle package is F32.
//
// BinaryFloat is used by the database/sql package.
//
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, BinaryFloat is used.
func (resultSetConfig *ResultSetConfig) BinaryFloat() GoColumnType {
	return resultSetConfig.binaryFloat
}

// SetFloat sets a GoColumnType associated to an Oracle select-list
// FLOAT column.
//
// Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64,
// OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32.
//
// Returns an error if a non-numeric GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetFloat(gct GoColumnType) (err error) {
	err = checkNumericColumn(gct)
	if err == nil {
		resultSetConfig.float = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Float is used.
func (resultSetConfig *ResultSetConfig) Float() GoColumnType {
	return resultSetConfig.float
}

// SetDate sets a GoColumnType associated to an Oracle select-list
// DATE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetDate(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		resultSetConfig.date = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Date is used.
func (resultSetConfig *ResultSetConfig) Date() GoColumnType {
	return resultSetConfig.date
}

// SetTimestamp sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetTimestamp(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		resultSetConfig.timestamp = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Timestamp is used.
func (resultSetConfig *ResultSetConfig) Timestamp() GoColumnType {
	return resultSetConfig.timestamp
}

// SetTimestampTz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetTimestampTz(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		resultSetConfig.timestampTz = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, TimestampTz is used.
func (resultSetConfig *ResultSetConfig) TimestampTz() GoColumnType {
	return resultSetConfig.timestampTz
}

// SetTimestampLtz sets a GoColumnType associated to an Oracle select-list
// TIMESTAMP WITH LOCAL TIME ZONE column.
//
// Valid values are T and OraT.
//
// Returns an error if a non-time GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetTimestampLtz(gct GoColumnType) (err error) {
	err = checkTimeColumn(gct)
	if err == nil {
		resultSetConfig.timestampLtz = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, TimestampLtz is used.
func (resultSetConfig *ResultSetConfig) TimestampLtz() GoColumnType {
	return resultSetConfig.timestampLtz
}

// SetChar1 sets a GoColumnType associated to an Oracle select-list
// CHAR column with length 1 and NCHAR column with length 1.
//
// Valid values are B, OraB, S and OraS.
//
// Returns an error if a non-bool or non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetChar1(gct GoColumnType) (err error) {
	err = checkBoolOrStringColumn(gct)
	if err == nil {
		resultSetConfig.char1 = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Char1 is used.
func (resultSetConfig *ResultSetConfig) Char1() GoColumnType {
	return resultSetConfig.char1
}

// SetChar sets a GoColumnType associated to an Oracle select-list
// CHAR column and NCHAR column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetChar(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		resultSetConfig.char = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Char is used.
func (resultSetConfig *ResultSetConfig) Char() GoColumnType {
	return resultSetConfig.char
}

// SetVarchar sets a GoColumnType associated to an Oracle select-list
// VARCHAR column, VARCHAR2 column and NVARCHAR2 column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetVarchar(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		resultSetConfig.varchar = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Varchar is used.
func (resultSetConfig *ResultSetConfig) Varchar() GoColumnType {
	return resultSetConfig.varchar
}

// SetLong sets a GoColumnType associated to an Oracle select-list
// LONG column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetLong(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		resultSetConfig.long = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Long is used.
func (resultSetConfig *ResultSetConfig) Long() GoColumnType {
	return resultSetConfig.long
}

// SetClob sets a GoColumnType associated to an Oracle select-list
// CLOB column and NCLOB column.
//
// Valid values are S and OraS.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetClob(gct GoColumnType) (err error) {
	err = checkStringColumn(gct)
	if err == nil {
		resultSetConfig.clob = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Clob is used.
func (resultSetConfig *ResultSetConfig) Clob() GoColumnType {
	return resultSetConfig.clob
}

// SetBlob sets a GoColumnType associated to an Oracle select-list
// BLOB column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetBlob(gct GoColumnType) (err error) {
	err = checkBitsColumn(gct)
	if err == nil {
		resultSetConfig.blob = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Blob is used.
func (resultSetConfig *ResultSetConfig) Blob() GoColumnType {
	return resultSetConfig.blob
}

// SetRaw sets a GoColumnType associated to an Oracle select-list
// RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetRaw(gct GoColumnType) (err error) {
	err = checkBitsColumn(gct)
	if err == nil {
		resultSetConfig.raw = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, Raw is used.
func (resultSetConfig *ResultSetConfig) Raw() GoColumnType {
	return resultSetConfig.raw
}

// SetLongRaw sets a GoColumnType associated to an Oracle select-list
// LONG RAW column.
//
// Valid values are Bits and OraBits.
//
// Returns an error if a non-string GoColumnType is specified.
func (resultSetConfig *ResultSetConfig) SetLongRaw(gct GoColumnType) (err error) {
	err = checkBitsColumn(gct)
	if err == nil {
		resultSetConfig.longRaw = gct
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
// When using the oracle package directly, custom GoColumnType associations may
// be specified to the Session.Prepare method. If no custom GoColumnType association
// is specified, LongRaw is used.
func (resultSetConfig *ResultSetConfig) LongRaw() GoColumnType {
	return resultSetConfig.longRaw
}
