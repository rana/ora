// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

// Package date implements encoding of 7 byte Oracle DATE storage formats.
package date

import "time"

// Date is an OCIDate
//
// SQLT_ODT: 7 bytes
//
// http://www.orafaq.com/wiki/Date
//
/*
The internal format is the following one:

    century + 100
    year in the century + 100
    month
    day
    hour + 1
    minute + 1
    second + 1

So in the previous example the date was 19-DEC-2007 at 22:35:10.
*/
type Date [7]byte

func (dt Date) Set(t time.Time) {
	y := t.Year()
	dt[0] = byte(y/100 + 100)
	dt[1] = byte(y%100 + 100)
	dt[2] = byte(t.Month())
	dt[3] = byte(t.Day())
	dt[4] = byte(t.Hour() + 1)
	dt[5] = byte(t.Minute() + 1)
	dt[6] = byte(t.Second() + 1)
}

func (dt Date) Get() time.Time {
	return time.Date(
		int((dt[0]-100)*100+(dt[1]-100)),
		time.Month(dt[2]),
		int(dt[3]),
		int(dt[4]-1),
		int(dt[5]-1),
		int(dt[6]-1),
		0,
		time.Local,
	)
}
