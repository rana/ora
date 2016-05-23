// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
*/
import "C"
import (
	"container/list"
	"fmt"
	"io"
	"unsafe"
)

const (
	MaxFetchLen = 32
	MinFetchLen = 8

	byteWidth64 = 8
	byteWidth32 = 4
	byteWidth16 = 2
	byteWidth8  = 1
)

// LogRsetCfg represents Rset logging configuration values.
type LogRsetCfg struct {
	// Close determines whether the Rset.close method is logged.
	//
	// The default is true.
	Close bool

	// BeginRow determines whether the Rset.beginRow method is logged.
	//
	// The default is false.
	BeginRow bool

	// EndRow determines whether the Rset.endRow method is logged.
	//
	// The default is false.
	EndRow bool

	// Next determines whether the Rset.Next method is logged.
	//
	// The default is false.
	Next bool

	// Open determines whether the Rset.open method is logged.
	//
	// The default is true.
	Open bool

	// OpenDefs determines whether Select-list definitions with the Rset.open method are logged.
	//
	// The default is true.
	OpenDefs bool
}

// NewLogTxCfg creates a LogRsetCfg with default values.
func NewLogRsetCfg() LogRsetCfg {
	c := LogRsetCfg{}
	c.Close = true
	c.BeginRow = false
	c.EndRow = false
	c.Next = false
	c.Open = true
	c.OpenDefs = true
	return c
}

// Rset represents a result set used to obtain Go values from a SQL select statement.
//
// Opening and closing a Rset is managed internally. Rset doesn't have an Open
// method or Close method.
type Rset struct {
	id        uint64
	stmt      *Stmt
	ocistmt   *C.OCIStmt
	defs      []def
	autoClose bool
	genByPool bool

	Row             []interface{}
	ColumnNames     []string
	Index           int
	Err             error
	fetched, offset int
	fetchLen        int
	finished        bool
}

// Len returns the number of rows retrieved.
func (rset *Rset) Len() int {
	return rset.Index + 1
}

// checkIsOpen validates that the result set is open.
func (rset *Rset) checkIsOpen() error {
	if !rset.IsOpen() {
		return er("Rset is closed.")
	}
	return nil
}

// IsOpen returns true when a result set is open; otherwise, false.
func (rset *Rset) IsOpen() bool {
	return rset.stmt != nil
}

// closeWithRemove releases allocated resources and removes the Rset from the
// Stmt.openRsets list.
func (rset *Rset) closeWithRemove() (err error) {
	rset.stmt.openRsets.remove(rset)
	return rset.close()
}

// close releases allocated resources.
func (rset *Rset) close() (err error) {
	rset.log(_drv.cfg.Log.Rset.Close)
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
		if rset.genByPool { // recycle pool-generated Rset; don't recycle user-specfied Rset
			_drv.rsetPool.Put(rset)
		}
	}()
	if err := rset.checkIsOpen(); err != nil {
		return err
	}
	errs := _drv.listPool.Get().(*list.List)
	if len(rset.defs) > 0 { // close defines
		for _, def := range rset.defs {
			if def != nil {
				err0 := def.close()
				if err0 != nil {
					errs.PushBack(err0)
				}
			}
		}
	}
	rset.stmt = nil
	rset.ocistmt = nil
	rset.defs = nil
	rset.Index = -1
	rset.Row = nil
	rset.ColumnNames = nil
	// do not clear error in case of autoClose when error exists
	// clear error when rset in initialized
	//rset.Err = nil
	m := newMultiErrL(errs)
	if m != nil {
		err = *m
	}
	errs.Init()
	_drv.listPool.Put(errs)
	return err
}

// beginRow allocates a handle for each column and fetches one row.
func (rset *Rset) beginRow() (err error) {
	rset.log(_drv.cfg.Log.Rset.BeginRow)
	rset.logF(_drv.cfg.Log.Rset.BeginRow, "fetched=%d offset=%d finished=%t", rset.fetched, rset.offset, rset.finished)
	if rset.finished {
		rset.log(_drv.cfg.Log.Rset.BeginRow, "finished")
		if rset.fetched > 0 && rset.fetched > rset.offset {
			rset.Index++
			return nil
		}
		return io.EOF
	}
	// check is open
	if rset.ocistmt == nil {
		return errF("Rset is closed")
	}
	// allocate define descriptor handles
	for _, define := range rset.defs {
		//rset.logF(_drv.cfg.Log.Rset.BeginRow, "%#v", define)
		if define == nil {
			continue
		}
		err := define.alloc()
		if err != nil {
			return err
		}
	}
	rset.finished = false
	// fetch one row
	r := C.OCIStmtFetch2(
		rset.ocistmt,                 //OCIStmt     *stmthp,
		rset.stmt.ses.srv.env.ocierr, //OCIError    *errhp,
		C.ub4(rset.fetchLen),         //ub4         nrows,
		C.OCI_FETCH_NEXT,             //ub2         orientation,
		C.sb4(0),                     //sb4         fetchOffset,
		C.OCI_DEFAULT)                //ub4         mode );
	if r == C.OCI_ERROR {
		err := rset.stmt.ses.srv.env.ociError()
		return err
	} else if r == C.OCI_NO_DATA {
		rset.log(_drv.cfg.Log.Rset.BeginRow, "OCI_NO_DATA")
		rset.finished = true
		if rset.fetchLen == 1 {
			// return io.EOF to conform with database/sql/driver
			return io.EOF
		}
		// If OCIStmtFetch2 returns OCI_NO_DATA this does not mean that no data fetched,
		// this means that the number of fetched rows is less than the array size,
		// they are all fetched by this OCIStmtFetch2 call, and you do not need to
		// call OCIStmtFetch2 anymore.
		//
	}
	rset.offset = 0
	if rset.fetchLen == 1 {
		rset.fetched = 1
	} else {
		var rowsFetched C.ub4
		if err := rset.attr(unsafe.Pointer(&rowsFetched), 4, C.OCI_ATTR_ROWS_FETCHED); err != nil {
			return err
		}

		rset.fetched = int(rowsFetched)
		if rset.fetched == 0 {
			rset.finished = true
			return io.EOF
		}
	}
	rset.Index++
	return nil
}

// endRow deallocates a handle for each column.
func (rset *Rset) endRow() {
	rset.log(_drv.cfg.Log.Rset.EndRow)
	for _, define := range rset.defs {
		if define != nil {
			define.free()
		}
	}
	rset.offset++
}

// Next attempts to load a row of data from an Oracle buffer. True is returned
// when a row of data is retrieved. False is returned when no data is available.
//
// Retrieve the loaded row from the Rset.Row field. Rset.Row is updated
// on each call to Next. Rset.Row is set to nil when Next returns false.
//
// When Next returns false check Rset.Err for any error that may have occured.
func (rset *Rset) Next() bool {
	rset.log(_drv.cfg.Log.Rset.Next)
	if err := rset.checkIsOpen(); err != nil {
		rset.Err = err
		rset.Row = nil
		if rset.autoClose {
			rset.stmt.Close()
		}
		return false
	}
	err := rset.beginRow()
	defer rset.endRow()
	//rset.logF(_drv.cfg.Log.Rset.Next, "beginRow=%v", err)
	if err != nil {
		// io.EOF means no more data; return nil err
		if err == io.EOF {
			err = nil
		}
		rset.Err = err
		rset.Row = nil
		if rset.autoClose {
			rset.stmt.Close()
		}
		return false
	}
	// populate column values
	for n, define := range rset.defs {
		value, err := define.value(rset.offset)
		//rset.logF(_drv.cfg.Log.Rset.Next, "value[%d]=%v (%v)", n, value, err)
		if err != nil {
			rset.Err = err
			rset.Row = nil
			if rset.autoClose {
				rset.stmt.Close()
			}
			return false
		}
		rset.Row[n] = value
	}
	//rset.logF(_drv.cfg.Log.Rset.Next, "Row=%#v", rset.Row)
	return true
}

// NextRow attempts to load a row from the Oracle buffer and return the row.
// Nil is returned when there's no data.
//
// When NextRow returns nil check Rset.Err for any error that may have occured.
func (rset *Rset) NextRow() []interface{} {
	rset.Next()
	return rset.Row
}

// gets a define struct from a driver slice
func (rset *Rset) getDef(idx int) interface{} {
	return _drv.defPools[idx].Get()
}

// puts a bind struct in the driver slice
func (rset *Rset) putDef(idx int, def def) {
	_drv.defPools[idx].Put(def)
}

// Open defines select-list columns.
func (rset *Rset) open(stmt *Stmt, ocistmt *C.OCIStmt) error {
	rset.stmt = stmt
	rset.ocistmt = ocistmt
	rset.Index = -1
	rset.offset = 0
	rset.fetched = 0
	rset.finished = false
	rset.Err = nil
	rset.log(_drv.cfg.Log.Rset.Open) // call log after rset.stmt is set
	// get the implcit select-list describe information; no server round-trip
	r := C.OCIStmtExecute(
		rset.stmt.ses.ocisvcctx,      //OCISvcCtx           *svchp,
		rset.ocistmt,                 //OCIStmt             *stmtp,
		rset.stmt.ses.srv.env.ocierr, //OCIError            *errhp,
		C.ub4(1),                     //ub4                 iters,
		C.ub4(0),                     //ub4                 rowoff,
		nil,                          //const OCISnapshot   *snap_in,
		nil,                          //OCISnapshot         *snap_out,
		C.OCI_DESCRIBE_ONLY)          //ub4                 mode );
	if r == C.OCI_ERROR {
		return rset.stmt.ses.srv.env.ociError()
	}
	// get the parameter count
	var paramCount C.ub4
	err := rset.attr(unsafe.Pointer(&paramCount), 4, C.OCI_ATTR_PARAM_COUNT)
	if err != nil {
		return err
	}
	// make defines slice
	rset.defs = make([]def, int(paramCount))
	rset.ColumnNames = make([]string, int(paramCount))
	rset.Row = make([]interface{}, int(paramCount))
	//fmt.Printf("rset.open (paramCount %v)\n", paramCount)

	// create parameters for each select-list column
	type paramS struct {
		columnSize uint32
		typeCode   C.ub2
		param      *C.OCIParam
	}
	params := make([]paramS, len(rset.defs))
	defer func() {
		for _, param := range params {
			if param.param == nil {
				continue
			}
			C.OCIDescriptorFree(unsafe.Pointer(param.param), C.OCI_DTYPE_PARAM)
		}
	}()

	var gct GoColumnType
	for n := range rset.defs {
		// Create oci parameter handle; may be freed by OCIDescriptorFree()
		// parameter position is 1-based
		r := C.OCIParamGet(
			unsafe.Pointer(rset.ocistmt),                        //const void        *hndlp,
			C.OCI_HTYPE_STMT,                                    //ub4               htype,
			rset.stmt.ses.srv.env.ocierr,                        //OCIError          *errhp,
			(*unsafe.Pointer)(unsafe.Pointer(&params[n].param)), //void              **parmdpp,
			C.ub4(n+1)) //ub4               pos );
		if r == C.OCI_ERROR {
			return rset.stmt.ses.srv.env.ociError()
		}
		ocipar := params[n].param
		// Get column size in bytes
		err = rset.paramAttr(ocipar, unsafe.Pointer(&params[n].columnSize), nil, C.OCI_ATTR_DATA_SIZE)
		if err != nil {
			return err
		}
		// Get oci data type code
		err = rset.paramAttr(ocipar, unsafe.Pointer(&params[n].typeCode), nil, C.OCI_ATTR_DATA_TYPE)
		if err != nil {
			return err
		}
		// Get column name
		var columnName *C.char
		var colSize C.ub4
		err := rset.paramAttr(ocipar, unsafe.Pointer(&columnName), &colSize, C.OCI_ATTR_NAME)
		if err != nil {
			return err
		}
		rset.ColumnNames[n] = C.GoStringN(columnName, C.int(colSize))
		rset.logF(_drv.cfg.Log.Rset.OpenDefs, "%d. %s/%d", n+1, rset.ColumnNames[n], params[n].typeCode)
	}

	rset.fetchLen = MaxFetchLen
Loop:
	for _, param := range params {
		switch param.typeCode {
		case C.SQLT_LNG, C.SQLT_BFILE, C.SQLT_BLOB, C.SQLT_CLOB, C.SQLT_LBI:
			rset.fetchLen = MinFetchLen
			break Loop
		}
	}

	for n := range rset.defs {
		ocipar := params[n].param
		ociTypeCode := params[n].typeCode
		columnSize := params[n].columnSize

		switch ociTypeCode {
		case C.SQLT_NUM:
			// NUMBER
			// Get precision
			var precision C.sb2
			err = rset.paramAttr(ocipar, unsafe.Pointer(&precision), nil, C.OCI_ATTR_PRECISION)
			if err != nil {
				return err
			}
			// Get scale (the number of decimal places)
			var scale C.sb1
			err = rset.paramAttr(ocipar, unsafe.Pointer(&scale), nil, C.OCI_ATTR_SCALE)
			if err != nil {
				return err
			}
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.numericColumnType(int(precision), int(scale))
			} else {
				err = checkNumericColumn(stmt.gcts[n], rset.ColumnNames[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			rset.logF(_drv.cfg.Log.Rset.OpenDefs, "%d. prec=%d scale=%d => gct=%s", n+1, precision, scale, GctName(gct))
			err := rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_IBDOUBLE:
			// BINARY_DOUBLE
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.binaryDouble
			} else {
				err = checkNumericColumn(stmt.gcts[n], rset.ColumnNames[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			err := rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_IBFLOAT:
			// BINARY_FLOAT
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.binaryFloat
			} else {
				err = checkNumericColumn(stmt.gcts[n], rset.ColumnNames[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			err := rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_DAT, C.SQLT_TIMESTAMP, C.SQLT_TIMESTAMP_TZ, C.SQLT_TIMESTAMP_LTZ:
			// DATE, TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIMEZONE
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				switch ociTypeCode {
				case C.SQLT_DAT:
					gct = rset.stmt.cfg.Rset.date
				case C.SQLT_TIMESTAMP:
					gct = rset.stmt.cfg.Rset.timestamp
				case C.SQLT_TIMESTAMP_TZ:
					gct = rset.stmt.cfg.Rset.timestampTz
				case C.SQLT_TIMESTAMP_LTZ:
					gct = rset.stmt.cfg.Rset.timestampLtz
				}
			} else {
				err = checkTimeColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			isNullable := false
			if gct == OraT {
				isNullable = true
			}
			def := rset.getDef(defIdxTime).(*defTime)
			rset.defs[n] = def
			err = def.define(n+1, isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_CHR:
			// VARCHAR, VARCHAR2, NVARCHAR2
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.varchar
			} else {
				err = checkStringColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			err = rset.defineString(n, columnSize, gct)
			if err != nil {
				return err
			}
		case C.SQLT_AFC:
			rset.logF(_drv.cfg.Log.Rset.OpenDefs, "%d. AFC size=%d", n+1, columnSize)
			//Log.Infof("rset AFC size=%d gct=%v", columnSize, gct)
			// CHAR, NCHAR
			// for char(1 char) columns, columnSize is 4 (AL32UTF8 charset)
			if columnSize == 1 || columnSize == 4 {
				if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
					gct = rset.stmt.cfg.Rset.char1
					rset.logF(_drv.cfg.Log.Rset.OpenDefs, "%d. AFC no gct, char1=%s", n+1, gct)
				} else {
					err = checkBoolOrStringColumn(stmt.gcts[n])
					if err != nil {
						return err
					}
					gct = stmt.gcts[n]
				}
				rset.logF(_drv.cfg.Log.Rset.OpenDefs, "%d. AFC gct=%s", n+1, gct)
				switch gct {
				case B, OraB:
					// Interpret single char as bool
					isNullable := false
					if gct == OraB {
						isNullable = true
					}
					def := rset.getDef(defIdxBool).(*defBool)
					rset.defs[n] = def
					err = def.define(n+1, int(columnSize), isNullable, rset)
					if err != nil {
						return err
					}
				case S, OraS:
					// Interpret single char as string
					rset.defineString(n, columnSize, gct)
				}
			} else {
				// Interpret as string
				if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
					gct = rset.stmt.cfg.Rset.char
				} else {
					err = checkStringColumn(stmt.gcts[n])
					if err != nil {
						return err
					}
					gct = stmt.gcts[n]
				}
				err = rset.defineString(n, columnSize, gct)
				if err != nil {
					return err
				}
			}
		case C.SQLT_LNG:
			// LONG
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.long
			} else {
				err = checkStringColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}

			// longBufferSize: Use a moderate default buffer size; 2GB max buffer may not be feasible on all clients
			err = rset.defineString(n, stmt.cfg.longBufferSize, gct)
			if err != nil {
				return err
			}
		case C.SQLT_CLOB:
			// CLOB, NCLOB
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.clob
			} else {
				err = checkStringColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			// Get character set form
			var charsetForm C.ub1
			err = rset.paramAttr(ocipar, unsafe.Pointer(&charsetForm), nil, C.OCI_ATTR_CHARSET_FORM)
			if err != nil {
				return err
			}
			def := rset.getDef(defIdxLob).(*defLob)
			rset.defs[n] = def
			err = def.define(n+1, charsetForm, C.SQLT_CLOB, gct, rset)
			if err != nil {
				return err
			}
		case C.SQLT_BLOB:
			// BLOB
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.blob
			} else {
				err = checkBinColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			def := rset.getDef(defIdxLob).(*defLob)
			rset.defs[n] = def
			err = def.define(n+1, C.SQLCS_IMPLICIT, C.SQLT_BLOB, gct, rset)
			if err != nil {
				return err
			}
		case C.SQLT_BIN:
			// RAW
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.raw
			} else {
				err = checkBinColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			isNullable := false
			if gct == OraBin {
				isNullable = true
			}
			def := rset.getDef(defIdxRaw).(*defRaw)
			rset.defs[n] = def
			err = def.define(n+1, int(columnSize), isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_LBI:
			//log(true, "LONG RAW")
			// LONG RAW
			if stmt.gcts == nil || n >= len(stmt.gcts) || stmt.gcts[n] == D {
				gct = rset.stmt.cfg.Rset.longRaw
			} else {
				err = checkBinColumn(stmt.gcts[n])
				if err != nil {
					return err
				}
				gct = stmt.gcts[n]
			}
			isNullable := false
			if gct == OraBin {
				isNullable = true
			}
			def := rset.getDef(defIdxLongRaw).(*defLongRaw)
			rset.defs[n] = def
			err = def.define(n+1, rset.stmt.cfg.longRawBufferSize, isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_YM:
			def := rset.getDef(defIdxIntervalYM).(*defIntervalYM)
			rset.defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_DS:
			def := rset.getDef(defIdxIntervalDS).(*defIntervalDS)
			rset.defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_FILE:
			// BFILE
			def := rset.getDef(defIdxBfile).(*defBfile)
			rset.defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_RDD:
			// ROWID, UROWID
			def := rset.getDef(defIdxRowid).(*defRowid)
			rset.defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_RSET:
			def := rset.getDef(defIdxRset).(*defRset)
			rset.defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		default:
			return errF("unsupported select-list column type (ociTypeCode: %v)", ociTypeCode)
		}
	}
	return nil
}

func (rset *Rset) defineString(n int, columnSize uint32, gct GoColumnType) (err error) {
	isNullable := false
	if gct == OraS {
		isNullable = true
	}
	def := rset.getDef(defIdxString).(*defString)
	rset.defs[n] = def
	err = def.define(n+1, int(columnSize), isNullable, rset)
	return err
}

func (rset *Rset) defineNumeric(n int, gct GoColumnType) (err error) {
	switch gct {
	case I64:
		def := rset.getDef(defIdxInt64).(*defInt64)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case I32:
		def := rset.getDef(defIdxInt32).(*defInt32)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case I16:
		def := rset.getDef(defIdxInt16).(*defInt16)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case I8:
		def := rset.getDef(defIdxInt8).(*defInt8)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case U64:
		def := rset.getDef(defIdxUint64).(*defUint64)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case U32:
		def := rset.getDef(defIdxUint32).(*defUint32)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case U16:
		def := rset.getDef(defIdxUint16).(*defUint16)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case U8:
		def := rset.getDef(defIdxUint8).(*defUint8)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case F64:
		def := rset.getDef(defIdxFloat64).(*defFloat64)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case F32:
		def := rset.getDef(defIdxFloat32).(*defFloat32)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case N:
		def := rset.getDef(defIdxOCINum).(*defOCINum)
		rset.defs[n] = def
		err = def.define(n+1, false, rset)
	case OraI64:
		def := rset.getDef(defIdxInt64).(*defInt64)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraI32:
		def := rset.getDef(defIdxInt32).(*defInt32)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraI16:
		def := rset.getDef(defIdxInt16).(*defInt16)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraI8:
		def := rset.getDef(defIdxInt8).(*defInt8)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraU64:
		def := rset.getDef(defIdxUint64).(*defUint64)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraU32:
		def := rset.getDef(defIdxUint32).(*defUint32)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraU16:
		def := rset.getDef(defIdxUint16).(*defUint16)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraU8:
		def := rset.getDef(defIdxUint8).(*defUint8)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraF64:
		def := rset.getDef(defIdxFloat64).(*defFloat64)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraF32:
		def := rset.getDef(defIdxFloat32).(*defFloat32)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	case OraN:
		def := rset.getDef(defIdxOCINum).(*defOCINum)
		rset.defs[n] = def
		err = def.define(n+1, true, rset)
	}
	return err
}

// paramAttr gets an attribute from the parameter handle.
func (rset *Rset) paramAttr(ocipar *C.OCIParam, attrup unsafe.Pointer, attrSizep *C.ub4, attrType C.ub4) error {
	if attrSizep == nil {
		attrSizep = new(C.ub4)
	}
	r := C.OCIAttrGet(
		unsafe.Pointer(ocipar),       //const void     *trgthndlp,
		C.OCI_DTYPE_PARAM,            //ub4            trghndltyp,
		attrup,                       //void           *attributep,
		attrSizep,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		rset.stmt.ses.srv.env.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (rset *Rset) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	r := C.OCIAttrGet(
		unsafe.Pointer(rset.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4            trghndltyp,
		attrup,                       //void           *attributep,
		&attrSize,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		rset.stmt.ses.srv.env.ocierr) //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

// sysName returns a string representing the Rset.
func (rset *Rset) sysName() string {
	return fmt.Sprintf("E%vS%vS%vS%vR%v", rset.stmt.ses.srv.env.id, rset.stmt.ses.srv.id, rset.stmt.ses.id, rset.stmt.id, rset.id)
}

// log writes a message with an Rset system name and caller info.
func (rset *Rset) log(enabled bool, v ...interface{}) {
	if !_drv.cfg.Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		_drv.cfg.Log.Logger.Infof("%v %v", rset.sysName(), callInfo(1))
	} else {
		_drv.cfg.Log.Logger.Infof("%v %v %v", rset.sysName(), callInfo(1), fmt.Sprint(v...))
	}
}

// log writes a formatted message with an Rset system name and caller info.
func (rset *Rset) logF(enabled bool, format string, v ...interface{}) {
	if !_drv.cfg.Log.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		_drv.cfg.Log.Logger.Infof("%v %v", rset.sysName(), callInfo(1))
	} else {
		_drv.cfg.Log.Logger.Infof("%v %v %v", rset.sysName(), callInfo(1), fmt.Sprintf(format, v...))
	}
}
