/*
Package ora implements an Oracle database driver.

An Oracle database may be accessed through the database/sql package or through the
ora package directly. database/sql offers connection pooling, thread safety,
a consistent API to multiple database technologies and a common set of Go types.
The ora package offers additional features including pointers, slices, nullable
types, numerics of various sizes, Oracle-specific types, Go return type configuration,
and Oracle abstractions such as environment, server and session.

The ora package is written with the Oracle Call Interface (OCI) C-language
libraries provided by Oracle. The OCI libraries are a standard for client
application communication and driver communication with Oracle databases.

The ora package has been verified to work with Oracle Enterprise 12c (12.1.0.1.0),
Windows 8.1 and 64-bit x86.

Installation

Minimum requirements are Go 1.3 with CGO enabled, a GCC C compiler, and
Oracle 12c (12.1.0.1.0) or Oracle Instant Client (12.1.0.1.0).

Get the ora package from GitHub:

	go get github.com/ranaian/ora

Install Oracle 12c or Oracle Instant Client.

Set the CGO_CFLAGS and CGO_LDFLAGS environment variables to locate the OCI headers
and library. For example:

	// example OS environment variables for Oracle 12c on Windows
	CGO_CFLAGS=-Ic:/oracle/home/OCI/include/
	CGO_LDFLAGS=c:/oracle/home/BIN/oci.dll

CGO_CFLAGS identifies the location of the OCI header file. CGO_LDFLAGS identifies
the location of the OCI library. These locations will vary based on whether an Oracle
database is locally installed or whether the Oracle instant client libraries are
locally installed.

The ora package does not have any external Go package dependencies.

Data Types

The ora package supports all built-in Oracle data types. The supported Oracle
built-in data types are NUMBER, BINARY_DOUBLE, BINARY_FLOAT, FLOAT, DATE,
TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIME ZONE,
INTERVAL YEAR TO MONTH, INTERVAL DAY TO SECOND, CHAR, NCHAR, VARCHAR, VARCHAR2,
NVARCHAR2, LONG, CLOB, NCLOB, BLOB, LONG RAW, RAW, ROWID and BFILE.
SYS_REFCURSOR is also supported.

Oracle does not provide a built-in boolean type. Oracle provides a single-byte
character type. A common practice is to define two single-byte characters which
represent true and false. The ora package adopts this approach. The oracle
package associates a Go bool value to a Go rune and sends and receives the rune
to a CHAR(1 BYTE) column or CHAR(1 CHAR) column.

The default false rune is zero '0'. The default true rune is one '1'. The bool rune
association may be configured or disabled when directly using the ora package
but not with the database/sql package.

SQL Placeholder Syntax

Within a SQL string a placeholder may be specified to indicate where a Go variable
is placed. The SQL placeholder is an Oracle identifier, from 1 to 30
characters, prefixed with a colon (:). For example:

	// example Oracle placeholder uses a colon
	insert into t1 (c1) values (:c1)

Placeholders within a SQL statement are bound by position. The actual name is not
used by the ora package driver e.g., placeholder names :c1, :1, or :xyz are
treated equally.

Working With The Sql Package

You may access an Oracle database through the database/sql package. The database/sql
package offers a consistent API across different databases, connection
pooling, thread safety and a set of common Go types. database/sql makes working
with Oracle straight-forward.

The ora package implements interfaces in the database/sql/driver package enabling
database/sql to communicate with an Oracle database. Using database/sql
ensures you never have to call the ora package directly.

When using database/sql, the mapping between Go types and Oracle types is immutable.
The Go-to-Oracle type mapping for database/sql is:

	Go type		Oracle type

	int64		NUMBER°, BINARY_DOUBLE, BINARY_FLOAT, FLOAT

	float64		NUMBER¹, BINARY_DOUBLE, BINARY_FLOAT, FLOAT

	time.Time	TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIME ZONE, DATE

	string		CHAR², NCHAR, VARCHAR, VARCHAR2, NVARCHAR2, LONG, CLOB, NCLOB

	bool		CHAR(1 BYTE)³, CHAR(1 CHAR)³

	[]byte		BLOB, LONG RAW, RAW


	° A select-list column defined as an Oracle NUMBER with zero scale e.g.,
	NUMBER(10,0) is returned as an int64. Either int64 or float64 may be inserted
	into a NUMBER column with zero scale. float64 insertion will have its fractional
	part truncated.

	¹ A select-list column defined as an Oracle NUMBER with a scale greater than
	zero e.g., NUMBER(10,4) is returned as a float64. Either int64 or float64 may
	be inserted into a NUMBER column with a scale greater than zero.

	² A select-list column defined as an Oracle CHAR with a length greater than 1
	e.g., CHAR(2 BYTE) or CHAR(2 CHAR) is returned as a string. A Go string of any
	length up to the column max length may be inserted into the CHAR column.

	³ The Go bool value false is mapped to the zero rune '0'. The Go bool value
	true is mapped to the one rune '1'.

Working With The Oracle Package Directly

The ora package allows programming with pointers, slices, nullable types,
numerics of various sizes, Oracle-specific types, Go return type configuration, and
Oracle abstractions such as environment, server and session. When working with the
ora package directly, the API is slightly different than database/sql.

When using the ora package directly, the mapping between Go types and Oracle types 
is mutable. The Go-to-Oracle type mapping for the ora package is:

	Go type				Oracle type

	int64, int32, int16, int8	NUMBER°, BINARY_DOUBLE, BINARY_FLOAT, FLOAT
	uint64, uint32, uint16, uint8
	Int64, Int32, Int16, Int8
	Uint64, Uint32, Uint16, Uint8
	
	float64, float32		NUMBER¹, BINARY_DOUBLE, BINARY_FLOAT, FLOAT
	Float64, Float32
	
	time.Time			TIMESTAMP, TIMESTAMP WITH TIME ZONE, 
	Time				TIMESTAMP WITH LOCAL TIME ZONE, DATE
	
	string				CHAR², NCHAR, VARCHAR, VARCHAR2, 
	String				NVARCHAR2, LONG, CLOB, NCLOB, ROWID
	
	bool				CHAR(1 BYTE)³, CHAR(1 CHAR)³
	Bool
	
	[]byte				BLOB, LONG RAW, RAW
	Bytes

	IntervalYM			INTERVAL MONTH TO YEAR
	
	IntervalDS			INTERVAL DAY TO SECOND
	
	Bfile				BFILE
	
	° A select-list column defined as an Oracle NUMBER with zero scale e.g.,
	NUMBER(10,0) is returned as an int64 by default. Integer and floating point 
	numerics may be inserted into a NUMBER column with zero scale. Inserting a 
	floating point numeric will have its fractional part truncated.

	¹ A select-list column defined as an Oracle NUMBER with a scale greater than
	zero e.g., NUMBER(10,4) is returned as a float64 by default. Integer and 
	floating point numerics may be inserted into a NUMBER column with a scale 
	greater than zero.

	² A select-list column defined as an Oracle CHAR with a length greater than 1
	e.g., CHAR(2 BYTE) or CHAR(2 CHAR) is returned as a string. A Go string of any
	length up to the column max length may be inserted into the CHAR column.

	³ The Go bool value false is mapped to the zero rune '0'. The Go bool value
	true is mapped to the one rune '1'.

An example of using the ora package directly:

	package main

	import (
		"fmt"
		"github.com/ranaian/ora"
	)

	func main() {
		// example usage of the oracle package driver
		// connect to a server and open a session
		env := ora.NewEnvironment()
		env.Open()
		defer env.Close()
		srv, err := env.OpenServer("orcl")
		defer srv.Close()
		if err != nil {
			panic(err)
		}
		ses, err := srv.OpenSession("test", "test")
		defer ses.Close()
		if err != nil {
			panic(err)
		}

		// create table
		stmtTbl, err := ses.Prepare("create table t1 " +
			"(c1 number(19,0) generated always as identity (start with 1 increment by 1), " +
			"c2 varchar2(48 char))")
		defer stmtTbl.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err := stmtTbl.Execute()
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)

		// begin first transaction
		tx1, err := ses.BeginTransaction()
		if err != nil {
			panic(err)
		}

		// insert record
		var id uint64
		str := "Go is expressive, concise, clean, and efficient."
		stmtIns, err := ses.Prepare("insert into t1 (c2) values (:c2) returning c1 into :c1")
		defer stmtIns.Close()
		rowsAffected, err = stmtIns.Execute(str, &id)
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)

		// insert nullable String slice
		a := make([]ora.String, 4)
		a[0] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
		a[1] = ora.String{IsNull: true}
		a[2] = ora.String{Value: "It's a fast, statically typed, compiled"}
		a[3] = ora.String{Value: "One of Go's key design goals is code"}
		stmtSliceIns, err := ses.Prepare("insert into t1 (c2) values (:c2)")
		defer stmtSliceIns.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err = stmtSliceIns.Execute(a)
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)

		// fetch records
		stmtFetch, err := ses.Prepare("select c1, c2 from t1")
		defer stmtFetch.Close()
		if err != nil {
			panic(err)
		}
		resultSet, err := stmtFetch.Fetch()
		if err != nil {
			panic(err)
		}
		for resultSet.Next() {
			fmt.Println(resultSet.Row[0], resultSet.Row[1])
		}
		if resultSet.Err != nil {
			panic(resultSet.Err)
		}

		// commit first transaction
		err = tx1.Commit()
		if err != nil {
			panic(err)
		}

		// begin second transaction
		tx2, err := ses.BeginTransaction()
		if err != nil {
			panic(err)
		}
		// insert null String
		nullableStr := ora.String{IsNull: true}
		stmtTrans, err := ses.Prepare("insert into t1 (c2) values (:c2)")
		defer stmtTrans.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err = stmtTrans.Execute(nullableStr)
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)
		// rollback second transaction
		err = tx2.Rollback()
		if err != nil {
			panic(err)
		}

		// fetch and specify return type
		stmtCount, err := ses.Prepare("select count(c1) from t1 where c2 is null", ora.U8)
		defer stmtCount.Close()
		if err != nil {
			panic(err)
		}
		resultSet, err = stmtCount.Fetch()
		if err != nil {
			panic(err)
		}
		row := resultSet.NextRow()
		if row != nil {
			fmt.Println(row[0])
		}
		if resultSet.Err != nil {
			panic(resultSet.Err)
		}

		// create stored procedure with sys_refcursor
		stmtProcCreate, err := ses.Prepare(
			"create or replace procedure proc1(p1 out sys_refcursor) as begin " +
			"open p1 for select c1, c2 from t1 where c1 > 2 order by c1; " +
			"end proc1;")
		defer stmtProcCreate.Close()
		rowsAffected, err = stmtProcCreate.Execute()
		if err != nil {
			panic(err)
		}

		// call stored procedure
		// pass *ResultSet to Execute to receive the results of a sys_refcursor
		stmtProcCall, err := ses.Prepare("call proc1(:1)")
		defer stmtProcCall.Close()
		if err != nil {
			panic(err)
		}
		procResultSet := &ora.ResultSet{}
		rowsAffected, err = stmtProcCall.Execute(procResultSet)
		if err != nil {
			panic(err)
		}
		if procResultSet.IsOpen() {
			for procResultSet.Next() {
				fmt.Println(procResultSet.Row[0], procResultSet.Row[1])
			}
			if procResultSet.Err != nil {
				panic(procResultSet.Err)
			}
			fmt.Println(procResultSet.Len())
		}

		// Output:
		// 0
		// 1
		// 4
		// 1 Go is expressive, concise, clean, and efficient.
		// 2 Its concurrency mechanisms make it easy to
		// 3 <nil>
		// 4 It's a fast, statically typed, compiled
		// 5 One of Go's key design goals is code
		// 1
		// 1
		// 3 <nil>
		// 4 It's a fast, statically typed, compiled
		// 5 One of Go's key design goals is code
		// 3
	}

Pointers may be used to capture out-bound values from a SQL statement such as
an insert or stored procedure call. For example, a numeric pointer captures an
identity value:
	
	// given:
	// create table t1 (
	// c1 number(19,0) generated always as identity (start with 1 increment by 1),
	// c2 varchar2(48 char))
	var id int64
	stmt, err = ses.Prepare("insert into t1 (c2) values ('go') returning c1 into :c1")
	stmt.Execute(&id)

A string pointer captures an out parameter from a stored procedure:

	// given:
	// create or replace procedure proc1 (p1 out varchar2) as begin p1 := 'go'; end proc1;
	var str string
	stmt, err = ses.Prepare("call proc1(:1)")
	stmt.Execute(&str)

Slices may be used to insert multiple records with a single insert statement:

	// insert one million rows with single insert statement
	// given: create table t1 (c1 number)
	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Execute(values)

The ora package provides nullable Go types to support DML operations such as
insert and select. The nullable Go types provided by the ora package are Int64,
Int32, Int16, Int8, Uint64, Uint32, Uint16, Uint8, Float64, Float32, Time,
IntervalYM, IntervalDS, String, Bool, Bytes and Bfile. For example, you may insert
nullable Strings and select nullable Strings:

	// insert String slice
	// given: create table t1 (c1 varchar2(48 char))
	a := make([]ora.String, 5)
	a[0] = ora.String{Value: "Go is expressive, concise, clean, and efficient."}
	a[1] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	a[2] = ora.String{IsNull: true}
	a[3] = ora.String{Value: "It's a fast, statically typed, compiled"}
	a[4] = ora.String{Value: "One of Go's key design goals is code"}
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Execute(a)

	// Specify OraS to Prepare method to return ora.String values
	// fetch records
	stmt, err = ses.Prepare("select c1 from t1", OraS)
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0])
	}

The Statement.Prepare method is variadic accepting zero or more GoColumnType
which define a Go return type for a select-list column. For example, a Prepare
call can be configured to return an int64 and a nullable Int64 from the same
column:

	// given: create table t1 (c1 number)
	stmt, err = ses.Prepare("select c1, c1 from t1", ora.I64, ora.OraI64)
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0], resultSet.Row[1])
	}

Go numerics of various sizes are supported in DML operations. The ora package
supports int64, int32, int16, int8, uint64, uint32, uint16, uint8, float64 and
float32. For example, you may insert a uint16 and select numerics of various sizes:

	// insert uint16
	// given: create table t1 (c1 number)
	value := uint16(9)
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Execute(value)

	// select numerics of various sizes from the same column
	stmt, err = ses.Prepare(
		"select c1, c1, c1, c1, c1, c1, c1, c1, c1, c1, from t1",
		ora.I64, ora.I32, ora.I16, ora.I8, ora.U64, ora.U32, ora.U16, ora.U8, 
		ora.F64, ora.F32)
	resultSet, err := stmt.Fetch()
	row := resultSet.NextRow()

If a non-nullable type is defined for a nullable column returning null, the Go
type's zero value is returned.

GoColumnTypes defined by the ora package are:

	Go type		GoColumnType

	int64		I64

	int32		I32

	int16		I16

	int8		I8

	uint64		U64

	uint32		U32

	uint16		U16

	uint8		U8

	float64		F64

	Int64		OraI64

	Int32		OraI32

	Int16		OraI16

	Int8		OraI8

	Uint64		OraU64

	Uint32		OraU32

	Uint16		OraU16

	Uint8		OraU8

	Float64		OraF64

	Float32		OraF32

	time.Time	T

	Time		OraT

	string		S

	String		OraS

	bool		B

	Bool		OraB

	[]byte		Bits

	Bytes		OraBits

	default°	D

	° D represents a default mapping between a select-list column and a Go type.
	The default mapping is defined in ResultSetConfig.

When Statement.Prepare doesn't receive a GoColumnType, or receives an incorrect GoColumnType, 
the default value defined in ResultSetConfig is used. 
	
There are two configuration structs, StatementConfig and ResultSetConfig.
StatementConfig configures various aspects of a Statement. ResultSetConfig configures
various aspects of a ResultSet, including the default mapping between an Oracle select-list
column and a Go type. StatementConfig may be set in an Environment, Server, Session
and Statement. ResultSetConfig may be set in a StatementConfig.

Setting StatementConfig on Environment, Server, Session
or Statement cascades the StatementConfig to all current and future descendent structs.
An Environment may contain multiple Servers. A Server may contain multiple Sessions.
A Session may contain multiple Statements. A Statement may contain multiple ResultSets.

	// setting StatementConfig cascades to descendent structs
	// Environment -> Server -> Session -> Statement -> ResultSet

Setting a ResultSetConfig on a StatementConfig does not cascade through descendent structs.
Configuration of Statement.Config takes effect prior to calls to Statement.Execute and
Statement.Fetch; consequently, any updates to Statement.Config after a call to Statement.Execute
or Statement.Fetch are not observed.

One configuration scenario may be to set a server's select statements to return nullable Go types by 
default:

	sc := NewStatementConfig()
	sc.ResultSet.SetNumberScaless(ora.OraI64)
	sc.ResultSet.SetNumberScaled(ora.OraF64)
	sc.ResultSet.SetBinaryDouble(ora.OraF64)
	sc.ResultSet.SetBinaryFloat(ora.OraF64)
	sc.ResultSet.SetFloat(ora.OraF64)
	sc.ResultSet.SetDate(ora.OraT)
	sc.ResultSet.SetTimestamp(ora.OraT)
	sc.ResultSet.SetTimestampTz(ora.OraT)
	sc.ResultSet.SetTimestampLtz(ora.OraT)
	sc.ResultSet.SetChar1(ora.OraB)
	sc.ResultSet.SetVarchar(ora.OraS)
	sc.ResultSet.SetLong(ora.OraS)
	sc.ResultSet.SetClob(ora.OraS)
	sc.ResultSet.SetBlob(ora.OraBits)
	sc.ResultSet.SetRaw(ora.OraBits)
	sc.ResultSet.SetLongRaw(ora.OraBits)
	srv, err := env.OpenServer("orcl")
	// setting the server StatementConfig will cascade to any open Sessions, Statements
	// any new Session, Statement will receive this StatementConfig
	// any new ResultSet will receive the StatementConfig.ResultSet configuration
	srv.SetStatementConfig(sc)

Another scenario may be to configure the runes mapped to bool values:
	
	// update StatementConfig to change the FalseRune and TrueRune inserted into the database
	// given: create table t1 (c1 char(1 byte))
	
	// insert 'false' record
	var falseValue bool = false
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Config.FalseRune = 'N'
	stmt.Execute(falseValue)
	
	// insert 'true' record
	var trueValue bool = true
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Config.TrueRune = 'Y'
	stmt.Execute(trueValue)

	// update ResultSetConfig to change the TrueRune
	// used to translate an Oracle char to a Go bool
	// fetch inserted records
	stmt, err = ses.Prepare("select c1 from t1")
	stmt.Config.TrueRune = 'Y'
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0])
	}

Oracle-specific types offered by the ora package are ResultSet, IntervalYM, IntervalDS, and Bfile.
ResultSet represents an Oracle SYS_REFCURSOR. IntervalYM represents an Oracle INTERVAL YEAR TO MONTH.
IntervalDS represents an Oracle INTERVAL DAY TO SECOND. And Bfile represents an Oracle BFILE. ROWID 
columns are returned as strings and don't have a unique Go type. 

ResultSet is used to obtain Go values from a SQL select statement. Methods ResultSet.Next, 
ResultSet.NextRow, and ResultSet.Len are available. Fields ResultSet.Row, ResultSet.Err, 
ResultSet.Index, and ResultSet.ColumnNames are also available. The Next method attempts to 
load data from an Oracle buffer into Row, returning true when successful. When no data is available, 
or if an error occurs, Next returns false setting Row to nil. Any error in Next is assigned to Err. 
Calling Next increments Index and method Len returns the total number of rows processed. The NextRow 
method is convenient for returning a single row. NextRow calls Next and returns Row. 

ResultSet has two usages. ResultSet may be returned from Statement.Fetch when prepared with a SQL select 
statement:

	// given: create table t1 (c1 number, c2, char(1 byte), c3 varchar2(48 char))
	stmt, err = ses.Prepare("select c1, c2, c3 from t1")
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Index, resultSet.Row[0], resultSet.Row[1], resultSet.Row[2])
	}

Or, a *ResultSet may be passed to Statement.Execute when prepared with a stored procedure accepting 
an OUT SYS_REFCURSOR parameter:
	
	// given:
	// create table t1 (c1 number, c2 varchar2(48 char))
	// create or replace procedure proc1(p1 out sys_refcursor) as 
	// begin open p1 for select c1, c2 from t1 order by c1; end proc1;
	stmt, err = ses.Prepare("call proc1(:1)")
	resultSet := &ora.ResultSet{}
	stmt.Execute(resultSet)
	if resultSet.IsOpen() {
		for resultSet.Next() {
			fmt.Println(resultSet.Row[0], resultSet.Row[1])
		}
	}

Stored procedures with multiple OUT SYS_REFCURSOR parameters enable a single Execute call to obtain 
multiple ResultSets:

	// given:
	// create table t1 (c1 number, c2 varchar2(48 char))
	// create or replace procedure proc1(p1 out sys_refcursor, p2 out sys_refcursor) as 
	// begin open p1 for select c1 from t1 order by c1; open p2 for select c2 from t1 order by c2; 
	// end proc1;
	stmt, err = ses.Prepare("call proc1(:1, :2)")
	resultSet1 := &ora.ResultSet{}
	resultSet2 := &ora.ResultSet{}
	stmt.Execute(resultSet1, resultSet2)
	// read from first cursor
	if resultSet1.IsOpen() {
		for resultSet1.Next() {
			fmt.Println(resultSet1.Row[0])
		}
	}
	// read from second cursor
	if resultSet2.IsOpen() {
		for resultSet2.Next() {
			fmt.Println(resultSet2.Row[0])
		}
	}

The types of values assigned to Row may be configured in the StatementConfig.ResultSet field. For configuration 
to take effect, assign StatementConfig.ResultSet prior to calling Statement.Fetch or Statement.Execute. 

ResultSet prefetching may be controlled by StatementConfig.PrefetchRowCount and
StatementConfig.PrefetchMemorySize. PrefetchRowCount works in coordination with 
PrefetchMemorySize. When PrefetchRowCount is set to zero only PrefetchMemorySize is used;
otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
The default uses a PrefetchMemorySize of 134MB.

Opening and closing ResultSets is managed internally. ResultSet doesn't have an Open method or Close method.

IntervalYM may be be inserted and selected:

	// insert IntervalYM slice
	// given: create table t1 (c1 interval year to month)
	a := make([]ora.IntervalYM, 5)
	a[0] = ora.IntervalYM{Year: 1, Month: 1}
	a[1] = ora.IntervalYM{Year: 99, Month: 9}
	a[2] = ora.IntervalYM{IsNull: true}
	a[3] = ora.IntervalYM{Year: -1, Month: -1}
	a[4] = ora.IntervalYM{Year: -99, Month: -9}
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Execute(a)

	// fetch IntervalYM
	stmt, err = ses.Prepare("select c1 from t1")
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0])
	}

IntervalDS may be be inserted and selected:

	// insert IntervalDS slice
	// given: create table t1 (c1 interval day to second)
	a := make([]ora.IntervalDS, 5)
	a[0] = ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	a[1] = ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}
	a[2] = ora.IntervalDS{IsNull: true}
	a[3] = ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}
	a[4] = ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}
	stmt, err = ses.Prepare("insert into t1 (c1) values (:c1)")
	stmt.Execute(a)

	// fetch IntervalDS
	stmt, err = ses.Prepare("select c1 from t1")
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0])
	}
	
Transactions on an Oracle server are supported:
	
	// given: create table t1 (c1 number)
	
	// rollback
	tx, err := ses.BeginTransaction()
	stmt, err = ses.Prepare("insert into t1 (c1) values (3)")
	stmt.Execute()
	stmt, err = ses.Prepare("insert into t1 (c1) values (5)")
	stmt.Execute()
	tx.Rollback()

	// commit
	tx, err = ses.BeginTransaction()
	stmt, err = ses.Prepare("insert into t1 (c1) values (7)")
	stmt.Execute()
	stmt, err = ses.Prepare("insert into t1 (c1) values (9)")
	stmt.Execute()
	tx.Commit()

	// fetch records
	stmt, err = ses.Prepare("select c1 from t1")
	resultSet, err := stmt.Fetch()
	for resultSet.Next() {
		fmt.Println(resultSet.Row[0])
	}

The Server.Ping method is available to check whether the client's connection to the 
Oracle server is valid. A call to Ping requires an open Session. Ping 
will return a nil error when the connection is fine:

	// open a session before calling Ping
	ses, _ := srv.OpenSession("username", "password")
	err := srv.Ping()
	if err == nil {
		fmt.Println("Ping sucessful")
	}

The Server.Version method is available to obtain the Oracle server version. A call 
to Version requires an open Session:

	// open a session before calling Version
	ses, err := srv.OpenSession("username", "password")
	version, err := srv.Version()
	if version != "" && err == nil {
		fmt.Println("Received version from server")
	}

Further code examples are available in the samples folder, example file and test files.
	
Test Database Setup

Tests are available and require some setup. Setup varies depending on whether
the Oracle server is configured as a container database or non-container database.
It's simpler to setup a non-container database. An example for each setup is
explained.

Non-container test database setup steps:

	// 1. login to an Oracle server with SqlPlus as sysdba:
	sqlplus / as sysdba

	// 2. create a file for the test database use
	CREATE TABLESPACE test_ts NOLOGGING DATAFILE 'test.dat' SIZE 100M AUTOEXTEND ON;

	// 3. create a test database
	CREATE USER test IDENTIFIED BY test DEFAULT TABLESPACE test_ts;

	// 4. grant permissions to the database
	GRANT CREATE SESSION, CREATE TABLE, CREATE SEQUENCE, 
	CREATE PROCEDURE, UNLIMITED TABLESPACE TO test;

	// 5. create OS environment variables
	// specify your_database_name; varies based on installation; may be 'orcl'
	GO_ORA_DRV_TEST_DB = your_database_name
	GO_ORA_DRV_TEST_USERNAME = test
	GO_ORA_DRV_TEST_PASSWORD = test


Container test database setup steps:

	// 1. login to an Oracle server with SqlPlus as sysdba:
	sqlplus / as sysdba

	// 2. create a test pluggable database and permissions
	// you will need to change the FILE_NAME_CONVERT file paths for your database installation
	CREATE PLUGGABLE DATABASE go_driver_test
	ADMIN USER test IDENTIFIED BY test
	ROLES = (DBA)
	FILE_NAME_CONVERT = ('d:\oracle\data\orcl\pdbseed\', 'd:\oracle\data\go_driver_test\');

	// 3. modify the pluggable database settings
	ALTER PLUGGABLE DATABASE go_driver_test OPEN;
	ALTER SESSION SET CONTAINER = go_driver_test;
	GRANT DBA TO test;

	// 4. add new database service to the tnsnames.ora file:
	// located on your client machine in $ORACLE_HOME\network\admin\tnsnames.ora
	GO_DRIVER_TEST =
	  (DESCRIPTION =
	    (ADDRESS = (PROTOCOL = TCP)(HOST = localhost)(PORT = 1521))
	    (CONNECT_DATA =
	      (SERVER = DEDICATED)
	      (SERVICE_NAME = go_driver_test)
	    )
	  )

	// 5. create OS environment variables
	GO_ORA_DRIVER_TEST_DB = go_driver_test
	GO_ORA_DRIVER_TEST_USERNAME = test
	GO_ORA_DRIVER_TEST_PASSWORD = test

Some helpful SQL maintenance statements:

	// delete all tables in a non-container database
	BEGIN
	FOR c IN (SELECT table_name FROM user_tables) LOOP
	EXECUTE IMMEDIATE ('DROP TABLE "' || c.table_name || '" CASCADE CONSTRAINTS');
	END LOOP;
	END;

	// delete the non-container test database; use SqlPlus as sysdba
	DROP USER test CASCADE;

Run the tests.
	
Limitations

database/sql method Stmt.QueryRow is not supported.

License

Copyright 2014 Rana Ian. All rights reserved.
Use of this source code is governed by The MIT License
found in the accompanying LICENSE file.

*/
package ora
