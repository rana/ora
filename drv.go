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
	"container/list"
	"database/sql/driver"
	"errors"
	"fmt"
)

// Drv is an Oracle database driver.
//
// Drv is not meant to be called by user-code.
//
// Drv implements the driver.Driver interface.
type Drv struct {
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

	// TODO: make setter,/getter and cascade to env when set
	// TODO: decide where to load the config file? From env may be best
	stmtCfg StmtCfg

	envId  Id
	srvId  Id
	conId  Id
	sesId  Id
	txId   Id
	stmtId Id
	rsetId Id
	envs   *list.List

	// An environment for use by the database/sql package.
	sqlEnv *Env

	// LogOpenEnv determines whether the Drv.OpenEnv method is logged.
	//
	// The default is true.
	LogOpenEnv bool

	// LogOpen determines whether the Drv.Open method is logged.
	//
	// The default is true.
	LogOpen bool
}

// log writes a message with caller info.
func (drv *Drv) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			Log.Infof("%v", callInfo(1))
		} else {
			Log.Infof("%v %v", callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with caller info.
func (drv *Drv) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			Log.Infof("%v", callInfo(1))
		} else {
			Log.Infof("%v %v", callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}

// err creates an error with caller info.
func (drv *Drv) err(v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprint(v...)))
	Log.Errorln(err)
	return err
}

// errF creates a formatted error with caller info.
func (drv *Drv) errF(format string, v ...interface{}) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), fmt.Sprintf(format, v...)))
	Log.Errorln(err)
	return err
}

// errE wraps an error with caller info.
func (drv *Drv) errE(e error) (err error) {
	err = errors.New(fmt.Sprintf("%v %v", errInfo(1), e.Error()))
	Log.Errorln(err)
	return err
}

// Open opens a connection to an Oracle server with the database/sql environment.
//
// This is meant to be called by the database/sql package only.
//
// As an alternative, create your own Env and call Env.OpenCon.
//
// Open is a member of the driver.Driver interface.
func (drv *Drv) Open(conStr string) (driver.Conn, error) {
	drv.log(drv.LogOpen)
	con, err := drv.sqlEnv.OpenCon(conStr)
	if err != nil {
		return nil, drv.errE(err)
	}
	return con, nil
}
