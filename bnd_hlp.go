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
import "unsafe"

type nullp struct {
	p *C.sb2
}

func (np *nullp) Pointer() *C.sb2 {
	if np.p == nil {
		np.p = (*C.sb2)(C.malloc(C.sizeof_sb2))
	}
	return np.p
}

func (np *nullp) IsNull() bool {
	if np.p == nil {
		return true
	}
	return *(np.p) < 0
}

func (np *nullp) Free() {
	if np.p != nil {
		C.free(unsafe.Pointer(np.p))
		np.p = nil
	}
}
func (np *nullp) Set(isNull bool) {
	x := C.sb2(0)
	if isNull {
		x = -1
	}
	*(np.Pointer()) = x
}

type lobLocatorp struct {
	p **C.OCILobLocator
}

func (ll *lobLocatorp) Pointer() **C.OCILobLocator {
	if ll.p == nil {
		ll.p = (**C.OCILobLocator)(C.malloc(C.size_t(ll.Size())))
	}
	return ll.p
}
func (ll *lobLocatorp) Value() *C.OCILobLocator {
	if ll.p == nil {
		return nil
	}
	return *ll.p
}
func (ll *lobLocatorp) Size() int {
	return C.sizeof_dvoid
}
func (ll *lobLocatorp) Free() {
	if ll.p != nil {
		C.free(unsafe.Pointer(ll.p))
		ll.p = nil
	}
}

type dateTimep struct {
	p **C.OCIDateTime
}

func (dt *dateTimep) Pointer() **C.OCIDateTime {
	if dt.p == nil {
		dt.p = (**C.OCIDateTime)(C.malloc(C.size_t(dt.Size())))
	}
	return dt.p
}
func (dt *dateTimep) Value() *C.OCIDateTime {
	if dt.p == nil {
		return nil
	}
	return *dt.p
}
func (dt *dateTimep) Size() int {
	return C.sizeof_dvoid
}
func (dt *dateTimep) Free() {
	if dt.p != nil {
		C.free(unsafe.Pointer(dt.p))
		dt.p = nil
	}
}

type numberp struct {
	p *C.OCINumber
}

func (np numberp) Pointer() *C.OCINumber {
	if np.p == nil {
		np.p = (*C.OCINumber)(C.malloc(C.sizeof_OCINumber))
	}
	return np.p
}
func (np numberp) Value() C.OCINumber {
	return *np.p
}
func (np numberp) Size() int {
	return C.sizeof_OCINumber
}
func (np *numberp) Free() {
	if np.p != nil {
		C.free(unsafe.Pointer(np.p))
		np.p = nil
	}
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
func (ip intervalp) Size() int { return C.sizeof_dvoid }
func (ip *intervalp) Free() {
	if ip.p != nil {
		C.free(unsafe.Pointer(ip.p))
		ip.p = nil
	}
}
