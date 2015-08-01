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
		return errF("Invalid go column type (%v) specified for numeric sql column. Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", GctName(gct))
	} else {
		return errF("Invalid go column type (%v) specified for numeric sql column (%v). Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64 or OraF32.", GctName(gct), columnName)
	}
}

// checkTimeColumn returns nil when the column type is time; otherwise, an error.
func checkTimeColumn(gct GoColumnType) error {
	switch gct {
	case T, OraT:
		return nil
	}
	return errF("Invalid go column type (%v) specified for time-based sql column. Expected go column type T or OraT.", GctName(gct))
}

// checkStringColumn returns nil when the column type is string; otherwise, an error.
func checkStringColumn(gct GoColumnType) error {
	switch gct {
	case S, OraS:
		return nil
	}
	return errF("Invalid go column type (%v) specified for string-based sql column. Expected go column type S or OraS.", GctName(gct))
}

// checkBoolOrStringColumn returns nil when the column type is bool; otherwise, an error.
func checkBoolOrStringColumn(gct GoColumnType) error {
	switch gct {
	case B, OraB, S, OraS:
		return nil
	}
	return errF("Invalid go column type (%v) specified. Expected go column type B, OraB, S, or OraS.", GctName(gct))
}

// checkBinOrU8Column returns nil when the column type is Bin or U8; otherwise, an error.
func checkBinOrU8Column(gct GoColumnType) error {
	switch gct {
	case Bin, U8:
		return nil
	}
	return errF("Invalid go column type (%v) specified. Expected go column type Bin or U8.", GctName(gct))
}

// checkBitsColumn returns nil when the column type is Bin or OraBits; otherwise, an error.
func checkBinColumn(gct GoColumnType) error {
	switch gct {
	case Bin, OraBin:
		return nil
	}
	return errF("Invalid go column type (%v) specified. Expected go column type Bits or OraBits.", GctName(gct))
}

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
	}
	return ""
}

func clear(buffer []byte, fill byte) {
	for n := range buffer {
		buffer[n] = fill
	}
}

func errNew(str string) error {
	return errors.New("ora: " + str)
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

// log writes a message with caller info.
func log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v", callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v", callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with caller info.
func logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v", callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v", callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}

// err creates an error with caller info.
func er(v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprint(v...)))
	_drv.cfg.Log.Logger.Errorln(err)
	return err
}

// errF creates a formatted error with caller info.
func errF(format string, v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprintf(format, v...)))
	_drv.cfg.Log.Logger.Errorln(err)
	return err
}

// errR creates a recovered error with caller info.
func errR(v ...interface{}) (err error) {
	trace := make([]byte, 4096)
	n := runtime.Stack(trace, false)
	err = errors.New(fmt.Sprintf("%v recovered: %v\n%s",
		errInfo(1), fmt.Sprint(v...), trace[:n]))
	_drv.cfg.Log.Logger.Errorln(err)
	return err
}

// errE wraps an error with caller info.
func errE(e error) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), e.Error()))
	_drv.cfg.Log.Logger.Errorln(err)
	return err
}

// CompileError represents a compile-time error as in user_errors view.
type CompileError struct {
	Owner, Name, Type    string
	Line, Position, Code int64
	Text                 string
	Warning              bool
}

func (ce CompileError) Error() string {
	prefix := "ERROR "
	if ce.Warning {
		prefix = "WARN  "
	}
	return fmt.Sprintf("%s %s.%s %s %d:%d [%d] %s",
		prefix, ce.Owner, ce.Name, ce.Type, ce.Line, ce.Position, ce.Code, ce.Text)
}

// GetCompileErrors returns the slice of the errors in user_errors.
//
// If all is false, only errors are returned; otherwise, warnings, too.
func GetCompileErrors(ses *Ses, all bool) ([]CompileError, error) {
	rset, err := ses.PrepAndQry(`
	SELECT USER owner, name, type, line, position, message_number, text, attribute
		FROM user_errors
		ORDER BY name, sequence`)
	if err != nil {
		return nil, err
	}
	var errors []CompileError
	for rset.Next() {
		errors = append(errors,
			CompileError{
				Owner:    rset.Row[0].(string),
				Name:     rset.Row[1].(string),
				Type:     rset.Row[2].(string),
				Line:     rset.Row[3].(int64),
				Position: rset.Row[4].(int64),
				Code:     rset.Row[5].(int64),
				Text:     rset.Row[6].(string),
				Warning:  rset.Row[7].(string) == "WARNING"})
	}
	return errors, nil
}
