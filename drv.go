// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
*/
import "C"
import (
	"database/sql/driver"
	"sync"
	"sync/atomic"
	"time"
)

// DrvCfg represents configuration values for the ora package.
type DrvCfg struct {
	StmtCfg
	Log LogDrvCfg
}

// NewDrvCfg creates a DrvCfg with default values.
func NewDrvCfg() DrvCfg {
	return DrvCfg{StmtCfg: NewStmtCfg(), Log: NewLogDrvCfg()}
}

func (cfg DrvCfg) SetStmtCfg(stmtCfg StmtCfg) DrvCfg {
	cfg.StmtCfg = stmtCfg
	return cfg
}

func (c DrvCfg) SetPrefetchRowCount(prefetchRowCount uint32) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetPrefetchRowCount(prefetchRowCount)
	return c
}
func (c DrvCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetPrefetchMemorySize(prefetchMemorySize)
	return c
}
func (c DrvCfg) SetLongBufferSize(size uint32) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetLongBufferSize(size)
	return c
}
func (c DrvCfg) SetLongRawBufferSize(size uint32) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetLongRawBufferSize(size)
	return c
}
func (c DrvCfg) SetLobBufferSize(size int) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetLobBufferSize(size)
	return c
}
func (c DrvCfg) SetStringPtrBufferSize(size int) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetStringPtrBufferSize(size)
	return c
}
func (c DrvCfg) SetByteSlice(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetByteSlice(gct)
	return c
}
func (c DrvCfg) SetNumberInt(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetNumberInt(gct)
	return c
}
func (c DrvCfg) SetNumberBigInt(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetNumberBigInt(gct)
	return c
}
func (c DrvCfg) SetNumberFloat(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetNumberFloat(gct)
	return c
}
func (c DrvCfg) SetNumberBigFloat(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetNumberBigFloat(gct)
	return c
}
func (c DrvCfg) SetBinaryDouble(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetBinaryDouble(gct)
	return c
}
func (c DrvCfg) SetBinaryFloat(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetBinaryFloat(gct)
	return c
}
func (c DrvCfg) SetFloat(gct GoColumnType) DrvCfg { c.StmtCfg = c.StmtCfg.SetFloat(gct); return c }
func (c DrvCfg) SetDate(gct GoColumnType) DrvCfg  { c.StmtCfg = c.StmtCfg.SetDate(gct); return c }
func (c DrvCfg) SetTimestamp(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetTimestamp(gct)
	return c
}
func (c DrvCfg) SetTimestampTz(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetTimestampTz(gct)
	return c
}
func (c DrvCfg) SetTimestampLtz(gct GoColumnType) DrvCfg {
	c.StmtCfg = c.StmtCfg.SetTimestampLtz(gct)
	return c
}
func (c DrvCfg) SetChar1(gct GoColumnType) DrvCfg   { c.StmtCfg = c.StmtCfg.SetChar1(gct); return c }
func (c DrvCfg) SetChar(gct GoColumnType) DrvCfg    { c.StmtCfg = c.StmtCfg.SetChar(gct); return c }
func (c DrvCfg) SetVarchar(gct GoColumnType) DrvCfg { c.StmtCfg = c.StmtCfg.SetVarchar(gct); return c }
func (c DrvCfg) SetLong(gct GoColumnType) DrvCfg    { c.StmtCfg = c.StmtCfg.SetLong(gct); return c }
func (c DrvCfg) SetClob(gct GoColumnType) DrvCfg    { c.StmtCfg = c.StmtCfg.SetClob(gct); return c }
func (c DrvCfg) SetBlob(gct GoColumnType) DrvCfg    { c.StmtCfg = c.StmtCfg.SetBlob(gct); return c }
func (c DrvCfg) SetRaw(gct GoColumnType) DrvCfg     { c.StmtCfg = c.StmtCfg.SetRaw(gct); return c }
func (c DrvCfg) SetLongRaw(gct GoColumnType) DrvCfg { c.StmtCfg = c.StmtCfg.SetLongRaw(gct); return c }

func (c DrvCfg) SetLogger(lgr Logger) DrvCfg { c.Log.Logger = lgr; return c }

// LogDrvCfg represents package-level logging configuration values.
type LogDrvCfg struct {
	// Logger writes log messages.
	// Logger can be replaced with any type implementing the Logger interface.
	//
	// The default implementation uses the standard lib's log package.
	//
	// For a glog-based implementation, see gopkg.in/rana/ora.v4/glg.
	// LogDrvCfg.Logger = glg.Log
	//
	// For an gopkg.in/inconshreveable/log15.v2-based, see gopkg.in/rana/ora.v4/lg15.
	// LogDrvCfg.Logger = lg15.Log
	Logger Logger

	// OpenEnv determines whether the ora.OpenEnv method is logged.
	//
	// The default is true.
	OpenEnv bool

	// Ins determines whether the ora.Ins method is logged.
	//
	// The default is true.
	Ins bool

	// Upd determines whether the ora.Upd method is logged.
	//
	// The default is true.
	Upd bool

	// Del determines whether the ora.Del method is logged.
	//
	// The default is true.
	Del bool

	// Sel determines whether the ora.Sel method is logged.
	//
	// The default is true.
	Sel bool

	// AddTbl determines whether the ora.AddTbl method is logged.
	//
	// The default is true.
	AddTbl bool

	Env  LogEnvCfg
	Srv  LogSrvCfg
	Ses  LogSesCfg
	Stmt LogStmtCfg
	Tx   LogTxCfg
	Con  LogConCfg
	Rset LogRsetCfg
}

// NewLogDrvCfg creates a LogDrvCfg with default values.
func NewLogDrvCfg() LogDrvCfg {
	c := LogDrvCfg{}
	c.Logger = EmpLgr{}
	c.OpenEnv = true
	c.Ins = true
	c.Upd = true
	c.Del = true
	c.Sel = true
	c.AddTbl = true
	c.Env = NewLogEnvCfg()
	c.Srv = NewLogSrvCfg()
	c.Ses = NewLogSesCfg()
	c.Stmt = NewLogStmtCfg()
	c.Tx = NewLogTxCfg()
	c.Con = NewLogConCfg()
	c.Rset = NewLogRsetCfg()
	return c
}

// IsEnabled returns whether the logger is enabled (and enabled is true).
func (c LogDrvCfg) IsEnabled(enabled bool) bool {
	if !enabled || c.Logger == nil {
		return false
	}
	_, ok := c.Logger.(EmpLgr)
	return !ok
}

// Drv represents an Oracle database driver.
//
// Drv is not meant to be called by user-code.
//
// Drv implements the driver.Driver interface.
type Drv struct {
	sync.RWMutex

	cfg atomic.Value

	envId  Id
	srvId  Id
	conId  Id
	sesId  Id
	txId   Id
	stmtId Id
	rsetId Id

	listPool *sync.Pool
	envPool  *sync.Pool
	conPool  *sync.Pool
	srvPool  *sync.Pool
	sesPool  *sync.Pool
	stmtPool *sync.Pool
	txPool   *sync.Pool
	rsetPool *sync.Pool
	bndPools []*sync.Pool
	defPools []*sync.Pool

	locationsMu sync.RWMutex
	locations   map[string]*time.Location

	sqlPkgEnv *Env // An environment for use by the database/sql package.
	openEnvs  *envList
}

func (drv *Drv) Cfg() DrvCfg {
	c := drv.cfg.Load()
	//fmt.Fprintf(os.Stderr, "%p.Cfg=%#v\n", drv, c)
	if c == nil || c.(DrvCfg).IsZero() {
		return NewDrvCfg()
	}
	return c.(DrvCfg)
}
func (drv *Drv) SetCfg(cfg DrvCfg) {
	//fmt.Fprintf(os.Stderr, "%p.SetCfg(%#v)\n", drv, cfg)
	drv.cfg.Store(cfg)
}

// Open opens a connection to an Oracle server with the database/sql environment.
//
// This is intended to be called by the database/sql package only.
//
// Alternatively, you may call Env.OpenCon to create an *ora.Con.
//
// Open is a member of the driver.Driver interface.
func (drv *Drv) Open(conStr string) (driver.Conn, error) {
	log(true)

	drv.RLock()
	env := drv.sqlPkgEnv
	drv.RUnlock()
	con, err := env.OpenCon(conStr)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return con, nil
}
