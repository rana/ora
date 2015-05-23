// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

import (
	"fmt"
	"log"
	"os"
)

// Log can be replaced with any type implementing the Logger interface.
//
// The default implementation uses the standard lib's log package.
//
// For a glog-based implementation, see github.com/ranaian/ora/glg.
// ora.Log = glg.Log
//
// For an gopkg.in/inconshreveable/log15.v2-based, see github.com/ranaian/ora/lg15.
// ora.Log = lg15.Log
var Log Logger = stdLog{l: log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)}

// Logger interface is for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}

type stdLog struct {
	l *log.Logger
}

func (s stdLog) Infof(format string, v ...interface{}) {
	s.l.SetPrefix("ORA I ")
	s.l.Output(2, fmt.Sprintf(format, v...))
}
func (s stdLog) Infoln(v ...interface{}) {
	s.l.SetPrefix("ORA I ")
	s.l.Output(2, fmt.Sprintln(v...))
}
func (s stdLog) Errorf(format string, v ...interface{}) {
	s.l.SetPrefix("ORA E ")
	s.l.Output(2, fmt.Sprintf(format, v...))
}
func (s stdLog) Errorln(v ...interface{}) {
	s.l.SetPrefix("ORA E ")
	s.l.Output(2, fmt.Sprintln(v...))
}
