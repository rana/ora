// Copyright 2016 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <stdlib.h>
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

type namedPos struct {
	Ordinal int
	Name    string
}

func (np namedPos) CString() (cstring *C.OraText, length C.sb4, free func()) {
	if np.Name == "" {
		return
	}
	cstring = (*C.OraText)(unsafe.Pointer(C.CString(np.Name)))
	return cstring, C.sb4(len(np.Name)), func() { C.free(unsafe.Pointer(cstring)) }
}

type nullp struct {
	p []C.sb2
}

func (np *nullp) Pointer() *C.sb2 {
	if np.p == nil {
		np.p = (*((*[1]C.sb2)(C.malloc(2))))[:1]
	}
	return &np.p[0]
}

func (np *nullp) IsNull() bool {
	return np.p == nil || np.p[0] < 0
}

func (np *nullp) Free() {
	if np.p != nil {
		C.free(unsafe.Pointer(&np.p[0]))
		np.p = nil
	}
}

func (np *nullp) Set(isNull bool) {
	p := np.Pointer()
	*p = 0
	if isNull {
		*p = -1
	}
}

type lobLocatorp struct {
	p []*C.OCILobLocator
}

func (ll *lobLocatorp) Pointer() **C.OCILobLocator {
	if ll.p == nil {
		ll.p = (*((*[1]*C.OCILobLocator)(C.malloc(C.size_t(ll.Size())))))[:1]
	}
	return &ll.p[0]
}
func (ll *lobLocatorp) Value() *C.OCILobLocator {
	if ll.p == nil {
		return nil
	}
	return ll.p[0]
}
func (ll *lobLocatorp) Size() int {
	return int(C.sof_LobLocatorp)
}
func (ll *lobLocatorp) Free() {
	if ll.p != nil {
		C.free(unsafe.Pointer(&ll.p[0]))
		ll.p = nil
	}
}

type dateTimep struct {
	p    []*C.OCIDateTime
	zone []byte
}

func (dt *dateTimep) Pointer() **C.OCIDateTime {
	if dt.p == nil {
		dt.p = (*((*[1]*C.OCIDateTime)(C.malloc(C.size_t(dt.Size())))))[:1]
	}
	return &dt.p[0]
}
func (dt *dateTimep) Value() *C.OCIDateTime {
	if dt.p == nil {
		return nil
	}
	return dt.p[0]
}
func (dt *dateTimep) Size() int { return int(C.sof_DateTimep) }
func (dt *dateTimep) Free() {
	if dt.p != nil {
		if dt.p[0] != nil {
			C.OCIDescriptorFree(
				unsafe.Pointer(dt.p[0]),  //void     *descp,
				C.OCI_DTYPE_TIMESTAMP_TZ) //ub4      type );
			dt.p[0] = nil
		}
		C.free(unsafe.Pointer(&dt.p[0]))
		dt.p = nil
	}
}
func (dt *dateTimep) Alloc(env *Env) error {
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(env.ocienv),                      //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(dt.Pointer())), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                        //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	return nil
}
func (dt *dateTimep) Set(env *Env, value time.Time) error {
	if dt.Value() == nil {
		if err := dt.Alloc(env); err != nil {
			return err
		}
	}
	dt.zone = zoneOffset(dt.zone[:0], value)
	r := C.OCIDateTimeConstruct(
		unsafe.Pointer(env.ocienv),                //dvoid         *hndl,
		env.ocierr,                                //OCIError      *err,
		dt.Value(),                                //OCIDateTime   *datetime,
		C.sb2(value.Year()),                       //sb2           year,
		C.ub1(int32(value.Month())),               //ub1           month,
		C.ub1(value.Day()),                        //ub1           day,
		C.ub1(value.Hour()),                       //ub1           hour,
		C.ub1(value.Minute()),                     //ub1           min,
		C.ub1(value.Second()),                     //ub1           sec,
		C.ub4(value.Nanosecond()),                 //ub4           fsec,
		(*C.OraText)(unsafe.Pointer(&dt.zone[0])), //OraText       *timezone,
		C.size_t(len(dt.zone)))                    //size_t        timezone_length );
	if r == C.OCI_ERROR {
		return env.ociError()
	}
	return nil
}

func zoneOffset(buf []byte, value time.Time) []byte {
	if cap(buf) < 6 {
		n := len(buf)
		buf = append(buf, make([]byte, 6)...)[:n]
	}
	_, zoneOffsetInSeconds := value.Zone()
	if zoneOffsetInSeconds < 0 {
		buf = append(buf, '-')
		zoneOffsetInSeconds *= -1
	} else {
		buf = append(buf, '+')
	}
	hourOffset := zoneOffsetInSeconds / 3600
	zoneOffsetInSeconds -= hourOffset * 3600
	minuteOffset := zoneOffsetInSeconds / 60
	buf = printTwoDigits(buf, hourOffset)
	buf = append(buf, ':')
	buf = printTwoDigits(buf, minuteOffset)
	return buf
}

func printTwoDigits(buf []byte, num int) []byte {
	if num == 0 {
		return append(buf, '0', '0')
	}
	if num < 0 {
		num *= -1
	}
	if num < 10 {
		return append(buf, '0', byte('0'+num))
	}
	return append(buf, byte('0'+num/10), byte('0'+(num%10)))
}

type intervalp struct {
	p **C.OCIInterval
}

func (ip *intervalp) Pointer() **C.OCIInterval {
	if ip.p == nil {
		ip.p = (**C.OCIInterval)(C.malloc(C.size_t(ip.Size())))
	}
	return ip.p
}
func (ip *intervalp) Value() *C.OCIInterval {
	if ip.p == nil {
		return nil
	}
	return *ip.p
}
func (ip intervalp) Size() int { return int(C.sof_Intervalp) }
func (ip *intervalp) Free() {
	if ip.p != nil {
		C.free(unsafe.Pointer(ip.p))
		ip.p = nil
	}
}

func intSixtyFour(v interface{}) int64 {
	switch i := v.(type) {
	case int64:
		return i
	case int8:
		return int64(i)
	case int16:
		return int64(i)
	case int32:
		return int64(i)
	case uint8:
		return int64(i)
	case uint16:
		return int64(i)
	case uint32:
		return int64(i)
	case uint64:
		return int64(i)
	}
	panic(fmt.Sprintf("want int, got %T", v))
}
func floatSixtyFour(v interface{}) float64 {
	switch f := v.(type) {
	case float64:
		return f
	case float32:
		return float64(f)
	}
	panic(fmt.Sprintf("wanted float, got %T", v))
}
