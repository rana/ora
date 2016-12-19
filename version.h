// Copyright 2015 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

#include <oci.h>

// define simple way to respresent Oracle version
#define ORACLE_VERSION(major, minor) \
        ((major << 8) | minor)

// define what version of Oracle we are building as 2 byte hex number
#if !defined(OCI_MAJOR_VERSION) && defined(OCI_ATTR_MODULE)
#define OCI_MAJOR_VERSION 10
#define OCI_MINOR_VERSION 1
#endif

#if defined(OCI_MAJOR_VERSION) && defined(OCI_MINOR_VERSION)
#define ORACLE_VERSION_HEX \
        ORACLE_VERSION(OCI_MAJOR_VERSION, OCI_MINOR_VERSION)
#else
#error Unsupported version of OCI.
#endif

#if ORACLE_VERSION_HEX >= ORACLE_VERSION(12,1)
	#define OCIBINDBYNAME               OCIBindByName2
	#define OCIBINDBYPOS                OCIBindByPos2
	#define OCIDEFINEBYPOS              OCIDefineByPos2
	#define ACTUAL_LENGTH_TYPE          ub4
	#define ACTUAL_LENGTH_LENGTH        4
	#define MAX_BINARY_BYTES			32767
	#define LENGTH_TYPE                 sb8
	#define LENGTH_LENGTH               8

	#define ROW_COUNT_TYPE				ub8
	#define ROW_COUNT_LENGTH			8
#else
	#define OCIBINDBYNAME               OCIBindByName
	#define OCIBINDBYPOS                OCIBindByPos
	#define OCIDEFINEBYPOS              OCIDefineByPos
	#define ACTUAL_LENGTH_TYPE          ub2
	#define ACTUAL_LENGTH_LENGTH        2
	#define MAX_BINARY_BYTES            4000
	#define LENGTH_TYPE                 sb4
	#define LENGTH_LENGTH               4

	#define OCI_ATTR_UB8_ROW_COUNT		OCI_ATTR_ROW_COUNT
	#define ROW_COUNT_TYPE				ub4
	#define ROW_COUNT_LENGTH			4
#endif

#if ORACLE_VERSION_HEX >= ORACLE_VERSION(10,1)
	#define LOB_LENGTH_TYPE             oraub8
	#define OCILOBGETLENGTH             OCILobGetLength2
	#define OCILOBTRIM                  OCILobTrim2
	#define OCILOBWRITE                 OCILobWrite2
#else
	#define LOB_LENGTH_TYPE             ub4
	#define OCILOBGETLENGTH             OCILobGetLength
	#define OCILOBTRIM                  OCILobTrim
	#define OCILOBWRITE                 OCILobWrite
#endif

#define sof_DateTimep sizeof(OCIDateTime*)
#define sof_Intervalp sizeof(OCIInterval*)
#define sof_LobLocatorp sizeof(OCILobLocator*)
#define sof_Stmtp sizeof(OCIStmt*)

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
);

sword
numberFromIntSlice(
	OCIError *err,
	void *inum,
	uword inum_length,
	uword inum_s_flag,
	OCINumber *numbers,
	ub4 arr_length
);

sword
numberFromFloatSlice(
	OCIError *err,
	void *inum,
	uword inum_length,
	OCINumber *numbers,
	ub4 arr_length
);

sword
decriptorAllocSlice(
	OCIEnv *env,
	void *dest,
	ub4 elem_size,
	ub4 type,
	size_t length
);
