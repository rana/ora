// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
)

// ColumnGoType defines the Go type returned from a sql select column.
type GoColumnType uint

const (
	// D defines a sql select column as a default Go type.
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
	// OraI64 defines a sql select column as a nullable Go oracle.Int64.
	OraI64
	// OraI32 defines a sql select column as a nullable Go oracle.Int32.
	OraI32
	// OraI16 defines a sql select column as a nullable Go oracle.Int16.
	OraI16
	// OraI8 defines a sql select column as a nullable Go oracle.Int8.
	OraI8
	// OraU64 defines a sql select column as a nullable Go oracle.Uint64.
	OraU64
	// OraU32 defines a sql select column as a nullable Go oracle.Uint32.
	OraU32
	// OraU16 defines a sql select column as a nullable Go oracle.Uint16.
	OraU16
	// OraU8 defines a sql select column as a nullable Go oracle.Uint8.
	OraU8
	// OraF64 defines a sql select column as a nullable Go oracle.Float64.
	OraF64
	// OraF32 defines a sql select column as a nullable Go oracle.Float32.
	OraF32
	// T defines a sql select column as a Go time.Time.
	T
	// OraT defines a sql select column as a nullable Go oracle.Time.
	OraT
	// S defines a sql select column as a Go string.
	S
	// OraS defines a sql select column as a nullable Go string.
	OraS
	// B defines a sql select column as a Go bool.
	B
	// OraB defines a sql select column as a nullable Go bool.
	OraB
	// Bits defines a sql select column or bind parmeter as a Go byte slice.
	Bits
	// OraBits defines a sql select column as a nullable Go oracle.Bytes.
	OraBits
)

const (
	// The Oracle driver name registered with the database/sql package.
	DriverName string = "oracle"
)

// An Oracle database driver.
//
// Implements the driver.Driver interface.
type Driver struct {
	environment *Environment
}

// Initalizes the driver
func init() {
	driver := &Driver{environment: NewEnvironment()}
	// database/sql/driver expects binaryFloat to return float64
	driver.environment.statementConfig.ResultSet.binaryFloat = F64
	sql.Register(DriverName, driver)
}

// Open starts a connection to an Oracle server.
//
// The connection string has the form username/password@dbname.
// dbname is a connection identifier such as a net service name,
// full connection identifier, or a simple connection identifier.
//
// Open is a member of the driver.Driver interface.
func (driver *Driver) Open(connStr string) (driver.Conn, error) {
	if !driver.environment.IsOpen() {
		driver.environment.Open()
	}
	connection, err := driver.environment.OpenConnection(connStr)
	if err != nil {
		return nil, err
	}

	return connection, nil
}

// checkNumericColumn returns nil when the column type is numeric; otherwise, an error.
func checkNumericColumn(gct GoColumnType) error {
	switch gct {
	case I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", gctName(gct))
}

// checkTimeColumn returns nil when the column type is time; otherwise, an error.
func checkTimeColumn(gct GoColumnType) error {
	switch gct {
	case T, OraT:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected T or OraT.", gctName(gct))
}

// checkStringColumn returns nil when the column type is string; otherwise, an error.
func checkStringColumn(gct GoColumnType) error {
	switch gct {
	case S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected S or OraS.", gctName(gct))
}

// checkBoolOrStringColumn returns nil when the column type is bool; otherwise, an error.
func checkBoolOrStringColumn(gct GoColumnType) error {
	switch gct {
	case B, OraB, S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected B, OraB, S, or OraS.", gctName(gct))
}

// checkBitsOrU8Column returns nil when the column type is Bits or U8; otherwise, an error.
func checkBitsOrU8Column(gct GoColumnType) error {
	switch gct {
	case Bits, U8:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected Bits or U8.", gctName(gct))
}

// checkBitsColumn returns nil when the column type is Bits or OraBits; otherwise, an error.
func checkBitsColumn(gct GoColumnType) error {
	switch gct {
	case Bits, OraBits:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected Bits or OraBits.", gctName(gct))
}

func gctName(gct GoColumnType) string {
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
	case Bits:
		return "Bits"
	case OraBits:
		return "OraBits"
	}
	return ""
}

func stringTrimmed(buffer []byte, pad byte) string {
	// Find length of non-padded string value
	// String buffer returned from Oracle is padded with Space char (32)
	var n int
	for n = len(buffer) - 1; n > -1; n-- {
		if buffer[n] != pad {
			n++
			break
		}
	}
	return string(buffer[:n])
}

func clear(buffer []byte, fill byte) {
	for n, _ := range buffer {
		buffer[n] = fill
	}
}

func errNew(str string) error {
	return errors.New("ora: " + str)
}

func errNewF(format string, a ...interface{}) error {
	return errNew(fmt.Sprintf(format, a...))
}
