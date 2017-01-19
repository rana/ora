# Changelog #

## master ##

## v3.8.4 ##

## v3.8.3 ##
  * Fix open/close memory leak by adding more handle freeing - see #148.

## v3.8.1 ##
  * Fix #141 (nil panic in ses.Close).
  * Remove connection pooling from Con.

## v3.8.0 ##
  * go1.8: NO support for additional features, use v4!
  * Change default column types for BLOB and CLOB: Bin and S, instead of D and D.
    With this change the driver will return a string/[]byte from [CB]LOB columns,
	instead of an io.ReadCloser.
  * Return driver.ErrBadConn for connection errors, to allow database/sql to reconnect.
  * Fix RowsAffected for Oracle 11 client (#139).

## v3.7.4 ##
  * Add RTrimChar to StmtCfg, default true. This makes the strip of right padding of CHAR columns configurable.

## v3.7.3 ##
  * Fix a panic with Con.sysName.
  * Be more forgiving for empty/non-open/NULL result sets returned by Oracle.
  * Treat SQLT_INT columns returned by TimesTen just as SQLT_NUM.
  * Make *[]string, *[]{,u}int{16,32,64} and *[]{Ui,I}nt{16,32,64}, *[]Date bindable.
  * Use a bytesArena (sync.Pool-based []byte arena) for []byte allocations.

## v3.7.2 ##
  * Fix panic in Pool.Get - #118.
  * Make returned LOB colums obey the given GoColumnType, to be able to force the into string/[]byte - see #117.
  * Return nil for everywhere in def*.value, if possible (string is an exception, see #105, #106).

## v3.7.1 ##
  * Fix defDate NULL handling error resulting in reading garbage as Time.
  * Add Ses.SetAction to be able to set the session's Module and Action attributes.

## v3.7.0 ##
  * Add "C" field to Lob to allow setting bind parameters to CLOB (default is BLOB).
  * Add date.Date, an implementation for encode/decode Oracle 7 byte DATE format, to avoid the overhead of calling C.OCIDateTime... functions. And use it in Rsets.
  * Make ora.Date use date.Date.
  * Fix nested rsets (rset with rset as field).
  * Add a new Pool implementation, which has a simple 1-1 pairing between ses and srv - easier to use,
    and "automatically" correct for parallel execution.
  * get rid of NewSesCfg() and NewSrvCfg()

## v3.6 ##
  * Refactor def* (resultset columns) to be more common, and allow multiple row fetch.
  * Use multiple row fetch to speed SELECTs - closes issue #86.

## v3.5 ##
  * Modify default for CHAR(1) columns: use ora.S (string), NOT ora.B (bool).
  * Add connect " AS SYSDBA" functionality.

## v3.4 ##
  * Implement PL/SQL TABLE support to be able to call PL/SQL blocks/stored procedures with slices of simple types.
  * Add numberBigInt and numberBigFloat colum types.
    This allows tweaking the used column types for NUMBER columns with unknown scale/precision.
  * OCINum instead of OraNum - space saving and less C calls (implemented pure Go read/write of Oracle OCINumber, in num directory).

## v3.3 ##
  * Introduce new `Num` and `OraNum` data types and `N` Go Column Type to represent Oracle numbers fully, by exchanging them with Go as strings.
    Make `N` the default column type for numbers unrepresentable by int64 or float64 (more digits than 19 or 15).

## v3.2 ##
  * Rewrite for Go 1.6 cgo restrictions.

## v3.1 ##

  * Use the returned length info for several places when reading attributes - such as column names.
  * Use float64 for non-integer numeric column types.

## v3.0 ##

  * Big rewrite of handles to pool them properly.
