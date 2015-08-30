// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
	"fmt"
)

// LogConCfg represents Con logging configuration values.
type LogConCfg struct {
	// Close determines whether the Con.Close method is logged.
	//
	// The default is true.
	Close bool

	// Prepare determines whether the Con.Prepare method is logged.
	//
	// The default is true.
	Prepare bool

	// Begin determines whether the Con.Begin method is logged.
	//
	// The default is true.
	Begin bool

	// Ping determines whether the Con.Ping method is logged.
	//
	// The default is true.
	Ping bool
}

// NewLogConCfg creates a LogTxCfg with default values.
func NewLogConCfg() LogConCfg {
	c := LogConCfg{}
	c.Close = true
	c.Prepare = true
	c.Begin = true
	c.Ping = true
	return c
}

// Con is an Oracle connection associated with a server and session.
//
// Implements the driver.Conn interface.
type Con struct {
	id uint64

	env *Env
	srv *Srv
	ses *Ses
}

// checkIsOpen validates that the connection is open.
func (con *Con) checkIsOpen() error {
	if !con.IsOpen() {
		return er("Con is closed.")
	}
	return nil
}

// IsOpen returns true when the connection to the Oracle server is open;
// otherwise, false.
//
// Calling Close will cause IsOpen to return false.
// Once closed, a connection cannot be re-opened.
// To open a new connection call Open on a driver.
func (con *Con) IsOpen() bool {
	return con.env != nil
}

// Close ends a session and disconnects from an Oracle server.
//
// Close is a member of the driver.Conn interface.
func (con *Con) Close() (err error) {
	con.env.openCons.remove(con)
	return con.close()
}

// close ends a session and disconnects from an Oracle server.
// does not remove Con from Ses.openCons
func (con *Con) close() (err error) {
	con.log(_drv.cfg.Log.Con.Close)
	if err := con.checkIsOpen(); err != nil {
		return err
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
		con.env = nil
		con.srv = nil
		con.ses = nil
		_drv.conPool.Put(con)
	}()

	// TODO(rana): RECONSIDER HOW SRV.CLOSE IS CALLED
	err1 := con.ses.Close()
	err2 := con.srv.Close()
	m := newMultiErr(err1, err2)
	if m != nil {
		err = *m
	}
	return err
}

// Prepare readies a sql string for use.
//
// Prepare is a member of the driver.Conn interface.
func (con *Con) Prepare(sql string) (driver.Stmt, error) {
	con.log(_drv.cfg.Log.Con.Prepare)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	stmt, err := con.ses.Prep(sql)
	if err != nil {
		return nil, err
	}
	return &DrvStmt{stmt: stmt}, err
}

// Begin starts a transaction.
//
// Begin is a member of the driver.Conn interface.
func (con *Con) Begin() (driver.Tx, error) {
	con.log(_drv.cfg.Log.Con.Begin)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	tx, err := con.ses.StartTx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Ping makes a round-trip call to an Oracle server to confirm that the connection is active.
func (con *Con) Ping() error {
	con.log(_drv.cfg.Log.Con.Ping)
	if err := con.checkIsOpen(); err != nil {
		return err
	}
	return con.ses.Ping()
}

// sysName returns a string representing the Con.
func (con *Con) sysName() string {
	return fmt.Sprintf("E%vS%vS%vC%v", con.ses.srv.env.id, con.ses.srv.id, con.ses.id, con.id)
}

// log writes a message with an Con system name and caller info.
func (con *Con) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", con.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", con.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

// log writes a formatted message with an Con system name and caller info.
func (con *Con) logF(enabled bool, format string, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.cfg.Log.Logger.Infof("%v %v", con.sysName(), callInfo(1))
		} else {
			_drv.cfg.Log.Logger.Infof("%v %v %v", con.sysName(), callInfo(1), fmt.Sprintf(format, v...))
		}
	}
}
