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

	"gopkg.in/rana/ora.v5/date"
	"gopkg.in/rana/ora.v5/num"
)

const fetchRowCount = 1 << 7
const maxArraySize = 1 << 10

type rows struct {
	*statement
	columns        []Column
	bufferRowIndex C.uint32_t
	fetched        C.uint32_t
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
	if r.fetched == 0 {
		var moreRows C.int
		if C.dpiStmt_fetchRows(r.dpiStmt, fetchRowCount, &r.bufferRowIndex, &r.fetched, &moreRows) == C.DPI_FAILURE {
			return r.getError()
		}
		if r.fetched == 0 {
			return io.EOF
		}
	}

	for i, col := range r.columns {
		d := r.data[i][r.bufferRowIndex]
		if d.isNull == 1 {
			dest[i] = nil
			continue
		}
		/*
			// structure used for transferring data to/from ODPI-C
			struct dpiData {
			    int isNull;
			    union {
			        int asBoolean;
			        int64_t asInt64;
			        uint64_t asUint64;
			        float asFloat;
			        double asDouble;
			        dpiBytes asBytes;
			        dpiTimestamp asTimestamp;
			        dpiIntervalDS asIntervalDS;
			        dpiIntervalYM asIntervalYM;
			        dpiLob *asLOB;
			        dpiObject *asObject;
			        dpiStmt *asStmt;
			        dpiRowid *asRowid;
			    } value;
			};
		*/
		switch col.Type {
		case C.DPI_ORACLE_TYPE_VARCHAR, C.DPI_ORACLE_TYPE_NVARCHAR, C.DPI_ORACLE_TYPE_CHAR, C.DPI_ORACLE_TYPE_NCHAR, C.DPI_ORACLE_TYPE_LONG_VARCHAR:
			b := C.dpiData_getBytes(d)
			dest[i] = C.GoStringN(b.ptr, C.int(b.length))
		case C.DPI_ORACLE_TYPE_ROWID, C.DPI_ORACLE_TYPE_RAW, C.DPI_ORACLE_TYPE_LONG_RAW:
			b := C.dpiData_getBytes(d)
			dest[i] = C.GoBytes(unsafe.Pointer(b.ptr), C.int(b.length))
		case C.DPI_ORACLE_TYPE_NATIVE_FLOAT:
			dest[i] = float32(C.dpiData_getFloat(d))
		case C.DPI_ORACLE_TYPE_NATIVE_DOUBLE:
			dest[i] = float64(C.dpiData_getDouble(d))
		case C.DPI_ORACLE_TYPE_NATIVE_INT:
			dest[i] = int64(C.dpiData_getInt64(d))
		case C.DPI_ORACLE_TYPE_NATIVE_UINT:
			dest[i] = uint64(C.dpiData_getUint64(d))
		case C.DPI_ORACLE_TYPE_NUMBER: //Default type used for NUMBER columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
			b := C.dpiData_getBytes(d)
			dest[i] = append(make(num.OCINum, 0, 22), C.GoBytes(unsafe.Pointer(b), C.int(b.length))...)
		case C.DPI_ORACLE_TYPE_DATE:
			var dt date.Date
			b := C.dpiData_getBytes(d)
			copy(dt[:], (*[7]byte)(unsafe.Pointer(&b.ptr))[:len(dt)])
			dest[i] = dt.Get()
		case C.DPI_ORACLE_TYPE_TIMESTAMP, C.DPI_ORACLE_TYPE_TIMESTAMP_TZ, C.DPI_ORACLE_TYPE_TIMESTAMP_LTZ: //Default type used for TIMESTAMP columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
			ts := C.dpiData_getTimestamp(d)
			tz := time.Local
			if col.Type != C.DPI_ORACLE_TYPE_TIMESTAMP {
				tz = time.FixedZone(
					fmt.Sprintf("%02d:%02d", ts.tzHourOffset, ts.tzMinuteOffset),
					int(ts.tzHourOffset)*3600+int(ts.tzMinuteOffset)*60,
				)
			}
			dest[i] = time.Date(int(ts.year), time.Month(ts.month), int(ts.day), int(ts.hour), int(ts.minute), int(ts.second), int(ts.fsecond), tz)
		case C.DPI_ORACLE_TYPE_INTERVAL_DS: //Default type used for INTERVAL DAY TO SECOND columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
			ds := C.dpiData_getIntervalDS(d)
			dest[i] = time.Duration(ds.days)*24*time.Hour +
				time.Duration(ds.hours)*time.Hour +
				time.Duration(ds.minutes)*time.Minute +
				time.Duration(ds.seconds)*time.Second +
				time.Duration(ds.fseconds)
		case C.DPI_ORACLE_TYPE_INTERVAL_YM: //Default type used for INTERVAL YEAR TO MONTH columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
			ym := C.dpiData_getIntervalYM(d)
			dest[i] = fmt.Sprintf("%dy%dm", ym.years, ym.months)
		case C.DPI_ORACLE_TYPE_CLOB, C.DPI_ORACLE_TYPE_NCLOB, C.DPI_ORACLE_TYPE_BLOB, C.DPI_ORACLE_TYPE_BFILE: //Default type used for CLOB columns in the database. Only a locator is transferred to/from Oracle, which can subsequently be used via dpiLob references to read/write from that locator.
			dest[i] = &Lob{dpiLob: C.dpiData_getLOB(d)}
		case C.DPI_ORACLE_TYPE_STMT: //Used within PL/SQL for REF CURSOR or within SQL for querying a CURSOR. Only a handle is transferred to/from Oracle, which can subsequently be used via dpiStmt for querying.
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
		case C.DPI_ORACLE_TYPE_BOOLEAN: //Used within PL/SQL for boolean values. This is only available in 12.1. Earlier releases simply use the integer values 0 and 1 to represent a boolean value. Data is transferred to/from Oracle as an integer.
			dest[i] = C.dpiData_getBool(d) == 1
			//case C.DPI_ORACLE_TYPE_OBJECT: //Default type used for named type columns in the database. Data is transferred to/from Oracle in Oracle's internal format.
		default:
			return errors.Errorf("unsupported column type %d", col.Type)
		}
	}

	return nil
}

type Lob struct {
	dpiLob *C.dpiLob
}
