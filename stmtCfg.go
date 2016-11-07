// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

// StmtCfg affects various aspects of a SQL statement.
//
// Assign values to StmtCfg prior to calling Stmt.Exe
// and Stmt.Qry for the configuration values to take effect.
type StmtCfg struct {
	prefetchRowCount    uint32
	prefetchMemorySize  uint32
	longBufferSize      uint32
	longRawBufferSize   uint32
	lobBufferSize       int
	stringPtrBufferSize int
	byteSlice           GoColumnType

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
	Rset RsetCfg
}

// NewStmtCfg returns a StmtCfg with default values.
func NewStmtCfg() *StmtCfg {
	c := &StmtCfg{}
	if _drv.cfg.Env != nil && _drv.cfg.Env.StmtCfg != nil {
		*c = *_drv.cfg.Env.StmtCfg
		return c
	}
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
	c.Rset = NewRsetCfg()
	return c
}

// SetPrefetchRowCount sets the number of rows to prefetch during a select query.
func (c *StmtCfg) SetPrefetchRowCount(prefetchRowCount uint32) error {
	c.prefetchRowCount = prefetchRowCount
	return nil
}

// PrefetchRowCount returns the number of rows to prefetch during a select query.
//
// The default is 0.
//
// PrefetchRowCount works in coordination with PrefetchMemorySize. When
// PrefetchRowCount is set to zero only PrefetchMemorySize is used;
// otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
func (c *StmtCfg) PrefetchRowCount() uint32 {
	return c.prefetchRowCount
}

// SetPrefetchMemorySize sets the prefetch memory size in bytes used during a SQL
// select command.
func (c *StmtCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) error {
	c.prefetchMemorySize = prefetchMemorySize
	return nil
}

// PrefetchMemorySize returns the prefetch memory size in bytes used during a SQL
// select command.
//
// The default is 134,217,728 bytes.
//
// PrefetchMemorySize works in coordination with PrefetchRowCount. When
// PrefetchRowCount is set to zero only PrefetchMemorySize is used;
// otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
func (c *StmtCfg) PrefetchMemorySize() uint32 {
	return c.prefetchMemorySize
}

// SetLongBufferSize sets the long buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is less than 1 or greater than 2,147,483,642.
func (c *StmtCfg) SetLongBufferSize(size uint32) error {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		return errNew("long buffer size too large")
	}
	if size < 1 {
		return errNew("SetLongBufferSize parameter 'size' must be greater than zero")
	}
	c.longBufferSize = size
	return nil
}

// LongBufferSize returns the long buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LONG type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c *StmtCfg) LongBufferSize() uint32 {
	return c.longBufferSize
}

// SetLongRawBufferSize sets the LONG RAW buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is greater than 2,147,483,642.
func (c *StmtCfg) SetLongRawBufferSize(size uint32) error {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		return errNew("long raw buffer size too large")
	}
	c.longRawBufferSize = size
	return nil
}

// LongRawBufferSize returns the LONG RAW buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LONG RAW type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c *StmtCfg) LongRawBufferSize() uint32 {
	return c.longRawBufferSize
}

// SetLobBufferSize sets the LOB buffer size in bytes.
//
// The maximum is 2,147,483,642 bytes.
//
// Returns an error if the specified size is greater than 2,147,483,642.
func (c *StmtCfg) SetLobBufferSize(size int) error {
	// OCI-22140: given size must be in the range of 0 to [2147483643]
	// Subtact one to account for the offset made within function stringDefine.bind.
	if size > 2147483642 {
		return errNew("lob buffer size too large")
	}
	c.lobBufferSize = size
	return nil
}

// LobBufferSize returns the LOB buffer size in bytes used to define the sql select-column
// buffer size of an Oracle LOB type.
//
// The default is 16,777,216 bytes.
//
// The default is considered a moderate buffer where the 2GB max buffer may not
// be feasible on all clients.
func (c *StmtCfg) LobBufferSize() int {
	return c.lobBufferSize
}

// SetStringPtrBufferSize sets the size of a buffer used to store a string during
// *string parameter binding and []*string parameter binding in a SQL statement.
func (c *StmtCfg) SetStringPtrBufferSize(size int) error {
	if size < 1 {
		return errNew("SetStringPtrBufferSize parameter 'size' must be greater than zero")
	}
	c.stringPtrBufferSize = size
	return nil
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
func (c *StmtCfg) StringPtrBufferSize() int {
	return c.stringPtrBufferSize
}

// SetByteSlice sets a GoColumnType associated to SQL statement []byte parameter.
//
// Valid values are U8 and Bits.
//
// Returns an error if U8 or Bits is not specified.
func (c *StmtCfg) SetByteSlice(gct GoColumnType) (err error) {
	err = checkBinOrU8Column(gct)
	if err == nil {
		c.byteSlice = gct
	}
	return err
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
func (c *StmtCfg) ByteSlice() GoColumnType {
	return c.byteSlice
}
