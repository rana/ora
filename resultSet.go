// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"io"
	"unsafe"
)

// ResultSet is used to obtain Go values from a SQL select statement.
//
// Opening and closing a ResultSet is managed internally.
// ResultSet doesn't have an Open method or Close method.
type ResultSet struct {
	ocistmt *C.OCIStmt
	stmt    *Statement
	defines []define

	Row         []interface{}
	ColumnNames []string
	Index       int
	Err         error
}

// Len returns the number of rows retrieved.
func (rst *ResultSet) Len() int {
	return rst.Index + 1
}

// beginRow allocates handles for each column and fetches one row.
func (rst *ResultSet) beginRow() error {
	rst.Index++
	// check is open
	if rst.ocistmt == nil {
		return errNew("ResultSet is closed")
	}
	// allocate define descriptor handles
	for _, define := range rst.defines {
		err := define.alloc()
		if err != nil {
			return err
		}
	}
	// fetch one row
	r := C.OCIStmtFetch2(
		rst.ocistmt,                 //OCIStmt     *stmthp,
		rst.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(1),                    //ub4         nrows,
		C.OCI_FETCH_NEXT,            //ub2         orientation,
		C.sb4(0),                    //sb4         fetchOffset,
		C.OCI_DEFAULT)               //ub4         mode );
	if r == C.OCI_ERROR {
		return rst.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_NO_DATA {
		// Adjust Index so that Len() returns correct value when all rows read
		rst.Index--
		return io.EOF
	}
	return nil
}

// endRow deallocates a handle for each column.
func (rst *ResultSet) endRow() {
	for _, define := range rst.defines {
		define.free()
	}
}

// Next attempts to load a row of data from an Oracle buffer. True is returned
// when a row of data is retrieved. False is returned when no data is available.
//
// Retrieve the loaded row from the ResultSet.Row field. ResultSet.Row is updated
// on each call to Next. ResultSet.Row is set to nil when Next returns false.
//
// When Next returns false check ResultSet.Err for any error that may have occured.
func (rst *ResultSet) Next() bool {
	err := rst.beginRow()
	defer rst.endRow()
	if err != nil {
		// io.EOF means no more data; return nil err
		if err == io.EOF {
			err = nil
		}
		rst.Err = err
		rst.Row = nil
		return false
	}
	// populate column values
	for n, define := range rst.defines {
		value, err := define.value()
		if err != nil {
			rst.Err = err
			rst.Row = nil
			return false
		}
		rst.Row[n] = value
	}
	return true
}

// NextRow attempts to load a row from the Oracle buffer and return the row.
// Nil is returned when there's no data.
//
// When NextRow returns nil check ResultSet.Err for any error that may have occured.
func (rst *ResultSet) NextRow() []interface{} {
	rst.Next()
	return rst.Row
}

// IsOpen returns true when a result set is open; otherwise, false.
func (rst *ResultSet) IsOpen() bool {
	return rst.ocistmt != nil
}

// Open defines select-list columns.
func (rst *ResultSet) open(stmt *Statement, ocistmt *C.OCIStmt) error {
	rst.stmt = stmt
	rst.ocistmt = ocistmt
	rst.Index = -1
	// Get the implcit select-list describe information; no server round-trip
	r := C.OCIStmtExecute(
		rst.stmt.ses.srv.ocisvcctx,  //OCISvcCtx           *svchp,
		rst.ocistmt,                 //OCIStmt             *stmtp,
		rst.stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(1),                    //ub4                 iters,
		C.ub4(0),                    //ub4                 rowoff,
		nil,                         //const OCISnapshot   *snap_in,
		nil,                         //OCISnapshot         *snap_out,
		C.OCI_DESCRIBE_ONLY)         //ub4                 mode );
	if r == C.OCI_ERROR {
		return rst.stmt.ses.srv.env.ociError()
	}

	// Get the parameter count
	var paramCount C.ub4
	err := rst.attr(unsafe.Pointer(&paramCount), 4, C.OCI_ATTR_PARAM_COUNT)
	if err != nil {
		return err
	}

	// Make defines slice
	rst.defines = make([]define, int(paramCount))
	rst.ColumnNames = make([]string, int(paramCount))
	rst.Row = make([]interface{}, int(paramCount))
	//fmt.Printf("rst.open (paramCount %v)\n", paramCount)

	// Create parameters for each select-list column
	var goColumnType GoColumnType
	for n := 0; n < int(paramCount); n++ {

		// Create oci parameter handle; may be freed by OCIDescriptorFree()
		// parameter position is 1-based
		var ocipar *C.OCIParam
		r := C.OCIParamGet(
			unsafe.Pointer(rst.ocistmt),                //const void        *hndlp,
			C.OCI_HTYPE_STMT,                           //ub4               htype,
			rst.stmt.ses.srv.env.ocierr,                //OCIError          *errhp,
			(*unsafe.Pointer)(unsafe.Pointer(&ocipar)), //void              **parmdpp,
			C.ub4(n+1))                                 //ub4               pos );
		if r == C.OCI_ERROR {
			return rst.stmt.ses.srv.env.ociError()
		}

		// Get column size in bytes
		var columnSize uint32
		err = rst.paramAttr(ocipar, unsafe.Pointer(&columnSize), 0, C.OCI_ATTR_DATA_SIZE)
		if err != nil {
			return err
		}

		// Get oci data type code
		var ociTypeCode C.ub2
		err = rst.paramAttr(ocipar, unsafe.Pointer(&ociTypeCode), 0, C.OCI_ATTR_DATA_TYPE)
		if err != nil {
			return err
		}

		//fmt.Printf("statement.defineSelectList: ociTypeCode (%v)\n", ociTypeCode)
		switch ociTypeCode {
		case C.SQLT_NUM:
			// NUMBER
			// Get precision
			var precision C.sb2
			err = rst.paramAttr(ocipar, unsafe.Pointer(&precision), 0, C.OCI_ATTR_PRECISION)
			if err != nil {
				return err
			}
			// Get scale (the number of decimal places)
			var numericScale C.sb1
			err = rst.paramAttr(ocipar, unsafe.Pointer(&numericScale), 0, C.OCI_ATTR_SCALE)
			if err != nil {
				return err
			}
			// If the precision is nonzero and scale is -127, then it is a FLOAT;
			// otherwise, it's a NUMBER(precision, scale).
			if precision != 0 && (numericScale > 0 || numericScale == -127) {
				if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
					if numericScale == -127 {
						goColumnType = rst.stmt.Config.ResultSet.float
					} else {
						goColumnType = rst.stmt.Config.ResultSet.numberScaled
					}
				} else {
					err = checkNumericColumn(stmt.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = stmt.goColumnTypes[n]
				}
				err := rst.defineNumeric(n, goColumnType)
				if err != nil {
					return err
				}
			} else {
				if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
					goColumnType = rst.stmt.Config.ResultSet.numberScaless
				} else {
					err = checkNumericColumn(stmt.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = stmt.goColumnTypes[n]
				}
				err := rst.defineNumeric(n, goColumnType)
				if err != nil {
					return err
				}
			}
		case C.SQLT_IBDOUBLE:
			// BINARY_DOUBLE
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.binaryDouble
			} else {
				err = checkNumericColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			err := rst.defineNumeric(n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_IBFLOAT:
			// BINARY_FLOAT
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.binaryFloat
			} else {
				err = checkNumericColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			err := rst.defineNumeric(n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_DAT, C.SQLT_TIMESTAMP, C.SQLT_TIMESTAMP_TZ, C.SQLT_TIMESTAMP_LTZ:
			// DATE, TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIMEZONE
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				switch ociTypeCode {
				case C.SQLT_DAT:
					goColumnType = rst.stmt.Config.ResultSet.date
				case C.SQLT_TIMESTAMP:
					goColumnType = rst.stmt.Config.ResultSet.timestamp
				case C.SQLT_TIMESTAMP_TZ:
					goColumnType = rst.stmt.Config.ResultSet.timestampTz
				case C.SQLT_TIMESTAMP_LTZ:
					goColumnType = rst.stmt.Config.ResultSet.timestampLtz
				}
			} else {
				err = checkTimeColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			if goColumnType == T {
				timeDefine := rst.stmt.ses.srv.env.timeDefinePool.Get().(*timeDefine)
				rst.defines[n] = timeDefine
				err = timeDefine.define(n+1, rst.ocistmt)
				if err != nil {
					return err
				}
			} else {
				oraTimeDefine := rst.stmt.ses.srv.env.oraTimeDefinePool.Get().(*oraTimeDefine)
				rst.defines[n] = oraTimeDefine
				err = oraTimeDefine.define(n+1, rst.ocistmt)
				if err != nil {
					return err
				}
			}
		case C.SQLT_INTERVAL_YM:
			intervalYMDefine := rst.stmt.ses.srv.env.intervalYMDefinePool.Get().(*intervalYMDefine)
			rst.defines[n] = intervalYMDefine
			err = intervalYMDefine.define(n+1, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_DS:
			intervalDSDefine := rst.stmt.ses.srv.env.intervalDSDefinePool.Get().(*intervalDSDefine)
			rst.defines[n] = intervalDSDefine
			err = intervalDSDefine.define(n+1, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_CHR:
			// VARCHAR, VARCHAR2, NVARCHAR2
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.varchar
			} else {
				err = checkStringColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			err = rst.defineString(columnSize, n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_AFC:
			// CHAR, NCHAR
			if columnSize == 1 {
				if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
					goColumnType = rst.stmt.Config.ResultSet.char1
				} else {
					err = checkBoolOrStringColumn(stmt.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = stmt.goColumnTypes[n]
				}
				switch goColumnType {
				case B:
					// Interpret single char as bool
					boolDefine := rst.stmt.ses.srv.env.boolDefinePool.Get().(*boolDefine)
					rst.defines[n] = boolDefine
					err = boolDefine.define(int(columnSize), n+1, rst, rst.ocistmt)
					if err != nil {
						return err
					}
				case OraB:
					// Interpret single char as nullable bool
					oraBoolDefine := rst.stmt.ses.srv.env.oraBoolDefinePool.Get().(*oraBoolDefine)
					rst.defines[n] = oraBoolDefine
					err = oraBoolDefine.define(int(columnSize), n+1, rst, rst.ocistmt)
					if err != nil {
						return err
					}
				case S, OraS:
					err = rst.defineString(columnSize, n, goColumnType)
					if err != nil {
						return err
					}
				}
			} else {
				// Interpret as string
				if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
					goColumnType = rst.stmt.Config.ResultSet.char
				} else {
					err = checkStringColumn(stmt.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = stmt.goColumnTypes[n]
				}
				err = rst.defineString(columnSize, n, goColumnType)
				if err != nil {
					return err
				}
			}
		case C.SQLT_LNG:
			// LONG
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.long
			} else {
				err = checkStringColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			// longBufferSize: Use a moderate default buffer size; 2GB max buffer may not be feasible on all clients
			err = rst.defineString(stmt.Config.longBufferSize, n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_CLOB:
			// CLOB, NCLOB
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.clob
			} else {
				err = checkStringColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			// Get character set form
			var charsetForm C.ub1
			err = rst.paramAttr(ocipar, unsafe.Pointer(&charsetForm), 0, C.OCI_ATTR_CHARSET_FORM)
			if err != nil {
				return err
			}
			lobDefine := rst.stmt.ses.srv.env.lobDefinePool.Get().(*lobDefine)
			rst.defines[n] = lobDefine
			err = lobDefine.define(C.SQLT_CLOB, charsetForm, int(columnSize), n+1, goColumnType, rst.stmt.ses.srv.ocisvcctx, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_LBI:
			// LONG RAW
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.longRaw
			} else {
				err = checkBitsColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			longRawDefine := rst.stmt.ses.srv.env.longRawDefinePool.Get().(*longRawDefine)
			rst.defines[n] = longRawDefine
			err = longRawDefine.define(int(columnSize), n+1, goColumnType, rst.stmt.Config.longRawBufferSize, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_BIN:
			// RAW
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.raw
			} else {
				err = checkBitsColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			rawDefine := rst.stmt.ses.srv.env.rawDefinePool.Get().(*rawDefine)
			rst.defines[n] = rawDefine
			err = rawDefine.define(int(columnSize), n+1, goColumnType, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_BLOB:
			// BLOB
			if stmt.goColumnTypes == nil || n >= len(stmt.goColumnTypes) || stmt.goColumnTypes[n] == D {
				goColumnType = rst.stmt.Config.ResultSet.blob
			} else {
				err = checkBitsColumn(stmt.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = stmt.goColumnTypes[n]
			}
			lobDefine := rst.stmt.ses.srv.env.lobDefinePool.Get().(*lobDefine)
			rst.defines[n] = lobDefine
			err = lobDefine.define(C.SQLT_BLOB, C.SQLCS_IMPLICIT, int(columnSize), n+1, goColumnType, rst.stmt.ses.srv.ocisvcctx, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_FILE:
			// BFILE
			bfileDefine := rst.stmt.ses.srv.env.bfileDefinePool.Get().(*bfileDefine)
			rst.defines[n] = bfileDefine
			err = bfileDefine.define(int(columnSize), n+1, rst.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_RDD:
			// ROWID, UROWID
			rowidDefine := rst.stmt.ses.srv.env.rowidDefinePool.Get().(*rowidDefine)
			rst.defines[n] = rowidDefine
			err = rowidDefine.define(int(columnSize), n+1, rst.ocistmt)
			if err != nil {
				return err
			}
			break
		default:
			return errNewF("unsupported select-list column type (ociTypeCode: %v)", ociTypeCode)
		}

		// Get column name
		if rst.defines[n] != nil {
			var columnName *C.char
			err := rst.paramAttr(ocipar, unsafe.Pointer(&columnName), 0, C.OCI_ATTR_NAME)
			if err != nil {
				return err
			}
			rst.ColumnNames[n] = C.GoString(columnName)
		}
	}

	return nil
}

// close releases allocated resources.
func (rst *ResultSet) close() {
	// Close defines
	if rst.ocistmt != nil {
		if len(rst.defines) > 0 {
			for _, define := range rst.defines {
				//fmt.Printf("close define %v\n", define)
				if define != nil {
					define.close()
				}
			}
		}

		rst.stmt = nil
		rst.ocistmt = nil
		rst.defines = nil
		rst.Err = nil
		rst.Index = -1
		rst.Row = nil
		rst.ColumnNames = nil
	}
}

func (rst *ResultSet) defineString(columnSize uint32, n int, goColumnType GoColumnType) (err error) {
	if goColumnType == S {
		stringDefine := rst.stmt.ses.srv.env.stringDefinePool.Get().(*stringDefine)
		rst.defines[n] = stringDefine
		err = stringDefine.define(int(columnSize), n+1, rst.ocistmt)
	} else {
		oraStringDefine := rst.stmt.ses.srv.env.oraStringDefinePool.Get().(*oraStringDefine)
		rst.defines[n] = oraStringDefine
		err = oraStringDefine.define(int(columnSize), n+1, rst.ocistmt)
	}
	return err
}

func (rst *ResultSet) defineNumeric(n int, goColumnType GoColumnType) (err error) {
	switch goColumnType {
	case I64:
		int64Define := rst.stmt.ses.srv.env.int64DefinePool.Get().(*int64Define)
		rst.defines[n] = int64Define
		err = int64Define.define(n+1, rst.ocistmt)
	case I32:
		int32Define := rst.stmt.ses.srv.env.int32DefinePool.Get().(*int32Define)
		rst.defines[n] = int32Define
		err = int32Define.define(n+1, rst.ocistmt)
	case I16:
		int16Define := rst.stmt.ses.srv.env.int16DefinePool.Get().(*int16Define)
		rst.defines[n] = int16Define
		err = int16Define.define(n+1, rst.ocistmt)
	case I8:
		int8Define := rst.stmt.ses.srv.env.int8DefinePool.Get().(*int8Define)
		rst.defines[n] = int8Define
		err = int8Define.define(n+1, rst.ocistmt)
	case U64:
		uint64Define := rst.stmt.ses.srv.env.uint64DefinePool.Get().(*uint64Define)
		rst.defines[n] = uint64Define
		err = uint64Define.define(n+1, rst.ocistmt)
	case U32:
		uint32Define := rst.stmt.ses.srv.env.uint32DefinePool.Get().(*uint32Define)
		rst.defines[n] = uint32Define
		err = uint32Define.define(n+1, rst.ocistmt)
	case U16:
		uint16Define := rst.stmt.ses.srv.env.uint16DefinePool.Get().(*uint16Define)
		rst.defines[n] = uint16Define
		err = uint16Define.define(n+1, rst.ocistmt)
	case U8:
		uint8Define := rst.stmt.ses.srv.env.uint8DefinePool.Get().(*uint8Define)
		rst.defines[n] = uint8Define
		err = uint8Define.define(n+1, rst.ocistmt)
	case F64:
		float64Define := rst.stmt.ses.srv.env.float64DefinePool.Get().(*float64Define)
		rst.defines[n] = float64Define
		err = float64Define.define(n+1, rst.ocistmt)
	case F32:
		float32Define := rst.stmt.ses.srv.env.float32DefinePool.Get().(*float32Define)
		rst.defines[n] = float32Define
		err = float32Define.define(n+1, rst.ocistmt)
	case OraI64:
		oraInt64Define := rst.stmt.ses.srv.env.oraInt64DefinePool.Get().(*oraInt64Define)
		rst.defines[n] = oraInt64Define
		err = oraInt64Define.define(n+1, rst.ocistmt)
	case OraI32:
		oraInt32Define := rst.stmt.ses.srv.env.oraInt32DefinePool.Get().(*oraInt32Define)
		rst.defines[n] = oraInt32Define
		err = oraInt32Define.define(n+1, rst.ocistmt)
	case OraI16:
		oraInt16Define := rst.stmt.ses.srv.env.oraInt16DefinePool.Get().(*oraInt16Define)
		rst.defines[n] = oraInt16Define
		err = oraInt16Define.define(n+1, rst.ocistmt)
	case OraI8:
		oraInt8Define := rst.stmt.ses.srv.env.oraInt8DefinePool.Get().(*oraInt8Define)
		rst.defines[n] = oraInt8Define
		err = oraInt8Define.define(n+1, rst.ocistmt)
	case OraU64:
		oraUint64Define := rst.stmt.ses.srv.env.oraUint64DefinePool.Get().(*oraUint64Define)
		rst.defines[n] = oraUint64Define
		err = oraUint64Define.define(n+1, rst.ocistmt)
	case OraU32:
		oraUint32Define := rst.stmt.ses.srv.env.oraUint32DefinePool.Get().(*oraUint32Define)
		rst.defines[n] = oraUint32Define
		err = oraUint32Define.define(n+1, rst.ocistmt)
	case OraU16:
		oraUint16Define := rst.stmt.ses.srv.env.oraUint16DefinePool.Get().(*oraUint16Define)
		rst.defines[n] = oraUint16Define
		err = oraUint16Define.define(n+1, rst.ocistmt)
	case OraU8:
		oraUint8Define := rst.stmt.ses.srv.env.oraUint8DefinePool.Get().(*oraUint8Define)
		rst.defines[n] = oraUint8Define
		err = oraUint8Define.define(n+1, rst.ocistmt)
	case OraF64:
		oraFloat64Define := rst.stmt.ses.srv.env.oraFloat64DefinePool.Get().(*oraFloat64Define)
		rst.defines[n] = oraFloat64Define
		err = oraFloat64Define.define(n+1, rst.ocistmt)
	case OraF32:
		oraFloat32Define := rst.stmt.ses.srv.env.oraFloat32DefinePool.Get().(*oraFloat32Define)
		rst.defines[n] = oraFloat32Define
		err = oraFloat32Define.define(n+1, rst.ocistmt)
	}
	return err
}

// paramAttr gets an attribute from the parameter handle.
func (rst *ResultSet) paramAttr(ocipar *C.OCIParam, attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(ocipar),      //const void     *trgthndlp,
		C.OCI_DTYPE_PARAM,           //ub4            trghndltyp,
		attrup,                      //void           *attributep,
		&attrSize,                   //ub4            *sizep,
		attrType,                    //ub4            attrtype,
		rst.stmt.ses.srv.env.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return rst.stmt.ses.srv.env.ociError()
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (rst *ResultSet) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(rst.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,            //ub4            trghndltyp,
		attrup,                      //void           *attributep,
		&attrSize,                   //ub4            *sizep,
		attrType,                    //ub4            attrtype,
		rst.stmt.ses.srv.env.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return rst.stmt.ses.srv.env.ociError()
	}
	return nil
}
