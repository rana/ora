package ora

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// checkNumericColumn returns nil when the column type is numeric; otherwise, an error.
func checkNumericColumn(gct GoColumnType, columnName string) error {
	switch gct {
	case I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32:
		return nil
	}
	if columnName == "" {
		return errNewF("invalid go column type (%v) specified for numeric sql column. Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", gctName(gct))
	} else {
		return errNewF("invalid go column type (%v) specified for numeric sql column (%v). Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", gctName(gct), columnName)
	}
}

// checkTimeColumn returns nil when the column type is time; otherwise, an error.
func checkTimeColumn(gct GoColumnType) error {
	switch gct {
	case T, OraT:
		return nil
	}
	return errNewF("invalid go column type (%v) specified for time-based sql column. Expected go column type T or OraT.", gctName(gct))
}

// checkStringColumn returns nil when the column type is string; otherwise, an error.
func checkStringColumn(gct GoColumnType) error {
	switch gct {
	case S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified for string-based sql column. Expected go column type S or OraS.", gctName(gct))
}

// checkBoolOrStringColumn returns nil when the column type is bool; otherwise, an error.
func checkBoolOrStringColumn(gct GoColumnType) error {
	switch gct {
	case B, OraB, S, OraS:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type B, OraB, S, or OraS.", gctName(gct))
}

// checkBitsOrU8Column returns nil when the column type is Bits or U8; otherwise, an error.
func checkBitsOrU8Column(gct GoColumnType) error {
	switch gct {
	case Bin, U8:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type Bits or U8.", gctName(gct))
}

// checkBitsColumn returns nil when the column type is Bits or OraBits; otherwise, an error.
func checkBitsColumn(gct GoColumnType) error {
	switch gct {
	case Bin, OraBin:
		return nil
	}
	return errNewF("invalid go column type (%v) specified. Expected go column type Bits or OraBits.", gctName(gct))
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
	case Bin:
		return "Bits"
	case OraBin:
		return "OraBits"
	}
	return ""
}

func stringTrimmed(buffer []byte, pad byte) string {
	// Find length of non-padded string value
	// String buffer returned from Oracle is padded with Space char (32)
	//fmt.Println("stringTrimmed: len(buffer): ", len(buffer))
	var n int
	for n = len(buffer) - 1; n > -1; n-- {
		if buffer[n] != pad {
			n++
			break
		}
	}
	if n > 0 {
		return string(buffer[:n])
	}
	return ""
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

func errRecover(value interface{}) error {
	return errors.New(fmt.Sprintf("ora recovered: %v", value))
}

func recoverMsg(value interface{}) string {
	return fmt.Sprintf("recovered: %v", value)
}

func callInfo(depth int) string {
	// get caller method name; remove main. prefix
	pc, file, _, _ := runtime.Caller(depth + 1)
	// get file name without path or suffix
	file = file[strings.LastIndex(file, "/")+1 : len(file)-3]
	method := runtime.FuncForPC(pc).Name()
	n := strings.LastIndex(method, ")")
	m := strings.LastIndex(method, "*")
	if n < 0 {
		m = strings.LastIndex(method, "(")
	}
	if n < 0 { // main.func·015
		return fmt.Sprintf("[%v]", method)
	} else { // main.(*core).open
		return fmt.Sprintf("[%v.%v]", method[m+1:n], method[n+2:])
	}
}
func errInfo(depth int) string {
	// get caller method name; remove main. prefix
	pc, file, line, _ := runtime.Caller(depth + 1)
	file = file[strings.LastIndex(file, "/")+1:]
	method := runtime.FuncForPC(pc).Name()
	n := strings.LastIndex(method, ")")
	m := strings.LastIndex(method, "*")
	if n < 0 {
		m = strings.LastIndex(method, "(")
	}
	if n > -1 { // main.(*core).open
		return fmt.Sprintf("%v.%v", method[m+1:n], method[n+2:])
	}
	return fmt.Sprintf("%v:%v:%v", file, line, method)
}
