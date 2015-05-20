// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import "log"

// Log can be replaced with any type implementing the Logger interface.
//
// The default implementation uses the standard lib's log package.
//
// For a glog-based implementation, see github.com/ranaian/ora/glg.
// ora.Log = glg.Log
//
// For an gopkg.in/inconshreveable/log15.v2-based, see github.com/ranaian/ora/lg15.
// ora.Log = lg15.Log
var Log Logger = stdLog{}

// Logger interface is for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}

var _ Logger = stdLog{}

type stdLog struct{}

func (s stdLog) Infof(format string, args ...interface{})  { log.Printf(format, args...) }
func (s stdLog) Infoln(args ...interface{})                { log.Println(args...) }
func (s stdLog) Errorf(format string, args ...interface{}) { log.Printf("ERROR "+format, args...) }
func (s stdLog) Errorln(args ...interface{})               { log.Println(append([]interface{}{"ERROR "}, args...)...) }
