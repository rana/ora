package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"sync"
	"unsafe"
)

type arrHlp struct {
	sync.Mutex
	curlen     C.ub4
	nullInds   []C.sb2
	alen       []C.ACTUAL_LENGTH_TYPE
	rcode      []C.ub2
	allocated  []bool
	isAssocArr bool
}

type ociDef struct {
	ocidef *C.OCIDefine
	rset   *Rset
	arrHlp
}

func (d *ociDef) defineByPos(position int, valuep unsafe.Pointer, valueSize int, dty int) error {
	d.ensureFetchLength(d.rset.stmt.Cfg().FetchLen())
	// If you omit the rlenp parameter of OCIDefineByPos(), returned values are blank-padded to the buffer length, and NULLs are returned as a string of blank characters. If rlenp is included, returned values are not blank-padded. Instead, their actual lengths are returned in the rlenp parameter.
	if r := C.OCIDEFINEBYPOS(
		d.rset.ocistmt,    //OCIStmt     *stmtp,
		&d.ocidef,         //OCIDefine   **defnpp,
		d.rset.env.ocierr, //OCIError    *errhp,
		C.ub4(position),   //ub4         position,
		valuep,            //void        *valuep,
		C.LENGTH_TYPE(valueSize),       //sb8         value_sz,
		C.ub2(dty),                     //ub2         dty,
		unsafe.Pointer(&d.nullInds[0]), //void        *indp,
		&d.alen[0],                     //ub4         *rlenp,
		&d.rcode[0],                    //ub2         *rcodep,
		C.OCI_DEFAULT,                  //ub4         mode );
	); r == C.OCI_ERROR {
		return d.rset.stmt.ses.srv.env.ociError()
	}
	if r := C.OCIDefineArrayOfStruct(
		d.ocidef,          //OCIDefine *defnp
		d.rset.env.ocierr, //OCIError *errhp,
		C.ub4(valueSize),  //ub4 pvskip,
		2,                 //ub4 indskip,
		C.ACTUAL_LENGTH_LENGTH, //ub4 rlskip,
		2, //ub4 rcskip
	); r == C.OCI_ERROR {
		return d.rset.env.ociError()
	}
	return nil
}

var (
	sb2Pool  = sync.Pool{New: func() interface{} { z := []C.sb2{}; return &z }}
	ub2Pool  = sync.Pool{New: func() interface{} { z := []C.ub2{}; return &z }}
	alenPool = sync.Pool{New: func() interface{} { z := []C.ACTUAL_LENGTH_TYPE{}; return &z }}
)

func (d *arrHlp) ensureAllocatedLength(length int) {
	if cap(d.allocated) < length {
		d.allocated = make([]bool, length)
		return
	}
	d.allocated = d.allocated[:length]
	for i := range d.allocated {
		d.allocated[i] = false
	}
}

func (a *arrHlp) ensureFetchLength(length int) {
	if length <= 0 || length >= MaxFetchLen {
		length = MaxFetchLen
	}
	if cap(a.nullInds) >= length {
		a.nullInds = a.nullInds[:length]
	} else {
		if a.nullInds = *(sb2Pool.Get().(*[]C.sb2)); cap(a.nullInds) < length {
			a.nullInds = make([]C.sb2, length)
		} else {
			a.nullInds = a.nullInds[:length]
		}
	}
	if cap(a.alen) >= length {
		a.alen = a.alen[:length]
	} else {
		if a.alen = *(alenPool.Get().(*[]C.ACTUAL_LENGTH_TYPE)); cap(a.alen) < length {
			a.alen = make([]C.ACTUAL_LENGTH_TYPE, length)
		} else {
			a.alen = a.alen[:length]
		}
	}
	if cap(a.rcode) >= length {
		a.rcode = a.rcode[:length]
	} else {
		if a.rcode = *(ub2Pool.Get().(*[]C.ub2)); cap(a.rcode) < length {
			a.rcode = make([]C.ub2, length)
		} else {
			a.rcode = a.rcode[:length]
		}
	}
}

// ensureBindArrLength calculates the needed length and capacity,
// and sets up the helper arrays for binding PL/SQL Table.
//
// Returns whether and element is needed to be appended to the value slice.
func (a *arrHlp) ensureBindArrLength(
	length, capacity *int,
	isAssocArray bool,
) (iterations uint32, curlenp *C.ub4, needsAppend bool) {
	a.curlen = C.ub4(*length) // the real length, not L!
	if isAssocArray {
		// for PL/SQL associative arrays
		curlenp = &a.curlen
		iterations = 1
		a.isAssocArr = true
		if *length == 0 {
			*length = 1
			if *capacity == 0 {
				needsAppend = true
				*capacity = 1
			}
		}
	} else {
		curlenp = nil
		iterations = uint32(*length)
		a.isAssocArr = false
	}
	L, C := *length, *capacity
	if cap(a.nullInds) >= C {
		a.nullInds = (a.nullInds)[:L]
	} else {
		if a.nullInds = *(sb2Pool.Get().(*[]C.sb2)); cap(a.nullInds) < C {
			a.nullInds = make([]C.sb2, L, C)
		} else {
			a.nullInds = (a.nullInds)[:L]
		}
	}
	if cap(a.alen) >= C {
		a.alen = a.alen[:L]
	} else {
		if a.alen = *(alenPool.Get().(*[]C.ACTUAL_LENGTH_TYPE)); cap(a.alen) < C {
			a.alen = make([]C.ACTUAL_LENGTH_TYPE, L, C)
		} else {
			a.alen = a.alen[:L]
		}
	}
	if cap(a.rcode) >= C {
		a.rcode = a.rcode[:L]
	} else {
		if a.rcode = *(ub2Pool.Get().(*[]C.ub2)); cap(a.rcode) < C {
			a.rcode = make([]C.ub2, L, C)
		} else {
			a.rcode = a.rcode[:L]
		}
	}
	return iterations, curlenp, needsAppend
}

// IsAssocArr returns true if the bind uses PL/SQL Table.
func (a arrHlp) IsAssocArr() bool {
	return a.isAssocArr
}

// close nils the slices, except when this is a PL/SQL Table.
//
// The reason for this is that for PL/SQL Tables, after exe returns,
// the bound slices can be reused; otherwise, they are still in use for
// the subsequent iterations!
func (a *arrHlp) close() error {
	if a == nil {
		return nil
	}
	if a.isAssocArr {
		return nil
	}
	if a.nullInds != nil {
		sb2Pool.Put(&a.nullInds)
		a.nullInds = nil
	}
	if a.alen != nil {
		alenPool.Put(&a.alen)
		a.alen = nil
	}
	if a.rcode != nil {
		ub2Pool.Put(&a.rcode)
		a.rcode = nil
	}
	if a.allocated != nil {
		a.allocated = nil
	}
	return nil
}

func getMaxarrLen(C int, isAssocArray bool) C.ub4 {
	if !isAssocArray {
		return 0
	}
	return C.ub4(C)
}
