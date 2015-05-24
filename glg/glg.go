// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package glg

import (
	"fmt"
	"github.com/golang/glog"
)

var Log = gLgr{}

type gLgr struct{}

func (l gLgr) Infof(format string, v ...interface{}) {
	glog.InfoDepth(2, fmt.Sprintf(format, v...))
}
func (l gLgr) Infoln(v ...interface{}) {
	glog.InfoDepth(2, v...)
}
func (l gLgr) Errorf(format string, v ...interface{}) {
	glog.ErrorDepth(2, fmt.Sprintf(format, v...))
}
func (l gLgr) Errorln(v ...interface{}) {
	glog.ErrorDepth(2, v...)
}
