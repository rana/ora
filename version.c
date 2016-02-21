#include <oci.h>

sword
numberFromIntSlice(
	OCIError *err,
	void *inum,
	uword inum_length,
	uword inum_s_flag,
	OCINumber *numbers,
	ub4 arr_length
) {
	sword rc;
	int i;
	for(i=0; i < arr_length; i++) {
		rc = OCINumberFromInt(err, inum + (i * inum_length), inum_length, inum_s_flag, &(numbers[i]));
		if(rc == OCI_ERROR) {
			return rc;
	    }
	}
	return OCI_SUCCESS;
}

sword
numberFromFloatSlice(
	OCIError *err,
	void *inum,
	uword inum_length,
	OCINumber *numbers,
	ub4 arr_length
) {
	sword rc;
	int i;
	for(i=0; i < arr_length; i++) {
		rc = OCINumberFromReal(err, inum + (i * inum_length), inum_length, &(numbers[i]));
		if(rc == OCI_ERROR) {
			return rc;
	    }
	}
	return OCI_SUCCESS;
}


sword
decriptorAllocSlice(
	OCIEnv *env,
	void *dest,
	ub4 elem_size,
	ub4 type,
	size_t length
) {
	sword rc;
	int i;
	for(i=0; i < length; i++) {
		rc = OCIDescriptorAlloc(
			env,             //CONST dvoid   *parenth,
			dest + i * elem_size, //dvoid         **descpp,
			type,                                 //ub4           type,
			0,   //size_t        xtramem_sz,
			0); //dvoid         **usrmempp);
		if(rc == OCI_ERROR) {
			return rc;
	    }
	}
	return OCI_SUCCESS;
}
