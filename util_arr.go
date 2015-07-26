package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"

type arrHlp struct {
	curlen   C.ACTUAL_LENGTH_TYPE
	nullInds []C.sb2
	alen     []C.ACTUAL_LENGTH_TYPE
	rcode    []C.ub2
}

// ensureBindArrLength calculates the needed length and capacity,
// and sets up the helper arrays for binding PL/SQL Table.
//
// Returns whether and element is needed to be appended to the value slice.
func (a *arrHlp) ensureBindArrLength(
	length, capacity *int,
) (needsAppend bool) {
	if *length == 0 {
		*length = 1
		if *capacity == 0 {
			needsAppend = true
			*capacity = 1
		}
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
	return
}
