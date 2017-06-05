package ora

/*
#cgo CFLAGS: -Iodpi/src -Iodpi/include
#cgo LDFLAGS: -Lodpi/lib -lodpic -ldl

#include "dpiImpl.h"
*/
import "C"
import (
	"database/sql/driver"
	"fmt"
	"io"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

const fetchRowCount = 1 //<< 7
const maxArraySize = 1 << 10

type rows struct {
	*statement
	columns        []Column
	bufferRowIndex C.uint32_t
	fetched        C.uint32_t
	finished       bool
	vars           []*C.dpiVar
	data           [][]*C.dpiData
}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (r *rows) Columns() []string {
	names := make([]string, len(r.columns))
	for i, col := range r.columns {
		names[i] = col.Name
	}
	return names
}

// Close closes the rows iterator.
func (r *rows) Close() error {
	for _, v := range r.vars {
		C.dpiVar_release(v)
	}
	if C.dpiStmt_release(r.statement.dpiStmt) == C.DPI_FAILURE {
		return r.getError()
	}
	return nil
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
func (r *rows) Next(dest []driver.Value) error {
	if r.finished {
		return io.EOF
	}
	if r.fetched == 0 {
		var moreRows C.int
		if C.dpiStmt_fetchRows(r.dpiStmt, fetchRowCount, &r.bufferRowIndex, &r.fetched, &moreRows) == C.DPI_FAILURE {
			return r.getError()
		}
		if r.fetched == 0 {
			r.finished = moreRows == 0
			return io.EOF
		}
		//fmt.Printf("data=%#v\n", r.data)
	}
	//fmt.Printf("data=%#v\n", r.data)

	//fmt.Printf("bri=%d fetched=%d\n", r.bufferRowIndex, r.fetched)
	//fmt.Printf("data=%#v\n", r.data[0][r.bufferRowIndex])
	//fmt.Printf("VC=%d\n", C.DPI_ORACLE_TYPE_VARCHAR)
	for i, col := range r.columns {
		typ := col.Type
		d := r.data[i][r.bufferRowIndex]
		//fmt.Printf("data=%#v typ=%d\n", d, typ)
		if d.isNull == 1 {
			dest[i] = nil
			continue
		}

		switch typ {
		case C.DPI_ORACLE_TYPE_VARCHAR, C.DPI_ORACLE_TYPE_NVARCHAR,
			C.DPI_ORACLE_TYPE_CHAR, C.DPI_ORACLE_TYPE_NCHAR,
			C.DPI_ORACLE_TYPE_LONG_VARCHAR,
			C.DPI_NATIVE_TYPE_BYTES:
			//fmt.Printf("CHAR\n")
			b := C.dpiData_getBytes(d)
			if b.ptr == nil {
				dest[i] = ""
				continue
			}
			dest[i] = C.GoStringN(b.ptr, C.int(b.length))

		case C.DPI_ORACLE_TYPE_NUMBER:
			switch col.DefaultNumType {
			case C.DPI_NATIVE_TYPE_INT64:
				dest[i] = C.dpiData_getInt64(d)
			case C.DPI_NATIVE_TYPE_UINT64:
				dest[i] = C.dpiData_getUint64(d)
			case C.DPI_NATIVE_TYPE_FLOAT:
				dest[i] = C.dpiData_getFloat(d)
			case C.DPI_NATIVE_TYPE_DOUBLE:
				dest[i] = C.dpiData_getDouble(d)
			default:
				b := C.dpiData_getBytes(d)
				//fmt.Printf("b=%p[%d] t=%d i=%d\n", b.ptr, b.length, col.DefaultNumType, C.dpiData_getInt64(d))
				dest[i] = C.GoStringN(b.ptr, C.int(b.length))
			}

		case C.DPI_ORACLE_TYPE_ROWID, C.DPI_NATIVE_TYPE_ROWID,
			C.DPI_ORACLE_TYPE_RAW, C.DPI_ORACLE_TYPE_LONG_RAW:
			fmt.Printf("RAW\n")
			b := C.dpiData_getBytes(d)
			dest[i] = C.GoBytes(unsafe.Pointer(b.ptr), C.int(b.length))
		case C.DPI_ORACLE_TYPE_NATIVE_FLOAT, C.DPI_NATIVE_TYPE_FLOAT:
			fmt.Printf("FLOAT\n")
			dest[i] = float32(C.dpiData_getFloat(d))
		case C.DPI_ORACLE_TYPE_NATIVE_DOUBLE, C.DPI_NATIVE_TYPE_DOUBLE:
			fmt.Printf("DOUBLE\n")
			dest[i] = float64(C.dpiData_getDouble(d))
		case C.DPI_ORACLE_TYPE_NATIVE_INT, C.DPI_NATIVE_TYPE_INT64:
			fmt.Printf("INT\n")
			dest[i] = int64(C.dpiData_getInt64(d))
		case C.DPI_ORACLE_TYPE_NATIVE_UINT, C.DPI_NATIVE_TYPE_UINT64:
			fmt.Printf("UINT\n")
			dest[i] = uint64(C.dpiData_getUint64(d))
		case C.DPI_ORACLE_TYPE_TIMESTAMP,
			C.DPI_ORACLE_TYPE_TIMESTAMP_TZ, C.DPI_ORACLE_TYPE_TIMESTAMP_LTZ,
			C.DPI_NATIVE_TYPE_TIMESTAMP,
			C.DPI_ORACLE_TYPE_DATE:
			//fmt.Printf("TS\n")
			ts := C.dpiData_getTimestamp(d)
			tz := time.Local
			if col.Type != C.DPI_ORACLE_TYPE_TIMESTAMP && col.Type != C.DPI_ORACLE_TYPE_DATE {
				tz = time.FixedZone(
					fmt.Sprintf("%02d:%02d", ts.tzHourOffset, ts.tzMinuteOffset),
					int(ts.tzHourOffset)*3600+int(ts.tzMinuteOffset)*60,
				)
			}
			dest[i] = time.Date(int(ts.year), time.Month(ts.month), int(ts.day), int(ts.hour), int(ts.minute), int(ts.second), int(ts.fsecond), tz)
		case C.DPI_ORACLE_TYPE_INTERVAL_DS, C.DPI_NATIVE_TYPE_INTERVAL_DS:
			fmt.Printf("INTERVAL_DS\n")
			ds := C.dpiData_getIntervalDS(d)
			dest[i] = time.Duration(ds.days)*24*time.Hour +
				time.Duration(ds.hours)*time.Hour +
				time.Duration(ds.minutes)*time.Minute +
				time.Duration(ds.seconds)*time.Second +
				time.Duration(ds.fseconds)
		case C.DPI_ORACLE_TYPE_INTERVAL_YM, C.DPI_NATIVE_TYPE_INTERVAL_YM:
			fmt.Printf("FLOAT\n")
			ym := C.dpiData_getIntervalYM(d)
			dest[i] = fmt.Sprintf("%dy%dm", ym.years, ym.months)
		case C.DPI_ORACLE_TYPE_CLOB, C.DPI_ORACLE_TYPE_NCLOB,
			C.DPI_ORACLE_TYPE_BLOB,
			C.DPI_ORACLE_TYPE_BFILE,
			C.DPI_NATIVE_TYPE_LOB:
			fmt.Printf("INTERVAL_YM\n")
			dest[i] = &Lob{dpiLob: C.dpiData_getLOB(d)}
		case C.DPI_ORACLE_TYPE_STMT, C.DPI_NATIVE_TYPE_STMT:
			fmt.Printf("STMT\n")
			st := &statement{dpiStmt: C.dpiData_getStmt(d)}
			var colCount C.uint32_t
			if C.dpiStmt_getNumQueryColumns(st.dpiStmt, &colCount) == C.DPI_FAILURE {
				return r.getError()
			}
			r2, err := st.openRows(int(colCount))
			if err != nil {
				return err
			}
			dest[i] = r2
		case C.DPI_ORACLE_TYPE_BOOLEAN, C.DPI_NATIVE_TYPE_BOOLEAN:
			fmt.Printf("BOOL\n")
			dest[i] = C.dpiData_getBool(d) == 1
			//case C.DPI_ORACLE_TYPE_OBJECT: //Default type used for named type columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
		default:
			fmt.Printf("OTHER(%d)\n", typ)
			return errors.Errorf("unsupported column type %d", typ)
		}

		//fmt.Printf("dest[%d]=%#v\n", i, dest[i])
	}
	r.bufferRowIndex++
	r.fetched--

	return nil
}

type Lob struct {
	dpiLob *C.dpiLob
}
