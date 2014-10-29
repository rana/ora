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
	environment *Environment
	server      *Server
	session     *Session
}

// Ping makes a round-trip call to an Oracle server to confirm that the connection is active.
func (connection *Connection) Ping() error {
	// Validate that the connection is open
	err := connection.checkIsOpen()
	if err != nil {
		return err
	}
	return connection.server.Ping()
}

// Prepare readies a sql string for use.
//
// Prepare is a member of the driver.Conn interface.
func (connection *Connection) Prepare(sql string) (driver.Stmt, error) {
	// Validate that the connection is open
	err := connection.checkIsOpen()
	if err != nil {
		return nil, err
	}

	statement, err := connection.session.Prepare(sql)
	if err != nil {
		return nil, err
	}
	return statement, err
}

// Begin starts a transaction.
//
// Begin is a member of the driver.Conn interface.
func (connection *Connection) Begin() (driver.Tx, error) {
	// Validate that the connection is open
	err := connection.checkIsOpen()
	if err != nil {
		return nil, err
	}

	transaction, err := connection.session.BeginTransaction()
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// checkIsOpen validates that the connection is open.
//
// ErrClosedConnection is returned if the connection is closed.
func (connection *Connection) checkIsOpen() error {
	if !connection.IsOpen() {
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
func (connection *Connection) IsOpen() bool {
	return connection.server != nil
}

// Close ends a session and disconnects from an Oracle server.
//
// Close is a member of the driver.Conn interface.
func (connection *Connection) Close() (err error) {
	if connection.IsOpen() {
		err = connection.session.Close()
		if err != nil {
			return err
		}
		err = connection.server.Close()
		if err != nil {
			return err
		}
		connection.server = nil
		connection.session = nil

		// Put connection in pool
		connection.environment.connectionPool.Put(connection)
	}
	return nil
}
