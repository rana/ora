// Copyright 2015 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package lg

import (
	"fmt"
	"log"
	"os"
)

var Log = Std{L: log.New(os.Stderr, "[ora] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)}

type Std struct {
	L *log.Logger
}

func (s Std) Infof(format string, v ...interface{}) {
	s.L.SetPrefix("ORA I ")
	s.L.Output(2, fmt.Sprintf(format, v...))
}
func (s Std) Infoln(v ...interface{}) {
	s.L.SetPrefix("ORA I ")
	s.L.Output(2, fmt.Sprintln(v...))
}
func (s Std) Errorf(format string, v ...interface{}) {
	s.L.SetPrefix("ORA E ")
	s.L.Output(2, fmt.Sprintf(format, v...))
}
func (s Std) Errorln(v ...interface{}) {
	s.L.SetPrefix("ORA E ")
	s.L.Output(2, fmt.Sprintln(v...))
}
