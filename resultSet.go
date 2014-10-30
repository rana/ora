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
	statement   *Statement
	ocistmt     *C.OCIStmt
	defines     []define
	Err         error
	Index       int
	Row         []interface{}
	ColumnNames []string
}

// Len returns the number of rows retrieved.
func (resultSet *ResultSet) Len() int {
	return resultSet.Index + 1
}

// beginRow allocates handles for each column and fetches one row.
func (resultSet *ResultSet) beginRow() error {
	resultSet.Index++
	// check is open
	if resultSet.ocistmt == nil {
		return errNew("ResultSet is closed")
	}
	// allocate define descriptor handles
	for _, define := range resultSet.defines {
		err := define.alloc()
		if err != nil {
			return err
		}
	}
	// fetch one row
	r := C.OCIStmtFetch2(
		resultSet.ocistmt,                                     //OCIStmt     *stmthp,
		resultSet.statement.session.server.environment.ocierr, //OCIError    *errhp,
		C.ub4(1),         //ub4         nrows,
		C.OCI_FETCH_NEXT, //ub2         orientation,
		C.sb4(0),         //sb4         fetchOffset,
		C.OCI_DEFAULT)    //ub4         mode );
	if r == C.OCI_ERROR {
		return resultSet.statement.session.server.environment.ociError()
	} else if r == C.OCI_NO_DATA {
		// Adjust Index so that Len() returns correct value when all rows read
		resultSet.Index--
		return io.EOF
	}
	return nil
}

// endRow deallocates a handle for each column.
func (resultSet *ResultSet) endRow() {
	for _, define := range resultSet.defines {
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
func (resultSet *ResultSet) Next() bool {
	err := resultSet.beginRow()
	defer resultSet.endRow()
	if err != nil {
		// io.EOF means no more data; return nil err
		if err == io.EOF {
			err = nil
		}
		resultSet.Err = err
		resultSet.Row = nil
		return false
	}
	// populate column values
	for n, define := range resultSet.defines {
		value, err := define.value()
		if err != nil {
			resultSet.Err = err
			resultSet.Row = nil
			return false
		}
		resultSet.Row[n] = value
	}
	return true
}

// NextRow attempts to load a row from the Oracle buffer and return the row.
// Nil is returned when there's no data.
//
// When NextRow returns nil check ResultSet.Err for any error that may have occured.
func (resultSet *ResultSet) NextRow() []interface{} {
	resultSet.Next()
	return resultSet.Row
}

// IsOpen returns true when a result set is open; otherwise, false.
func (resultSet *ResultSet) IsOpen() bool {
	return resultSet.ocistmt != nil
}

// Open defines select-list columns.
func (resultSet *ResultSet) open(statement *Statement, ocistmt *C.OCIStmt) error {
	resultSet.statement = statement
	resultSet.ocistmt = ocistmt
	resultSet.Index = -1
	// Get the implcit select-list describe information; no server round-trip
	r := C.OCIStmtExecute(
		resultSet.statement.session.server.ocisvcctx,          //OCISvcCtx           *svchp,
		resultSet.ocistmt,                                     //OCIStmt             *stmtp,
		resultSet.statement.session.server.environment.ocierr, //OCIError            *errhp,
		C.ub4(1),            //ub4                 iters,
		C.ub4(0),            //ub4                 rowoff,
		nil,                 //const OCISnapshot   *snap_in,
		nil,                 //OCISnapshot         *snap_out,
		C.OCI_DESCRIBE_ONLY) //ub4                 mode );
	if r == C.OCI_ERROR {
		return resultSet.statement.session.server.environment.ociError()
	}

	// Get the parameter count
	var paramCount C.ub4
	err := resultSet.attr(unsafe.Pointer(&paramCount), 4, C.OCI_ATTR_PARAM_COUNT)
	if err != nil {
		return err
	}

	// Make defines slice
	resultSet.defines = make([]define, int(paramCount))
	resultSet.ColumnNames = make([]string, int(paramCount))
	resultSet.Row = make([]interface{}, int(paramCount))
	//fmt.Printf("resultSet.open (paramCount %v)\n", paramCount)

	// Create parameters for each select-list column
	var goColumnType GoColumnType
	for n := 0; n < int(paramCount); n++ {

		// Create oci parameter handle; may be freed by OCIDescriptorFree()
		// parameter position is 1-based
		var ocipar *C.OCIParam
		r := C.OCIParamGet(
			unsafe.Pointer(resultSet.ocistmt),                     //const void        *hndlp,
			C.OCI_HTYPE_STMT,                                      //ub4               htype,
			resultSet.statement.session.server.environment.ocierr, //OCIError          *errhp,
			(*unsafe.Pointer)(unsafe.Pointer(&ocipar)),            //void              **parmdpp,
			C.ub4(n+1))                                            //ub4               pos );
		if r == C.OCI_ERROR {
			return resultSet.statement.session.server.environment.ociError()
		}

		// Get column size in bytes
		var columnSize uint32
		err = resultSet.paramAttr(ocipar, unsafe.Pointer(&columnSize), 0, C.OCI_ATTR_DATA_SIZE)
		if err != nil {
			return err
		}

		// Get oci data type code
		var ociTypeCode C.ub2
		err = resultSet.paramAttr(ocipar, unsafe.Pointer(&ociTypeCode), 0, C.OCI_ATTR_DATA_TYPE)
		if err != nil {
			return err
		}

		//fmt.Printf("statement.defineSelectList: ociTypeCode (%v)\n", ociTypeCode)
		switch ociTypeCode {
		case C.SQLT_NUM:
			// NUMBER
			// Get precision
			var precision C.sb2
			err = resultSet.paramAttr(ocipar, unsafe.Pointer(&precision), 0, C.OCI_ATTR_PRECISION)
			if err != nil {
				return err
			}
			// Get scale (the number of decimal places)
			var numericScale C.sb1
			err = resultSet.paramAttr(ocipar, unsafe.Pointer(&numericScale), 0, C.OCI_ATTR_SCALE)
			if err != nil {
				return err
			}
			// If the precision is nonzero and scale is -127, then it is a FLOAT;
			// otherwise, it's a NUMBER(precision, scale).
			if precision != 0 && (numericScale > 0 || numericScale == -127) {
				if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
					if numericScale == -127 {
						goColumnType = resultSet.statement.Config.ResultSet.float
					} else {
						goColumnType = resultSet.statement.Config.ResultSet.numberScaled
					}
				} else {
					err = checkNumericColumn(statement.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = statement.goColumnTypes[n]
				}
				err := resultSet.defineNumeric(n, goColumnType)
				if err != nil {
					return err
				}
			} else {
				if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
					goColumnType = resultSet.statement.Config.ResultSet.numberScaless
				} else {
					err = checkNumericColumn(statement.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = statement.goColumnTypes[n]
				}
				err := resultSet.defineNumeric(n, goColumnType)
				if err != nil {
					return err
				}
			}
		case C.SQLT_IBDOUBLE:
			// BINARY_DOUBLE
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.binaryDouble
			} else {
				err = checkNumericColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			err := resultSet.defineNumeric(n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_IBFLOAT:
			// BINARY_FLOAT
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.binaryFloat
			} else {
				err = checkNumericColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			err := resultSet.defineNumeric(n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_DAT, C.SQLT_TIMESTAMP, C.SQLT_TIMESTAMP_TZ, C.SQLT_TIMESTAMP_LTZ:
			// DATE, TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIMEZONE
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				switch ociTypeCode {
				case C.SQLT_DAT:
					goColumnType = resultSet.statement.Config.ResultSet.date
				case C.SQLT_TIMESTAMP:
					goColumnType = resultSet.statement.Config.ResultSet.timestamp
				case C.SQLT_TIMESTAMP_TZ:
					goColumnType = resultSet.statement.Config.ResultSet.timestampTz
				case C.SQLT_TIMESTAMP_LTZ:
					goColumnType = resultSet.statement.Config.ResultSet.timestampLtz
				}
			} else {
				err = checkTimeColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			if goColumnType == T {
				timeDefine := resultSet.statement.session.server.environment.timeDefinePool.Get().(*timeDefine)
				resultSet.defines[n] = timeDefine
				err = timeDefine.define(n+1, resultSet.ocistmt)
				if err != nil {
					return err
				}
			} else {
				oraTimeDefine := resultSet.statement.session.server.environment.oraTimeDefinePool.Get().(*oraTimeDefine)
				resultSet.defines[n] = oraTimeDefine
				err = oraTimeDefine.define(n+1, resultSet.ocistmt)
				if err != nil {
					return err
				}
			}
		case C.SQLT_INTERVAL_YM:
			intervalYMDefine := resultSet.statement.session.server.environment.intervalYMDefinePool.Get().(*intervalYMDefine)
			resultSet.defines[n] = intervalYMDefine
			err = intervalYMDefine.define(n+1, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_DS:
			intervalDSDefine := resultSet.statement.session.server.environment.intervalDSDefinePool.Get().(*intervalDSDefine)
			resultSet.defines[n] = intervalDSDefine
			err = intervalDSDefine.define(n+1, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_CHR:
			// VARCHAR, VARCHAR2, NVARCHAR2
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.varchar
			} else {
				err = checkStringColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			err = resultSet.defineString(columnSize, n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_AFC:
			// CHAR, NCHAR
			if columnSize == 1 {
				if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
					goColumnType = resultSet.statement.Config.ResultSet.char1
				} else {
					err = checkBoolOrStringColumn(statement.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = statement.goColumnTypes[n]
				}
				switch goColumnType {
				case B:
					// Interpret single char as bool
					boolDefine := resultSet.statement.session.server.environment.boolDefinePool.Get().(*boolDefine)
					resultSet.defines[n] = boolDefine
					err = boolDefine.define(int(columnSize), n+1, resultSet, resultSet.ocistmt)
					if err != nil {
						return err
					}
				case OraB:
					// Interpret single char as nullable bool
					oraBoolDefine := resultSet.statement.session.server.environment.oraBoolDefinePool.Get().(*oraBoolDefine)
					resultSet.defines[n] = oraBoolDefine
					err = oraBoolDefine.define(int(columnSize), n+1, resultSet, resultSet.ocistmt)
					if err != nil {
						return err
					}
				case S, OraS:
					err = resultSet.defineString(columnSize, n, goColumnType)
					if err != nil {
						return err
					}
				}
			} else {
				// Interpret as string
				if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
					goColumnType = resultSet.statement.Config.ResultSet.char
				} else {
					err = checkStringColumn(statement.goColumnTypes[n])
					if err != nil {
						return err
					}
					goColumnType = statement.goColumnTypes[n]
				}
				err = resultSet.defineString(columnSize, n, goColumnType)
				if err != nil {
					return err
				}
			}
		case C.SQLT_LNG:
			// LONG
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.long
			} else {
				err = checkStringColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			// longBufferSize: Use a moderate default buffer size; 2GB max buffer may not be feasible on all clients
			err = resultSet.defineString(statement.Config.longBufferSize, n, goColumnType)
			if err != nil {
				return err
			}
		case C.SQLT_CLOB:
			// CLOB, NCLOB
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.clob
			} else {
				err = checkStringColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			// Get character set form
			var charsetForm C.ub1
			err = resultSet.paramAttr(ocipar, unsafe.Pointer(&charsetForm), 0, C.OCI_ATTR_CHARSET_FORM)
			if err != nil {
				return err
			}
			lobDefine := resultSet.statement.session.server.environment.lobDefinePool.Get().(*lobDefine)
			resultSet.defines[n] = lobDefine
			err = lobDefine.define(C.SQLT_CLOB, charsetForm, int(columnSize), n+1, goColumnType, resultSet.statement.session.server.ocisvcctx, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_LBI:
			// LONG RAW
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.longRaw
			} else {
				err = checkBitsColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			longRawDefine := resultSet.statement.session.server.environment.longRawDefinePool.Get().(*longRawDefine)
			resultSet.defines[n] = longRawDefine
			err = longRawDefine.define(int(columnSize), n+1, goColumnType, resultSet.statement.Config.longRawBufferSize, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_BIN:
			// RAW
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.raw
			} else {
				err = checkBitsColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			rawDefine := resultSet.statement.session.server.environment.rawDefinePool.Get().(*rawDefine)
			resultSet.defines[n] = rawDefine
			err = rawDefine.define(int(columnSize), n+1, goColumnType, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_BLOB:
			// BLOB
			if statement.goColumnTypes == nil || n >= len(statement.goColumnTypes) || statement.goColumnTypes[n] == D {
				goColumnType = resultSet.statement.Config.ResultSet.blob
			} else {
				err = checkBitsColumn(statement.goColumnTypes[n])
				if err != nil {
					return err
				}
				goColumnType = statement.goColumnTypes[n]
			}
			lobDefine := resultSet.statement.session.server.environment.lobDefinePool.Get().(*lobDefine)
			resultSet.defines[n] = lobDefine
			err = lobDefine.define(C.SQLT_BLOB, C.SQLCS_IMPLICIT, int(columnSize), n+1, goColumnType, resultSet.statement.session.server.ocisvcctx, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_FILE:
			// BFILE
			bfileDefine := resultSet.statement.session.server.environment.bfileDefinePool.Get().(*bfileDefine)
			resultSet.defines[n] = bfileDefine
			err = bfileDefine.define(int(columnSize), n+1, resultSet.ocistmt)
			if err != nil {
				return err
			}
		case C.SQLT_RDD:
			// ROWID, UROWID
			rowidDefine := resultSet.statement.session.server.environment.rowidDefinePool.Get().(*rowidDefine)
			resultSet.defines[n] = rowidDefine
			err = rowidDefine.define(int(columnSize), n+1, resultSet.ocistmt)
			if err != nil {
				return err
			}
			break
		default:
			return errNewF("unsupported select-list column type (ociTypeCode: %v)", ociTypeCode)
		}

		// Get column name
		if resultSet.defines[n] != nil {
			var columnName *C.char
			err := resultSet.paramAttr(ocipar, unsafe.Pointer(&columnName), 0, C.OCI_ATTR_NAME)
			if err != nil {
				return err
			}
			resultSet.ColumnNames[n] = C.GoString(columnName)
		}
	}

	return nil
}

// close releases allocated resources.
func (resultSet *ResultSet) close() {
	// Close defines
	if resultSet.ocistmt != nil {
		if len(resultSet.defines) > 0 {
			for _, define := range resultSet.defines {
				//fmt.Printf("close define %v\n", define)
				if define != nil {
					define.close()
				}
			}
		}

		resultSet.statement = nil
		resultSet.ocistmt = nil
		resultSet.defines = nil
		resultSet.Err = nil
		resultSet.Index = -1
		resultSet.Row = nil
		resultSet.ColumnNames = nil
	}
}

func (resultSet *ResultSet) defineString(columnSize uint32, n int, goColumnType GoColumnType) (err error) {
	if goColumnType == S {
		stringDefine := resultSet.statement.session.server.environment.stringDefinePool.Get().(*stringDefine)
		resultSet.defines[n] = stringDefine
		err = stringDefine.define(int(columnSize), n+1, resultSet.ocistmt)
	} else {
		oraStringDefine := resultSet.statement.session.server.environment.oraStringDefinePool.Get().(*oraStringDefine)
		resultSet.defines[n] = oraStringDefine
		err = oraStringDefine.define(int(columnSize), n+1, resultSet.ocistmt)
	}
	return err
}

func (resultSet *ResultSet) defineNumeric(n int, goColumnType GoColumnType) (err error) {
	switch goColumnType {
	case I64:
		int64Define := resultSet.statement.session.server.environment.int64DefinePool.Get().(*int64Define)
		resultSet.defines[n] = int64Define
		err = int64Define.define(n+1, resultSet.ocistmt)
	case I32:
		int32Define := resultSet.statement.session.server.environment.int32DefinePool.Get().(*int32Define)
		resultSet.defines[n] = int32Define
		err = int32Define.define(n+1, resultSet.ocistmt)
	case I16:
		int16Define := resultSet.statement.session.server.environment.int16DefinePool.Get().(*int16Define)
		resultSet.defines[n] = int16Define
		err = int16Define.define(n+1, resultSet.ocistmt)
	case I8:
		int8Define := resultSet.statement.session.server.environment.int8DefinePool.Get().(*int8Define)
		resultSet.defines[n] = int8Define
		err = int8Define.define(n+1, resultSet.ocistmt)
	case U64:
		uint64Define := resultSet.statement.session.server.environment.uint64DefinePool.Get().(*uint64Define)
		resultSet.defines[n] = uint64Define
		err = uint64Define.define(n+1, resultSet.ocistmt)
	case U32:
		uint32Define := resultSet.statement.session.server.environment.uint32DefinePool.Get().(*uint32Define)
		resultSet.defines[n] = uint32Define
		err = uint32Define.define(n+1, resultSet.ocistmt)
	case U16:
		uint16Define := resultSet.statement.session.server.environment.uint16DefinePool.Get().(*uint16Define)
		resultSet.defines[n] = uint16Define
		err = uint16Define.define(n+1, resultSet.ocistmt)
	case U8:
		uint8Define := resultSet.statement.session.server.environment.uint8DefinePool.Get().(*uint8Define)
		resultSet.defines[n] = uint8Define
		err = uint8Define.define(n+1, resultSet.ocistmt)
	case F64:
		float64Define := resultSet.statement.session.server.environment.float64DefinePool.Get().(*float64Define)
		resultSet.defines[n] = float64Define
		err = float64Define.define(n+1, resultSet.ocistmt)
	case F32:
		float32Define := resultSet.statement.session.server.environment.float32DefinePool.Get().(*float32Define)
		resultSet.defines[n] = float32Define
		err = float32Define.define(n+1, resultSet.ocistmt)
	case OraI64:
		oraInt64Define := resultSet.statement.session.server.environment.oraInt64DefinePool.Get().(*oraInt64Define)
		resultSet.defines[n] = oraInt64Define
		err = oraInt64Define.define(n+1, resultSet.ocistmt)
	case OraI32:
		oraInt32Define := resultSet.statement.session.server.environment.oraInt32DefinePool.Get().(*oraInt32Define)
		resultSet.defines[n] = oraInt32Define
		err = oraInt32Define.define(n+1, resultSet.ocistmt)
	case OraI16:
		oraInt16Define := resultSet.statement.session.server.environment.oraInt16DefinePool.Get().(*oraInt16Define)
		resultSet.defines[n] = oraInt16Define
		err = oraInt16Define.define(n+1, resultSet.ocistmt)
	case OraI8:
		oraInt8Define := resultSet.statement.session.server.environment.oraInt8DefinePool.Get().(*oraInt8Define)
		resultSet.defines[n] = oraInt8Define
		err = oraInt8Define.define(n+1, resultSet.ocistmt)
	case OraU64:
		oraUint64Define := resultSet.statement.session.server.environment.oraUint64DefinePool.Get().(*oraUint64Define)
		resultSet.defines[n] = oraUint64Define
		err = oraUint64Define.define(n+1, resultSet.ocistmt)
	case OraU32:
		oraUint32Define := resultSet.statement.session.server.environment.oraUint32DefinePool.Get().(*oraUint32Define)
		resultSet.defines[n] = oraUint32Define
		err = oraUint32Define.define(n+1, resultSet.ocistmt)
	case OraU16:
		oraUint16Define := resultSet.statement.session.server.environment.oraUint16DefinePool.Get().(*oraUint16Define)
		resultSet.defines[n] = oraUint16Define
		err = oraUint16Define.define(n+1, resultSet.ocistmt)
	case OraU8:
		oraUint8Define := resultSet.statement.session.server.environment.oraUint8DefinePool.Get().(*oraUint8Define)
		resultSet.defines[n] = oraUint8Define
		err = oraUint8Define.define(n+1, resultSet.ocistmt)
	case OraF64:
		oraFloat64Define := resultSet.statement.session.server.environment.oraFloat64DefinePool.Get().(*oraFloat64Define)
		resultSet.defines[n] = oraFloat64Define
		err = oraFloat64Define.define(n+1, resultSet.ocistmt)
	case OraF32:
		oraFloat32Define := resultSet.statement.session.server.environment.oraFloat32DefinePool.Get().(*oraFloat32Define)
		resultSet.defines[n] = oraFloat32Define
		err = oraFloat32Define.define(n+1, resultSet.ocistmt)
	}
	return err
}

// paramAttr gets an attribute from the parameter handle.
func (resultSet *ResultSet) paramAttr(ocipar *C.OCIParam, attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(ocipar), //const void     *trgthndlp,
		C.OCI_DTYPE_PARAM,      //ub4            trghndltyp,
		attrup,                 //void           *attributep,
		&attrSize,              //ub4            *sizep,
		attrType,               //ub4            attrtype,
		resultSet.statement.session.server.environment.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return resultSet.statement.session.server.environment.ociError()
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (resultSet *ResultSet) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(resultSet.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,                  //ub4            trghndltyp,
		attrup,                            //void           *attributep,
		&attrSize,                         //ub4            *sizep,
		attrType,                          //ub4            attrtype,
		resultSet.statement.session.server.environment.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return resultSet.statement.session.server.environment.ociError()
	}
	return nil
}
