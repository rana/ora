// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package lg15

import (
	"fmt"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

var Log = lgr{log15.New("lib", "ora")}

type lgr struct {
	log15.Logger
}

func (s lgr) Infof(format string, args ...interface{})  { s.Debug(fmt.Sprintf(format, args...)) }
func (s lgr) Infoln(args ...interface{})                { s.Debug(strings.Join(asStrings(args), " ")) }
func (s lgr) Errorf(format string, args ...interface{}) { s.Error(fmt.Sprintf(format, args...)) }
func (s lgr) Errorln(args ...interface{})               { s.Error(strings.Join(asStrings(args), " ")) }

func asStrings(args ...interface{}) []string {
	arr := make([]string, len(args))
	for i, a := range args {
		if s, ok := a.(string); ok {
			arr[i] = s
			continue
		}
		if s, ok := a.(fmt.Stringer); ok {
			arr[i] = s.String()
			continue
		}
		arr[i] = fmt.Sprintf("%v", a)
	}
	return arr
}
