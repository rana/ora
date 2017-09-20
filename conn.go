// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"context"
	"database/sql/driver"
	"fmt"

	"golang.org/x/sync/errgroup"
)

/*
#include <oci.h>
*/
import "C"

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
	ses *Ses

	sysNamer
}

// checkIsOpen validates that the connection is open.
func (con *Con) checkIsOpen() error {
	if !con.IsOpen() {
		return driver.ErrBadConn
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
	con.log(_drv.Cfg().Log.Con.Close)
	if err = con.checkIsOpen(); err != nil {
		return err
	}
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
		con.env = nil
		con.ses = nil
		_drv.conPool.Put(con)
	}()

	// Close the session, and its srv, too!
	if ses := con.ses; ses != nil {
		srv := ses.srv
		err := ses.Close()
		if srv != nil {
			srv.Close()
		}
		return err
	}
	return nil
}

// Begin starts a transaction.
//
// Begin is a member of the driver.Conn interface.
func (con *Con) Begin() (driver.Tx, error) {
	con.log(_drv.Cfg().Log.Con.Begin)
	if err := con.checkIsOpen(); err != nil {
		return nil, err
	}
	tx, err := con.ses.StartTx()
	if err != nil {
		return nil, maybeBadConn(err)
	}
	return tx, nil
}

// Ping makes a round-trip call to an Oracle server to confirm that the connection is active.
func (con *Con) Ping(ctx context.Context) error {
	con.log(_drv.Cfg().Log.Con.Ping)
	if err := con.checkIsOpen(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return maybeBadConn(con.ses.Ping())
	})
	if err := ctx.Err(); err != nil {
		if isCanceled(err) {
			con.ses.Break()
		}
		return err
	}
	return grp.Wait()
}

// sysName returns a string representing the Con.
func (con *Con) sysName() string {
	if con == nil {
		return "E_S_S_C_"
	}
	return con.sysNamer.Name(func() string {
		return fmt.Sprintf("%sC%v", con.ses.sysName(), con.id)
	})
}

// log writes a message with an Con system name and caller info.
func (con *Con) log(enabled bool, v ...interface{}) {
	if enabled {
		if len(v) == 0 {
			_drv.Cfg().Log.Logger.Infof("%v %v", con.sysName(), callInfo(1))
		} else {
			_drv.Cfg().Log.Logger.Infof("%v %v %v", con.sysName(), callInfo(1), fmt.Sprint(v...))
		}
	}
}

func isCanceled(err error) bool {
	return err != nil && (err == context.Canceled || err == context.DeadlineExceeded)
}
func maybeBadConn(err error) error {
	if err == nil {
		return nil
	}
	// database/sql API expect driver.ErrBadConn to reconnect to the database
	if cd, ok := err.(interface {
		Code() int
	}); ok {
		switch cd.Code() {
		case 3113, 3114, 12528, 12545, 28547:
			// ORA-03113: end-of-file on communication channel
			// ORA-03114: not connected to ORACLE
			// ORA-12528: TNS:listener: all appropriate instances are blocking new connections
			// ORA-12545: Connect failed because target host or object does not exist
			// ORA-28547: connection to server failed, probable Oracle Net admin error
			return driver.ErrBadConn
		}
	}
	return err
}
