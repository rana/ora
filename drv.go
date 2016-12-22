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
	"time"
)

// DrvCfg represents configuration values for the ora package.
type DrvCfg struct {
	Env *EnvCfg
	Log LogDrvCfg
}

// NewDrvCfg creates a DrvCfg with default values.
func NewDrvCfg() *DrvCfg {
	c := &DrvCfg{}
	c.Env = NewEnvCfg()
	c.Log = NewLogDrvCfg()
	return c
}

// LogDrvCfg represents package-level logging configuration values.
type LogDrvCfg struct {
	// Logger writes log messages.
	// Logger can be replaced with any type implementing the Logger interface.
	//
	// The default implementation uses the standard lib's log package.
	//
	// For a glog-based implementation, see gopkg.in/rana/ora.v3/glg.
	// LogDrvCfg.Logger = glg.Log
	//
	// For an gopkg.in/inconshreveable/log15.v2-based, see gopkg.in/rana/ora.v3/lg15.
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
	cfg      DrvCfg
	mu       sync.Mutex
	insMu    sync.Mutex
	updMu    sync.Mutex
	delMu    sync.Mutex
	selMu    sync.Mutex
	addTblMu sync.Mutex

	envId  Id
	srvId  Id
	conId  Id
	sesId  Id
	txId   Id
	stmtId Id
	rsetId Id

	listPool *pool
	envPool  *pool
	conPool  *pool
	srvPool  *pool
	sesPool  *pool
	stmtPool *pool
	txPool   *pool
	rsetPool *pool
	bndPools []*pool
	defPools []*pool

	locationsMu sync.RWMutex
	locations   map[string]*time.Location

	sqlPkgEnv *Env // An environment for use by the database/sql package.
	openEnvs  *envList
}

// Open opens a connection to an Oracle server with the database/sql environment.
//
// This is intended to be called by the database/sql package only.
//
// Alternatively, you may call Env.OpenCon to create an *ora.Con.
//
// Open is a member of the driver.Driver interface.
func (drv *Drv) Open(conStr string) (driver.Conn, error) {
	logL("Open", true)
	con, err := _drv.sqlPkgEnv.OpenCon(conStr)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return con, nil
}
