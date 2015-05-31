// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// Log can be replaced with any type implementing the Logger interface.
//
// The default implementation uses the standard lib's log package.
//
// For a glog-based implementation, see github.com/rana/ora/glg.
// ora.Log = glg.Log
//
// For an gopkg.in/inconshreveable/log15.v2-based, see github.com/rana/ora/lg15.
// ora.Log = lg15.Log
var Log Logger = empLgr{}

// Logger interface is for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}

type empLgr struct{}

func (e empLgr) Infof(format string, v ...interface{})  {}
func (e empLgr) Infoln(v ...interface{})                {}
func (e empLgr) Errorf(format string, v ...interface{}) {}
func (e empLgr) Errorln(v ...interface{})               {}
