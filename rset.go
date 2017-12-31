// Copyright 2017 The Ora Authors. All rights reserved.
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
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	MaxFetchLen        = 1024
	DefaultFetchLen    = 128
	DefaultLOBFetchLen = 8

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
	sync.RWMutex

	id uint64
	// cached
	env       *Env
	stmt      *Stmt
	ocistmt   *C.OCIStmt
	defs      []def
	autoClose bool // whether the close of the rset shall close its parent stmt
	genByPool bool

	Row             []interface{}
	Columns         []Column
	index           int32
	err             error
	fetched, offset int64
	fetchLen        int
	finished        bool

	sysNamer
}

type Column struct {
	Name      string
	Type      C.ub2
	Length    uint32
	Precision C.sb2
	Scale     C.sb1
}

// Err returns the last error of the reesult set.
func (rset *Rset) Err() error {
	rset.RLock()
	err := rset.err
	rset.RUnlock()
	return err
}

// Len returns the number of rows retrieved.
func (rset *Rset) Len() int {
	return int(atomic.LoadInt32(&rset.index)) + 1
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
	if rset == nil {
		return false
	}
	rset.RLock()
	defer rset.RUnlock()
	return rset.stmt != nil && rset.ocistmt != nil && rset.env != nil
}

// ColumnIndex returns a map of column names to their respective indexes.
// Duplicate column names are not treated specially, the biggest index returned
// for each column name.
func (rset *Rset) ColumnIndex() map[string]int {
	if rset == nil {
		return nil
	}
	rset.RLock()
	defer rset.RUnlock()

	if rset.Columns == nil || len(rset.Columns) == 0 {
		return nil
	}

	index := make(map[string]int, len(rset.Columns))
	for i, c := range rset.Columns {
		index[c.Name] = i
	}
	return index
}

// closeWithRemove releases allocated resources and removes the Rset from the
// Stmt.openRsets list.
func (rset *Rset) closeWithRemove() (err error) {
	rset.RLock()
	rset.stmt.openRsets.remove(rset)
	rset.RUnlock()
	return rset.close()
}

// close releases allocated resources.
func (rset *Rset) close() (err error) {
	rset.log(_drv.Cfg().Log.Rset.Close)

	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
		if rset.genByPool { // recycle pool-generated Rset; don't recycle user-specfied Rset
			*rset = Rset{}
			_drv.rsetPool.Put(rset)
		}
	}()
	if err := rset.checkIsOpen(); err != nil {
		return er("Rset is closed.")
	}
	rset.Lock()
	defer rset.Unlock()

	errs := _drv.listPool.Get().(*list.List)
	// close defines
	for _, def := range rset.defs {
		if def == nil {
			continue
		}
		err0 := def.close()
		if err0 != nil {
			errs.PushBack(err0)
		}
	}

	rset.env = nil
	rset.stmt = nil
	rset.ocistmt = nil
	rset.defs = nil
	rset.Row = nil
	rset.Columns = nil
	// do not clear error in case of autoClose when error exists
	// clear error when rset in initialized
	//rset.err = nil
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
	rset.log(_drv.Cfg().Log.Rset.BeginRow)
	rset.Lock()
	defer rset.Unlock()

	fetched, offset, finished := rset.fetched, rset.offset, rset.finished
	ocistmt := rset.ocistmt

	rset.logF(_drv.Cfg().Log.Rset.BeginRow, "fetched=%d offset=%d finished=%t", fetched, offset, finished)
	if fetched > 0 && fetched > offset {
		atomic.AddInt32(&rset.index, 1)
		return nil
	}
	if finished {
		rset.log(_drv.Cfg().Log.Rset.BeginRow, "finished")
		return io.EOF
	}
	// check is open
	if ocistmt == nil {
		rset.log(_drv.Cfg().Log.Rset.BeginRow, "Rset is closed")
		return io.EOF
	}
	// allocate define descriptor handles
	if rset.env == nil {
		return errF("Rset env is closed")
	}
	env := rset.env
	for _, define := range rset.defs {
		//rset.logF(_drv.Cfg().Log.Rset.BeginRow, "defs[%d]=%#v", i, define)
		if define == nil {
			continue
		}
		err := define.alloc()
		if err != nil {
			return err
		}
	}

	rset.finished = false
	// fetch rset.fetchLen rows
	r := C.OCIStmtFetch2(
		rset.ocistmt,         //OCIStmt     *stmthp,
		env.ocierr,           //OCIError    *errhp,
		C.ub4(rset.fetchLen), //ub4         nrows,
		C.OCI_FETCH_NEXT,     //ub2         orientation,
		C.sb4(0),             //sb4         fetchOffset,
		C.OCI_DEFAULT)        //ub4         mode );
	if r == C.OCI_ERROR {
		err := env.ociError()
		return err
	} else if r == C.OCI_NO_DATA {
		rset.log(_drv.Cfg().Log.Rset.BeginRow, "OCI_NO_DATA")
		rset.finished = true
		fetchLen := rset.fetchLen
		if fetchLen == 1 {
			// return io.EOF to conform with database/sql/driver
			return io.EOF
		}
		// If OCIStmtFetch2 returns OCI_NO_DATA this does not mean that no data fetched,
		// this means that the number of fetched rows is less than the array size,
		// they are all fetched by this OCIStmtFetch2 call, and you do not need to
		// call OCIStmtFetch2 anymore.
		//
	}
	var rowsFetched C.ub4
	if err := rset.attr(unsafe.Pointer(&rowsFetched), 4, C.OCI_ATTR_ROWS_FETCHED); err != nil {
		return err
	}

	rset.fetched = int64(rowsFetched)
	rset.offset = 0
	err = nil
	if rset.fetched == 0 {
		rset.finished = true
		err = io.EOF
	} else {
		atomic.AddInt32(&rset.index, 1)
	}

	return err
}

// endRow deallocates a handle for each column.
func (rset *Rset) endRow() {
	rset.log(_drv.Cfg().Log.Rset.EndRow)
	rset.Lock()
	defer rset.Unlock()
	done := rset.finished && !(rset.fetched > 0 && rset.fetched > rset.offset)
	defs := rset.defs
	rset.offset++
	if !done {
		return
	}
	for _, define := range defs {
		if define != nil {
			define.free()
		}
	}
}

// Exhaust will cycle to the end of the Rset, to autoclose it.
func (rset *Rset) Exhaust() {
	if rset == nil {
		return
	}
	if !rset.IsOpen() {
		return
	}
	for {
		err := rset.beginRow()
		rset.endRow()
		if err != nil {
			return
		}
	}
}

// Next attempts to load a row of data from an Oracle buffer. True is returned
// when a row of data is retrieved. False is returned when no data is available.
//
// Retrieve the loaded row from the Rset.Row field. Rset.Row is updated
// on each call to Next. Rset.Row is set to nil when Next returns false.
//
// When Next returns false check Rset.Err() for any error that may have occured.
func (rset *Rset) Next() bool {
	rset.log(_drv.Cfg().Log.Rset.Next)
	erase := func(err error) {
		rset.Lock()
		rset.err = err
		rset.Row = nil
		autoClose := rset.autoClose
		rset.Unlock()
		// closing the Stmt will close this (and all) Rsets under it!
		if !autoClose {
			rset.close()
		} else {
			rset.RLock()
			stmt := rset.stmt
			rset.RUnlock()
			rset.closeWithRemove()
			stmt.Close()
		}
	}

	if err := rset.checkIsOpen(); err != nil {
		erase(err)
		return false
	}
	err := rset.beginRow()
	defer rset.endRow()
	rset.logF(_drv.Cfg().Log.Rset.Next, "beginRow=%v", err)
	if err != nil {
		// io.EOF means no more data; return nil err
		if err == io.EOF {
			err = nil
		}
		erase(err)
		return false
	}
	// populate column values
	rset.RLock()
	Row := rset.Row
	defs := rset.defs
	offset := rset.offset
	rset.RUnlock()
	for n, define := range defs {
		value, err := define.value(int(offset))
		//rset.logF(_drv.Cfg().Log.Rset.Next, "value[%d]=%v (%v)", n, value, err)
		if err != nil {
			erase(err)
			return false
		}
		Row[n] = value
	}
	rset.Lock()
	rset.defs = defs
	rset.Row = Row
	rset.Unlock()
	//rset.logF(_drv.Cfg().Log.Rset.Next, "Row=%#v", rset.Row)
	return true
}

// NextRow attempts to load a row from the Oracle buffer and return the row.
// Nil is returned when there's no data.
//
// When NextRow returns nil check Rset.Err() for any error that may have occured.
func (rset *Rset) NextRow() []interface{} {
	rset.Next()
	rset.RLock()
	defer rset.RUnlock()
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
	logCfg := _drv.Cfg().Log
	rset.Lock()
	defer rset.Unlock()

	env := rset.env
	if env == nil {
		rset.log(logCfg.Rset.Open, "env is nil")
		return io.EOF
	}
	rset.stmt = stmt
	rset.ocistmt = ocistmt
	atomic.StoreInt32(&rset.index, -1)
	rset.offset = 0
	rset.fetched = 0
	rset.finished = false
	rset.err = nil
	defs, Columns, Row := rset.defs, rset.Columns, rset.Row
	rset.defs, rset.Columns, rset.Row = nil, nil, nil

	stmt.RLock()
	ses := stmt.ses
	stmt.RUnlock()
	rset.log(logCfg.Rset.Open) // call log after rset.stmt is set
	// get the implcit select-list describe information; no server round-trip
	//fmt.Fprintf(os.Stdout, "stmt=%#v rset=%#v\n", stmt, rset)
	r := C.OCIStmtExecute(
		ses.ocisvcctx,       //OCISvcCtx           *svchp,
		ocistmt,             //OCIStmt             *stmtp,
		env.ocierr,          //OCIError            *errhp,
		C.ub4(1),            //ub4                 iters,
		C.ub4(0),            //ub4                 rowoff,
		nil,                 //const OCISnapshot   *snap_in,
		nil,                 //OCISnapshot         *snap_out,
		C.OCI_DESCRIBE_ONLY) //ub4                 mode );
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	// get the parameter count
	var paramCount C.ub4
	err := rset.attr(unsafe.Pointer(&paramCount), 4, C.OCI_ATTR_PARAM_COUNT)
	if err != nil {
		return err
	}
	// make defines slice
	if cap(defs) < int(paramCount) {
		defs = make([]def, int(paramCount))
	} else {
		defs = defs[:int(paramCount)]
	}
	if cap(Columns) < len(defs) {
		Columns = make([]Column, len(defs))
	} else {
		Columns = Columns[:len(defs)]
	}
	if cap(Row) < len(defs) {
		Row = make([]interface{}, len(defs))
	} else {
		Row = Row[:len(Row)]
	}

	//fmt.Printf("rset.open (paramCount %v)\n", paramCount)

	// create parameters for each select-list column
	type paramS struct {
		columnSize uint32
		typeCode   C.ub2
		param      *C.OCIParam
	}
	params := make([]paramS, len(defs))
	defer func() {
		for _, param := range params {
			if param.param == nil {
				continue
			}
			C.OCIDescriptorFree(unsafe.Pointer(param.param), C.OCI_DTYPE_PARAM)
		}
	}()

	var gct GoColumnType

	for n := range defs {
		// Create oci parameter handle; may be freed by OCIDescriptorFree()
		// parameter position is 1-based
		r := C.OCIParamGet(
			unsafe.Pointer(rset.ocistmt), //const void        *hndlp,
			C.OCI_HTYPE_STMT,             //ub4               htype,
			env.ocierr,                   //OCIError          *errhp,
			(*unsafe.Pointer)(unsafe.Pointer(&params[n].param)), //void              **parmdpp,
			C.ub4(n+1)) //ub4               pos );
		if r == C.OCI_ERROR {
			return env.ociError()
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

		Columns[n] = Column{
			Name:   C.GoStringN(columnName, C.int(colSize)),
			Type:   params[n].typeCode,
			Length: params[n].columnSize,
		}

		rset.logF(logCfg.Rset.OpenDefs, "%d. %s/%d", n+1, Columns[n].Name, params[n].typeCode)
	}

	fetchLen, lobFetchLen := DefaultFetchLen, DefaultLOBFetchLen
	cfg := rset.stmt.Cfg()
	if cfg.fetchLen > 0 {
		fetchLen = cfg.fetchLen
	}
	if cfg.lobFetchLen > 0 {
		lobFetchLen = cfg.lobFetchLen
	}

	if fetchLen != lobFetchLen {
	Loop:
		for _, param := range params {
			switch param.typeCode {
			// These can consume a lot of memory.
			case C.SQLT_LNG, C.SQLT_BFILE, C.SQLT_BLOB, C.SQLT_CLOB, C.SQLT_LBI:
				fetchLen = lobFetchLen
				break Loop
			}
		}
	}
	//rset.logF(true, "fetchLen=%d", fetchLen)

	rset.defs, rset.Columns, rset.Row = defs, Columns, Row
	rset.fetchLen = fetchLen

	//rset.logF(logCfg.Rset.Open, "cfg=%#v", cfg)
	stmt.RLock()
	gcts := stmt.gcts
	stmt.RUnlock()
	for n := range defs {
		ocipar := params[n].param
		ociTypeCode := params[n].typeCode
		columnSize := params[n].columnSize

		switch ociTypeCode {
		case C.SQLT_NUM, C.SQLT_INT: // TimesTen may return an SQLT_INT
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
			rset.Columns[n].Precision = precision
			rset.Columns[n].Scale = scale
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.numericColumnType(int(precision), int(scale))
			} else {
				err = checkNumericColumn(gcts[n], rset.Columns[n].Name)
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			rset.logF(logCfg.Rset.OpenDefs, "%d. prec=%d scale=%d => gct=%s", n+1, precision, scale, GctName(gct))
			defs[n], err = rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_IBDOUBLE:
			// BINARY_DOUBLE
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.binaryDouble
			} else {
				err = checkNumericColumn(gcts[n], rset.Columns[n].Name)
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			defs[n], err = rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_IBFLOAT:
			// BINARY_FLOAT
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.binaryFloat
			} else {
				err = checkNumericColumn(gcts[n], rset.Columns[n].Name)
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			defs[n], err = rset.defineNumeric(n, gct)
			if err != nil {
				return err
			}
		case C.SQLT_DAT:
			// DATE
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.date
			} else {
				err = checkTimeColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			isNullable := false
			if gct == OraT {
				isNullable = true
			}
			def := rset.getDef(defIdxDate).(*defDate)
			defs[n] = def
			err = def.define(n+1, isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_TIMESTAMP, C.SQLT_TIMESTAMP_TZ, C.SQLT_TIMESTAMP_LTZ:
			// TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIMEZONE
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				switch ociTypeCode {
				case C.SQLT_TIMESTAMP:
					gct = cfg.timestamp
				case C.SQLT_TIMESTAMP_TZ:
					gct = cfg.timestampTz
				case C.SQLT_TIMESTAMP_LTZ:
					gct = cfg.timestampLtz
				}
			} else {
				err = checkTimeColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			isNullable := false
			if gct == OraT {
				isNullable = true
			}
			def := rset.getDef(defIdxTime).(*defTime)
			defs[n] = def
			err = def.define(n+1, isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_CHR:
			// VARCHAR, VARCHAR2, NVARCHAR2
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.varchar
			} else {
				err = checkStringColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			defs[n], err = rset.defineString(n, columnSize, gct, false)
			if err != nil {
				return err
			}
		case C.SQLT_AFC:
			// CHAR, NCHAR
			rset.logF(logCfg.Rset.OpenDefs, "%d. AFC size=%d", n+1, columnSize)
			//Log.Infof("rset AFC size=%d gct=%v", columnSize, gct)
			// for char(1 char) columns, columnSize is 4 (AL32UTF8 charset)
			if columnSize == 1 || columnSize == 4 {
				if gcts == nil || n >= len(gcts) || gcts[n] == D {
					gct = cfg.char1
					rset.logF(logCfg.Rset.OpenDefs, "%d. AFC no gct, char1=%s", n+1, gct)
				} else {
					err = checkBoolOrStringColumn(gcts[n])
					if err != nil {
						return err
					}
					gct = gcts[n]
				}
				rset.logF(logCfg.Rset.OpenDefs, "%d. AFC gct=%s", n+1, gct)
				switch gct {
				case B, OraB:
					// Interpret single char as bool
					isNullable := false
					if gct == OraB {
						isNullable = true
					}
					def := rset.getDef(defIdxBool).(*defBool)
					defs[n] = def
					err = def.define(n+1, int(columnSize), isNullable, rset)
					if err != nil {
						return err
					}
				case S, OraS:
					// Interpret single char as string
					defs[n], err = rset.defineString(n, columnSize, gct, true)
					if err != nil {
						return err
					}
				}
			} else {
				// Interpret as string
				if gcts == nil || n >= len(gcts) || gcts[n] == D {
					gct = cfg.char
				} else {
					err = checkStringColumn(gcts[n])
					if err != nil {
						return err
					}
					gct = gcts[n]
				}
				defs[n], err = rset.defineString(n, columnSize, gct, true)
				if err != nil {
					return err
				}
			}
		case C.SQLT_LNG:
			// LONG
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.long
			} else {
				err = checkStringColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}

			// longBufferSize: Use a moderate default buffer size; 2GB max buffer may not be feasible on all clients
			defs[n], err = rset.defineString(n, stmt.Cfg().longBufferSize, gct, false)
			if err != nil {
				return err
			}
		case C.SQLT_CLOB:
			// CLOB, NCLOB
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.clob
			} else if gcts[n] == L {
				gct = L
			} else {
				err = checkStringColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}

			def := rset.getDef(defIdxLob).(*defLob)
			defs[n] = def
			err = def.define(n+1, C.SQLT_CLOB, gct, rset)
			if err != nil {
				return err
			}
		case C.SQLT_BLOB:
			// BLOB
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.blob
			} else if gcts[n] == L {
				gct = L
			} else {
				err = checkBinColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			def := rset.getDef(defIdxLob).(*defLob)
			defs[n] = def
			err = def.define(n+1, C.SQLT_BLOB, gct, rset)
			if err != nil {
				return err
			}
		case C.SQLT_BIN:
			// RAW
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.raw
			} else {
				err = checkBinColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			isNullable := false
			if gct == OraBin {
				isNullable = true
			}
			def := rset.getDef(defIdxRaw).(*defRaw)
			defs[n] = def
			err = def.define(n+1, int(columnSize), isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_LBI:
			//log(true, "LONG RAW")
			// LONG RAW
			if gcts == nil || n >= len(gcts) || gcts[n] == D {
				gct = cfg.longRaw
			} else {
				err = checkBinColumn(gcts[n])
				if err != nil {
					return err
				}
				gct = gcts[n]
			}
			isNullable := false
			if gct == OraBin {
				isNullable = true
			}
			def := rset.getDef(defIdxLongRaw).(*defLongRaw)
			defs[n] = def
			err = def.define(n+1, cfg.longRawBufferSize, isNullable, rset)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_YM:
			def := rset.getDef(defIdxIntervalYM).(*defIntervalYM)
			defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_INTERVAL_DS:
			def := rset.getDef(defIdxIntervalDS).(*defIntervalDS)
			defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_FILE:
			// BFILE
			def := rset.getDef(defIdxBfile).(*defBfile)
			defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_RDD:
			// ROWID, UROWID
			def := rset.getDef(defIdxRowid).(*defRowid)
			defs[n] = def
			err = def.define(n+1, rset)
			if err != nil {
				return err
			}
		case C.SQLT_RSET:
			def := rset.getDef(defIdxRset).(*defRset)
			defs[n] = def
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

func (rset *Rset) defineString(n int, columnSize uint32, gct GoColumnType, rTrim bool) (def, error) {
	isNullable := false
	if gct == OraS {
		isNullable = true
	}
	rTrim = rTrim && rset.stmt.Cfg().RTrimChar
	D := rset.getDef(defIdxString).(*defString)
	return D, D.define(n+1, int(columnSize), isNullable, rTrim, rset)
}

func (rset *Rset) defineNumeric(n int, gct GoColumnType) (def, error) {
	var nullable bool
	var D def
	switch gct {
	case I64:
		D = rset.getDef(defIdxInt64).(*defInt64)
	case I32:
		D = rset.getDef(defIdxInt32).(*defInt32)
	case I16:
		D = rset.getDef(defIdxInt16).(*defInt16)
	case I8:
		D = rset.getDef(defIdxInt8).(*defInt8)
	case U64:
		D = rset.getDef(defIdxUint64).(*defUint64)
	case U32:
		D = rset.getDef(defIdxUint32).(*defUint32)
	case U16:
		D = rset.getDef(defIdxUint16).(*defUint16)
	case U8:
		D = rset.getDef(defIdxUint8).(*defUint8)
	case F64:
		D = rset.getDef(defIdxFloat64).(*defFloat64)
	case F32:
		D = rset.getDef(defIdxFloat32).(*defFloat32)
	case N:
		D = rset.getDef(defIdxOCINum).(*defOCINum)
	case OraI64:
		D = rset.getDef(defIdxInt64).(*defInt64)
		nullable = true
	case OraI32:
		D = rset.getDef(defIdxInt32).(*defInt32)
		nullable = true
	case OraI16:
		D = rset.getDef(defIdxInt16).(*defInt16)
		nullable = true
	case OraI8:
		D = rset.getDef(defIdxInt8).(*defInt8)
		nullable = true
	case OraU64:
		D = rset.getDef(defIdxUint64).(*defUint64)
		nullable = true
	case OraU32:
		D = rset.getDef(defIdxUint32).(*defUint32)
		nullable = true
	case OraU16:
		D = rset.getDef(defIdxUint16).(*defUint16)
		nullable = true
	case OraU8:
		D = rset.getDef(defIdxUint8).(*defUint8)
		nullable = true
	case OraF64:
		D = rset.getDef(defIdxFloat64).(*defFloat64)
		nullable = true
	case OraF32:
		D = rset.getDef(defIdxFloat32).(*defFloat32)
		nullable = true
	case OraN:
		D = rset.getDef(defIdxOCINum).(*defOCINum)
		nullable = true
	case S:
		D = rset.getDef(defIdxNumString).(*defNumString)
	}
	return D, D.(interface {
		define(int, bool, *Rset) error
	}).define(n+1, nullable, rset)
}

// paramAttr gets an attribute from the parameter handle.
func (rset *Rset) paramAttr(ocipar *C.OCIParam, attrup unsafe.Pointer, attrSizep *C.ub4, attrType C.ub4) error {
	if attrSizep == nil {
		attrSizep = new(C.ub4)
	}
	env := rset.env
	r := C.OCIAttrGet(
		unsafe.Pointer(ocipar), //const void     *trgthndlp,
		C.OCI_DTYPE_PARAM,      //ub4            trghndltyp,
		attrup,                 //void           *attributep,
		attrSizep,              //ub4            *sizep,
		attrType,               //ub4            attrtype,
		env.ocierr)             //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

// attr gets an attribute from the statement handle.
func (rset *Rset) attr(attrup unsafe.Pointer, attrSize C.ub4, attrType C.ub4) error {
	env := rset.env
	r := C.OCIAttrGet(
		unsafe.Pointer(rset.ocistmt), //const void     *trgthndlp,
		C.OCI_HTYPE_STMT,             //ub4            trghndltyp,
		attrup,                       //void           *attributep,
		&attrSize,                    //ub4            *sizep,
		attrType,                     //ub4            attrtype,
		rset.env.ocierr)              //OCIError       *errhp );
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

// sysName returns a string representing the Rset.
func (rset *Rset) sysName() string {
	if rset == nil {
		return "E_S_S_S_"
	}
	return rset.sysNamer.Name(func() string { return fmt.Sprintf("%sS%v", rset.stmt.sysName(), rset.id) })
}

// log writes a message with an Rset system name and caller info.
func (rset *Rset) log(enabled bool, v ...interface{}) {
	logCfg := _drv.Cfg().Log
	if !logCfg.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		logCfg.Logger.Infof("%v %v", rset.sysName(), callInfo(1))
	} else {
		logCfg.Logger.Infof("%v %v %v", rset.sysName(), callInfo(1), fmt.Sprint(v...))
	}
}

// log writes a formatted message with an Rset system name and caller info.
func (rset *Rset) logF(enabled bool, format string, v ...interface{}) {
	logCfg := _drv.Cfg().Log
	if !logCfg.IsEnabled(enabled) {
		return
	}
	if len(v) == 0 {
		logCfg.Logger.Infof("%v %v", rset.sysName(), callInfo(1))
	} else {
		logCfg.Logger.Infof("%v %v %v", rset.sysName(), callInfo(1), fmt.Sprintf(format, v...))
	}
}
