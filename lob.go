// Copyright 2017 Tamás Gulácsi
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package ora

/*
#cgo CFLAGS: -Iodpi/src -Iodpi/include
#cgo LDFLAGS: -Lodpi/lib -lodpic -ldl

#include "dpiImpl.h"
*/
import "C"
import (
	"fmt"
	"io"
	"unsafe"

	"github.com/pkg/errors"
)

const CheckLOBWrite = true

// Lob is for reading/writing a LOB.
type Lob struct {
	io.Reader
	IsClob bool
}

// Scan assigns a value from a database driver.
//
// The src value will be of one of the following types:
//
//    int64
//    float64
//    bool
//    []byte
//    string
//    time.Time
//    nil - for NULL values
//
// An error should be returned if the value cannot be stored
// without loss of information.
func (dlr *dpiLobReader) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return errors.Errorf("cannot convert LOB to %T", src)
	}
	_ = b
	return nil
}

var _ = io.Reader((*dpiLobReader)(nil))

type dpiLobReader struct {
	*conn
	dpiLob *C.dpiLob
	offset C.uint64_t
	size   C.uint64_t
}

func (dlr *dpiLobReader) Read(p []byte) (int, error) {
	if dlr == nil {
		return 0, errors.New("read on nil dpiLobReader")
	}
	if dlr.size == 0 {
		if C.dpiLob_getSize(dlr.dpiLob, &dlr.size) == C.DPI_FAILURE {
			return 0, errors.Wrap(dlr.getError(), "getSize")
		}
	}
	//fmt.Printf("offset=%d size=%d\n", dlr.offset, dlr.size)
	if dlr.size <= dlr.offset {
		return 0, io.EOF
	}
	n := C.uint64_t(len(p))
	var eof bool
	if eof := dlr.size < dlr.offset+n; eof {
		n = dlr.size - dlr.offset
	}
	fmt.Printf("%p.Read offset=%d n=%d size=%d\n", dlr.dpiLob, dlr.offset, n, dlr.size)
	if C.dpiLob_readBytes(dlr.dpiLob, dlr.offset+1, n, (*C.char)(unsafe.Pointer(&p[0])), &n) == C.DPI_FAILURE {
		err := dlr.getError()
		fmt.Printf("read %d, got ORA-%d\n", n, err.Code())
		if n > 0 && err.Code() == 1403 {
			dlr.offset += n
			return int(n), io.EOF
		}
		return int(n), errors.Wrapf(err, "lob=%p offset=%d n=%d size=%d", dlr.dpiLob, dlr.offset, len(p), dlr.size)
	}
	fmt.Printf("read %d\n", n)
	dlr.offset += n
	var err error
	if eof {
		err = io.EOF
	}
	return int(n), err
}

type dpiLobWriter struct {
	*conn
	dpiLob *C.dpiLob
	offset C.uint64_t
	opened bool
}

func (dlw *dpiLobWriter) Write(p []byte) (int, error) {
	if !dlw.opened {
		if C.dpiLob_openResource(dlw.dpiLob) == C.DPI_FAILURE {
			return 0, errors.Wrapf(dlw.getError(), "openResources(%p)", dlw.dpiLob)
		}
		dlw.opened = true
	}

	n := C.uint64_t(len(p))
	if C.dpiLob_writeBytes(dlw.dpiLob, dlw.offset+1, (*C.char)(unsafe.Pointer(&p[0])), n) == C.DPI_FAILURE {
		lob := dlw.dpiLob
		dlw.dpiLob = nil
		C.dpiLob_closeResource(lob)
		return 0, errors.Wrapf(dlw.getError(), "writeBytes(%p, offset=%d, data=%d)", lob, dlw.offset, n)
	}
	fmt.Printf("written %q into %p@%d\n", p, dlw.dpiLob, dlw.offset)
	dlw.offset += n

	if CheckLOBWrite {
		var size C.uint64_t
		if C.dpiLob_getSize(dlw.dpiLob, &size); size != dlw.offset {
			return int(n), errors.Errorf("%p size=%d, offset=%d", dlw.dpiLob, size, dlw.offset)
		}
	}
	return int(n), nil
}

func (dlw *dpiLobWriter) Close() error {
	if dlw == nil || dlw.dpiLob == nil {
		return nil
	}
	lob := dlw.dpiLob
	if C.dpiLob_flushBuffer(lob) == C.DPI_FAILURE {
		return errors.Wrapf(dlw.getError(), "flushBuffer(%p)", lob)
	}
	dlw.dpiLob = nil
	if C.dpiLob_closeResource(lob) == C.DPI_FAILURE {
		return errors.Wrapf(dlw.getError(), "closeResource(%p)", lob)
	}
	return nil
}
