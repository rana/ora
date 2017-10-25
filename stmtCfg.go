// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// StmtCfg affects various aspects of a SQL statement.
//
// Assign values to StmtCfg prior to calling Stmt.Exe
// and Stmt.Qry for the configuration values to take effect.
//
// StmtCfg is immutable, so every Set method returns a new
// instance, maybe with Err set, too.
type StmtCfg struct {
	prefetchRowCount      uint32
	prefetchMemorySize    uint32
	longBufferSize        uint32
	longRawBufferSize     uint32
	lobBufferSize         int
	stringPtrBufferSize   int
	fetchLen, lobFetchLen int
	byteSlice             GoColumnType

	// IsAutoCommitting determines whether DML statements are automatically
	// committed.
	//
	// The default is true.
	//
	// IsAutoCommitting is not observed during a transaction.
	IsAutoCommitting bool

	// RTrimChar makes returning from CHAR colums trim the blanks (spaces)
	// from the end of the string, added by Oracle.
	//
	// The default is true.
	RTrimChar bool

	// FalseRune represents the false Go bool value sent to an Oracle server
	// during a parameter bind.
	//
	// The is default is '0'.
	FalseRune rune

	// TrueRune represents the true Go bool value sent to an Oracle server
	// during a parameter bind.
	//
	// The is default is '1'.
	TrueRune rune

	// Rset represents configuration options for an Rset struct.
	RsetCfg

	// Err is the error from the last Set... call, if there's any.
	Err error
}

// NewStmtCfg returns a StmtCfg with default values.
func NewStmtCfg() StmtCfg {
	var c StmtCfg
	c.fetchLen = DefaultFetchLen
	c.lobFetchLen = DefaultLOBFetchLen
	c.prefetchRowCount = 128
	c.prefetchMemorySize = 128 << 20 // 134,217,728
	c.longBufferSize = 16 << 20      // 16,777,216
	c.longRawBufferSize = 16 << 20   // 16,777,216
	c.lobBufferSize = 16 << 20       // 16,777,216
	c.stringPtrBufferSize = 4000

	c.IsAutoCommitting = true
	c.RTrimChar = true
	c.FalseRune = '0'
	c.TrueRune = '1'
	c.RsetCfg = NewRsetCfg()
	return c
}

func (c StmtCfg) IsZero() bool { return c.prefetchRowCount == 0 && c.prefetchMemorySize == 0 }

// SetPrefetchRowCount sets the number of rows to prefetch during a select query.
func (c StmtCfg) SetPrefetchRowCount(prefetchRowCount uint32) StmtCfg {
	c.prefetchRowCount = prefetchRowCount
	return c
}

// PrefetchRowCount returns the number of rows to prefetch during a select query.
//
// The default is 0.
//
// PrefetchRowCount works in coordination with PrefetchMemorySize. When
// PrefetchRowCount is set to zero only PrefetchMemorySize is used;
// otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
func (c StmtCfg) PrefetchRowCount() uint32 {
	return c.prefetchRowCount
}

// SetPrefetchMemorySize sets the prefetch memory size in bytes used during a SQL
// select command.
func (c StmtCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) StmtCfg {
	c.prefetchMemorySize = prefetchMemorySize
	return c
}

// PrefetchMemorySize returns the prefetch memory size in bytes used during a SQL
// select command.
//
// The default is 134,217,728 bytes.
//
// PrefetchMemorySize works in coordination with PrefetchRowCount. When
// PrefetchRowCount is set to zero only PrefetchMemorySize is used;
// otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
func (c StmtCfg) PrefetchMemorySize() uint32 {
	return c.prefetchMemorySize
}

// SetLongBufferSize sets the long buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is less than 1 or greater than 2,147,483,642.
func (c StmtCfg) SetLongBufferSize(size uint32) StmtCfg {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		if c.Err == nil {
			c.Err = errNew("long buffer size too large")
		}
		return c
	}
	if size < 1 {
		if c.Err == nil {
			c.Err = errNew("SetLongBufferSize parameter 'size' must be greater than zero")
		}
		return c
	}
	c.longBufferSize = size
	return c
}

// LongBufferSize returns the long buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LONG type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c StmtCfg) LongBufferSize() uint32 {
	return c.longBufferSize
}

// SetLongRawBufferSize sets the LONG RAW buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is greater than 2,147,483,642.
func (c StmtCfg) SetLongRawBufferSize(size uint32) StmtCfg {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		if c.Err == nil {
			c.Err = errNew("long raw buffer size too large")
		}
		return c
	}
	c.longRawBufferSize = size
	return c
}

// LongRawBufferSize returns the LONG RAW buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LONG RAW type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c StmtCfg) LongRawBufferSize() uint32 {
	return c.longRawBufferSize
}

// SetLobBufferSize sets the LOB buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is greater than 2,147,483,642.
func (c StmtCfg) SetLobBufferSize(size int) StmtCfg {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		if c.Err == nil {
			c.Err = errNew("lob buffer size too large")
		}
		return c
	}
	c.lobBufferSize = size
	return c
}

// LobBufferSize returns the LOB buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LOB type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c StmtCfg) LobBufferSize() int {
	return c.lobBufferSize
}

// SetStringPtrBufferSize sets the size of a buffer used to store a string during
// *string parameter binding and []*string parameter binding in a SQL statement.
func (c StmtCfg) SetStringPtrBufferSize(size int) StmtCfg {
	if size < 1 {
		if c.Err == nil {
			c.Err = errNew("SetStringPtrBufferSize parameter 'size' must be greater than zero")
		}
		return c
	}
	c.stringPtrBufferSize = size
	return c
}

// StringPtrBufferSize returns the size of a buffer in bytes used to store a string
// during *string parameter binding and []*string parameter binding in a SQL statement.
//
// The default is 4000 bytes.
//
// For a *string parameter binding, you may wish to increase the size of
// StringPtrBufferSize depending on the Oracle column type. For VARCHAR2,
// NVARCHAR2, and RAW oracle columns the Oracle MAX_STRING_SIZE is usually 4000
// but may be set up to 32767.
func (c StmtCfg) StringPtrBufferSize() int {
	return c.stringPtrBufferSize
}

// SetByteSlice sets a GoColumnType associated to SQL statement []byte parameter.
//
// Valid values are U8 and Bits.
//
// Returns an error if U8 or Bits is not specified.
func (c StmtCfg) SetByteSlice(gct GoColumnType) StmtCfg {
	if err := checkBinOrU8Column(gct); err != nil {
		if c.Err == nil {
			c.Err = err
		}
		return c
	}
	c.byteSlice = gct
	return c
}

// ByteSlice returns a GoColumnType associated to SQL statement []byte parameter.
//
// The default is Bits.
//
// ByteSlice is used by the database/sql package.
//
// Sending a byte slice to an Oracle server as a parameter in a SQL statement
// requires knowing the destination column type ahead of time. Set ByteSlice to
// Bits if the destination column is BLOB, RAW or LONG RAW. Set ByteSlice to U8
// if the destination column is NUMBER, BINARY_DOUBLE, BINARY_FLOAT or FLOAT.
func (c StmtCfg) ByteSlice() GoColumnType {
	return c.byteSlice
}

// returns a value of the lobFetchLen
func (c StmtCfg) LOBFetchLen() int {
	return c.lobFetchLen
}

// returns a value of the fetchLen
func (c StmtCfg) FetchLen() int {
	return c.fetchLen
}

// SetFetchLen overrides DefaultFetchLen for prefetch lengths.
func (c StmtCfg) SetFetchLen(length int) StmtCfg {
	if length <= 0 {
		return c
	}
	if length >= MaxFetchLen {
		length = MaxFetchLen
	}
	c.fetchLen = length
	return c
}

// SetLOBFetchLen overrides DefaultLOBFetchLen for prefetch LOB lengths.
//
// This affects result sets with any of the following column types:
// C.SQLT_LNG, C.SQLT_BFILE, C.SQLT_BLOB, C.SQLT_CLOB, C.SQLT_LBI
//
// Caution: the default buffer size for blob is 1MB. So, for example a single
// fetch from the result set that contains just one blob will consume 128MB of RAM
func (c StmtCfg) SetLOBFetchLen(length int) StmtCfg {
	if length <= 0 {
		return c
	}
	if length >= MaxFetchLen {
		length = MaxFetchLen
	}
	c.lobFetchLen = length
	return c
}

func (c StmtCfg) SetNumberInt(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetNumberInt(gct)
	return c
}
func (c StmtCfg) SetNumberBigInt(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetNumberBigInt(gct)
	return c
}
func (c StmtCfg) SetNumberFloat(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetNumberFloat(gct)
	return c
}
func (c StmtCfg) SetNumberBigFloat(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetNumberBigFloat(gct)
	return c
}
func (c StmtCfg) SetBinaryDouble(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetBinaryDouble(gct)
	return c
}
func (c StmtCfg) SetBinaryFloat(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetBinaryFloat(gct)
	return c
}
func (c StmtCfg) SetFloat(gct GoColumnType) StmtCfg { c.RsetCfg = c.RsetCfg.SetFloat(gct); return c }
func (c StmtCfg) SetDate(gct GoColumnType) StmtCfg  { c.RsetCfg = c.RsetCfg.SetDate(gct); return c }
func (c StmtCfg) SetTimestamp(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetTimestamp(gct)
	return c
}
func (c StmtCfg) SetTimestampTz(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetTimestampTz(gct)
	return c
}
func (c StmtCfg) SetTimestampLtz(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetTimestampLtz(gct)
	return c
}
func (c StmtCfg) SetChar1(gct GoColumnType) StmtCfg   { c.RsetCfg = c.RsetCfg.SetChar1(gct); return c }
func (c StmtCfg) SetChar(gct GoColumnType) StmtCfg    { c.RsetCfg = c.RsetCfg.SetChar(gct); return c }
func (c StmtCfg) SetVarchar(gct GoColumnType) StmtCfg { c.RsetCfg = c.RsetCfg.SetVarchar(gct); return c }
func (c StmtCfg) SetLong(gct GoColumnType) StmtCfg    { c.RsetCfg = c.RsetCfg.SetLong(gct); return c }
func (c StmtCfg) SetClob(gct GoColumnType) StmtCfg    { c.RsetCfg = c.RsetCfg.SetClob(gct); return c }
func (c StmtCfg) SetBlob(gct GoColumnType) StmtCfg    { c.RsetCfg = c.RsetCfg.SetBlob(gct); return c }
func (c StmtCfg) SetRaw(gct GoColumnType) StmtCfg     { c.RsetCfg = c.RsetCfg.SetRaw(gct); return c }
func (c StmtCfg) SetLongRaw(gct GoColumnType) StmtCfg {
	c.RsetCfg = c.RsetCfg.SetLongRaw(gct)
	return c
}
