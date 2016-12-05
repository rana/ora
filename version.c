#include <oci.h>
#include "version.h"

sword
bindByNameOrPos(
	OCIStmt       *stmtp,
	OCIBind       **bindpp,
	OCIError      *errhp,
	ub4           position,
	const OraText *placeholder,
	sb4           placeholder_length,
	void          *valuep,
	LENGTH_TYPE   value_sz,
	ub2           dty,
	void          *indp,
	ACTUAL_LENGTH_TYPE *alenp,
	ub2           *rcodep,
	ub4           maxarr_len,
	ub4           *curelep, 
	ub4           mode
) {
	if( placeholder != NULL && placeholder_length > 0 ) {
		return OCIBINDBYNAME(
			stmtp,
			bindpp,
			errhp,
			placeholder,
			placeholder_length,
			valuep,
			value_sz,
			dty,
			indp,
			alenp,
			rcodep,
			maxarr_len,
			curelep,
			mode);
	}
	return OCIBINDBYPOS(
		stmtp,
		bindpp,
		errhp,
		position,
		valuep,
		value_sz,
		dty,
		indp,
		alenp,
		rcodep,
		maxarr_len,
		curelep,
		mode);
}

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
