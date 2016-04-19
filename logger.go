// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

var (
	Infof  func(format string, args ...interface{})
	Debugf func(format string, args ...interface{})
)

// Logger interface is for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}

type EmpLgr struct{}

func (e EmpLgr) Infof(format string, v ...interface{}) {
	if Debugf == nil {
		return
	}
	Debugf(format, v)
}
func (e EmpLgr) Infoln(v ...interface{}) {
	if Debugf == nil {
		return
	}
	Debugf("%v", v)
}
func (e EmpLgr) Errorf(format string, v ...interface{}) {
	if Infof == nil {
		return
	}
	Infof(format, v)
}
func (e EmpLgr) Errorln(v ...interface{}) {
	if Infof == nil {
		return
	}
	Infof("%v", v)
}
