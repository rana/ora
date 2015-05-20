// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package glg

import "github.com/golang/glog"

var Log = gLogger{}

type gLogger struct{}

func (s gLogger) Infof(format string, args ...interface{})  { glog.Infof(format, args...) }
func (s gLogger) Infoln(args ...interface{})                { glog.Infoln(args...) }
func (s gLogger) Errorf(format string, args ...interface{}) { glog.Errorf(format, args...) }
func (s gLogger) Errorln(args ...interface{})               { glog.Errorln(args...) }
