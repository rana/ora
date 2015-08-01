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
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"sync"
	"time"
)

// When a parent handle is freed, all child handles associated with it are also
// freed, and can no longer be used. For example, when a statement handle is freed,
// any bind and define handles associated with it are also freed.
//
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

var _ = (json.Marshaler)(Int64{})
var _ = (json.Unmarshaler)((*Int64)(nil))

func (this Int64) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Int64) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Int32{})
var _ = (json.Unmarshaler)((*Int32)(nil))

func (this Int32) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Int32) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	err := json.Unmarshal(p, &this.Value)
	return err
}

// Int16 is a nullable int16.
type Int16 struct {
	IsNull bool
	Value  int16
}

var _ = (json.Marshaler)(Int16{})
var _ = (json.Unmarshaler)((*Int16)(nil))

func (this Int16) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Int16) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Int8{})
var _ = (json.Unmarshaler)((*Int8)(nil))

func (this Int8) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Int8) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, (*int8)(&this.Value))
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

var _ = (json.Marshaler)(Uint64{})
var _ = (json.Unmarshaler)((*Uint64)(nil))

func (this Uint64) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Uint64) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Uint32{})
var _ = (json.Unmarshaler)((*Uint32)(nil))

func (this Uint32) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Uint32) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Uint16{})
var _ = (json.Unmarshaler)((*Uint16)(nil))

func (this Uint16) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Uint16) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Uint8{})
var _ = (json.Unmarshaler)((*Uint8)(nil))

func (this Uint8) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Uint8) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Float64{})
var _ = (json.Unmarshaler)((*Float64)(nil))

func (this Float64) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Float64) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Float32{})
var _ = (json.Unmarshaler)((*Float32)(nil))

func (this Float32) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Float32) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Time{})
var _ = (json.Unmarshaler)((*Time)(nil))

func (this Time) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Time) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
}

// Date is a nullable time.Time, but only with second precision.
type Date struct {
	IsNull bool
	Value  time.Time
}

// Equals returns true when the receiver and specified Date are both null,
// or when the receiver and specified Date are both not null and Values are equal.
func (this Date) Equals(other Date) bool {
	return (this.IsNull && other.IsNull) ||
		(this.IsNull == other.IsNull && this.Value.Equal(other.Value))
}

var _ = (json.Marshaler)(Date{})
var _ = (json.Unmarshaler)((*Date)(nil))

func (this Date) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Date) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, (*time.Time)(&this.Value))
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
func (this String) String() string {
	if this.IsNull {
		return ""
	}
	return this.Value
}

var _ = (json.Marshaler)(String{})
var _ = (json.Unmarshaler)((*String)(nil))

func (this String) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	if this.Value == "" {
		return []byte(`""`), nil
	}
	return json.Marshal(this.Value)
}
func (this *String) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Bool{})
var _ = (json.Unmarshaler)((*Bool)(nil))

func (this Bool) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	if this.Value {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}
func (this *Bool) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
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

var _ = (json.Marshaler)(Raw{})
var _ = (json.Unmarshaler)((*Raw)(nil))

func (this Raw) MarshalJSON() ([]byte, error) {
	if this.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(this.Value)
}
func (this *Raw) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.IsNull = true
		return nil
	}
	this.IsNull = false
	return json.Unmarshal(p, &this.Value)
}

// Lob's Reader is sent to the DB on bind, if not nil.
// The Reader can read the LOB if we bind a *Lob, Closer will close the LOB.
type Lob struct {
	io.Reader
	io.Closer
}

func (this Lob) Close() error {
	if this.Closer != nil {
		return this.Closer.Close()
	}
	return nil
}

// Equals returns true when the receiver and specified Lob are both null,
// or when they both not null and share the same Reader.
func (this Lob) Equals(other Lob) bool {
	return this.Reader == other.Reader // this is a quite strict equality...
}

// Bytes will read the contents of the Lob.Reader, and will keep that for future.
func (this *Lob) Bytes() ([]byte, error) {
	if this.Reader == nil {
		return nil, io.EOF
	}
	if br, ok := this.Reader.(bytesPeeker); ok {
		return br.PeekBytes(), nil
	}
	p, err := ioutil.ReadAll(this.Reader)
	if err != nil {
		return p, err
	}
	this.Reader = bytesReader{p: p, Reader: bytes.NewReader(p)}
	return p, nil
}

var _ = (json.Marshaler)((*Lob)(nil))
var _ = (json.Unmarshaler)((*Lob)(nil))

func (this *Lob) MarshalJSON() ([]byte, error) {
	if this.Reader == nil {
		return []byte("null"), nil
	}
	p, err := this.Bytes()
	if err != nil {
		return nil, err
	}
	return json.Marshal(p)
}
func (this *Lob) UnmarshalJSON(p []byte) error {
	if bytes.Equal(p, []byte("null")) || bytes.Equal(p, []byte(`""`)) {
		this.Reader = nil
		return nil
	}
	var b []byte
	err := json.Unmarshal(p, &b)
	this.Reader = bytesReader{p: p, Reader: bytes.NewReader(p)}
	return err
}

type bytesReader struct {
	p []byte
	io.Reader
}

var _ = bytesPeeker(bytesReader{})
var _ = io.Reader(bytesReader{})

func (br bytesReader) PeekBytes() []byte {
	return br.p
}

type bytesPeeker interface {
	PeekBytes() []byte
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

////////////////////////////////////////////////////////////////////////////////
// pool
////////////////////////////////////////////////////////////////////////////////
// prefer simple pool instead of sync.Pool to control lifetime of instances
// sync.Pool eliminates instances at its own discretion
// would prefer to maintain instances indefinitely; or, eventually clean on a long timer
type pool struct {
	l       *list.List
	mu      sync.Mutex
	genItem func() interface{}
}

func newPool(genItem func() interface{}) *pool {
	p := &pool{}
	p.l = list.New()
	p.genItem = genItem
	return p
}

func (p *pool) Get() interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.l.Len() > 0 {
		return p.l.Remove(p.l.Front())
	} else {
		return p.genItem()
	}
}

func (p *pool) Put(v interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.l.PushFront(v)
}

type Id struct {
	val uint64
	mu  sync.Mutex
}

func (id *Id) nextId() (result uint64) {
	id.mu.Lock()
	defer id.mu.Unlock()
	if id.val == math.MaxUint64 {
		id.val = 1
	} else {
		id.val++
	}
	return id.val
}

////////////////////////////////////////////////////////////////////////////////
// envList
////////////////////////////////////////////////////////////////////////////////
type envList struct {
	items []*Env
	mu    sync.Mutex
}

func newEnvList() *envList {
	return &envList{items: make([]*Env, 0, 2)}
}

func (l *envList) add(e *Env) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Env, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, e) // append item
}

func (l *envList) remove(e *Env) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == e {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *envList) setAllCfg(cfg *EnvCfg) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		l.items[n].SetCfg(cfg)
	}
}

func (l *envList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *envList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// srvList
////////////////////////////////////////////////////////////////////////////////
type srvList struct {
	items []*Srv
	mu    sync.Mutex
}

func newSrvList() *srvList {
	return &srvList{items: make([]*Srv, 0, 2)}
}

func (l *srvList) add(s *Srv) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Srv, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, s) // append item
}

func (l *srvList) remove(s *Srv) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == s {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *srvList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Srv from openSrvs
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Srvs from srvList
}

func (l *srvList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *srvList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// conList
////////////////////////////////////////////////////////////////////////////////
type conList struct {
	items []*Con
	mu    sync.Mutex
}

func newConList() *conList {
	return &conList{items: make([]*Con, 0, 8)}
}

func (l *conList) add(c *Con) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Con, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, c) // append item
}

func (l *conList) remove(c *Con) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == c {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *conList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Con from openCons
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Cons from conList
}

func (l *conList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *conList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// sesList
////////////////////////////////////////////////////////////////////////////////
type sesList struct {
	items []*Ses
	mu    sync.Mutex
}

func newSesList() *sesList {
	return &sesList{items: make([]*Ses, 0, 8)}
}

func (l *sesList) add(s *Ses) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Ses, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, s) // append item
}

func (l *sesList) remove(s *Ses) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == s {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *sesList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Ses from openSess
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Sess from sesList
}

func (l *sesList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *sesList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// txList
////////////////////////////////////////////////////////////////////////////////
type txList struct {
	items []*Tx
	mu    sync.Mutex
}

func newTxList() *txList {
	return &txList{items: make([]*Tx, 0, 2)}
}

func (l *txList) add(t *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Tx, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, t) // append item
}

func (l *txList) remove(t *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == t {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *txList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Tx from openTxs
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Txs from txList
}

func (l *txList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *txList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// stmtList
////////////////////////////////////////////////////////////////////////////////
type stmtList struct {
	items []*Stmt
	mu    sync.Mutex
}

func newStmtList() *stmtList {
	return &stmtList{items: make([]*Stmt, 0, 8)}
}

func (l *stmtList) add(s *Stmt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Stmt, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, s) // append item
}

func (l *stmtList) remove(s *Stmt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == s {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *stmtList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Stmt from openStmts
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Stmts from stmtList
}

func (l *stmtList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *stmtList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}

////////////////////////////////////////////////////////////////////////////////
// rsetList
////////////////////////////////////////////////////////////////////////////////
type rsetList struct {
	items []*Rset
	mu    sync.Mutex
}

func newRsetList() *rsetList {
	return &rsetList{items: make([]*Rset, 0, 8)}
}

func (l *rsetList) add(r *Rset) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) == cap(l.items) { // double capacity if needed
		cap := cap(l.items)
		if cap == 0 {
			cap = 4
		}
		tmp := make([]*Rset, 0, 2*cap)
		copy(tmp, l.items)
		l.items = tmp
	}
	l.items = append(l.items, r) // append item
}

func (l *rsetList) remove(r *Rset) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		if l.items[n] == r {
			l.items = append(l.items[:n], l.items[n+1:]...)
			break
		}
	}
}

func (l *rsetList) closeAll(errs *list.List) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for n := 0; n < len(l.items); n++ {
		err := l.items[n].close() // close will not remove Rset from openRsets
		if err != nil {
			errs.PushBack(errE(err))
		}
	}
	l.items = l.items[:0] // clear all Rsets from rsetList
}

func (l *rsetList) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = l.items[:0]
}

func (l *rsetList) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.items)
}
