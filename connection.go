// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"database/sql/driver"
)

// An Oracle connection associated with a session and server.
//
// Implements the driver.Conn interface.
type Connection struct {
	env *Environment
	srv *Server
	ses *Session
}

// Ping makes a round-trip call to an Oracle server to confirm that the connection is active.
func (con *Connection) Ping() error {
	// Validate that the connection is open
	err := con.checkIsOpen()
	if err != nil {
		return err
	}
	return con.srv.Ping()
}

// Prepare readies a sql string for use.
//
// Prepare is a member of the driver.Conn interface.
func (con *Connection) Prepare(sql string) (driver.Stmt, error) {
	// Validate that the connection is open
	err := con.checkIsOpen()
	if err != nil {
		return nil, err
	}

	stmt, err := con.ses.Prepare(sql)
	if err != nil {
		return nil, err
	}
	return stmt, err
}

// Begin starts a transaction.
//
// Begin is a member of the driver.Conn interface.
func (con *Connection) Begin() (driver.Tx, error) {
	// Validate that the connection is open
	err := con.checkIsOpen()
	if err != nil {
		return nil, err
	}

	tx, err := con.ses.BeginTransaction()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// checkIsOpen validates that the connection is open.
//
// ErrClosedConnection is returned if the connection is closed.
func (con *Connection) checkIsOpen() error {
	if !con.IsOpen() {
		return errNew("open connection prior to method call")
	}
	return nil
}

// IsOpen returns true when the connection to the Oracle server is open;
// otherwise, false.
//
// Calling Close will cause IsOpen to return false.
// Once closed, a connection cannot be re-opened.
// To open a new connection call Open on a driver.
func (con *Connection) IsOpen() bool {
	return con.srv != nil
}

// Close ends a session and disconnects from an Oracle server.
//
// Close is a member of the driver.Conn interface.
func (con *Connection) Close() (err error) {
	if con.IsOpen() {
		err = con.ses.Close()
		if err != nil {
			return err
		}
		err = con.srv.Close()
		if err != nil {
			return err
		}
		con.srv = nil
		con.ses = nil

		// Put connection in pool
		con.env.conPool.Put(con)
	}
	return nil
}
