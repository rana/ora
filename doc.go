//go:generate go get github.com/robertkrimen/godocdown/godocdown
//go:generate godocdown -output README.md

/*
Package ora implements an Oracle database driver.


### Golang Oracle Database Driver ###

#### TL;DR; just use it ####

	import (
		"database/sql"

		_ "gopkg.in/rana/ora.v4"
	)

	func main() {
		db, err := sql.Open("ora", "user/passw@host:port/sid")
		defer db.Close()

		// Set timeout (Go 1.8)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// Set prefetch count (Go 1.8)
		ctx = ora.WithStmtCfg(ctx, ora.Cfg().StmtCfg.SetPrefetchCount(50000))
		rows, err := db.QueryContext(ctx, "SELECT * FROM user_objects")
		defer rows.Close()
	}

Call stored procedure with OUT parameters:

	import (
		"gopkg.in/rana/ora.v4"
	)

	func main() {
		env, srv, ses, err := ora.NewEnvSrvSes("user/passw@host:port/sid")
		if err != nil {
			log.Fatal(err)
		}
		defer env.Close()
		defer srv.Close()
		defer ses.Close()

		var user string
		if _, err = ses.PrepAndExe("BEGIN :1 := SYS_CONTEXT('USERENV', :2); END;", &res, "SESSION_USER"); err != nil {
			log.Fatal(err)
		}
		log.Printf("user: %q", user)
	}

Background

An Oracle database may be accessed through the [database/sql](http://golang.org/pkg/database/sql) package or through the
ora package directly. database/sql offers connection pooling, thread safety,
a consistent API to multiple database technologies and a common set of Go types.
The ora package offers additional features including pointers, slices, nullable
types, numerics of various sizes, Oracle-specific types, Go return type configuration,
and Oracle abstractions such as environment, server and session.

The ora package is written with the Oracle Call Interface (OCI) C-language
libraries provided by Oracle. The OCI libraries are a standard for client
application communication and driver communication with Oracle databases.

The ora package has been verified to work with:

* Oracle Standard 11g (11.2.0.4.0), Linux x86_64 (RHEL6)

* Oracle Enterprise 12c (12.1.0.1.0), Windows 8.1 and AMD64.

---

* [Installation](https://github.com/rana/ora#installation)

* [Data Types](https://github.com/rana/ora#data-types)

* [SQL Placeholder Syntax](https://github.com/rana/ora#sql-placeholder-syntax)

* [Working With The Sql Package](https://github.com/rana/ora#working-with-the-sql-package)

* [Working With The Oracle Package Directly](https://github.com/rana/ora#working-with-the-oracle-package-directly)

* [Logging](https://github.com/rana/ora#logging)

* [Test Database Setup](https://github.com/rana/ora#test-database-setup)

* [Limitations](https://github.com/rana/ora#limitations)

* [License](https://github.com/rana/ora#license)

* [API Reference](http://godoc.org/github.com/rana/ora#pkg-index)

* [Examples](./examples)

---


Installation

Minimum requirements are Go 1.3 with CGO enabled, a GCC C compiler, and
Oracle 11g (11.2.0.4.0) or Oracle Instant Client (11.2.0.4.0).

Install Oracle or Oracle Instant Client.

Copy the [oci8.pc](contrib/oci8.pc) from the `contrib` folder
(or the one for your system, maybe tailored to your specific locations) to a folder
in `$PKG_CONFIG_PATH` or a system folder, such as

	cp -aL contrib/oci8.pc /usr/local/lib/pkgconfig/oci8.pc

The ora package has no external Go dependencies and is available on GitHub and
gopkg.in:

	go get gopkg.in/rana/ora.v4

*WARNING*: If you have Oracle Instant Client 11.2, you'll need to add "=lnnz11"
to the list of linked libs!
Otherwise, you may encounter "undefined reference to `nzosSCSP_SetCertSelectionParams' "
errors.
Oracle Instant Client 12.1 does not need this.

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
	INSERT INTO T1 (C1) VALUES (:C1)

Placeholders within a SQL statement are bound by position. The actual name is not
used by the ora package driver e.g., placeholder names :c1, :1, or :xyz are
treated equally.
*/
//
// LastInsertId
//
// The `database/sql` package provides a LastInsertId method to return the
// last inserted row's id. Oracle does not provide such functionality,
// but if you append `... RETURNING col /*LastInsertId*/` to your SQL, then it will
// be presented as LastInsertId. Note that you have to mark with a `/*LastInsertId*/`
// (case insensitive) your `RETURNING` part, to allow ora to return the last column
// as `LastInsertId()`. That column must fit in `int64`, though!
/*

Working With The Sql Package

You may access an Oracle database through the database/sql package. The database/sql
package offers a consistent API across different databases, connection
pooling, thread safety and a set of common Go types. database/sql makes working
with Oracle straight-forward.

The ora package implements interfaces in the database/sql/driver package enabling
database/sql to communicate with an Oracle database. Using database/sql
ensures you never have to call the ora package directly.

When using database/sql, the mapping between Go types and Oracle types may be
changed slightly. The database/sql package has strict expectations on Go return
types. The Go-to-Oracle type mapping for database/sql is:

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

The "ora" driver is automatically registered for use with sql.Open, but you can
call ora.SetCfg to set the used configuration options including
statement configuration and Rset configuration.

    func init() {
		drvCfg := ora.Cfg()
		drvCfg.FalseRune = 'N'
		drvCfg.TrueRune = 'Y'
		drvCfg.TrueRune = 'Y'
		ora.SetCfg(drvCfg)
	}

When configuring the driver for use with database/sql, keep in mind that
database/sql has strict Go type-to-Oracle type mapping expectations.

Working With The Oracle Package Directly

The ora package allows programming with pointers, slices, nullable types,
numerics of various sizes, Oracle-specific types, Go return type configuration, and
Oracle abstractions such as environment, server and session. When working with the
ora package directly, the API is slightly different than database/sql.

When using the ora package directly, the mapping between Go types and Oracle types
may be changed. The Go-to-Oracle type mapping for the ora package is:

	Go type				Oracle type

	int64, int32, int16, int8	NUMBER°, BINARY_DOUBLE, BINARY_FLOAT, FLOAT
	uint64, uint32, uint16, uint8
	Int64, Int32, Int16, Int8
	Uint64, Uint32, Uint16, Uint8
	*int64, *int32, *int16, *int8
	*uint64, *uint32, *uint16, *uint8
	[]int64, []int32, []int16, []int8
	[]uint64, []uint32, []uint16, []uint8
	[]Int64, []Int32, []Int16, []Int8
	[]Uint64, []Uint32, []Uint16, []Uint8


	float64, float32		NUMBER¹, BINARY_DOUBLE, BINARY_FLOAT, FLOAT
	Float64, Float32
	*float64, *float32
	[]float64, []float32
	[]Float64, []Float32

	time.Time			TIMESTAMP, TIMESTAMP WITH TIME ZONE,
	Time				TIMESTAMP WITH LOCAL TIME ZONE, DATE
	*time.Time
	[]time.Time
	[]Time

	string				CHAR², NCHAR, VARCHAR, VARCHAR2,
	String				NVARCHAR2, LONG, CLOB, NCLOB, ROWID
	*string
	[]string
	[]String

	bool				CHAR(1 BYTE)³, CHAR(1 CHAR)³
	Bool
	*bool
	[]bool
	[]Bool

	[]byte, [][]byte	BLOB

	Lob, []Lob, *Lob	BLOB, CLOB

	Raw, []Raw			RAW, LONG RAW

	IntervalYM			INTERVAL MONTH TO YEAR
	[]IntervalYM

	IntervalDS			INTERVAL DAY TO SECOND
	[]IntervalDS

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
		"gopkg.in/rana/ora.v4"
	)

	func main() {
		// example usage of the ora package driver
		// connect to a server and open a session
		env, err := ora.OpenEnv()
		defer env.Close()
		if err != nil {
			panic(err)
		}
		srvCfg := ora.SrvCfg{Dblink: "orcl"}
		srv, err := env.OpenSrv(&srvCfg)
		defer srv.Close()
		if err != nil {
			panic(err)
		}
		sesCfg := ora.SesCfg{
			Username: "test",
			Password: "test",
		}
		ses, err := srv.OpenSes(sesCfg)
		defer ses.Close()
		if err != nil {
			panic(err)
		}

		// create table
		tableName := "t1"
		stmtTbl, err := ses.Prep(fmt.Sprintf("CREATE TABLE %v "+
			"(C1 NUMBER(19,0) GENERATED ALWAYS AS IDENTITY "+
			"(START WITH 1 INCREMENT BY 1), C2 VARCHAR2(48 CHAR))", tableName))
		defer stmtTbl.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err := stmtTbl.Exe()
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)

		// begin first transaction
		tx1, err := ses.StartTx()
		if err != nil {
			panic(err)
		}

		// insert record
		var id uint64
		str := "Go is expressive, concise, clean, and efficient."
		stmtIns, err := ses.Prep(fmt.Sprintf(
			"INSERT INTO %v (C2) VALUES (:C2) RETURNING C1 INTO :C1", tableName))
		defer stmtIns.Close()
		rowsAffected, err = stmtIns.Exe(str, &id)
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
		stmtSliceIns, err := ses.Prep(fmt.Sprintf(
			"INSERT INTO %v (C2) VALUES (:C2)", tableName))
		defer stmtSliceIns.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err = stmtSliceIns.Exe(a)
		if err != nil {
			panic(err)
		}
		fmt.Println(rowsAffected)

		// fetch records
		stmtQry, err := ses.Prep(fmt.Sprintf(
			"SELECT C1, C2 FROM %v", tableName))
		defer stmtQry.Close()
		if err != nil {
			panic(err)
		}
		rset, err := stmtQry.Qry()
		if err != nil {
			panic(err)
		}
		for rset.Next() {
			fmt.Println(rset.Row[0], rset.Row[1])
		}
		if err := rset.Err(); err != nil {
			panic(err)
		}

		// commit first transaction
		err = tx1.Commit()
		if err != nil {
			panic(err)
		}

		// begin second transaction
		tx2, err := ses.StartTx()
		if err != nil {
			panic(err)
		}
		// insert null String
		nullableStr := ora.String{IsNull: true}
		stmtTrans, err := ses.Prep(fmt.Sprintf(
			"INSERT INTO %v (C2) VALUES (:C2)", tableName))
		defer stmtTrans.Close()
		if err != nil {
			panic(err)
		}
		rowsAffected, err = stmtTrans.Exe(nullableStr)
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
		stmtCount, err := ses.Prep(fmt.Sprintf(
			"SELECT COUNT(C1) FROM %v WHERE C2 IS NULL", tableName), ora.U8)
		defer stmtCount.Close()
		if err != nil {
			panic(err)
		}
		rset, err = stmtCount.Qry()
		if err != nil {
			panic(err)
		}
		row := rset.NextRow()
		if row != nil {
			fmt.Println(row[0])
		}
		if err := rset.Err(); err != nil {
			panic(err)
		}

		// create stored procedure with sys_refcursor
		stmtProcCreate, err := ses.Prep(fmt.Sprintf(
			"CREATE OR REPLACE PROCEDURE PROC1(P1 OUT SYS_REFCURSOR) AS BEGIN "+
				"OPEN P1 FOR SELECT C1, C2 FROM %v WHERE C1 > 2 ORDER BY C1; "+
				"END PROC1;",
			tableName))
		defer stmtProcCreate.Close()
		rowsAffected, err = stmtProcCreate.Exe()
		if err != nil {
			panic(err)
		}

		// call stored procedure
		// pass *Rset to Exe to receive the results of a sys_refcursor
		stmtProcCall, err := ses.Prep("CALL PROC1(:1)")
		defer stmtProcCall.Close()
		if err != nil {
			panic(err)
		}
		procRset := &ora.Rset{}
		rowsAffected, err = stmtProcCall.Exe(procRset)
		if err != nil {
			panic(err)
		}
		if procRset.IsOpen() {
			for procRset.Next() {
				fmt.Println(procRset.Row[0], procRset.Row[1])
			}
			if err := procRset.Err(); err != nil {
				panic(err)
			}
			fmt.Println(procRset.Len())
		}

		// Output:
		// 0
		// 1
		// 4
		// 1 Go is expressive, concise, clean, and efficient.
		// 2 Its concurrency mechanisms make it easy to
		// 3
		// 4 It's a fast, statically typed, compiled
		// 5 One of Go's key design goals is code
		// 1
		// 1
		// 3
		// 4 It's a fast, statically typed, compiled
		// 5 One of Go's key design goals is code
		// 3
	}


Pointers may be used to capture out-bound values from a SQL statement such as
an insert or stored procedure call. For example, a numeric pointer captures an
identity value:

	// given:
	// CREATE TABLE T1 (
	// C1 NUMBER(19,0) GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1),
	// C2 VARCHAR2(48 CHAR))
	var id int64
	stmt, err = ses.Prep("INSERT INTO T1 (C2) VALUES ('GO') RETURNING C1 INTO :C1")
	stmt.Exe(&id)

A string pointer captures an out parameter from a stored procedure:

	// given:
	// CREATE OR REPLACE PROCEDURE PROC1 (P1 OUT VARCHAR2) AS BEGIN P1 := 'GO'; END PROC1;
	var str string
	stmt, err = ses.Prep("CALL PROC1(:1)")
	stmt.Exe(&str)

Slices may be used to insert multiple records with a single insert statement:

	// insert one million rows with single insert statement
	// given: CREATE TABLE T1 (C1 NUMBER)
	values := make([]int64, 1000000)
	for n, _ := range values {
		values[n] = int64(n)
	}
	rowsAffected, err := ses.PrepAndExe("INSERT INTO T1 (C1) VALUES (:C1)", values)

The ora package provides nullable Go types to support DML operations such as
insert and select. The nullable Go types provided by the ora package are Int64,
Int32, Int16, Int8, Uint64, Uint32, Uint16, Uint8, Float64, Float32, Time,
IntervalYM, IntervalDS, String, Bool, Binary and Bfile. For example, you may insert
nullable Strings and select nullable Strings:

	// insert String slice
	// given: CREATE TABLE T1 (C1 VARCHAR2(48 CHAR))
	a := make([]ora.String, 5)
	a[0] = ora.String{Value: "Go is expressive, concise, clean, and efficient."}
	a[1] = ora.String{Value: "Its concurrency mechanisms make it easy to"}
	a[2] = ora.String{IsNull: true}
	a[3] = ora.String{Value: "It's a fast, statically typed, compiled"}
	a[4] = ora.String{Value: "One of Go's key design goals is code"}
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Exe(a)

	// Specify OraS to Prep method to return ora.String values
	// fetch records
	stmt, err = ses.Prep("SELECT C1 FROM T1", OraS)
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}

The `Stmt.Prep` method is variadic accepting zero or more `GoColumnType`
which define a Go return type for a select-list column. For example, a Prep
call can be configured to return an int64 and a nullable Int64 from the same
column:

	// given: create table t1 (c1 number)
	stmt, err = ses.Prep("SELECT C1, C1 FROM T1", ora.I64, ora.OraI64)
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0], rset.Row[1])
	}

Go numerics of various sizes are supported in DML operations. The ora package
supports int64, int32, int16, int8, uint64, uint32, uint16, uint8, float64 and
float32. For example, you may insert a uint16 and select numerics of various sizes:

	// insert uint16
	// given: create table t1 (c1 number)
	value := uint16(9)
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Exe(value)

	// select numerics of various sizes from the same column
	stmt, err = ses.Prep(
		"SELECT C1, C1, C1, C1, C1, C1, C1, C1, C1, C1, FROM T1",
		ora.I64, ora.I32, ora.I16, ora.I8, ora.U64, ora.U32, ora.U16, ora.U8,
		ora.F64, ora.F32)
	rset, err := stmt.Qry()
	row := rset.NextRow()

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

	[]byte		Bin

	Raw			Bin

	Lob°		Bin or S

	default¹	D

	° Lob will return binary data if the Oracle column is a BLOB; otherwise, Lob
	  will return a string if the Oracle column is a CLOB.

	¹ D represents a default mapping between a select-list column and a Go type.
	The default mapping is defined in RsetCfg.

When Stmt.Prep doesn't receive a GoColumnType, or receives an incorrect GoColumnType,
the default value defined in RsetCfg is used.

EnvCfg, SrvCfg, SesCfg, StmtCfg and RsetCfg are the main configuration structs.
EnvCfg configures aspects of an Env. SrvCfg configures aspects of a Srv. SesCfg
configures aspects of a Ses. StmtCfg configures aspects of a Stmt. RsetCfg
configures aspects of Rset. StmtCfg and RsetCfg have the most options to
configure. RsetCfg defines the default mapping between an Oracle select-list
column and a Go type. StmtCfg may be set in an EnvCfg, SrvCfg, SesCfg and StmtCfg.
RsetCfg may be set in a Stmt.

EnvCfg.StmtCfg, SrvCfg.StmtCfg, SesCfg.StmtCfg may optionally be specified to
configure a statement. If StmtCfg isn't specified default values are applied.
EnvCfg.StmtCfg, SrvCfg.StmtCfg, SesCfg.StmtCfg cascade to new descendent structs.
When ora.OpenEnv() is called a specified EnvCfg is used or a default EnvCfg is
created. Creating a Srv with env.OpenSrv() will use SrvCfg.StmtCfg if
it is specified; otherwise, EnvCfg.StmtCfg is copied by value to SrvCfg.StmtCfg.
Creating a Ses with srv.OpenSes() will use SesCfg.StmtCfg if it is specified;
otherwise, SrvCfg.StmtCfg is copied by value to SesCfg.StmtCfg. Creating a Stmt
with ses.Prep() will use SesCfg.StmtCfg if it is specified; otherwise, a new
StmtCfg with default values is set on the Stmt. Call Stmt.Cfg() to change a Stmt's
configuration.

An Env may contain multiple Srv. A Srv may contain multiple Ses. A Ses may
contain multiple Stmt. A Stmt may contain multiple Rset.

	// StmtCfg cascades to descendent structs
	// EnvCfg -> SrvCfg -> SesCfg -> StmtCfg -> RsetCfg

Setting a RsetCfg on a StmtCfg does not cascade through descendent structs.
Configuration of Stmt.Cfg takes effect prior to calls to Stmt.Exe and
Stmt.Qry; consequently, any updates to Stmt.Cfg after a call to Stmt.Exe
or Stmt.Qry are not observed.

One configuration scenario may be to set a server's select statements to return
nullable Go types by default:

	sc := &ora.SrvCfg{Dblink: "orcl"}
	sc.Dblink = "orcl"
	cfg := NewStmtCfg().
		SetNumberInt(ora.OraI64).
		SetNumberFloat(ora.OraF64).
		SetBinaryDouble(ora.OraF64).
		SetBinaryFloat(ora.OraF64).
		SetFloat(ora.OraF64).
		SetDate(ora.OraT).
		SetTimestamp(ora.OraT).
		SetTimestampTz(ora.OraT).
		SetTimestampLtz(ora.OraT).
		SetChar1(ora.OraS).
		SetVarchar(ora.OraS).
		SetLong(ora.OraS).
		SetClob(ora.OraS).
		SetBlob(ora.OraBin).
		SetRaw(ora.OraBin).
		SetLongRaw(ora.OraBin).
		SetFetchLen(100).
		SetLOBFetchLen(100)
	sc.StmtCfg = cfg
	srv, err := env.OpenSrv(sc)
	// any new SesCfg.StmtCfg, StmtCfg.Cfg will receive this StmtCfg
	// any new Rset will receive the StmtCfg.Rset configuration

Another scenario may be to configure the runes mapped to bool values:

	// update StmtCfg to change the FalseRune and TrueRune inserted into the database
	// given: CREATE TABLE T1 (C1 CHAR(1 BYTE))

	stmt.Cfg().Char1(ora.OraB)

	// insert 'false' record
	var falseValue bool = false
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Cfg().FalseRune = 'N'
	stmt.Exe(falseValue)

	// insert 'true' record
	var trueValue bool = true
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Cfg().TrueRune = 'Y'
	stmt.Exe(trueValue)

	// update RsetCfg to change the TrueRune
	// used to translate an Oracle char to a Go bool
	// fetch inserted records
	stmt, err = ses.Prep("SELECT C1 FROM T1")
	stmt.Cfg().Rset.TrueRune = 'Y'
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}

Oracle-specific types offered by the ora package are ora.Rset, ora.IntervalYM,
ora.IntervalDS, ora.Raw, ora.Lob and ora.Bfile. ora.Rset represents an Oracle
SYS_REFCURSOR. ora.IntervalYM represents an Oracle INTERVAL YEAR TO MONTH.
ora.IntervalDS represents an Oracle INTERVAL DAY TO SECOND. ora.Raw represents
an Oracle RAW or LONG RAW. ora.Lob may represent an Oracle BLOB or Oracle CLOB.
And ora.Bfile represents an Oracle BFILE. ROWID columns are returned as strings and
don't have a unique Go type.

#### LOBs

The default for SELECTing [BC]LOB columns is a safe Bin or S,
which means all the contents of the LOB is slurped into memory and returned
as a []byte or string.

The DefaultLOBFetchLen says LOBs are prefetched only a minimal way, to minimize
extra memory usage - you can override this using
`stmt.SetCfg(stmt.Cfg().SetLOBFetchLen(100))`.

If you want more control, you can use ora.L in Prep, Qry or
`ses.SetCfg(ses.Cfg().SetBlob(ora.L))`. But keep in mind that Oracle restricts
the use of LOBs: it is forbidden to do ANYTHING while reading the LOB!
No another query, no exec, no close of the Rset - even *advance* to the next record
in the result set is forbidden!

Failing to adhere these rules results in "Invalid handle" and ORA-03127 errors.

You cannot start reading another LOB till you haven't finished reading the previous
LOB, not even in the same row! Failing this results in ORA-24804!

For examples, see [z_lob_test.go](z_lob_test.go).

#### Rset

Rset is used to obtain Go values from a SQL select statement. Methods Rset.Next,
Rset.NextRow, and Rset.Len are available. Fields Rset.Row, Rset.Err,
Rset.Index, and Rset.ColumnNames are also available. The Next method attempts to
load data from an Oracle buffer into Row, returning true when successful. When no data is available,
or if an error occurs, Next returns false setting Row to nil. Any error in Next is assigned to Err.
Calling Next increments Index and method Len returns the total number of rows processed. The NextRow
method is convenient for returning a single row. NextRow calls Next and returns Row.
ColumnNames returns the names of columns defined by the SQL select statement.

Rset has two usages. Rset may be returned from Stmt.Qry when prepared with a SQL select
statement:

	// given: CREATE TABLE T1 (C1 NUMBER, C2, CHAR(1 BYTE), C3 VARCHAR2(48 CHAR))
	stmt, err = ses.Prep("SELECT C1, C2, C3 FROM T1")
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Index, rset.Row[0], rset.Row[1], rset.Row[2])
	}

Or, *Rset may be passed to Stmt.Exe when prepared with a stored procedure accepting
an OUT SYS_REFCURSOR parameter:

	// given:
	// CREATE TABLE T1 (C1 NUMBER, C2 VARCHAR2(48 CHAR))
	// CREATE OR REPLACE PROCEDURE PROC1(P1 OUT SYS_REFCURSOR) AS
	// BEGIN OPEN P1 FOR SELECT C1, C2 FROM T1 ORDER BY C1; END PROC1;
	stmt, err = ses.Prep("CALL PROC1(:1)")
	rset := &ora.Rset{}
	stmt.Exe(rset)
	if rset.IsOpen() {
		for rset.Next() {
			fmt.Println(rset.Row[0], rset.Row[1])
		}
	}

Stored procedures with multiple OUT SYS_REFCURSOR parameters enable a single Exe call to obtain
multiple Rsets:

	// given:
	// CREATE TABLE T1 (C1 NUMBER, C2 VARCHAR2(48 CHAR))
	// CREATE OR REPLACE PROCEDURE PROC1(P1 OUT SYS_REFCURSOR, P2 OUT SYS_REFCURSOR) AS
	// BEGIN OPEN P1 FOR SELECT C1 FROM T1 ORDER BY C1; OPEN P2 FOR SELECT C2 FROM T1 ORDER BY C2;
	// END PROC1;
	stmt, err = ses.Prep("CALL PROC1(:1, :2)")
	rset1 := &ora.Rset{}
	rset2 := &ora.Rset{}
	stmt.Exe(rset1, rset2)
	// read from first cursor
	if rset1.IsOpen() {
		for rset1.Next() {
			fmt.Println(rset1.Row[0])
		}
	}
	// read from second cursor
	if rset2.IsOpen() {
		for rset2.Next() {
			fmt.Println(rset2.Row[0])
		}
	}

The types of values assigned to Row may be configured in StmtCfg.Rset. For configuration
to take effect, assign StmtCfg.Rset prior to calling Stmt.Qry or Stmt.Exe.

Rset prefetching may be controlled by StmtCfg.PrefetchRowCount and
StmtCfg.PrefetchMemorySize. PrefetchRowCount works in coordination with
PrefetchMemorySize. When PrefetchRowCount is set to zero only PrefetchMemorySize is used;
otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is used.
The default uses a PrefetchMemorySize of 134MB.

Opening and closing Rsets is managed internally. Rset does not have an Open
method or Close method.

IntervalYM may be be inserted and selected:

	// insert IntervalYM slice
	// given: CREATE TABLE T1 (C1 INTERVAL YEAR TO MONTH)
	a := make([]ora.IntervalYM, 5)
	a[0] = ora.IntervalYM{Year: 1, Month: 1}
	a[1] = ora.IntervalYM{Year: 99, Month: 9}
	a[2] = ora.IntervalYM{IsNull: true}
	a[3] = ora.IntervalYM{Year: -1, Month: -1}
	a[4] = ora.IntervalYM{Year: -99, Month: -9}
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Exe(a)

	// query IntervalYM
	stmt, err = ses.Prep("SELECT C1 FROM T1")
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}

IntervalDS may be be inserted and selected:

	// insert IntervalDS slice
	// given: CREATE TABLE T1 (C1 INTERVAL DAY TO SECOND)
	a := make([]ora.IntervalDS, 5)
	a[0] = ora.IntervalDS{Day: 1, Hour: 1, Minute: 1, Second: 1, Nanosecond: 123456789}
	a[1] = ora.IntervalDS{Day: 59, Hour: 59, Minute: 59, Second: 59, Nanosecond: 123456789}
	a[2] = ora.IntervalDS{IsNull: true}
	a[3] = ora.IntervalDS{Day: -1, Hour: -1, Minute: -1, Second: -1, Nanosecond: -123456789}
	a[4] = ora.IntervalDS{Day: -59, Hour: -59, Minute: -59, Second: -59, Nanosecond: -123456789}
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (:C1)")
	stmt.Exe(a)

	// query IntervalDS
	stmt, err = ses.Prep("SELECT C1 FROM T1")
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}

Transactions on an Oracle server are supported. DML statements auto-commit
unless a transaction has started:

	// given: CREATE TABLE T1 (C1 NUMBER)

	// rollback
	tx, err := ses.StartTx()
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (3)")
	stmt.Exe()
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (5)")
	stmt.Exe()
	tx.Rollback()

	// commit
	tx, err = ses.StartTx()
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (7)")
	stmt.Exe()
	stmt, err = ses.Prep("INSERT INTO T1 (C1) VALUES (9)")
	stmt.Exe()
	tx.Commit()

	// fetch records
	stmt, err = ses.Prep("SELECT C1 FROM T1")
	rset, err := stmt.Qry()
	for rset.Next() {
		fmt.Println(rset.Row[0])
	}

Ses.PrepAndExe, Ses.PrepAndQry, Ses.Ins, Ses.Upd, and Ses.Sel are convenient
one-line methods.

Ses.PrepAndExe offers a convenient one-line call to Ses.Prep and Stmt.Exe.

	rowsAffected, err := ses.PrepAndExe("CREATE TABLE T1 (C1 NUMBER)")

Ses.PrepAndQry offers a convenient one-line call to Ses.Prep and Stmt.Qry.

	rset, err := ses.PrepAndQry("SELECT CURRENT_TIMESTAMP FROM DUAL")

Ses.Ins composes, prepares and executes a sql INSERT statement. Ses.Ins is useful
when you have to create and maintain a simple INSERT statement with a long
list of columns. As table columns are added and dropped over the lifetime of
a table Ses.Ins is easy to read and revise.

	err = ses.Ins("T1",
		"C2", e.C2,
		"C3", e.C3,
		"C4", e.C4,
		"C5", e.C5,
		"C6", e.C6,
		"C7", e.C7,
		"C8", e.C8,
		"C9", e.C9,
		"C10", e.C10,
		"C11", e.C11,
		"C12", e.C12,
		"C13", e.C13,
		"C14", e.C14,
		"C15", e.C15,
		"C16", e.C16,
		"C17", e.C17,
		"C18", e.C18,
		"C19", e.C19,
		"C20", e.C20,
		"C21", e.C21,
		"C1", &e.C1)

Ses.Upd composes, prepares and executes a sql UPDATE statement. Ses.Upd is useful
when you have to create and maintain a simple UPDATE statement with a long list
of columns. As table columns are added and dropped over the lifetime of
a table Ses.Upd is easy to read and revise.

	err = ses.Upd("T1",
		"C2", e.C2*2,
		"C3", e.C3*2,
		"C4", e.C4*2,
		"C5", e.C5*2,
		"C6", e.C6*2,
		"C7", e.C7*2,
		"C8", e.C8*2,
		"C9", e.C9*2,
		"C10", e.C10*2,
		"C11", e.C11*2,
		"C12", e.C12*2,
		"C13", e.C13*2,
		"C14", e.C14*2,
		"C15", e.C15*2,
		"C16", e.C16*2,
		"C17", e.C17*2,
		"C18", e.C18*2,
		"C19", e.C19*2,
		"C20", e.C20*2,
		"C21", e.C21*2,
		"C1", e.C1)

Ses.Sel composes, prepares and queries a sql SELECT statement. Ses.Sel is useful
when you have to create and maintain a simple SELECT statement with a long
list of columns that have non-default GoColumnTypes. As table columns are added
and dropped over the lifetime of a table Ses.Sel is easy to read and revise.

	rset, err := ses.Sel("T1",
		"C1", ora.U64,
		"C2", ora.F64,
		"C3", ora.I8,
		"C4", ora.I16,
		"C5", ora.I32,
		"C6", ora.I64,
		"C7", ora.U8,
		"C8", ora.U16,
		"C9", ora.U32,
		"C10", ora.U64,
		"C11", ora.F32,
		"C12", ora.F64,
		"C13", ora.I8,
		"C14", ora.I16,
		"C15", ora.I32,
		"C16", ora.I64,
		"C17", ora.U8,
		"C18", ora.U16,
		"C19", ora.U32,
		"C20", ora.U64,
		"C21", ora.F32)

The Ses.Ping method checks whether the client's connection to an
Oracle server is valid. A call to Ping requires an open Ses. Ping
will return a nil error when the connection is fine:

	// open a session before calling Ping
	ses, _ := srv.OpenSes("username", "password")
	err := ses.Ping()
	if err == nil {
		fmt.Println("Ping successful")
	}

The Srv.Version method is available to obtain the Oracle server version. A call
to Version requires an open Ses:

	// open a session before calling Version
	ses, err := srv.OpenSes("username", "password")
	version, err := srv.Version()
	if version != "" && err == nil {
		fmt.Println("Received version from server")
	}

Further code examples are available in the [example file](https://github.com/rana/ora/blob/master/z_example_test.go), test files and [samples folder](https://github.com/rana/ora/tree/master/samples).

Logging

The ora package provides a simple ora.Logger interface for logging. Logging is
disabled by default. Specify one of three optional built-in logging packages to
enable logging; or, use your own logging package.

ora.Cfg().Log offers various options to enable or disable logging of specific
ora driver methods. For example:

	// enable logging of the Rset.Next method
	ora.Cfg().Log.Rset.Next = true

To use the standard Go log package:

	import (
		"gopkg.in/rana/ora.v4"
		"gopkg.in/rana/ora.v4/lg"
	)

	func main() {
		// use an optional log package for ora logging
		ora.Cfg().Log.Logger = lg.Log
	}

which produces a sample log of:

	ORA I 2015/05/23 16:54:44.615462 drv.go:411: OpenEnv 1
	ORA I 2015/05/23 16:54:44.626443 drv.go:411: OpenEnv 2
	ORA I 2015/05/23 16:54:44.627465 env.go:115: E2] OpenSrv (dbname orcl)
	ORA I 2015/05/23 16:54:44.643449 env.go:150: E2] OpenSrv (srvId 1)
	ORA I 2015/05/23 16:54:44.643449 srv.go:113: E2S1] OpenSes (username test)
	ORA I 2015/05/23 16:54:44.665451 ses.go:163: E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL
	ORA I 2015/05/23 16:54:44.666451 rset.go:205: E2S1S1S1R0] open
	ORA I 2015/05/23 16:54:44.666451 ses.go:74: E2S1S1] Close
	ORA I 2015/05/23 16:54:44.666451 stmt.go:78: E2S1S1S1] Close
	ORA I 2015/05/23 16:54:44.666451 rset.go:57: E2S1S1S1R0] close
	ORA I 2015/05/23 16:54:44.666451 srv.go:63: E2S1] Close
	ORA I 2015/05/23 16:54:44.667451 env.go:68: E2] Close

Messages are prefixed with 'ORA I' for information or 'ORA E'
for an error. The log package is configured to write to os.Stderr by default.
Use the ora/lg.Std type to configure an alternative io.Writer.

To use the glog package:

	import (
		"flag"
		"gopkg.in/rana/ora.v4"
		"gopkg.in/rana/ora.v4/glg"
	)

	func main() {

		// parse flags for glog (required)
		// consider specifying cmd line arg -alsologtostderr=true
		flag.Parse()

		// use the optional glog package for ora logging
		cfg := ora.Cfg()
		cfg.Log.Logger = glg.Log
		ora.SetCfg(cfg)
	}

which produces a sample log of:

	I0523 17:31:41.702365   97708 drv.go:411] OpenEnv 1
	I0523 17:31:41.728377   97708 drv.go:411] OpenEnv 2
	I0523 17:31:41.728377   97708 env.go:115] E2] OpenSrv (dbname orcl)
	I0523 17:31:41.741390   97708 env.go:150] E2] OpenSrv (srvId 1)
	I0523 17:31:41.741390   97708 srv.go:113] E2S1] OpenSes (username test)
	I0523 17:31:41.762366   97708 ses.go:163] E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL
	I0523 17:31:41.762366   97708 rset.go:205] E2S1S1S1R0] open
	I0523 17:31:41.762366   97708 ses.go:74] E2S1S1] Close
	I0523 17:31:41.762366   97708 stmt.go:78] E2S1S1S1] Close
	I0523 17:31:41.762366   97708 rset.go:57] E2S1S1S1R0] close
	I0523 17:31:41.763365   97708 srv.go:63] E2S1] Close
	I0523 17:31:41.763365   97708 env.go:68] E2] Close

To use the log15 package:

	import (
		"gopkg.in/rana/ora.v4"
		"gopkg.in/rana/ora.v4/lg15"
	)
	func main() {
		// use the optional log15 package for ora logging
		cfg := ora.Cfg()
		cfg.Log.Logger = lg15.Log
		ora.SetCfg(cfg)
	}

which produces a sample log of:

	t=2015-05-23T17:08:32-0700 lvl=info msg="OpenEnv 1" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="OpenEnv 2" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] OpenSrv (dbname orcl)" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] OpenSrv (srvId 1)" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1] OpenSes (username test)" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1] Prep: SELECT CURRENT_TIMESTAMP FROM DUAL" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1R0] open" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1] Close" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1] Close" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1S1S1R0] close" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2S1] Close" lib=ora
	t=2015-05-23T17:08:32-0700 lvl=info msg="E2] Close" lib=ora

See https://github.com/rana/ora/tree/master/samples/lg15/main.go for sample
code which uses the log15 package.

Test Database Setup

Tests are available and require some setup. Setup varies depending on whether
the Oracle server is configured as a container database or non-container database.
It's simpler to setup a non-container database. An example for each setup is
explained.

Non-container test database setup steps:

	// 1. login to an Oracle server with SqlPlus as sysdba:
	SQLPLUS / AS SYSDBA

	// 2. create a file for the test database use
	CREATE TABLESPACE test_ts NOLOGGING DATAFILE 'test.dat' SIZE 100M AUTOEXTEND ON;

	// 3. create a test database
	CREATE USER test IDENTIFIED BY test DEFAULT TABLESPACE test_ts;

	// 4. grant permissions to the database
	GRANT CREATE SESSION, CREATE TABLE, CREATE SEQUENCE,
	CREATE PROCEDURE, UNLIMITED TABLESPACE TO test;

	// 5. increase the number allowable open cursors
	ALTER SYSTEM SET OPEN_CURSORS = 400 SCOPE=BOTH;

	// 6. create OS environment variables
	// specify your_database_name; varies based on installation; may be 'orcl'
	GO_ORA_DRV_TEST_DB = your_database_name
	GO_ORA_DRV_TEST_USERNAME = test
	GO_ORA_DRV_TEST_PASSWORD = test


Container test database setup steps:

	// 1. login to an Oracle server with SqlPlus as sysdba:
	SQLPLUS / AS SYSDBA

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

Go 1.6 introduced stricter cgo (call C from Go) rules, and introduced runtime checks.
This is good, as the possibility of C code corrupting Go code is almost completely eliminated,
but it also means a severe call overhead grow.
[Sometimes](https://groups.google.com/forum/#!topic/golang-nuts/ccMkPG6Bi5k)
this can be 22x the go 1.5.3 call time!

So if you need performance more than correctness, start your programs with
"GODEBUG=cgocheck=0" environment setting.

License

Copyright 2017 Rana Ian, Tamás Gulácsi. All rights reserved.
Use of this source code is governed by The MIT License
found in the accompanying LICENSE file.

*/
package ora // import "gopkg.in/rana/ora.v4"
