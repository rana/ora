package ora

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// checkNumericColumn returns nil when the column type is numeric; otherwise, an error.
func checkNumericColumn(gct GoColumnType, columnName string) error {
	switch gct {
	case I64, I32, I16, I8,
		OraI64, OraI32, OraI16, OraI8,
		U64, U32, U16, U8,
		OraU64, OraU32, OraU16, OraU8,
		F64, F32,
		OraF64, OraF32,
		N, OraN,
		S:
		return nil
	}
	var s string
	if columnName != "" {
		s = fmt.Sprintf(" (%s)", columnName)
	}
	return errF("Invalid go column type (%v) specified for numeric sql column%s. Expected go column type I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32, OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N or OraN.", GctName(gct), s)
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
	pc, _, _, _ := runtime.Caller(depth + 1)
	// get file name without path or suffix
	method := runtime.FuncForPC(pc).Name()
	n := strings.LastIndex(method, ")")
	m := strings.LastIndex(method, "*")
	if n < 0 {
		m = strings.LastIndex(method, "(")
	}
	if n < 0 { // main.func·015
		return fmt.Sprintf("[%v]", method)
	}
	// main.(*core).open
	return fmt.Sprintf("[%v.%v]", method[m+1:n], method[n+2:])
}
func errInfo(depth int) fmt.Stringer {
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
		//return fmt.Sprintf("%v.%v", method[m+1:n], method[n+2:])
		return methodInfo{Type: method[m+1 : n], Method: method[n+2:]}
	}
	//return fmt.Sprintf("%v:%v:%v", file, line, method)
	return posInfo{File: file, Line: line, Method: method}
}

type posInfo struct {
	File   string
	Line   int
	Method string
}

func (p posInfo) String() string { return fmt.Sprintf("%v:%v:%v", p.File, p.Line, p.Method) }

type methodInfo struct {
	Type, Method string
}

func (m methodInfo) String() string { return m.Type + "." + m.Method }

// log writes a message with caller info.
func log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.Cfg().Log.Logger.Infof("%v", callInfo(1))
		} else {
			_drv.Cfg().Log.Logger.Infof("%v %v", callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with caller info.
func logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.Cfg().Log.Logger.Infof("%v", callInfo(1))
		} else {
			_drv.Cfg().Log.Logger.Infof("%v %v", callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}

// err creates an error with caller info.
func er(v ...interface{}) error {
	//err := errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprint(v...)))
	var err error
	if len(v) == 1 {
		err, _ = v[0].(error)
	}
	if err == nil {
		err = errors.New(fmt.Sprint(v...))
	}
	err = &oraErr{Caller: errInfo(1), Underlying: err}
	_drv.Cfg().Log.Logger.Errorln(err)
	return err
}

// errF creates a formatted error with caller info.
func errF(format string, v ...interface{}) error {
	//err := errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprintf(format, v...)))
	err := &oraErr{Caller: errInfo(1), Underlying: fmt.Errorf(format, v...)}
	_drv.Cfg().Log.Logger.Errorln(err)
	return err
}

var stackMu sync.Mutex
var stack = make([]byte, 4096)

func getStack() string {
	stackMu.Lock()
	defer stackMu.Unlock()
	n := runtime.Stack(stack, false)
	return string(stack[:n])
}

// errR creates a recovered error with caller info.
func errR(v ...interface{}) error {
	//err := errors.New(fmt.Sprintf("%v recovered: %v\n%s", errInfo(1), fmt.Sprint(v...), trace[:n]))
	err := &oraErr{
		Caller:     errInfo(1),
		Underlying: errors.New(fmt.Sprint(v...)),
		Trace:      getStack(),
	}
	_drv.Cfg().Log.Logger.Errorln(err)
	return err
}

// errE wraps an error with caller info.
func errE(e error) error {
	//err := errors.New(fmt.Sprintf("%v %v", errInfo(1), e.Error()))
	err := &oraErr{Caller: errInfo(1), Underlying: e}
	_drv.Cfg().Log.Logger.Errorln(err)
	return err
}

type oraErr struct {
	Caller     fmt.Stringer
	Underlying error
	Trace      string
}

func (e *oraErr) Error() string {
	if e == nil {
		return ""
	}
	if e.Caller != nil {
		if len(e.Trace) > 0 {
			return fmt.Sprintf("%v recovered: %v\n%s", e.Caller, e.Underlying, e.Trace)
		}
		return fmt.Sprintf("%v %v", e.Caller, e.Underlying)
	}
	return e.Underlying.Error()
}

func (e oraErr) Code() int {
	if e.Underlying == nil {
		return 0
	}
	if coder, ok := e.Underlying.(interface {
		Code() int
	}); ok {
		return coder.Code()
	}
	errS := e.Error()
	i := strings.Index(errS, "ORA-")
	if i < 0 {
		return 0
	}
	var code int
	fmt.Scanf(errS[i+4:], "%d", &code)
	return code
}

// DescribedColumn type for describing a column (see DescribeQuery).
type DescribedColumn struct {
	Column

	Schema                 string
	Nullable               bool
	CharsetID, CharsetForm int
}

// DescribeQuery parses the query and returns the column types, as
// DBMS_SQL.describe_column does.
func DescribeQuery(db *sql.DB, qry string) ([]DescribedColumn, error) {
	res := bytesPool.Get(32766)
	defer bytesPool.Put(res)
	for i := range res {
		res[i] = 0
	}
	res = make([]byte, 32766)
	if _, err := db.Exec(`DECLARE
  c INTEGER;
  col_cnt INTEGER;
  rec_tab DBMS_SQL.DESC_TAB;
  a DBMS_SQL.DESC_REC;
  v_idx PLS_INTEGER;
  res VARCHAR2(32767);
BEGIN
  c := DBMS_SQL.OPEN_CURSOR;
  BEGIN
    DBMS_SQL.PARSE(c, :1, DBMS_SQL.NATIVE);
    DBMS_SQL.DESCRIBE_COLUMNS(c, col_cnt, rec_tab);
    v_idx := rec_tab.FIRST;
    WHILE v_idx IS NOT NULL LOOP
      a := rec_tab(v_idx);
      res := res||a.col_schema_name||' '||a.col_name||' '||a.col_type||' '||
                  a.col_max_len||' '||a.col_precision||' '||a.col_scale||' '||
                  (CASE WHEN a.col_null_ok THEN 1 ELSE 0 END)||' '||
                  a.col_charsetid||' '||a.col_charsetform||
                  CHR(10);
      v_idx := rec_tab.NEXT(v_idx);
    END LOOP;
  EXCEPTION WHEN OTHERS THEN NULL;
    DBMS_SQL.CLOSE_CURSOR(c);
	RAISE;
  END;
  :2 := UTL_RAW.CAST_TO_RAW(res);
END;`, qry, &res,
	); err != nil {
		return nil, err
	}
	if i := bytes.IndexByte(res, 0); i >= 0 {
		res = res[:i]
	}
	lines := bytes.Split(res, []byte{'\n'})
	cols := make([]DescribedColumn, 0, len(lines))
	var nullable int
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var col DescribedColumn
		switch j := bytes.IndexByte(line, ' '); j {
		case -1:
			continue
		case 0:
			line = line[1:]
		default:
			col.Schema, line = string(line[:j]), line[j+1:]
		}
		if n, err := fmt.Sscanf(string(line), "%s %d %d %d %d %d %d %d",
			&col.Name, &col.Type, &col.Length, &col.Precision, &col.Scale, &nullable, &col.CharsetID, &col.CharsetForm,
		); err != nil {
			return cols, fmt.Errorf("parsing %q (parsed: %d): %v", line, n, err)
		}
		col.Nullable = nullable != 0
		cols = append(cols, col)
	}
	return cols, nil
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
	SELECT user owner, name, type, line, position, message_number, text, attribute
		FROM user_errors
		ORDER BY name, sequence`)
	if err != nil {
		return nil, err
	}
	var errors []CompileError
	for rset.Next() {
		warn := rset.Row[7].(string) == "WARNING"
		if warn && !all {
			continue
		}
		if len(rset.Row) != 8 {
			panic(fmt.Sprintf("rset.Row=%#v", rset.Row))
		}
		errors = append(errors,
			CompileError{
				Owner:    rset.Row[0].(string),
				Name:     rset.Row[1].(string),
				Type:     rset.Row[2].(string),
				Line:     int64(rset.Row[3].(float64)),
				Position: int64(rset.Row[4].(float64)),
				Code:     int64(rset.Row[5].(float64)),
				Text:     rset.Row[6].(string),
				Warning:  warn,
			})
	}
	return errors, nil
}

type sysNamer struct {
	once sync.Once
	name string
}

// Name sets the name to the result of calc once,
// then returns that result forever.
// (Effectively caches the result of calc().)
func (s *sysNamer) Name(calc func() string) string {
	s.once.Do(func() { s.name = calc() })
	return s.name
}

var bytesPool bytesArena

const (
	bytesArenaOffset       = 10
	bytesArenaMaxPoolCount = 16
)

type bytesArena struct {
	sync.Mutex
	pools []*sync.Pool
}

func (bp *bytesArena) Get(n int) []byte {
	p := bp.poolOf(boundingPower(n))
	b := *(p.Get().(*[]byte))
	if cap(b) >= n {
		return b[:n]
	}
	p.Put(&b)
	return make([]byte, n)
}

func (bp *bytesArena) Put(p []byte) {
	if cap(p) < 1<<bytesArenaOffset {
		return
	}
	i := boundingPower(cap(p))
	if 1<<uint(i) > cap(p) {
		i--
	}
	if i < bytesArenaOffset {
		return
	}
	bp.poolOf(i).Put(&p)
}
func (bp *bytesArena) poolOf(j int) *sync.Pool {
	j -= bytesArenaOffset
	if j < 0 {
		j = 0
	} else if j > bytesArenaMaxPoolCount {
		j = bytesArenaMaxPoolCount
	}
	bp.Lock()
	defer bp.Unlock()
	if len(bp.pools) <= j {
		bp.pools = append(bp.pools, make([]*sync.Pool, j-len(bp.pools)+1)...)
	}
	p := bp.pools[j]
	if p != nil {
		return p
	}
	p = &sync.Pool{New: func() interface{} { z := make([]byte, 1<<(uint(j)+bytesArenaOffset)); return &z }}
	bp.pools[j] = p
	return p
}

func boundingPower(n int) int {
	var i int
	for j := 1; j < n; j <<= 1 {
		i++
	}
	return i
}
