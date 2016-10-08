// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

// Package date implements encoding of 7 byte Oracle DATE storage formats.
package date

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

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

func (dt *Date) Set(t time.Time) {
	if t.IsZero() {
		for i := range dt[:] {
			dt[i] = 0
		}
		return
	}
	y := t.Year()
	if y < -4711 {
		y = -4711
	} else if y > 9999 {
		y = 9999
	}
	dt[0] = byte(y/100 + 100)
	dt[1] = byte(y%100 + 100)
	dt[2] = byte(t.Month())
	dt[3] = byte(t.Day())
	dt[4] = byte(t.Hour() + 1)
	dt[5] = byte(t.Minute() + 1)
	dt[6] = byte(t.Second() + 1)
}

func (dt Date) Bytes() []byte {
	return dt[:]
}

func (dt Date) IsNull() bool {
	for _, b := range dt[:] {
		if b != 0 {
			return false
		}
	}
	return true
}
func (dt Date) MarshalJSON() ([]byte, error) {
	if dt.IsNull() {
		return []byte("null"), nil
	}
	return dt.Get().MarshalJSON()
}
func (dt *Date) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		for i := range dt[:] {
			dt[i] = 0
		}
		return nil
	}
	var t time.Time
	if err := json.Unmarshal(p, &t); err != nil {
		return err
	}
	dt.Set(t)
	return nil
}

// FromTime returns a Date from a time.Time
// Does the allocation inside, so easier to use.
func FromTime(t time.Time) Date {
	var dt Date
	dt.Set(t)
	return dt
}

func (dt Date) Equal(other Date) bool {
	return bytes.Equal(dt[:], other[:])
}

func (dt Date) String() string {
	if dt.IsNull() {
		return (time.Time{}).Format("2006-01-02T15:04:05")
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d",
		(int(dt[0])-100)*100+(int(dt[1])-100),
		time.Month(dt[2]),
		int(dt[3]),
		int(dt[4]-1),
		int(dt[5]-1),
		int(dt[6]-1),
	)
}

func (dt Date) Get() time.Time {
	return dt.GetIn(nil)
}
func (dt Date) GetIn(zone *time.Location) time.Time {
	if dt.IsNull() {
		return time.Time{}
	}
	if zone == nil {
		zone = time.Local
	}
	return time.Date(
		(int(dt[0])-100)*100+(int(dt[1])-100),
		time.Month(dt[2]),
		int(dt[3]),
		int(dt[4]-1),
		int(dt[5]-1),
		int(dt[6]-1),
		0,
		zone,
	)
}
