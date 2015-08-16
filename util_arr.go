package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"

type arrHlp struct {
	curlen     C.ub4
	nullInds   []C.sb2
	alen       []C.ACTUAL_LENGTH_TYPE
	rcode      []C.ub2
	isAssocArr bool
}

// ensureBindArrLength calculates the needed length and capacity,
// and sets up the helper arrays for binding PL/SQL Table.
//
// Returns whether and element is needed to be appended to the value slice.
func (a *arrHlp) ensureBindArrLength(
	length, capacity *int,
	stmtType C.ub4,
) (iterations uint32, curlenp *C.ub4, needsAppend bool) {
	a.curlen = C.ub4(*length) // the real length, not L!
	if stmtType == C.OCI_STMT_BEGIN || stmtType == C.OCI_STMT_DECLARE {
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
	if cap(a.nullInds) < C {
		a.nullInds = make([]C.sb2, L, C)
	} else {
		a.nullInds = (a.nullInds)[:L]
	}
	if cap(a.alen) < C {
		a.alen = make([]C.ACTUAL_LENGTH_TYPE, L, C)
	} else {
		a.alen = a.alen[:L]
	}
	if cap(a.rcode) < C {
		a.rcode = make([]C.ub2, L, C)
	} else {
		a.rcode = a.rcode[:L]
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
	if a.isAssocArr {
		return nil
	}
	a.nullInds = nil
	a.alen = nil
	a.rcode = nil
	return nil
}
