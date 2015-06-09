// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"container/list"
	"io"
	"io/ioutil"
	"time"
)

// When a parent handle is freed, all child handles associated with it are also
// freed, and can no longer be used. For example, when a statement handle is freed,
// any bind and define handles associated with it are also freed.

// bnd represents an between a Go parameter and a sql statement placeholder and
// contains logic to transfer a Go type to an Oracle OCI type.
type bnd interface {
	// setPtr enables some bind types to set out-bound pointers for some types such as time.Time, etc.
	setPtr() error
	// close releases resources and resets fields.
	close() error
}

// def represents a select-list column definition containing logic to transfer
// an Oracle OCI type to a Go type.
type def interface {
	// value gets a Go value from an Oracle buffer.
	value() (interface{}, error)
	// alloc allocates an OCI descriptor.
	alloc() error
	// free releases an OCI descriptor.
	free()
	// close releases resources and resets fields.
	close() error
}

// Int64 is a nullable int64.
type Int64 struct {
	IsNull bool
	Value  int64
}

// Equals returns true when the receiver and specified Int64 are both null,
// or when the receiver and specified Int64 are both not null and Values are equal.
func (this Int64) Equals(other Int64) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Int32 is a nullable int32.
type Int32 struct {
	IsNull bool
	Value  int32
}

// Equals returns true when the receiver and specified Int32 are both null,
// or when the receiver and specified Int32 are both not null and Values are equal.
func (this Int32) Equals(other Int32) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Int16 is a nullable int16.
type Int16 struct {
	IsNull bool
	Value  int16
}

// Equals returns true when the receiver and specified Int16 are both null,
// or when the receiver and specified Int16 are both not null and Values are equal.
func (this Int16) Equals(other Int16) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Int8 is a nullable int8.
type Int8 struct {
	IsNull bool
	Value  int8
}

// Equals returns true when the receiver and specified Int8 are both null,
// or when the receiver and specified Int8 are both not null and Values are equal.
func (this Int8) Equals(other Int8) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Uint64 is a nullable uint64.
type Uint64 struct {
	IsNull bool
	Value  uint64
}

// Equals returns true when the receiver and specified Uint64 are both null,
// or when the receiver and specified Uint64 are both not null and Values are equal.
func (this Uint64) Equals(other Uint64) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Uint32 is a nullable uint32.
type Uint32 struct {
	IsNull bool
	Value  uint32
}

// Equals returns true when the receiver and specified Uint32 are both null,
// or when the receiver and specified Uint32 are both not null and Values are equal.
func (this Uint32) Equals(other Uint32) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Uint16 is a nullable uint16.
type Uint16 struct {
	IsNull bool
	Value  uint16
}

// Equals returns true when the receiver and specified Uint16 are both null,
// or when the receiver and specified Uint16 are both not null and Values are equal.
func (this Uint16) Equals(other Uint16) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Uint8 is a nullable uint8.
type Uint8 struct {
	IsNull bool
	Value  uint8
}

// Equals returns true when the receiver and specified Uint8 are both null,
// or when the receiver and specified Uint8 are both not null and Values are equal.
func (this Uint8) Equals(other Uint8) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Float64 is a nullable float64.
type Float64 struct {
	IsNull bool
	Value  float64
}

// Equals returns true when the receiver and specified Float64 are both null,
// or when the receiver and specified Float64 are both not null and Values are equal.
func (this Float64) Equals(other Float64) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Float32 is a nullable float32.
type Float32 struct {
	IsNull bool
	Value  float32
}

// Equals returns true when the receiver and specified Float32 are both null,
// or when the receiver and specified Float32 are both not null and Values are equal.
func (this Float32) Equals(other Float32) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Time is a nullable time.Time.
type Time struct {
	IsNull bool
	Value  time.Time
}

// Equals returns true when the receiver and specified Time are both null,
// or when the receiver and specified Time are both not null and Values are equal.
func (this Time) Equals(other Time) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value.Equal(other.Value))
}

// String is a nullable string.
type String struct {
	IsNull bool
	Value  string
}

// Equals returns true when the receiver and specified String are both null,
// or when the receiver and specified String are both not null and Values are equal.
func (this String) Equals(other String) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Bool is a nullable bool.
type Bool struct {
	IsNull bool
	Value  bool
}

// Equals returns true when the receiver and specified Bool are both null,
// or when the receiver and specified Bool are both not null and Values are equal.
func (this Bool) Equals(other Bool) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value == other.Value)
}

// Raw represents a nullable byte slice for RAW or LONG RAW Oracle values.
type Raw struct {
	IsNull bool
	Value  []byte
}

// Equals returns true when the receiver and specified Raw are both null,
// or when the receiver and specified Raw are both not null and Values are equal.
func (this Raw) Equals(other Raw) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull &&
			bytes.Equal(this.Value, other.Value))
}

type Lob struct {
	io.Reader
	io.Closer
}

// Equals returns true when the receiver and specified Lob are both null,
// or when they both not null and share the same Reader.
func (this Lob) Equals(other Lob) bool {
	return this.Reader == other.Reader // this is a quite strict equality...
}

func (this Lob) Bytes() ([]byte, error) {
	if this.Reader == nil {
		return nil, io.EOF
	}
	return ioutil.ReadAll(this.Reader)
}

// Bfile represents a nullable BFILE Oracle value.
type Bfile struct {
	IsNull         bool
	DirectoryAlias string
	Filename       string
}

// Equals returns true when the receiver and specified Bfile are both null,
// or when the receiver and specified Bfile are both not null, DirectoryAlias are equal
// and Filename are equal.
func (this Bfile) Equals(other Bfile) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.DirectoryAlias == other.DirectoryAlias && this.Filename == other.Filename)
}

// IntervalYM represents a nullable INTERVAL YEAR TO MONTH Oracle value.
type IntervalYM struct {
	IsNull bool
	Year   int32
	Month  int32
}

// Equals returns true when the receiver and specified IntervalYM are both null,
// or when the receiver and specified IntervalYM are both not null, Year are equal
// and Month are equal.
func (this IntervalYM) Equals(other IntervalYM) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Year == other.Year && this.Month == other.Month)
}

// ShiftTime returns a new Time with IntervalYM applied.
func (this IntervalYM) ShiftTime(t time.Time) time.Time {
	return t.AddDate(int(this.Year), int(this.Month), 0)
}

// IntervalDS represents a nullable INTERVAL DAY TO SECOND Oracle value.
type IntervalDS struct {
	IsNull     bool
	Day        int32
	Hour       int32
	Minute     int32
	Second     int32
	Nanosecond int32
}

// Equals returns true when the receiver and specified IntervalDS are both null,
// or when the receiver and specified IntervalDS are both not null, and all other
// fields are equal.
func (this IntervalDS) Equals(other IntervalDS) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull &&
			this.Day == other.Day &&
			this.Hour == other.Hour &&
			this.Minute == other.Minute &&
			this.Second == other.Second &&
			this.Nanosecond == other.Nanosecond)
}

// ShiftTime returns a new Time with IntervalDS applied.
func (this IntervalDS) ShiftTime(t time.Time) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	return time.Date(year, month, day+int(this.Day), hour+int(this.Hour), min+int(this.Minute), sec+int(this.Second), t.Nanosecond()+int(this.Nanosecond), t.Location())
}

// MultiErr holds multiple errors in a single string.
type MultiErr struct {
	str string
}

// Error returns one or more errors.
//
// Error is a member of the 'error' interface.
func (m MultiErr) Error() string {
	return m.str
}

// newMultiErr returns a MultiErr or nil.
// It is valid to pass nil errors to newMultiErr.
// Nil errors will be filtered out. If all errors
// are nil newMultiError will return nil.
func newMultiErr(errs ...error) *MultiErr {
	var buf bytes.Buffer
	for _, err := range errs {
		if err != nil {
			buf.WriteString(err.Error())
			buf.WriteString(", ")
		}
	}
	if buf.Len() > 0 {
		return &MultiErr{str: buf.String()}
	} else {
		return nil
	}
}

// newMultiErrL returns a MultiErr or nil.
// It is valid to pass nil errors to newMultiErr.
// Nil errors will be filtered out. If all errors
// are nil newMultiError will return nil.
func newMultiErrL(errs *list.List) *MultiErr {
	var buf bytes.Buffer
	for e := errs.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			err := e.Value.(error)
			buf.WriteString(err.Error())
			buf.WriteString(", ")
		}
	}
	if buf.Len() > 0 {
		return &MultiErr{str: buf.String()}
	} else {
		return nil
	}
}
