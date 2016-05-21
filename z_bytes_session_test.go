//Copyright 2014 Rana Ian. All rights reserved.
//Use of this source code is governed by The MIT License
//found in the accompanying LICENSE file.

package ora_test

import (
	"testing"

	"gopkg.in/rana/ora.v3"
	"gopkg.in/rana/ora.v3/lg"
)

//// bytes
//longRaw     oracleColumnType = "long raw not null"
//longRawNull oracleColumnType = "long raw null"
//raw2000     oracleColumnType = "raw(2000) not null"
//raw2000Null oracleColumnType = "raw(2000) null"
//blob        oracleColumnType = "blob not null"
//blobNull    oracleColumnType = "blob null"

//////////////////////////////////////////////////////////////////////////////////
//// longRaw
//////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_longRaw_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_bytes(9), longRaw, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_longRaw_session(t *testing.T) {
	testBindDefine(gen_OraBytes(9, false), longRaw, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_longRaw_session(t *testing.T) {
	testBindDefine(gen_bytesSlice(9), longRaw, t, nil)
}

func TestBindSlice_OraBytes_longRaw_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(9, false), longRaw, t, nil)
}

func TestMultiDefine_longRaw_session(t *testing.T) {
	//enableLogging(t)
	testMultiDefine(gen_bytes(9), longRaw, t)
}

//// Do not test workload of multiple Oracle LONG RAW types within the same table because
//// ORA-01754: a table may contain only one column of type LONG
//func TestWorkload_longRaw_session(t *testing.T) {
//	testWorkload(testWorkloadColumnCount, t, longRaw)
//}

////////////////////////////////////////////////////////////////////////////////
// longRawNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_longRawNull_session(t *testing.T) {
	testBindDefine(gen_bytes(9), longRawNull, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_longRawNull_session(t *testing.T) {
	testBindDefine(gen_OraBytes(9, true), longRawNull, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_longRawNull_session(t *testing.T) {
	testBindDefine(gen_bytesSlice(9), longRawNull, t, nil)
}

func TestBindSlice_OraBytes_longRawNull_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(9, true), longRawNull, t, nil)
}

func TestMultiDefine_longRawNull_session(t *testing.T) {
	testMultiDefine(gen_bytes(9), longRawNull, t)
}

//// Do not test workload of multiple Oracle LONG RAW types within the same table because
//// ORA-01754: a table may contain only one column of type LONG
//func TestWorkload_longRawNull_session(t *testing.T) {
//	testWorkload(testWorkloadColumnCount, t, longRawNull)
//}

func TestBindDefine_longRawNull_nil_session(t *testing.T) {
	testBindDefine(nil, longRawNull, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// raw2000
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_raw2000_session(t *testing.T) {
	testBindDefine(gen_bytes(2000), raw2000, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_raw2000_session(t *testing.T) {
	testBindDefine(gen_OraBytes(2000, false), raw2000, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_raw2000_session(t *testing.T) {
	testBindDefine(gen_bytesSlice(2000), raw2000, t, nil)
}

func TestBindSlice_OraBytes_raw2000_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(2000, false), raw2000, t, nil)
}

func TestMultiDefine_raw2000_session(t *testing.T) {
	testMultiDefine(gen_bytes(2000), raw2000, t)
}

func TestWorkload_raw2000_session(t *testing.T) {
	testWorkload(raw2000, t)
}

////////////////////////////////////////////////////////////////////////////////
// raw2000Null
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_raw2000Null_session(t *testing.T) {
	testBindDefine(gen_bytes(2000), raw2000Null, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_raw2000Null_session(t *testing.T) {
	testBindDefine(gen_OraBytes(2000, true), raw2000Null, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_raw2000Null_session(t *testing.T) {
	testBindDefine(gen_bytesSlice(2000), raw2000Null, t, nil)
}

func TestBindSlice_OraBytes_raw2000Null_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(2000, true), raw2000Null, t, nil)
}

func TestMultiDefine_raw2000Null_session(t *testing.T) {
	testMultiDefine(gen_bytes(2000), raw2000Null, t)
}

func TestWorkload_raw2000Null_session(t *testing.T) {
	testWorkload(raw2000Null, t)
}

func TestBindDefine_raw2000Null_nil_session(t *testing.T) {
	testBindDefine(nil, raw2000Null, t, nil)
}

////////////////////////////////////////////////////////////////////////////////
// blob
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_blob_session(t *testing.T) {
	testBindDefine(gen_bytes(9), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_blob_session(t *testing.T) {
	testBindDefine(gen_OraBytes(9, false), blob, t, nil, ora.OraBin)
}

func TestBindDefine_Lob_blob_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_OraBytesLob(9, false), blob, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_blob_session(t *testing.T) {
	ora.Cfg().Log.Rset.BeginRow = true
	ora.Cfg().Log.Rset.EndRow = true
	ora.Cfg().Log.Rset.Next = true
	//enableLogging(t)
	ora.Cfg().Log.Logger = lg.Log
	testBindDefine(gen_bytesSlice(9), blob, t, nil)
}

func TestBindSlice_OraBytes_blob_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(9, false), blob, t, nil)
}

func TestMultiDefine_blob_session(t *testing.T) {
	//enableLogging(t)
	testMultiDefine(gen_bytes(9), blob, t)
}

func TestWorkload_blob_session(t *testing.T) {
	testWorkload(blob, t)
}

func TestBindDefine_bytes_blob_bufferSize_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes(sc.LobBufferSize()), blob, t, nil, ora.Bin)
}

func TestBindDefine_bytes_blob_bufferSizeMinusOne_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes(sc.LobBufferSize()-1), blob, t, nil, ora.Bin)
}

func TestBindDefine_bytes_blob_bufferSizePlusOne_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes(sc.LobBufferSize()+1), blob, t, nil, ora.Bin)
}

func TestBindDefine_bytes_blob_bufferSizeMultiple_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes(sc.LobBufferSize()*3), blob, t, nil, ora.Bin)
}

func TestBindDefine_bytes_blob_bufferSizeMultipleMinusOne_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes((sc.LobBufferSize()*3)-1), blob, t, nil, ora.Bin)
}

func TestBindDefine_bytes_blob_bufferSizeMultiplePlusOne_session(t *testing.T) {
	sc := ora.NewStmtCfg()
	testBindDefine(gen_bytes((sc.LobBufferSize()*3)+1), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSize_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob(sc.LobBufferSize(), false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLobPtr_blob_bufferSize_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	lob := gen_OraBytesLob(sc.LobBufferSize(), false)
	testBindDefine(&lob, blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSizeMinusOne_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob(sc.LobBufferSize()-1, false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSizePlusOne_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob(sc.LobBufferSize()+1, false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSizeMultiple_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob(sc.LobBufferSize()*3, false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSizeMultipleMinusOne_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob((sc.LobBufferSize()*3)-1, false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLob_blob_bufferSizeMultiplePlusOne_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	testBindDefine(gen_OraBytesLob((sc.LobBufferSize()*3)+1, false), blob, t, nil, ora.Bin)
}

func TestBindDefine_OraBytesLobPtr_blob_bufferSizeMultiplePlusOne_session(t *testing.T) {
	//enableLogging(t)
	sc := ora.NewStmtCfg()
	lob := gen_OraBytesLob((sc.LobBufferSize()*3)+1, false)
	testBindDefine(&lob, blob, t, nil, ora.Bin)
}

////////////////////////////////////////////////////////////////////////////////
// blobNull
////////////////////////////////////////////////////////////////////////////////
func TestBindDefine_bytes_blobNull_session(t *testing.T) {
	testBindDefine(gen_bytes(9), blobNull, t, nil, ora.Bin)
}

func TestBindDefine_OraBytes_blobNull_session(t *testing.T) {
	testBindDefine(gen_OraBytes(9, true), blobNull, t, nil, ora.OraBin)
}

func TestBindDefine_OraBytesLob_blobNull_session(t *testing.T) {
	//enableLogging(t)
	testBindDefine(gen_OraBytesLob(9, true), blobNull, t, nil, ora.OraBin)
}

func TestBindDefine_OraBytesLobPtr_blobNull_session(t *testing.T) {
	//enableLogging(t)
	lob := gen_OraBytesLob(9, true)
	testBindDefine(&lob, blobNull, t, nil, ora.OraBin)
}

func TestBindSlice_bytes_blobNull_session(t *testing.T) {
	testBindDefine(gen_bytesSlice(9), blobNull, t, nil)
}

func TestBindSlice_OraBytes_blobNull_session(t *testing.T) {
	testBindDefine(gen_OraBytesSlice(9, true), blobNull, t, nil)
}

func TestMultiDefine_blobNull_session(t *testing.T) {
	testMultiDefine(gen_bytes(9), blobNull, t)
}

func TestWorkload_blobNull_session(t *testing.T) {
	testWorkload(blobNull, t)
}

func TestBindDefine_blobNull_nil_session(t *testing.T) {
	testBindDefine(nil, blobNull, t, nil)
}
