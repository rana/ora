// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"container/list"
	"database/sql/driver"
	"github.com/golang/glog"
)

// Con is an Oracle connection associated with a server and session.
//
// Implements the driver.Conn interface.
type Con struct {
	conId uint64

	env  *Env
	srv  *Srv
	ses  *Ses
	elem *list.Element
}

// checkIsOpen validates that the connection is open.
func (con *Con) checkIsOpen() error {
	if !con.IsOpen() {
		return errNewF("Con is closed (conId %v)", con.conId)
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
	if err := con.checkIsOpen(); err != nil {
		return err
	}
	glog.Infof("E%vC%v Close", con.env.envId, con.conId)
	defer func() {
		if value := recover(); value != nil {
			glog.Errorln(recoverMsg(value))
			err = errRecover(value)
		}

		env := con.env
		env.cons.Remove(con.elem)
		con.env = nil
		con.srv = nil
		con.ses = nil
		con.elem = nil
		env.drv.conPool.Put(con)
	}()

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
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vC%v Prepare", con.env.envId, con.conId)
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
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	glog.Infof("E%vC%v Begin", con.env.envId, con.conId)
	tx, err := con.ses.StartTx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Ping makes a round-trip call to an Oracle server to confirm that the connection is active.
func (con *Con) Ping() error {
	if err := con.checkIsOpen(); err != nil {
		return err
	}
	glog.Infof("E%vC%v Ping", con.env.envId, con.conId)
	return con.srv.Ping()
}
