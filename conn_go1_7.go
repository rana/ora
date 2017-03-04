// +build !go1.8

// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import "database/sql/driver"

// Prepare readies a sql string for use.
//
// Prepare is a member of the driver.Conn interface.
func (con *Con) Prepare(query string) (driver.Stmt, error) {
	con.log(_drv.Cfg().Log.Con.Prepare)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	stmt, err := con.ses.Prep(query)
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return &DrvStmt{stmt: stmt}, err
}
