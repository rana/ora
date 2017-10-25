# ora
--
    import "gopkg.in/rana/ora.v4"

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


### Background

An Oracle database may be accessed through the
[database/sql](http://golang.org/pkg/database/sql) package or through the ora
package directly. database/sql offers connection pooling, thread safety, a
consistent API to multiple database technologies and a common set of Go types.
The ora package offers additional features including pointers, slices, nullable
types, numerics of various sizes, Oracle-specific types, Go return type
configuration, and Oracle abstractions such as environment, server and session.

The ora package is written with the Oracle Call Interface (OCI) C-language
libraries provided by Oracle. The OCI libraries are a standard for client
application communication and driver communication with Oracle databases.

The ora package has been verified to work with:

* Oracle Standard 11g (11.2.0.4.0), Linux x86_64 (RHEL6)

* Oracle Enterprise 12c (12.1.0.1.0), Windows 8.1 and AMD64.

### ---

* [Installation](https://github.com/rana/ora#installation)

* [Data Types](https://github.com/rana/ora#data-types)

* [SQL Placeholder Syntax](https://github.com/rana/ora#sql-placeholder-syntax)

* [Working With The Sql
Package](https://github.com/rana/ora#working-with-the-sql-package)

* [Working With The Oracle Package
Directly](https://github.com/rana/ora#working-with-the-oracle-package-directly)

* [Logging](https://github.com/rana/ora#logging)

* [Test Database Setup](https://github.com/rana/ora#test-database-setup)

* [Limitations](https://github.com/rana/ora#limitations)

* [License](https://github.com/rana/ora#license)

* [API Reference](http://godoc.org/github.com/rana/ora#pkg-index)

* [Examples](./examples)

### ---


### Installation

Minimum requirements are Go 1.3 with CGO enabled, a GCC C compiler, and Oracle
11g (11.2.0.4.0) or Oracle Instant Client (11.2.0.4.0).

Install Oracle or Oracle Instant Client.

Copy the [oci8.pc](contrib/oci8.pc) from the `contrib` folder (or the one for
your system, maybe tailored to your specific locations) to a folder in
`$PKG_CONFIG_PATH` or a system folder, such as

    cp -aL contrib/oci8.pc /usr/local/lib/pkgconfig/oci8.pc

The ora package has no external Go dependencies and is available on GitHub and
gopkg.in:

    go get gopkg.in/rana/ora.v4

*WARNING*: If you have Oracle Instant Client 11.2, you'll need to add "=lnnz11"
to the list of linked libs! Otherwise, you may encounter "undefined reference to
`nzosSCSP_SetCertSelectionParams' " errors. Oracle Instant Client 12.1 does not
need this.


### Data Types

The ora package supports all built-in Oracle data types. The supported Oracle
built-in data types are NUMBER, BINARY_DOUBLE, BINARY_FLOAT, FLOAT, DATE,
TIMESTAMP, TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIME ZONE, INTERVAL
YEAR TO MONTH, INTERVAL DAY TO SECOND, CHAR, NCHAR, VARCHAR, VARCHAR2,
NVARCHAR2, LONG, CLOB, NCLOB, BLOB, LONG RAW, RAW, ROWID and BFILE.
SYS_REFCURSOR is also supported.

Oracle does not provide a built-in boolean type. Oracle provides a single-byte
character type. A common practice is to define two single-byte characters which
represent true and false. The ora package adopts this approach. The oracle
package associates a Go bool value to a Go rune and sends and receives the rune
to a CHAR(1 BYTE) column or CHAR(1 CHAR) column.

The default false rune is zero '0'. The default true rune is one '1'. The bool
rune association may be configured or disabled when directly using the ora
package but not with the database/sql package.


### SQL Placeholder Syntax

Within a SQL string a placeholder may be specified to indicate where a Go
variable is placed. The SQL placeholder is an Oracle identifier, from 1 to 30
characters, prefixed with a colon (:). For example:

    // example Oracle placeholder uses a colon
    INSERT INTO T1 (C1) VALUES (:C1)

Placeholders within a SQL statement are bound by position. The actual name is
not used by the ora package driver e.g., placeholder names :c1, :1, or :xyz are
treated equally.


### LastInsertId

The `database/sql` package provides a LastInsertId method to return the last
inserted row's id. Oracle does not provide such functionality, but if you append
`... RETURNING col /*LastInsertId*/` to your SQL, then it will be presented as
LastInsertId. Note that you have to mark with a `/*LastInsertId*/` (case
insensitive) your `RETURNING` part, to allow ora to return the last column as
`LastInsertId()`. That column must fit in `int64`, though!


### Working With The Sql Package

You may access an Oracle database through the database/sql package. The
database/sql package offers a consistent API across different databases,
connection pooling, thread safety and a set of common Go types. database/sql
makes working with Oracle straight-forward.

The ora package implements interfaces in the database/sql/driver package
enabling database/sql to communicate with an Oracle database. Using database/sql
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
call ora.SetCfg to set the used configuration options including statement
configuration and Rset configuration.

        func init() {
    		drvCfg := ora.Cfg()
    		drvCfg.FalseRune = 'N'
    		drvCfg.TrueRune = 'Y'
    		drvCfg.TrueRune = 'Y'
    		ora.SetCfg(drvCfg)
    	}

When configuring the driver for use with database/sql, keep in mind that
database/sql has strict Go type-to-Oracle type mapping expectations.


### Working With The Oracle Package Directly

The ora package allows programming with pointers, slices, nullable types,
numerics of various sizes, Oracle-specific types, Go return type configuration,
and Oracle abstractions such as environment, server and session. When working
with the ora package directly, the API is slightly different than database/sql.

When using the ora package directly, the mapping between Go types and Oracle
types may be changed. The Go-to-Oracle type mapping for the ora package is:

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

Pointers may be used to capture out-bound values from a SQL statement such as an
insert or stored procedure call. For example, a numeric pointer captures an
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
IntervalYM, IntervalDS, String, Bool, Binary and Bfile. For example, you may
insert nullable Strings and select nullable Strings:

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

The `Stmt.Prep` method is variadic accepting zero or more `GoColumnType` which
define a Go return type for a select-list column. For example, a Prep call can
be configured to return an int64 and a nullable Int64 from the same column:

    // given: create table t1 (c1 number)
    stmt, err = ses.Prep("SELECT C1, C1 FROM T1", ora.I64, ora.OraI64)
    rset, err := stmt.Qry()
    for rset.Next() {
    	fmt.Println(rset.Row[0], rset.Row[1])
    }

Go numerics of various sizes are supported in DML operations. The ora package
supports int64, int32, int16, int8, uint64, uint32, uint16, uint8, float64 and
float32. For example, you may insert a uint16 and select numerics of various
sizes:

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

When Stmt.Prep doesn't receive a GoColumnType, or receives an incorrect
GoColumnType, the default value defined in RsetCfg is used.

EnvCfg, SrvCfg, SesCfg, StmtCfg and RsetCfg are the main configuration structs.
EnvCfg configures aspects of an Env. SrvCfg configures aspects of a Srv. SesCfg
configures aspects of a Ses. StmtCfg configures aspects of a Stmt. RsetCfg
configures aspects of Rset. StmtCfg and RsetCfg have the most options to
configure. RsetCfg defines the default mapping between an Oracle select-list
column and a Go type. StmtCfg may be set in an EnvCfg, SrvCfg, SesCfg and
StmtCfg. RsetCfg may be set in a Stmt.

EnvCfg.StmtCfg, SrvCfg.StmtCfg, SesCfg.StmtCfg may optionally be specified to
configure a statement. If StmtCfg isn't specified default values are applied.
EnvCfg.StmtCfg, SrvCfg.StmtCfg, SesCfg.StmtCfg cascade to new descendent
structs. When ora.OpenEnv() is called a specified EnvCfg is used or a default
EnvCfg is created. Creating a Srv with env.OpenSrv() will use SrvCfg.StmtCfg if
it is specified; otherwise, EnvCfg.StmtCfg is copied by value to SrvCfg.StmtCfg.
Creating a Ses with srv.OpenSes() will use SesCfg.StmtCfg if it is specified;
otherwise, SrvCfg.StmtCfg is copied by value to SesCfg.StmtCfg. Creating a Stmt
with ses.Prep() will use SesCfg.StmtCfg if it is specified; otherwise, a new
StmtCfg with default values is set on the Stmt. Call Stmt.Cfg() to change a
Stmt's configuration.

An Env may contain multiple Srv. A Srv may contain multiple Ses. A Ses may
contain multiple Stmt. A Stmt may contain multiple Rset.

    // StmtCfg cascades to descendent structs
    // EnvCfg -> SrvCfg -> SesCfg -> StmtCfg -> RsetCfg

Setting a RsetCfg on a StmtCfg does not cascade through descendent structs.
Configuration of Stmt.Cfg takes effect prior to calls to Stmt.Exe and Stmt.Qry;
consequently, any updates to Stmt.Cfg after a call to Stmt.Exe or Stmt.Qry are
not observed.

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
And ora.Bfile represents an Oracle BFILE. ROWID columns are returned as strings
and don't have a unique Go type.

#### LOBs

The default for SELECTing [BC]LOB columns is a safe Bin or S, which means all
the contents of the LOB is slurped into memory and returned as a []byte or
string.

The DefaultLOBFetchLen says LOBs are prefetched only a minimal way, to minimize
extra memory usage - you can override this using
`stmt.SetCfg(stmt.Cfg().SetLOBFetchLen(100))`.

If you want more control, you can use ora.L in Prep, Qry or
`ses.SetCfg(ses.Cfg().SetBlob(ora.L))`. But keep in mind that Oracle restricts
the use of LOBs: it is forbidden to do ANYTHING while reading the LOB! No
another query, no exec, no close of the Rset - even *advance* to the next record
in the result set is forbidden!

Failing to adhere these rules results in "Invalid handle" and ORA-03127 errors.

You cannot start reading another LOB till you haven't finished reading the
previous LOB, not even in the same row! Failing this results in ORA-24804!

For examples, see [z_lob_test.go](z_lob_test.go).

#### Rset

Rset is used to obtain Go values from a SQL select statement. Methods Rset.Next,
Rset.NextRow, and Rset.Len are available. Fields Rset.Row, Rset.Err, Rset.Index,
and Rset.ColumnNames are also available. The Next method attempts to load data
from an Oracle buffer into Row, returning true when successful. When no data is
available, or if an error occurs, Next returns false setting Row to nil. Any
error in Next is assigned to Err. Calling Next increments Index and method Len
returns the total number of rows processed. The NextRow method is convenient for
returning a single row. NextRow calls Next and returns Row. ColumnNames returns
the names of columns defined by the SQL select statement.

Rset has two usages. Rset may be returned from Stmt.Qry when prepared with a SQL
select statement:

    // given: CREATE TABLE T1 (C1 NUMBER, C2, CHAR(1 BYTE), C3 VARCHAR2(48 CHAR))
    stmt, err = ses.Prep("SELECT C1, C2, C3 FROM T1")
    rset, err := stmt.Qry()
    for rset.Next() {
    	fmt.Println(rset.Index, rset.Row[0], rset.Row[1], rset.Row[2])
    }

Or, *Rset may be passed to Stmt.Exe when prepared with a stored procedure
accepting an OUT SYS_REFCURSOR parameter:

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

Stored procedures with multiple OUT SYS_REFCURSOR parameters enable a single Exe
call to obtain multiple Rsets:

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

The types of values assigned to Row may be configured in StmtCfg.Rset. For
configuration to take effect, assign StmtCfg.Rset prior to calling Stmt.Qry or
Stmt.Exe.

Rset prefetching may be controlled by StmtCfg.PrefetchRowCount and
StmtCfg.PrefetchMemorySize. PrefetchRowCount works in coordination with
PrefetchMemorySize. When PrefetchRowCount is set to zero only PrefetchMemorySize
is used; otherwise, the minimum of PrefetchRowCount and PrefetchMemorySize is
used. The default uses a PrefetchMemorySize of 134MB.

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

Ses.Ins composes, prepares and executes a sql INSERT statement. Ses.Ins is
useful when you have to create and maintain a simple INSERT statement with a
long list of columns. As table columns are added and dropped over the lifetime
of a table Ses.Ins is easy to read and revise.

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

Ses.Upd composes, prepares and executes a sql UPDATE statement. Ses.Upd is
useful when you have to create and maintain a simple UPDATE statement with a
long list of columns. As table columns are added and dropped over the lifetime
of a table Ses.Upd is easy to read and revise.

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
when you have to create and maintain a simple SELECT statement with a long list
of columns that have non-default GoColumnTypes. As table columns are added and
dropped over the lifetime of a table Ses.Sel is easy to read and revise.

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

The Ses.Ping method checks whether the client's connection to an Oracle server
is valid. A call to Ping requires an open Ses. Ping will return a nil error when
the connection is fine:

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

Further code examples are available in the [example
file](https://github.com/rana/ora/blob/master/z_example_test.go), test files and
[samples folder](https://github.com/rana/ora/tree/master/samples).


### Logging

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

Messages are prefixed with 'ORA I' for information or 'ORA E' for an error. The
log package is configured to write to os.Stderr by default. Use the ora/lg.Std
type to configure an alternative io.Writer.

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

See https://github.com/rana/ora/tree/master/samples/lg15/main.go for sample code
which uses the log15 package.


### Test Database Setup

Tests are available and require some setup. Setup varies depending on whether
the Oracle server is configured as a container database or non-container
database. It's simpler to setup a non-container database. An example for each
setup is explained.

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


### Limitations

database/sql method Stmt.QueryRow is not supported.

Go 1.6 introduced stricter cgo (call C from Go) rules, and introduced runtime
checks. This is good, as the possibility of C code corrupting Go code is almost
completely eliminated, but it also means a severe call overhead grow.
[Sometimes](https://groups.google.com/forum/#!topic/golang-nuts/ccMkPG6Bi5k)
this can be 22x the go 1.5.3 call time!

So if you need performance more than correctness, start your programs with
"GODEBUG=cgocheck=0" environment setting.


### License

Copyright 2017 Rana Ian, Tamás Gulácsi. All rights reserved. Use of this source
code is governed by The MIT License found in the accompanying LICENSE file.

## Usage

```go
const (
	// The driver name registered with the database/sql package.
	Name string = "ora"

	// The driver version sent to an Oracle server and visible in
	// V$SESSION_CONNECT_INFO or GV$SESSION_CONNECT_INFO.
	Version string = "v4.1.13"
)
```

```go
const (
	NoPool  = PoolType(0)
	DRCPool = PoolType(1)
	SPool   = PoolType(2)
	CPool   = PoolType(3)
)
```

```go
const (
	DefaultPoolSize      = 4
	DefaultEvictDuration = time.Minute
)
```

```go
const (
	MaxFetchLen        = 1024
	DefaultFetchLen    = 128
	DefaultLOBFetchLen = 8
)
```

```go
const (
	// SysDefault is the default, normal session mode.
	SysDefault = SessionMode(iota)
	// SysDba is for connecting as SYSDBA.
	SysDba
	// SysOper is for connectiong as SYSOPER.
	SysOper
)
```

```go
var Schema string
```
Schema may optionally be specified to prefix a table name in the sql generated
by the ora.Ins, ora.Upd, ora.Del, and ora.Sel methods.

#### func  AddTbl

```go
func AddTbl(v interface{}, tblName string) (err error)
```
AddTbl maps a table name to a struct type when a struct type name is not
identitcal to an Oracle table name.

AddTbl is optional and used by the orm-like methods ora.Ins, ora.Upd, ora.Del,
and ora.Sel.

AddTbl may be called once during the lifetime of the driver.

#### func  Del

```go
func Del(v interface{}, ses *Ses) (err error)
```
Del deletes a struct from an Oracle table returning a possible error.

Specify a struct, or struct pointer to parameter 'v' and an open Ses to
parameter 'ses'.

Del requires one struct field tagged with `db:"pk"`. The field tagged with
`db:"pk"` is used in a sql WHERE clause.

By default, Del generates and executes a sql DELETE statement based on the
struct name and one exported field name tagged with `db:"pk"`. A struct name is
used for the table name and a field name is used for a column name. Prior to
calling Del, you may specify an alternative table name to ora.AddTbl. An
alternative column name may be specified to the field tag `db:"column_name"`.

Set ora.Schema to specify an optional table name prefix.

#### func  DescribeQuery

```go
func DescribeQuery(db *sql.DB, qry string) ([]DescribedColumn, error)
```
DescribeQuery parses the query and returns the column types, as
DBMS_SQL.describe_column does.

#### func  GctName

```go
func GctName(gct GoColumnType) string
```

#### func  GetCompileErrors

```go
func GetCompileErrors(ses *Ses, all bool) ([]CompileError, error)
```
GetCompileErrors returns the slice of the errors in user_errors.

If all is false, only errors are returned; otherwise, warnings, too.

#### func  Ins

```go
func Ins(v interface{}, ses *Ses) (err error)
```
Ins inserts a struct into an Oracle table returning a possible error.

Specify a struct, or struct pointer to parameter 'v' and an open Ses to
parameter 'ses'.

Optional struct field tags `db:"column_name,id,-"` may be specified to control
how the sql INSERT statement is generated.

By default, Ins generates and executes a sql INSERT statement based on the
struct name and all exported field names. A struct name is used for the table
name and a field name is used for a column name. Prior to calling Ins, you may
specify an alternative table name to ora.AddTbl. An alternative column name may
be specified to the field tag `db:"column_name"`. Specifying the `db:"-"` tag
will remove a field from the INSERT statement.

The optional `db:"id"` field tag may combined with the `db:"pk"` tag. A field
tagged with `db:"pk,id"` indicates a field is a primary key backed by an Oracle
identity sequence. `db:"pk,id"` may be tagged to one field per struct. When
`db:"pk,id"` is tagged to a field Ins generates a RETURNING clause to recevie a
db generated identity value. The `db:"id"` tag is not required and Ins will
insert a struct to a table without returning an identity value.

Set ora.Schema to specify an optional table name prefix.

#### func  NumEnv

```go
func NumEnv() int
```
NumEnv returns the number of open Oracle environments.

#### func  Register

```go
func Register(cfg DrvCfg)
```
Register used to register the ora database driver with the database/sql package,
but this is automatic now - so this function is deprecated, has the same effect
as SetCfg.

#### func  Sel

```go
func Sel(v interface{}, rt ResType, ses *Ses, where string, whereParams ...interface{}) (result interface{}, err error)
```
Sel selects structs from an Oracle table returning a specified container of
structs and a possible error.

Specify a struct, or struct pointer to parameter 'v' to indicate the struct
return type. Specify a ResType to parameter 'rt' to indicate the container
return type. Possible container return types include a slice of structs, slice
of struct pointers, map of structs, and map of struct pointers. Specify an open
Ses to parameter 'ses'. Optionally specify a where clause to parameter 'where'
and where parameters to variadic parameter 'whereParams'.

Optional struct field tags `db:"column_name,omit"` may be specified to control
how the sql SELECT statement is generated. Optional struct field tags
`db:"pk,fk1,fk2,fk3,fk4"` control how a map return type is generated.

A slice may be returned by specifying one of the 'SliceOf' ResTypes to parameter
'rt'. Specify a SliceOfPtr to return a slice of struct pointers. Specify a
SliceOfVal to return a slice of structs.

A map may be returned by specifying one of the 'MapOf' ResTypes to parameter
'rt'. The map key type is based on a struct field type tagged with one of
`db:"pk"`, `db:"fk1"`, `db:"fk2"`, `db:"fk3"`, or `db:"fk4"` matching the
specified ResType suffix Pk, Fk1, Fk2, Fk3, or Fk4. The map value type is a
struct pointer when a 'MapOfPtr' ResType is specified. The map value type is a
struct when a 'MapOfVal' ResType is specified. For example, tagging a uint64
struct field with `db:"pk"` and specifying a MapOfPtrPk generates a map with a
key type of uint64 and a value type of struct pointer.

ResTypes available to specify to parameter 'rt' are MapOfPtrPk, MapOfPtrFk1,
MapOfPtrFk2, MapOfPtrFk3, MapOfPtrFk4, MapOfValPk, MapOfValFk1, MapOfValFk2,
MapOfValFk3, and MapOfValFk4.

Set ora.Schema to specify an optional table name prefix.

#### func  SetCfg

```go
func SetCfg(cfg DrvCfg)
```
SetCfg applies the specified cfg to the ora database driver.

#### func  SplitDSN

```go
func SplitDSN(dsn string) (username, password, sid string)
```
SplitDSN splits the user/password@dblink string to username, password and
dblink, to be used as SesCfg.Username, SesCfg.Password, SrvCfg.Dblink.

#### func  Upd

```go
func Upd(v interface{}, ses *Ses) (err error)
```
Upd updates a struct to an Oracle table returning a possible error.

Specify a struct, or struct pointer to parameter 'v' and an open Ses to
parameter 'ses'.

Upd requires one struct field tagged with `db:"pk"`. The field tagged with
`db:"pk"` is used in a sql WHERE clause. Optional struct field tags
`db:"column_name,-"` may be specified to control how the sql UPDATE statement is
generated.

By default, Upd generates and executes a sql UPDATE statement based on the
struct name and all exported field names. A struct name is used for the table
name and a field name is used for a column name. Prior to calling Upd, you may
specify an alternative table name to ora.AddTbl. An alternative column name may
be specified to the field tag `db:"column_name"`. Specifying the `db:"-"` tag
will remove a field from the UPDATE statement.

Set ora.Schema to specify an optional table name prefix.

#### func  WithStmtCfg

```go
func WithStmtCfg(ctx context.Context, cfg StmtCfg) context.Context
```
WithStmtCfg returns a new context, with the given cfg that can be used to
configure several parameters.

WARNING: the StmtCfg must be derived from Cfg(), or NewStmtCfg(), as an empty
StmtCfg is not usable!

#### type Bfile

```go
type Bfile struct {
	IsNull         bool
	DirectoryAlias string
	Filename       string
}
```

Bfile represents a nullable BFILE Oracle value.

#### func (Bfile) Equals

```go
func (this Bfile) Equals(other Bfile) bool
```
Equals returns true when the receiver and specified Bfile are both null, or when
the receiver and specified Bfile are both not null, DirectoryAlias are equal and
Filename are equal.

#### type Bool

```go
type Bool struct {
	IsNull bool
	Value  bool
}
```

Bool is a nullable bool.

#### func (Bool) Equals

```go
func (this Bool) Equals(other Bool) bool
```
Equals returns true when the receiver and specified Bool are both null, or when
the receiver and specified Bool are both not null and Values are equal.

#### func (Bool) MarshalJSON

```go
func (this Bool) MarshalJSON() ([]byte, error)
```

#### func (*Bool) UnmarshalJSON

```go
func (this *Bool) UnmarshalJSON(p []byte) error
```

#### type Column

```go
type Column struct {
	Name      string
	Type      C.ub2
	Length    uint32
	Precision C.sb2
	Scale     C.sb1
}
```


#### type CompileError

```go
type CompileError struct {
	Owner, Name, Type    string
	Line, Position, Code int64
	Text                 string
	Warning              bool
}
```

CompileError represents a compile-time error as in user_errors view.

#### func (CompileError) Error

```go
func (ce CompileError) Error() string
```

#### type Con

```go
type Con struct {
}
```

Con is an Oracle connection associated with a server and session.

Implements the driver.Conn interface.

#### func (*Con) Begin

```go
func (con *Con) Begin() (driver.Tx, error)
```
Begin starts a transaction.

Begin is a member of the driver.Conn interface.

#### func (*Con) BeginTx

```go
func (con *Con) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error)
```
BeginTx starts and returns a new transaction. The provided context should be
used to roll the transaction back if it is cancelled.

If the driver does not support setting the isolation level and one is set or if
there is a set isolation level but the set level is not supported, an error must
be returned.

If the read-only value is true to either set the read-only transaction property
if supported or return an error if it is not supported.

#### func (*Con) Close

```go
func (con *Con) Close() (err error)
```
Close ends a session and disconnects from an Oracle server.

Close is a member of the driver.Conn interface.

#### func (*Con) IsOpen

```go
func (con *Con) IsOpen() bool
```
IsOpen returns true when the connection to the Oracle server is open; otherwise,
false.

Calling Close will cause IsOpen to return false. Once closed, a connection
cannot be re-opened. To open a new connection call Open on a driver.

#### func (*Con) Name

```go
func (s *Con) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Con) Ping

```go
func (con *Con) Ping(ctx context.Context) error
```
Ping makes a round-trip call to an Oracle server to confirm that the connection
is active.

#### func (*Con) Prepare

```go
func (con *Con) Prepare(query string) (driver.Stmt, error)
```
Prepare readies a sql string for use.

Prepare is a member of the driver.Conn interface.

#### func (*Con) PrepareContext

```go
func (con *Con) PrepareContext(ctx context.Context, query string) (driver.Stmt, error)
```
PrepareContext returns a prepared statement, bound to this connection. context
is for the preparation of the statement, it must not store the context within
the statement itself.

#### type Date

```go
type Date struct {
	date.Date
}
```

Date is a nullable date, for low (second) precisions (OCIDate)

#### type DescribedColumn

```go
type DescribedColumn struct {
	Column

	Schema                 string
	Nullable               bool
	CharsetID, CharsetForm int
}
```

DescribedColumn type for describing a column (see DescribeQuery).

#### type Drv

```go
type Drv struct {
	sync.RWMutex
}
```

Drv represents an Oracle database driver.

Drv is not meant to be called by user-code.

Drv implements the driver.Driver interface.

#### func (*Drv) Cfg

```go
func (drv *Drv) Cfg() DrvCfg
```

#### func (*Drv) Open

```go
func (drv *Drv) Open(conStr string) (driver.Conn, error)
```
Open opens a connection to an Oracle server with the database/sql environment.

This is intended to be called by the database/sql package only.

Alternatively, you may call Env.OpenCon to create an *ora.Con.

Open is a member of the driver.Driver interface.

#### func (*Drv) SetCfg

```go
func (drv *Drv) SetCfg(cfg DrvCfg)
```

#### type DrvCfg

```go
type DrvCfg struct {
	StmtCfg
	Log LogDrvCfg
}
```

DrvCfg represents configuration values for the ora package.

#### func  Cfg

```go
func Cfg() DrvCfg
```
Cfg returns the ora database driver's cfg.

#### func  NewDrvCfg

```go
func NewDrvCfg() DrvCfg
```
NewDrvCfg creates a DrvCfg with default values.

#### func (DrvCfg) SetBinaryDouble

```go
func (c DrvCfg) SetBinaryDouble(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetBinaryFloat

```go
func (c DrvCfg) SetBinaryFloat(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetBlob

```go
func (c DrvCfg) SetBlob(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetByteSlice

```go
func (c DrvCfg) SetByteSlice(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetChar

```go
func (c DrvCfg) SetChar(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetChar1

```go
func (c DrvCfg) SetChar1(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetClob

```go
func (c DrvCfg) SetClob(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetDate

```go
func (c DrvCfg) SetDate(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetFloat

```go
func (c DrvCfg) SetFloat(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetLobBufferSize

```go
func (c DrvCfg) SetLobBufferSize(size int) DrvCfg
```

#### func (DrvCfg) SetLogger

```go
func (c DrvCfg) SetLogger(lgr Logger) DrvCfg
```

#### func (DrvCfg) SetLong

```go
func (c DrvCfg) SetLong(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetLongBufferSize

```go
func (c DrvCfg) SetLongBufferSize(size uint32) DrvCfg
```

#### func (DrvCfg) SetLongRaw

```go
func (c DrvCfg) SetLongRaw(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetLongRawBufferSize

```go
func (c DrvCfg) SetLongRawBufferSize(size uint32) DrvCfg
```

#### func (DrvCfg) SetNumberBigFloat

```go
func (c DrvCfg) SetNumberBigFloat(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetNumberBigInt

```go
func (c DrvCfg) SetNumberBigInt(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetNumberFloat

```go
func (c DrvCfg) SetNumberFloat(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetNumberInt

```go
func (c DrvCfg) SetNumberInt(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetPrefetchMemorySize

```go
func (c DrvCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) DrvCfg
```

#### func (DrvCfg) SetPrefetchRowCount

```go
func (c DrvCfg) SetPrefetchRowCount(prefetchRowCount uint32) DrvCfg
```

#### func (DrvCfg) SetRaw

```go
func (c DrvCfg) SetRaw(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetStmtCfg

```go
func (cfg DrvCfg) SetStmtCfg(stmtCfg StmtCfg) DrvCfg
```

#### func (DrvCfg) SetStringPtrBufferSize

```go
func (c DrvCfg) SetStringPtrBufferSize(size int) DrvCfg
```

#### func (DrvCfg) SetTimestamp

```go
func (c DrvCfg) SetTimestamp(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetTimestampLtz

```go
func (c DrvCfg) SetTimestampLtz(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetTimestampTz

```go
func (c DrvCfg) SetTimestampTz(gct GoColumnType) DrvCfg
```

#### func (DrvCfg) SetVarchar

```go
func (c DrvCfg) SetVarchar(gct GoColumnType) DrvCfg
```

#### type DrvExecResult

```go
type DrvExecResult struct {
}
```

DrvExecResult is an Oracle execution result.

DrvExecResult implements the driver.Result interface.

#### func (*DrvExecResult) LastInsertId

```go
func (er *DrvExecResult) LastInsertId() (int64, error)
```
LastInsertId returns the identity value from an insert statement.

There are two setup steps required to reteive the LastInsertId. One, specify a
'returning into' clause in the SQL insert statement. And, two, specify a nil
parameter to DB.Exec or DrvStmt.Exec.

For example:

    db, err := sql.Open("ora", "scott/tiger@orcl")

    db.Exec("CREATE TABLE T1 (C1 NUMBER(19,0) GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1), C2 VARCHAR2(48 CHAR))")

    result, err := db.Exec("INSERT INTO T1 (C2) VALUES ('GO') RETURNING C1 /*lastInsertId*/ INTO :C1", nil)

    id, err := result.LastInsertId()

#### func (*DrvExecResult) RowsAffected

```go
func (er *DrvExecResult) RowsAffected() (int64, error)
```
RowsAffected returns the number of rows affected by the exec statement.

#### type DrvQueryResult

```go
type DrvQueryResult struct {
}
```

DrvQueryResult contains methods to retrieve the results of a SQL select
statement.

DrvQueryResult implements the driver.Rows interface.

#### func (*DrvQueryResult) Close

```go
func (qr *DrvQueryResult) Close() error
```
Close performs no operations.

Close is a member of the driver.Rows interface.

#### func (*DrvQueryResult) ColumnTypeDatabaseTypeName

```go
func (qr *DrvQueryResult) ColumnTypeDatabaseTypeName(index int) string
```
ColumnTypeDatabaseTypeName returns the database system type name without the
length, in uppercase.

#### func (*DrvQueryResult) ColumnTypeLength

```go
func (qr *DrvQueryResult) ColumnTypeLength(index int) (length int64, ok bool)
```
ColumnTypeLength returns the length of the column type if the column is a
variable length type. If the column is not a variable length type ok should
return false. If length is not limited other than system limits, it should
return math.MaxInt64.

#### func (*DrvQueryResult) ColumnTypeNullable

```go
func (qr *DrvQueryResult) ColumnTypeNullable(index int) (nullable, ok bool)
```
ColumnTypeNullable returns true if it is known the column may be null, or false
if the column is known to be not nullable. If the column nullability is unknown,
ok should be false.

#### func (*DrvQueryResult) ColumnTypePrecisionScale

```go
func (qr *DrvQueryResult) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool)
```
ColumnTypePrecisionScale return the precision and scale for decimal types. If
not applicable, ok should be false.

#### func (*DrvQueryResult) ColumnTypeScanType

```go
func (qr *DrvQueryResult) ColumnTypeScanType(index int) reflect.Type
```

#### func (*DrvQueryResult) Columns

```go
func (qr *DrvQueryResult) Columns() []string
```
Columns returns query column names.

Columns is a member of the driver.Rows interface.

#### func (*DrvQueryResult) HasNextResultSet

```go
func (qr *DrvQueryResult) HasNextResultSet() bool
```
HasNextResultSet reports whether there is another result set after the current
one.

#### func (*DrvQueryResult) Next

```go
func (qr *DrvQueryResult) Next(dest []driver.Value) (err error)
```
Next populates the specified slice with the next row of data.

Returns io.EOF when there are no more rows.

Next is a member of the driver.Rows interface.

#### func (*DrvQueryResult) NextResultSet

```go
func (qr *DrvQueryResult) NextResultSet() error
```
NextResultSet advances the driver to the next result set even if there are
remaining rows in the current result set.

#### type DrvStmt

```go
type DrvStmt struct {
}
```

DrvStmt is an Oracle statement associated with a session.

DrvStmt wraps Stmt and is intended for use by the database/sql/driver package.

DrvStmt implements the driver.Stmt interface.

#### func (*DrvStmt) Close

```go
func (ds *DrvStmt) Close() error
```
Close closes the SQL statement.

Close is a member of the driver.Stmt interface.

#### func (*DrvStmt) Exec

```go
func (ds *DrvStmt) Exec(values []driver.Value) (driver.Result, error)
```
Exec executes an Oracle SQL statement on a server. Exec returns a driver.Result
and a possible error.

Exec is a member of the driver.Stmt interface.

#### func (*DrvStmt) ExecContext

```go
func (ds *DrvStmt) ExecContext(ctx context.Context, values []driver.NamedValue) (driver.Result, error)
```
ExecContext enhances the Stmt interface by providing Exec with context.
ExecContext must honor the context timeout and return when it is cancelled.

#### func (*DrvStmt) NumInput

```go
func (ds *DrvStmt) NumInput() int
```
NumInput returns the number of placeholders in a sql statement.

NumInput is a member of the driver.Stmt interface.

#### func (*DrvStmt) Query

```go
func (ds *DrvStmt) Query(values []driver.Value) (driver.Rows, error)
```
Query runs a SQL query on an Oracle server. Query returns driver.Rows and a
possible error.

Query is a member of the driver.Stmt interface.

#### func (*DrvStmt) QueryContext

```go
func (ds *DrvStmt) QueryContext(ctx context.Context, values []driver.NamedValue) (driver.Rows, error)
```
QueryContext enhances the Stmt interface by providing Query with context.
QueryContext must honor the context timeout and return when it is cancelled.

#### type EmpLgr

```go
type EmpLgr struct{}
```


#### func (EmpLgr) Errorf

```go
func (e EmpLgr) Errorf(format string, v ...interface{})
```

#### func (EmpLgr) Errorln

```go
func (e EmpLgr) Errorln(v ...interface{})
```

#### func (EmpLgr) Infof

```go
func (e EmpLgr) Infof(format string, v ...interface{})
```

#### func (EmpLgr) Infoln

```go
func (e EmpLgr) Infoln(v ...interface{})
```

#### type Env

```go
type Env struct {
	sync.RWMutex
}
```

Env represents an Oracle environment.

#### func  NewEnvSrvSes

```go
func NewEnvSrvSes(dsn string) (*Env, *Srv, *Ses, error)
```
NewEnvSrvSes is a comfort function which opens the environment, creates a
connection (Srv) to the server, and opens a session (Ses), in one call.

Ideal for simple use cases.

#### func  OpenEnv

```go
func OpenEnv() (env *Env, err error)
```
OpenEnv opens an Oracle environment.

Optionally specify a cfg parameter. If cfg is nil, default cfg values are
applied.

#### func (*Env) Cfg

```go
func (env *Env) Cfg() StmtCfg
```

#### func (*Env) Close

```go
func (env *Env) Close() (err error)
```
Close disconnects from servers and resets optional fields.

#### func (*Env) IsOpen

```go
func (env *Env) IsOpen() bool
```
IsOpen returns true when the environment is open; otherwise, false.

Calling Close will cause IsOpen to return false. Once closed, the environment
may be re-opened by calling Open.

#### func (*Env) Name

```go
func (s *Env) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Env) NewPool

```go
func (env *Env) NewPool(srvCfg SrvCfg, sesCfg SesCfg, size int) *Pool
```
NewPool returns an idle session pool, which evicts the idle sessions every
minute, and automatically manages the required new connections (Srv).

This is done by maintaining a 1-1 pairing between the Srv and its Ses.

This pool does NOT limit the number of active connections, just helps reuse
already established connections and sessions, lowering the resource usage on the
server.

If size <= 0, then DefaultPoolSize is used.

#### func (*Env) NewSrvPool

```go
func (env *Env) NewSrvPool(srvCfg SrvCfg, size int) *SrvPool
```
NewSrvPool returns a connection pool, which evicts the idle connections in every
minute. The pool holds at most size idle Srv. If size is zero, DefaultPoolSize
will be used.

#### func (*Env) NumCon

```go
func (env *Env) NumCon() int
```
NumCon returns the number of open Oracle connections.

#### func (*Env) NumSrv

```go
func (env *Env) NumSrv() int
```
NumSrv returns the number of open Oracle servers.

#### func (*Env) OCINumberFromFloat

```go
func (env *Env) OCINumberFromFloat(dest *C.OCINumber, value float64, byteLen int) error
```

#### func (*Env) OCINumberFromInt

```go
func (env *Env) OCINumberFromInt(dest *C.OCINumber, value int64, byteLen int) error
```

#### func (*Env) OCINumberFromUint

```go
func (env *Env) OCINumberFromUint(dest *C.OCINumber, value uint64, byteLen int) error
```

#### func (*Env) OCINumberToFloat

```go
func (env *Env) OCINumberToFloat(src *C.OCINumber, byteLen int) (float64, error)
```

#### func (*Env) OCINumberToInt

```go
func (env *Env) OCINumberToInt(src *C.OCINumber, byteLen int) (int64, error)
```

#### func (*Env) OCINumberToUint

```go
func (env *Env) OCINumberToUint(src *C.OCINumber, byteLen int) (uint64, error)
```

#### func (*Env) OpenCon

```go
func (env *Env) OpenCon(dsn string) (con *Con, err error)
```
OpenCon starts an Oracle session on a server returning a *Con and possible
error.

The connection string has the form username/password@dblink e.g.,
scott/tiger@orcl For connecting as SYSDBA or SYSOPER, append " AS SYSDBA" to the
end of the connection string: "sys/sys as sysdba".

dblink is a connection identifier such as a net service name, full connection
identifier, or a simple connection identifier. The dblink may be defined in the
client machine's tnsnames.ora file.

#### func (*Env) OpenSrv

```go
func (env *Env) OpenSrv(cfg SrvCfg) (srv *Srv, err error)
```
OpenSrv connects to an Oracle server returning a *Srv and possible error.

#### func (*Env) SetCfg

```go
func (env *Env) SetCfg(cfg StmtCfg)
```

#### type Float32

```go
type Float32 struct {
	IsNull bool
	Value  float32
}
```

Float32 is a nullable float32.

#### func (Float32) Equals

```go
func (this Float32) Equals(other Float32) bool
```
Equals returns true when the receiver and specified Float32 are both null, or
when the receiver and specified Float32 are both not null and Values are equal.

#### func (Float32) MarshalJSON

```go
func (this Float32) MarshalJSON() ([]byte, error)
```

#### func (*Float32) UnmarshalJSON

```go
func (this *Float32) UnmarshalJSON(p []byte) error
```

#### type Float64

```go
type Float64 struct {
	IsNull bool
	Value  float64
}
```

Float64 is a nullable float64.

#### func (Float64) Equals

```go
func (this Float64) Equals(other Float64) bool
```
Equals returns true when the receiver and specified Float64 are both null, or
when the receiver and specified Float64 are both not null and Values are equal.

#### func (Float64) MarshalJSON

```go
func (this Float64) MarshalJSON() ([]byte, error)
```

#### func (*Float64) UnmarshalJSON

```go
func (this *Float64) UnmarshalJSON(p []byte) error
```

#### type GoColumnType

```go
type GoColumnType uint
```

GoColumnType defines the Go type returned from a sql select column.

```go
const (
	// D defines a sql select column based on its default mapping.
	D GoColumnType = iota + 1
	// I64 defines a sql select column as a Go int64.
	I64
	// I32 defines a sql select column as a Go int32.
	I32
	// I16 defines a sql select column as a Go int16.
	I16
	// I8 defines a sql select column as a Go int8.
	I8
	// U64 defines a sql select column as a Go uint64.
	U64
	// U32 defines a sql select column as a Go uint32.
	U32
	// U16 defines a sql select column as a Go uint16.
	U16
	// U8 defines a sql select column as a Go uint8.
	U8
	// F64 defines a sql select column as a Go float64.
	F64
	// F32 defines a sql select column as a Go float32.
	F32
	// OraI64 defines a sql select column as a nullable Go ora.Int64.
	OraI64
	// OraI32 defines a sql select column as a nullable Go ora.Int32.
	OraI32
	// OraI16 defines a sql select column as a nullable Go ora.Int16.
	OraI16
	// OraI8 defines a sql select column as a nullable Go ora.Int8.
	OraI8
	// OraU64 defines a sql select column as a nullable Go ora.Uint64.
	OraU64
	// OraU32 defines a sql select column as a nullable Go ora.Uint32.
	OraU32
	// OraU16 defines a sql select column as a nullable Go ora.Uint16.
	OraU16
	// OraU8 defines a sql select column as a nullable Go ora.Uint8.
	OraU8
	// OraF64 defines a sql select column as a nullable Go ora.Float64.
	OraF64
	// OraF32 defines a sql select column as a nullable Go ora.Float32.
	OraF32
	// T defines a sql select column as a Go time.Time.
	T
	// OraT defines a sql select column as a nullable Go ora.Time.
	OraT
	// S defines a sql select column as a Go string.
	S
	// OraS defines a sql select column as a nullable Go ora.String.
	OraS
	// B defines a sql select column as a Go bool.
	B
	// OraB defines a sql select column as a nullable Go ora.Bool.
	OraB
	// Bin defines a sql select column or bind parmeter as a Go byte slice.
	Bin
	// OraBin defines a sql select column as a nullable Go ora.Binary.
	OraBin
	// N defines a sql select column as a Go string for number.
	N
	// OraN defines a sql select column as a nullable Go string for number.
	OraN
	// L defins an sql select column as an ora.Lob.
	L
)
```
go column types

#### func (GoColumnType) String

```go
func (gct GoColumnType) String() string
```

#### type Id

```go
type Id struct {
}
```


#### type Int16

```go
type Int16 struct {
	IsNull bool
	Value  int16
}
```

Int16 is a nullable int16.

#### func (Int16) Equals

```go
func (this Int16) Equals(other Int16) bool
```
Equals returns true when the receiver and specified Int16 are both null, or when
the receiver and specified Int16 are both not null and Values are equal.

#### func (Int16) MarshalJSON

```go
func (this Int16) MarshalJSON() ([]byte, error)
```

#### func (*Int16) UnmarshalJSON

```go
func (this *Int16) UnmarshalJSON(p []byte) error
```

#### type Int32

```go
type Int32 struct {
	IsNull bool
	Value  int32
}
```

Int32 is a nullable int32.

#### func (Int32) Equals

```go
func (this Int32) Equals(other Int32) bool
```
Equals returns true when the receiver and specified Int32 are both null, or when
the receiver and specified Int32 are both not null and Values are equal.

#### func (Int32) MarshalJSON

```go
func (this Int32) MarshalJSON() ([]byte, error)
```

#### func (*Int32) UnmarshalJSON

```go
func (this *Int32) UnmarshalJSON(p []byte) error
```

#### type Int64

```go
type Int64 struct {
	IsNull bool
	Value  int64
}
```

Int64 is a nullable int64.

#### func (Int64) Equals

```go
func (this Int64) Equals(other Int64) bool
```
Equals returns true when the receiver and specified Int64 are both null, or when
the receiver and specified Int64 are both not null and Values are equal.

#### func (Int64) MarshalJSON

```go
func (this Int64) MarshalJSON() ([]byte, error)
```

#### func (*Int64) UnmarshalJSON

```go
func (this *Int64) UnmarshalJSON(p []byte) error
```

#### type Int8

```go
type Int8 struct {
	IsNull bool
	Value  int8
}
```

Int8 is a nullable int8.

#### func (Int8) Equals

```go
func (this Int8) Equals(other Int8) bool
```
Equals returns true when the receiver and specified Int8 are both null, or when
the receiver and specified Int8 are both not null and Values are equal.

#### func (Int8) MarshalJSON

```go
func (this Int8) MarshalJSON() ([]byte, error)
```

#### func (*Int8) UnmarshalJSON

```go
func (this *Int8) UnmarshalJSON(p []byte) error
```

#### type IntervalDS

```go
type IntervalDS struct {
	IsNull     bool
	Day        int32
	Hour       int32
	Minute     int32
	Second     int32
	Nanosecond int32
}
```

IntervalDS represents a nullable INTERVAL DAY TO SECOND Oracle value.

#### func (IntervalDS) Equals

```go
func (this IntervalDS) Equals(other IntervalDS) bool
```
Equals returns true when the receiver and specified IntervalDS are both null, or
when the receiver and specified IntervalDS are both not null, and all other
fields are equal.

#### func (IntervalDS) ShiftTime

```go
func (this IntervalDS) ShiftTime(t time.Time) time.Time
```
ShiftTime returns a new Time with IntervalDS applied.

#### func (IntervalDS) String

```go
func (this IntervalDS) String() string
```

#### type IntervalYM

```go
type IntervalYM struct {
	IsNull bool
	Year   int32
	Month  int32
}
```

IntervalYM represents a nullable INTERVAL YEAR TO MONTH Oracle value.

#### func (IntervalYM) Equals

```go
func (this IntervalYM) Equals(other IntervalYM) bool
```
Equals returns true when the receiver and specified IntervalYM are both null, or
when the receiver and specified IntervalYM are both not null, Year are equal and
Month are equal.

#### func (IntervalYM) ShiftTime

```go
func (this IntervalYM) ShiftTime(t time.Time) time.Time
```
ShiftTime returns a new Time with IntervalYM applied.

#### func (IntervalYM) String

```go
func (this IntervalYM) String() string
```

#### type Lob

```go
type Lob struct {
	io.Reader
	io.Closer
	C bool
}
```

Lob Reader is sent to the DB on bind, if not nil. The Reader can read the LOB if
we bind a *Lob, Closer will close the LOB. Set Lob.C = true to make this a CLOB
reader!

#### func (*Lob) Bytes

```go
func (this *Lob) Bytes() ([]byte, error)
```
Bytes will read the contents of the Lob.Reader, and will keep that for future.

#### func (*Lob) Close

```go
func (this *Lob) Close() error
```

#### func (*Lob) Equals

```go
func (this *Lob) Equals(other Lob) bool
```
Equals returns true when the receiver and specified Lob are both null, or when
they both not null and share the same Reader.

#### func (*Lob) MarshalJSON

```go
func (this *Lob) MarshalJSON() ([]byte, error)
```

#### func (*Lob) Read

```go
func (this *Lob) Read(p []byte) (int, error)
```

#### func (*Lob) Scan

```go
func (this *Lob) Scan(src interface{}) error
```

#### func (*Lob) String

```go
func (this *Lob) String() string
```

#### func (*Lob) UnmarshalJSON

```go
func (this *Lob) UnmarshalJSON(p []byte) error
```

#### func (*Lob) Value

```go
func (this *Lob) Value() (driver.Value, error)
```
Value returns what Lob.Bytes returns.

#### type LogConCfg

```go
type LogConCfg struct {
	// Close determines whether the Con.Close method is logged.
	//
	// The default is true.
	Close bool

	// Prepare determines whether the Con.Prepare method is logged.
	//
	// The default is true.
	Prepare bool

	// Begin determines whether the Con.Begin method is logged.
	//
	// The default is true.
	Begin bool

	// Ping determines whether the Con.Ping method is logged.
	//
	// The default is true.
	Ping bool
}
```

LogConCfg represents Con logging configuration values.

#### func  NewLogConCfg

```go
func NewLogConCfg() LogConCfg
```
NewLogConCfg creates a LogTxCfg with default values.

#### type LogDrvCfg

```go
type LogDrvCfg struct {
	// Logger writes log messages.
	// Logger can be replaced with any type implementing the Logger interface.
	//
	// The default implementation uses the standard lib's log package.
	//
	// For a glog-based implementation, see gopkg.in/rana/ora.v4/glg.
	// LogDrvCfg.Logger = glg.Log
	//
	// For an gopkg.in/inconshreveable/log15.v2-based, see gopkg.in/rana/ora.v4/lg15.
	// LogDrvCfg.Logger = lg15.Log
	Logger Logger

	// OpenEnv determines whether the ora.OpenEnv method is logged.
	//
	// The default is true.
	OpenEnv bool

	// Ins determines whether the ora.Ins method is logged.
	//
	// The default is true.
	Ins bool

	// Upd determines whether the ora.Upd method is logged.
	//
	// The default is true.
	Upd bool

	// Del determines whether the ora.Del method is logged.
	//
	// The default is true.
	Del bool

	// Sel determines whether the ora.Sel method is logged.
	//
	// The default is true.
	Sel bool

	// AddTbl determines whether the ora.AddTbl method is logged.
	//
	// The default is true.
	AddTbl bool

	Env  LogEnvCfg
	Srv  LogSrvCfg
	Ses  LogSesCfg
	Stmt LogStmtCfg
	Tx   LogTxCfg
	Con  LogConCfg
	Rset LogRsetCfg
}
```

LogDrvCfg represents package-level logging configuration values.

#### func  NewLogDrvCfg

```go
func NewLogDrvCfg() LogDrvCfg
```
NewLogDrvCfg creates a LogDrvCfg with default values.

#### func (LogDrvCfg) IsEnabled

```go
func (c LogDrvCfg) IsEnabled(enabled bool) bool
```
IsEnabled returns whether the logger is enabled (and enabled is true).

#### type LogEnvCfg

```go
type LogEnvCfg struct {
	// Close determines whether the Env.Close method is logged.
	//
	// The default is true.
	Close bool

	// OpenSrv determines whether the Env.OpenSrv method is logged.
	//
	// The default is true.
	OpenSrv bool

	// OpenCon determines whether the Env.OpenCon method is logged.
	//
	// The default is true.
	OpenCon bool
}
```

LogEnvCfg represents Env logging configuration values.

#### func  NewLogEnvCfg

```go
func NewLogEnvCfg() LogEnvCfg
```
NewLogEnvCfg creates a LogEnvCfg with default values.

#### type LogRsetCfg

```go
type LogRsetCfg struct {
	// Close determines whether the Rset.close method is logged.
	//
	// The default is true.
	Close bool

	// BeginRow determines whether the Rset.beginRow method is logged.
	//
	// The default is false.
	BeginRow bool

	// EndRow determines whether the Rset.endRow method is logged.
	//
	// The default is false.
	EndRow bool

	// Next determines whether the Rset.Next method is logged.
	//
	// The default is false.
	Next bool

	// Open determines whether the Rset.open method is logged.
	//
	// The default is true.
	Open bool

	// OpenDefs determines whether Select-list definitions with the Rset.open method are logged.
	//
	// The default is true.
	OpenDefs bool
}
```

LogRsetCfg represents Rset logging configuration values.

#### func  NewLogRsetCfg

```go
func NewLogRsetCfg() LogRsetCfg
```
NewLogTxCfg creates a LogRsetCfg with default values.

#### type LogSesCfg

```go
type LogSesCfg struct {
	// Close determines whether the Ses.Close method is logged.
	//
	// The default is true.
	Close bool

	// PrepAndExe determines whether the Ses.PrepAndExe method is logged.
	//
	// The default is true.
	PrepAndExe bool

	// PrepAndQry determines whether the Ses.PrepAndQry method is logged.
	//
	// The default is true.
	PrepAndQry bool

	// Prep determines whether the Ses.Prep method is logged.
	//
	// The default is true.
	Prep bool

	// Ins determines whether the Ses.Ins method is logged.
	//
	// The default is true.
	Ins bool

	// Upd determines whether the Ses.Upd method is logged.
	//
	// The default is true.
	Upd bool

	// Sel determines whether the Ses.Sel method is logged.
	//
	// The default is true.
	Sel bool

	// StartTx determines whether the Ses.StartTx method is logged.
	//
	// The default is true.
	StartTx bool

	// Ping determines whether the Ses.Ping method is logged.
	//
	// The default is true.
	Ping bool

	// Break determines whether the Ses.Break method is logged.
	//
	// The default is true.
	Break bool
}
```

LogSesCfg represents Ses logging configuration values.

#### func  NewLogSesCfg

```go
func NewLogSesCfg() LogSesCfg
```
NewLogSesCfg creates a LogSesCfg with default values.

#### type LogSrvCfg

```go
type LogSrvCfg struct {
	// Close determines whether the Srv.Close method is logged.
	//
	// The default is true.
	Close bool

	// OpenSes determines whether the Srv.OpenSes method is logged.
	//
	// The default is true.
	OpenSes bool

	// Version determines whether the Srv.Version method is logged.
	//
	// The default is true.
	Version bool
}
```

LogSrvCfg represents Srv logging configuration values.

#### func  NewLogSrvCfg

```go
func NewLogSrvCfg() LogSrvCfg
```
NewLogSrvCfg creates a LogSrvCfg with default values.

#### type LogStmtCfg

```go
type LogStmtCfg struct {
	// Close determines whether the Stmt.Close method is logged.
	//
	// The default is true.
	Close bool

	// Exe determines whether the Stmt.Exe method is logged.
	//
	// The default is true.
	Exe bool

	// Qry determines whether the Stmt.Qry method is logged.
	//
	// The default is true.
	Qry bool

	// Bind determines whether the Stmt.bind method is logged.
	//
	// The default is true.
	Bind bool
}
```

LogStmtCfg represents Stmt logging configuration values.

#### func  NewLogStmtCfg

```go
func NewLogStmtCfg() LogStmtCfg
```
NewLogStmtCfg creates a LogStmtCfg with default values.

#### type LogTxCfg

```go
type LogTxCfg struct {
	// Commit determines whether the Tx.Commit method is logged.
	//
	// The default is true.
	Commit bool

	// Rollback determines whether the Tx.Rollback method is logged.
	//
	// The default is true.
	Rollback bool
}
```

LogTxCfg represents Tx logging configuration values.

#### func  NewLogTxCfg

```go
func NewLogTxCfg() LogTxCfg
```
NewLogTxCfg creates a LogTxCfg with default values.

#### type Logger

```go
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}
```

Logger interface is for logging.

#### type MultiErr

```go
type MultiErr struct {
}
```

MultiErr holds multiple errors in a single string.

#### func (MultiErr) Error

```go
func (m MultiErr) Error() string
```
Error returns one or more errors.

Error is a member of the 'error' interface.

#### type Num

```go
type Num string
```


#### type OCINum

```go
type OCINum struct {
	num.OCINum
}
```


#### func (*OCINum) FromC

```go
func (num *OCINum) FromC(x C.OCINumber)
```
FromC converts from the given C.OCINumber.

#### func (OCINum) String

```go
func (n OCINum) String() string
```

#### func (OCINum) ToC

```go
func (num OCINum) ToC(x *C.OCINumber)
```
ToC converts the OCINum into the given *C.OCINumber.

#### func (OCINum) Value

```go
func (n OCINum) Value() (driver.Value, error)
```
Value returns the driver.Value as required by database/sql. So OCINum is allowed
as a parameter to Scan.

#### type ORAError

```go
type ORAError struct {
}
```


#### func (ORAError) Code

```go
func (e ORAError) Code() int
```

#### func (*ORAError) Error

```go
func (e *ORAError) Error() string
```

#### type OraNum

```go
type OraNum struct {
	IsNull bool
	Value  string
}
```


#### func (OraNum) Equals

```go
func (this OraNum) Equals(other OraNum) bool
```
Equals returns true when the receiver and specified OraNum are both null, or
when the receiver and specified OraNum are both not null and Values are equal.

#### func (OraNum) MarshalJSON

```go
func (this OraNum) MarshalJSON() ([]byte, error)
```

#### func (OraNum) String

```go
func (this OraNum) String() string
```

#### func (*OraNum) UnmarshalJSON

```go
func (this *OraNum) UnmarshalJSON(p []byte) error
```

#### type OraOCINum

```go
type OraOCINum struct {
	IsNull bool
	Value  num.OCINum
}
```


#### func (OraOCINum) Equals

```go
func (this OraOCINum) Equals(other OraOCINum) bool
```
Equals returns true when the receiver and specified OraOCINum are both null, or
when the receiver and specified OraOCINum are both not null and Values are
equal.

#### func (OraOCINum) MarshalJSON

```go
func (this OraOCINum) MarshalJSON() ([]byte, error)
```

#### func (OraOCINum) String

```go
func (this OraOCINum) String() string
```

#### func (*OraOCINum) UnmarshalJSON

```go
func (this *OraOCINum) UnmarshalJSON(p []byte) error
```

#### type Pool

```go
type Pool struct {
	sync.Mutex
}
```


#### func  NewPool

```go
func NewPool(dsn string, size int) (*Pool, error)
```
NewPool returns a new session pool with default config.

#### func (*Pool) Close

```go
func (p *Pool) Close() (err error)
```
Close all idle sessions and connections.

#### func (*Pool) Get

```go
func (p *Pool) Get() (ses *Ses, err error)
```
Get a session - either an idle session, or if such does not exist, then a new
session on an idle connection; if such does not exist, then a new session on a
new connection.

#### func (*Pool) Put

```go
func (p *Pool) Put(ses *Ses)
```
Put the session back to the session pool. Ensure that on ses Close (eviction),
srv is put back on the idle pool.

#### func (Pool) SetEvictDuration

```go
func (p Pool) SetEvictDuration(dur time.Duration)
```
Set the eviction duration to the given. Also starts eviction if not yet started.

#### type PoolCfg

```go
type PoolCfg struct {
	Type           PoolType
	Name           string
	Username       string
	Password       string
	Min, Max, Incr uint32
}
```


#### func  DSNPool

```go
func DSNPool(str string) PoolCfg
```
DSNPool returns the Pool config from dsn.

#### type PoolType

```go
type PoolType uint8
```


#### type Raw

```go
type Raw struct {
	IsNull bool
	Value  []byte
}
```

Raw represents a nullable byte slice for RAW or LONG RAW Oracle values.

#### func (Raw) Equals

```go
func (this Raw) Equals(other Raw) bool
```
Equals returns true when the receiver and specified Raw are both null, or when
the receiver and specified Raw are both not null and Values are equal.

#### func (Raw) MarshalJSON

```go
func (this Raw) MarshalJSON() ([]byte, error)
```

#### func (*Raw) UnmarshalJSON

```go
func (this *Raw) UnmarshalJSON(p []byte) error
```

#### type ResType

```go
type ResType int
```

ResType represents a result type returned by the ora.Sel method.

```go
const (
	// SliceOfPtr indicates a slice of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	SliceOfPtr ResType = iota

	// SliceOfVal indicates a slice of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	SliceOfVal

	// MapOfPtrPk indicates a map of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"pk"`.
	MapOfPtrPk

	// MapOfPtrFk1 indicates a map of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk1"`.
	MapOfPtrFk1

	// MapOfPtrFk2 indicates a map of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk2"`.
	MapOfPtrFk2

	// MapOfPtrFk3 indicates a map of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk3"`.
	MapOfPtrFk3

	// MapOfPtrFk4 indicates a map of struct pointers will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk4"`.
	MapOfPtrFk4

	// MapOfValPk indicates a map of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"pk"`.
	MapOfValPk

	// MapOfValFk1 indicates a map of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk1"`.
	MapOfValFk1

	// MapOfValFk2 indicates a map of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk2"`.
	MapOfValFk2

	// MapOfValFk3 indicates a map of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk3"`.
	MapOfValFk3

	// MapOfValFk4 indicates a map of structs will be returned by the ora.Sel method.
	// The struct type is specified to ora.Sel by the user.
	// The map key is determined by a struct field tagged with `db:"fk4"`.
	MapOfValFk4
)
```

#### type Rset

```go
type Rset struct {
	sync.RWMutex

	Row     []interface{}
	Columns []Column
}
```

Rset represents a result set used to obtain Go values from a SQL select
statement.

Opening and closing a Rset is managed internally. Rset doesn't have an Open
method or Close method.

#### func (*Rset) Err

```go
func (rset *Rset) Err() error
```
Err returns the last error of the reesult set.

#### func (*Rset) Exhaust

```go
func (rset *Rset) Exhaust()
```
Exhaust will cycle to the end of the Rset, to autoclose it.

#### func (*Rset) IsOpen

```go
func (rset *Rset) IsOpen() bool
```
IsOpen returns true when a result set is open; otherwise, false.

#### func (*Rset) Len

```go
func (rset *Rset) Len() int
```
Len returns the number of rows retrieved.

#### func (*Rset) Name

```go
func (s *Rset) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Rset) Next

```go
func (rset *Rset) Next() bool
```
Next attempts to load a row of data from an Oracle buffer. True is returned when
a row of data is retrieved. False is returned when no data is available.

Retrieve the loaded row from the Rset.Row field. Rset.Row is updated on each
call to Next. Rset.Row is set to nil when Next returns false.

When Next returns false check Rset.Err() for any error that may have occured.

#### func (*Rset) NextRow

```go
func (rset *Rset) NextRow() []interface{}
```
NextRow attempts to load a row from the Oracle buffer and return the row. Nil is
returned when there's no data.

When NextRow returns nil check Rset.Err() for any error that may have occured.

#### type RsetCfg

```go
type RsetCfg struct {

	// TrueRune is rune a Go bool true value from SQL select-list character column.
	//
	// The is default is '1'.
	TrueRune rune

	// Err is the error from the last Set... method.
	Err error
}
```

RsetCfg affects the association of Oracle select-list columns to Go types.

Though it is unlucky, an empty RsetCfg is unusable! Please use NewRsetCfg().

RsetCfg is immutable, so all Set... methods returns a new copy!

#### func  NewRsetCfg

```go
func NewRsetCfg() RsetCfg
```
NewRsetCfg returns a RsetCfg with default values.

#### func (RsetCfg) BinaryDouble

```go
func (c RsetCfg) BinaryDouble() GoColumnType
```
BinaryDouble returns a GoColumnType associated to an Oracle select-list
BINARY_DOUBLE column.

The default is F64.

BinaryDouble is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, BinaryDouble is used.

#### func (RsetCfg) BinaryFloat

```go
func (c RsetCfg) BinaryFloat() GoColumnType
```
BinaryFloat returns a GoColumnType associated to an Oracle select-list
BINARY_FLOAT column.

The default for the database/sql package is F64.

The default for the ora package is F32.

BinaryFloat is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, BinaryFloat is used.

#### func (RsetCfg) Blob

```go
func (c RsetCfg) Blob() GoColumnType
```
Blob returns a GoColumnType associated to an Oracle select-list BLOB column.

The default is Bits.

Blob is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Blob is used.

#### func (RsetCfg) Char

```go
func (c RsetCfg) Char() GoColumnType
```
Char returns a GoColumnType associated to an Oracle select-list CHAR column and
NCHAR column.

The default is S.

Char is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Char is used.

#### func (RsetCfg) Char1

```go
func (c RsetCfg) Char1() GoColumnType
```
Char1 returns a GoColumnType associated to an Oracle select-list CHAR column
with length 1 and NCHAR column with length 1.

The default is B.

Char1 is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Char1 is used.

#### func (RsetCfg) Clob

```go
func (c RsetCfg) Clob() GoColumnType
```
Clob returns a GoColumnType associated to an Oracle select-list CLOB column and
NCLOB column.

The default is S.

Clob is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Clob is used.

#### func (RsetCfg) Date

```go
func (c RsetCfg) Date() GoColumnType
```
Date returns a GoColumnType associated to an Oracle select-list DATE column.

The default is T.

Date is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Date is used.

#### func (RsetCfg) Float

```go
func (c RsetCfg) Float() GoColumnType
```
Float returns a GoColumnType associated to an Oracle select-list FLOAT column.

The default is F64.

Float is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Float is used.

#### func (RsetCfg) IsZero

```go
func (c RsetCfg) IsZero() bool
```

#### func (RsetCfg) Long

```go
func (c RsetCfg) Long() GoColumnType
```
Long returns a GoColumnType associated to an Oracle select-list LONG column.

The default is S.

Long is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Long is used.

#### func (RsetCfg) LongRaw

```go
func (c RsetCfg) LongRaw() GoColumnType
```
LongRaw returns a GoColumnType associated to an Oracle select-list LONG RAW
column.

The default is Bits.

LongRaw is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, LongRaw is used.

#### func (RsetCfg) NumberBigFloat

```go
func (c RsetCfg) NumberBigFloat() GoColumnType
```
NumberBigFloat returns a GoColumnType associated to an Oracle select-list NUMBER
column defined with a scale greater than zero and precision unknown or > 15.

The default is N.

NumberBugFloat is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, NumberFloat is used.

#### func (RsetCfg) NumberBigInt

```go
func (c RsetCfg) NumberBigInt() GoColumnType
```
NumberBigInt returns a GoColumnType associated to an Oracle select-list NUMBER
column defined with scale zero and precision unknown or > 19.

The default is N.

The database/sql package uses NumberBigInt.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, NumberInt is used.

#### func (RsetCfg) NumberFloat

```go
func (c RsetCfg) NumberFloat() GoColumnType
```
NumberFloat returns a GoColumnType associated to an Oracle select-list NUMBER
column defined with a scale greater than zero.

The default is F64.

NumberFloat is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, NumberFloat is used.

#### func (RsetCfg) NumberInt

```go
func (c RsetCfg) NumberInt() GoColumnType
```
NumberInt returns a GoColumnType associated to an Oracle select-list NUMBER
column defined with scale zero and precision <= 19.

The default is I64.

The database/sql package uses NumberInt.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, NumberInt is used.

#### func (RsetCfg) Raw

```go
func (c RsetCfg) Raw() GoColumnType
```
Raw returns a GoColumnType associated to an Oracle select-list RAW column.

The default is Bits.

Raw is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Raw is used.

#### func (RsetCfg) SetBinaryDouble

```go
func (c RsetCfg) SetBinaryDouble(gct GoColumnType) RsetCfg
```
SetBinaryDouble sets a GoColumnType associated to an Oracle select-list
BINARY_DOUBLE column.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetBinaryFloat

```go
func (c RsetCfg) SetBinaryFloat(gct GoColumnType) RsetCfg
```
SetBinaryFloat sets a GoColumnType associated to an Oracle select-list
BINARY_FLOAT column.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, Num, OraNum.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetBlob

```go
func (c RsetCfg) SetBlob(gct GoColumnType) RsetCfg
```
SetBlob sets a GoColumnType associated to an Oracle select-list BLOB column.

Valid values are Bits and OraBits.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetChar

```go
func (c RsetCfg) SetChar(gct GoColumnType) RsetCfg
```
SetChar sets a GoColumnType associated to an Oracle select-list CHAR column and
NCHAR column.

Valid values are S and OraS.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetChar1

```go
func (c RsetCfg) SetChar1(gct GoColumnType) RsetCfg
```
SetChar1 sets a GoColumnType associated to an Oracle select-list CHAR column
with length 1 and NCHAR column with length 1.

Valid values are B, OraB, S and OraS.

Returns an error if a non-bool or non-string GoColumnType is specified.

#### func (RsetCfg) SetClob

```go
func (c RsetCfg) SetClob(gct GoColumnType) RsetCfg
```
SetClob sets a GoColumnType associated to an Oracle select-list CLOB column and
NCLOB column.

Valid values are S and OraS.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetDate

```go
func (c RsetCfg) SetDate(gct GoColumnType) RsetCfg
```
SetDate sets a GoColumnType associated to an Oracle select-list DATE column.

Valid values are T and OraT.

Returns an error if a non-time GoColumnType is specified.

#### func (RsetCfg) SetFloat

```go
func (c RsetCfg) SetFloat(gct GoColumnType) RsetCfg
```
SetFloat sets a GoColumnType associated to an Oracle select-list FLOAT column.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetLong

```go
func (c RsetCfg) SetLong(gct GoColumnType) RsetCfg
```
SetLong sets a GoColumnType associated to an Oracle select-list LONG column.

Valid values are S and OraS.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetLongRaw

```go
func (c RsetCfg) SetLongRaw(gct GoColumnType) RsetCfg
```
SetLongRaw sets a GoColumnType associated to an Oracle select-list LONG RAW
column.

Valid values are Bits and OraBits.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetNumberBigFloat

```go
func (c RsetCfg) SetNumberBigFloat(gct GoColumnType) RsetCfg
```
SetNumberBigFloat sets a GoColumnType associated to an Oracle select-list NUMBER
column defined with a scale greater than zero and precision unkonw or > 15.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetNumberBigInt

```go
func (c RsetCfg) SetNumberBigInt(gct GoColumnType) RsetCfg
```
SetNumberBigInt sets a GoColumnType associated to an Oracle select-list NUMBER
column defined with scale zero and precision unknown or > 19.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetNumberFloat

```go
func (c RsetCfg) SetNumberFloat(gct GoColumnType) RsetCfg
```
SetNumberFloat sets a GoColumnType associated to an Oracle select-list NUMBER
column defined with a scale greater than zero and precision <= 15.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetNumberInt

```go
func (c RsetCfg) SetNumberInt(gct GoColumnType) RsetCfg
```
SetNumberInt sets a GoColumnType associated to an Oracle select-list NUMBER
column defined with scale zero and precision <= 19.

Valid values are I64, I32, I16, I8, U64, U32, U16, U8, F64, F32, OraI64, OraI32,
OraI16, OraI8, OraU64, OraU32, OraU16, OraU8, OraF64, OraF32, N, OraN.

Returns an error if a non-numeric GoColumnType is specified.

#### func (RsetCfg) SetRaw

```go
func (c RsetCfg) SetRaw(gct GoColumnType) RsetCfg
```
SetRaw sets a GoColumnType associated to an Oracle select-list RAW column.

Valid values are Bits and OraBits.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) SetTimestamp

```go
func (c RsetCfg) SetTimestamp(gct GoColumnType) RsetCfg
```
SetTimestamp sets a GoColumnType associated to an Oracle select-list TIMESTAMP
column.

Valid values are T and OraT.

Returns an error if a non-time GoColumnType is specified.

#### func (RsetCfg) SetTimestampLtz

```go
func (c RsetCfg) SetTimestampLtz(gct GoColumnType) RsetCfg
```
SetTimestampLtz sets a GoColumnType associated to an Oracle select-list
TIMESTAMP WITH LOCAL TIME ZONE column.

Valid values are T and OraT.

Returns an error if a non-time GoColumnType is specified.

#### func (RsetCfg) SetTimestampTz

```go
func (c RsetCfg) SetTimestampTz(gct GoColumnType) RsetCfg
```
SetTimestampTz sets a GoColumnType associated to an Oracle select-list TIMESTAMP
WITH TIME ZONE column.

Valid values are T and OraT.

Returns an error if a non-time GoColumnType is specified.

#### func (RsetCfg) SetVarchar

```go
func (c RsetCfg) SetVarchar(gct GoColumnType) RsetCfg
```
SetVarchar sets a GoColumnType associated to an Oracle select-list VARCHAR
column, VARCHAR2 column and NVARCHAR2 column.

Valid values are S and OraS.

Returns an error if a non-string GoColumnType is specified.

#### func (RsetCfg) Timestamp

```go
func (c RsetCfg) Timestamp() GoColumnType
```
Timestamp returns a GoColumnType associated to an Oracle select-list TIMESTAMP
column.

The default is T.

Timestamp is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Timestamp is used.

#### func (RsetCfg) TimestampLtz

```go
func (c RsetCfg) TimestampLtz() GoColumnType
```
TimestampLtz returns a GoColumnType associated to an Oracle select-list
TIMESTAMP WITH LOCAL TIME ZONE column.

The default is T.

TimestampLtz is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, TimestampLtz is used.

#### func (RsetCfg) TimestampTz

```go
func (c RsetCfg) TimestampTz() GoColumnType
```
TimestampTz returns a GoColumnType associated to an Oracle select-list TIMESTAMP
WITH TIME ZONE column.

The default is T.

TimestampTz is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, TimestampTz is used.

#### func (RsetCfg) Varchar

```go
func (c RsetCfg) Varchar() GoColumnType
```
Varchar returns a GoColumnType associated to an Oracle select-list VARCHAR
column, VARCHAR2 column and NVARCHAR2 column.

The default is S.

Varchar is used by the database/sql package.

When using the ora package directly, custom GoColumnType associations may be
specified to the Ses.Prep method. If no custom GoColumnType association is
specified, Varchar is used.

#### type Ses

```go
type Ses struct {
	sync.RWMutex
}
```

Ses is an Oracle session associated with a server.

#### func (*Ses) Break

```go
func (ses *Ses) Break() (err error)
```
Break stops the currently running OCI function.

#### func (*Ses) Cfg

```go
func (ses *Ses) Cfg() SesCfg
```
Cfg returns the Ses's SesCfg, or it's Srv's, if not set. If the ses.srv.env is
the PkgSqlEnv, that will override StmtCfg!

#### func (*Ses) Close

```go
func (ses *Ses) Close() (err error)
```
Close ends a session on an Oracle server.

Any open statements associated with the session are closed.

Calling Close will cause Ses.IsOpen to return false. Once closed, a session
cannot be re-opened. Call Srv.OpenSes to open a new session.

#### func (*Ses) Env

```go
func (ses *Ses) Env() *Env
```

#### func (*Ses) Ins

```go
func (ses *Ses) Ins(tbl string, columnPairs ...interface{}) (err error)
```
Ins composes, prepares and executes a sql INSERT statement returning a possible
error.

Ins offers convenience when specifying a long list of sql columns.

Ins expects at least two column name-value pairs where the last pair will be a
part of a sql RETURNING clause. The last column name is expected to be an
identity column returning an Oracle-generated value. The last value specified to
the variadic parameter 'columnPairs' is expected to be a pointer capable of
receiving the identity value.

#### func (*Ses) IsOpen

```go
func (ses *Ses) IsOpen() bool
```
IsOpen returns true when a session is open; otherwise, false.

Calling Close will cause Ses.IsOpen to return false. Once closed, a session
cannot be re-opened. Call Srv.OpenSes to open a new session.

#### func (*Ses) Name

```go
func (s *Ses) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Ses) NumStmt

```go
func (ses *Ses) NumStmt() int
```
NumStmt returns the number of open Oracle statements.

#### func (*Ses) NumTx

```go
func (ses *Ses) NumTx() int
```
NumTx returns the number of open Oracle transactions.

#### func (*Ses) Ping

```go
func (ses *Ses) Ping() (err error)
```
Ping returns nil when an Oracle server is contacted; otherwise, an error.

#### func (*Ses) Prep

```go
func (ses *Ses) Prep(sql string, gcts ...GoColumnType) (stmt *Stmt, err error)
```
Prep prepares a sql statement returning a *Stmt and possible error.

#### func (*Ses) PrepAndExe

```go
func (ses *Ses) PrepAndExe(sql string, params ...interface{}) (rowsAffected uint64, err error)
```
PrepAndExe prepares and executes a SQL statement returning the number of rows
affected and a possible error, using Exe, calling in batch for arrays.

WARNING: just as sql.QueryRow, the prepared statement is closed right after
execution, with all its siblings (Lobs, Rsets...)!

So if you want to retrieve and use such objects, you have to first Prep, then
Exe separately (and close the Stmt returned by Prep after finishing with those
objects).

#### func (*Ses) PrepAndExeP

```go
func (ses *Ses) PrepAndExeP(sql string, params ...interface{}) (rowsAffected uint64, err error)
```
PrepAndExeP prepares and executes a SQL statement returning the number of rows
affected and a possible error, using ExeP, so passing arrays as is.

#### func (*Ses) PrepAndQry

```go
func (ses *Ses) PrepAndQry(sql string, params ...interface{}) (rset *Rset, err error)
```
PrepAndQry prepares a SQL statement and queries an Oracle server returning an
*Rset and a possible error.

If an error occurs during Prep or Qry a nil *Rset will be returned.

The *Stmt internal to this method is automatically closed when the *Rset
retrieves all rows or returns an error.

#### func (*Ses) Sel

```go
func (ses *Ses) Sel(sqlFrom string, columnPairs ...interface{}) (rset *Rset, err error)
```
Sel composes, prepares and queries a sql SELECT statement returning an *ora.Rset
and possible error.

Sel offers convenience when specifying a long list of sql columns with
non-default GoColumnTypes.

Specify a sql FROM clause with one or more pairs of sql column name-GoColumnType
pairs. The FROM clause may have additional SQL clauses such as WHERE, HAVING,
etc.

#### func (*Ses) SetAction

```go
func (ses *Ses) SetAction(module, action string) error
```
SetAction sets the MODULE and ACTION attribute of the session.

#### func (*Ses) SetCfg

```go
func (ses *Ses) SetCfg(cfg SesCfg)
```

#### func (*Ses) StartTx

```go
func (ses *Ses) StartTx(opts ...TxOption) (tx *Tx, err error)
```
StartTx starts an Oracle transaction returning a *Tx and possible error.

#### func (*Ses) Timezone

```go
func (ses *Ses) Timezone() (*time.Location, error)
```
Timezone return the current session's timezone.

#### func (*Ses) Upd

```go
func (ses *Ses) Upd(tbl string, columnPairs ...interface{}) (err error)
```
Upd composes, prepares and executes a sql UPDATE statement returning a possible
error.

Upd offers convenience when specifying a long list of sql columns.

#### type SesCfg

```go
type SesCfg struct {
	Username string
	Password string
	Mode     SessionMode

	StmtCfg
}
```


#### func  NewSesCfg

```go
func NewSesCfg() SesCfg
```

#### func (SesCfg) IsZero

```go
func (c SesCfg) IsZero() bool
```

#### func (SesCfg) SetBinaryDouble

```go
func (c SesCfg) SetBinaryDouble(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetBinaryFloat

```go
func (c SesCfg) SetBinaryFloat(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetBlob

```go
func (c SesCfg) SetBlob(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetByteSlice

```go
func (c SesCfg) SetByteSlice(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetChar

```go
func (c SesCfg) SetChar(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetChar1

```go
func (c SesCfg) SetChar1(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetClob

```go
func (c SesCfg) SetClob(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetDate

```go
func (c SesCfg) SetDate(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetFloat

```go
func (c SesCfg) SetFloat(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetLobBufferSize

```go
func (c SesCfg) SetLobBufferSize(size int) SesCfg
```

#### func (SesCfg) SetLong

```go
func (c SesCfg) SetLong(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetLongBufferSize

```go
func (c SesCfg) SetLongBufferSize(size uint32) SesCfg
```

#### func (SesCfg) SetLongRaw

```go
func (c SesCfg) SetLongRaw(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetLongRawBufferSize

```go
func (c SesCfg) SetLongRawBufferSize(size uint32) SesCfg
```

#### func (SesCfg) SetNumberBigFloat

```go
func (c SesCfg) SetNumberBigFloat(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetNumberBigInt

```go
func (c SesCfg) SetNumberBigInt(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetNumberFloat

```go
func (c SesCfg) SetNumberFloat(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetNumberInt

```go
func (c SesCfg) SetNumberInt(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetPrefetchMemorySize

```go
func (c SesCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) SesCfg
```

#### func (SesCfg) SetPrefetchRowCount

```go
func (c SesCfg) SetPrefetchRowCount(prefetchRowCount uint32) SesCfg
```

#### func (SesCfg) SetRaw

```go
func (c SesCfg) SetRaw(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetStmtCfg

```go
func (cfg SesCfg) SetStmtCfg(stmtCfg StmtCfg) SesCfg
```

#### func (SesCfg) SetStringPtrBufferSize

```go
func (c SesCfg) SetStringPtrBufferSize(size int) SesCfg
```

#### func (SesCfg) SetTimestamp

```go
func (c SesCfg) SetTimestamp(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetTimestampLtz

```go
func (c SesCfg) SetTimestampLtz(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetTimestampTz

```go
func (c SesCfg) SetTimestampTz(gct GoColumnType) SesCfg
```

#### func (SesCfg) SetVarchar

```go
func (c SesCfg) SetVarchar(gct GoColumnType) SesCfg
```

#### type SesPool

```go
type SesPool struct {
}
```


#### func (*SesPool) Close

```go
func (p *SesPool) Close() error
```

#### func (*SesPool) Get

```go
func (p *SesPool) Get() (*Ses, error)
```
Get a session from an idle Srv.

#### func (*SesPool) Put

```go
func (p *SesPool) Put(ses *Ses)
```
Put the session back to the session pool.

#### func (SesPool) SetEvictDuration

```go
func (p SesPool) SetEvictDuration(dur time.Duration)
```
Set the eviction duration to the given. Also starts eviction if not yet started.

#### type SessionMode

```go
type SessionMode uint8
```


#### func  DSNMode

```go
func DSNMode(str string) SessionMode
```
DSNMode returns the SessionMode (SysDefault/SysDba/SysOper).

#### type Srv

```go
type Srv struct {
	sync.RWMutex
}
```

Srv represents an Oracle server.

#### func (*Srv) Cfg

```go
func (srv *Srv) Cfg() SrvCfg
```
Cfg returns the Srv's SrvCfg, or it's Env's, if not set. If the env is the
PkgSqlEnv, that will override StmtCfg!

#### func (*Srv) Close

```go
func (srv *Srv) Close() (err error)
```
Close disconnects from an Oracle server.

Any open sessions associated with the server are closed.

Calling Close will cause Srv.IsOpen to return false. Once closed, a server
cannot be re-opened. Call Env.OpenSrv to open a new server.

#### func (*Srv) IsOpen

```go
func (srv *Srv) IsOpen() bool
```
IsOpen returns true when the server is open; otherwise, false.

Calling Close will cause Srv.IsOpen to return false. Once closed, a server
cannot be re-opened. Call Env.OpenSrv to open a new server.

#### func (*Srv) IsUTF8

```go
func (srv *Srv) IsUTF8() bool
```
IsUTF8 returns whether the DB uses AL32UTF8 encoding.

#### func (*Srv) Name

```go
func (s *Srv) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Srv) NewSesPool

```go
func (srv *Srv) NewSesPool(sesCfg SesCfg, size int) *SesPool
```
NewSesPool returns a session pool, which evicts the idle sessions in every
minute. The pool holds at most size idle Ses. If size is zero, DefaultPoolSize
will be used.

#### func (*Srv) NumSes

```go
func (srv *Srv) NumSes() int
```
NumSes returns the number of open Oracle sessions.

#### func (*Srv) OpenSes

```go
func (srv *Srv) OpenSes(cfg SesCfg) (ses *Ses, err error)
```
OpenSes opens an Oracle session returning a *Ses and possible error.

#### func (*Srv) SetCfg

```go
func (srv *Srv) SetCfg(cfg SrvCfg)
```

#### func (*Srv) Version

```go
func (srv *Srv) Version() (ver string, err error)
```
Version returns the Oracle database server version.

Version requires the server have at least one open session.

#### type SrvCfg

```go
type SrvCfg struct {
	// Dblink specifies an Oracle database server. Dblink is a connect string
	// or a service point.
	Dblink string

	Pool PoolCfg

	// StmtCfg configures new Stmts.
	StmtCfg
}
```

SrvCfg configures a new Srv.

#### func (SrvCfg) IsZero

```go
func (c SrvCfg) IsZero() bool
```

#### type SrvPool

```go
type SrvPool struct {
}
```


#### func (*SrvPool) Close

```go
func (p *SrvPool) Close() error
```

#### func (*SrvPool) Get

```go
func (p *SrvPool) Get() (*Srv, error)
```
Get a connection.

#### func (*SrvPool) Put

```go
func (p *SrvPool) Put(srv *Srv)
```
Put the connection back to the idle pool.

#### func (SrvPool) SetEvictDuration

```go
func (p SrvPool) SetEvictDuration(dur time.Duration)
```
Set the eviction duration to the given. Also starts eviction if not yet started.

#### type Stmt

```go
type Stmt struct {
	sync.RWMutex
}
```

Stmt represents an Oracle statement.

#### func (*Stmt) Cfg

```go
func (stmt *Stmt) Cfg() StmtCfg
```
Cfg returns the Stmt's StmtCfg, or it's Ses's, if not set. If the env is the
PkgSqlEnv, that will override StmtCfg!

#### func (*Stmt) Close

```go
func (stmt *Stmt) Close() (err error)
```
Close closes the SQL statement.

Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
cannot be re-opened. Call Stmt.Prep to create a new statement.

#### func (*Stmt) Env

```go
func (stmt *Stmt) Env() *Env
```

#### func (*Stmt) Exe

```go
func (stmt *Stmt) Exe(params ...interface{}) (rowsAffected uint64, err error)
```
Exe executes a SQL statement on an Oracle server returning the number of rows
affected and a possible error.

Slice arguments should have the same length, as they'll be called in batch mode.

#### func (*Stmt) ExeP

```go
func (stmt *Stmt) ExeP(params ...interface{}) (rowsAffected uint64, err error)
```
ExeP executes an (PL/)SQL statement on an Oracle server returning the number of
rows affected and a possible error.

All arguments are sent as is (esp. slices).

#### func (*Stmt) Gcts

```go
func (stmt *Stmt) Gcts() []GoColumnType
```
Gcts returns a slice of GoColumnType specified by Ses.Prep or Stmt.SetGcts.

Gcts is used by a Stmt.Qry *ora.Rset to determine which Go types are mapped to a
sql select-list.

#### func (*Stmt) IsOpen

```go
func (stmt *Stmt) IsOpen() bool
```
IsOpen returns true when a statement is open; otherwise, false.

Calling Close will cause Stmt.IsOpen to return false. Once closed, a statement
cannot be re-opened. Call Stmt.Prep to create a new statement.

#### func (*Stmt) Name

```go
func (s *Stmt) Name(calc func() string) string
```
Name sets the name to the result of calc once, then returns that result forever.
(Effectively caches the result of calc().)

#### func (*Stmt) NumInput

```go
func (stmt *Stmt) NumInput() int
```
NumInput returns the number of placeholders in a sql statement.

#### func (*Stmt) NumRset

```go
func (stmt *Stmt) NumRset() int
```
NumRset returns the number of open Oracle result sets.

#### func (*Stmt) Parse

```go
func (stmt *Stmt) Parse() (err error)
```
Parse the statement, and return the syntax errors - WITHOUT executing it.
Rejects ALTER statements, as they're executed anyway by Oracle...

#### func (*Stmt) Qry

```go
func (stmt *Stmt) Qry(params ...interface{}) (*Rset, error)
```
Qry runs a SQL query on an Oracle server returning a *Rset and possible error.

#### func (*Stmt) SelfCfg

```go
func (stmt *Stmt) SelfCfg() StmtCfg
```
returns the Stmt's StmtCfg only

#### func (*Stmt) SetCfg

```go
func (stmt *Stmt) SetCfg(cfg StmtCfg)
```

#### func (*Stmt) SetGcts

```go
func (stmt *Stmt) SetGcts(gcts []GoColumnType) []GoColumnType
```
SetGcts sets a slice of GoColumnType used in a Stmt.Qry *ora.Rset.

SetGcts is optional.

#### type StmtCfg

```go
type StmtCfg struct {

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
```

StmtCfg affects various aspects of a SQL statement.

Assign values to StmtCfg prior to calling Stmt.Exe and Stmt.Qry for the
configuration values to take effect.

StmtCfg is immutable, so every Set method returns a new instance, maybe with Err
set, too.

#### func  NewStmtCfg

```go
func NewStmtCfg() StmtCfg
```
NewStmtCfg returns a StmtCfg with default values.

#### func (StmtCfg) ByteSlice

```go
func (c StmtCfg) ByteSlice() GoColumnType
```
ByteSlice returns a GoColumnType associated to SQL statement []byte parameter.

The default is Bits.

ByteSlice is used by the database/sql package.

Sending a byte slice to an Oracle server as a parameter in a SQL statement
requires knowing the destination column type ahead of time. Set ByteSlice to
Bits if the destination column is BLOB, RAW or LONG RAW. Set ByteSlice to U8 if
the destination column is NUMBER, BINARY_DOUBLE, BINARY_FLOAT or FLOAT.

#### func (StmtCfg) FetchLen

```go
func (c StmtCfg) FetchLen() int
```
returns a value of the fetchLen

#### func (StmtCfg) IsZero

```go
func (c StmtCfg) IsZero() bool
```

#### func (StmtCfg) LOBFetchLen

```go
func (c StmtCfg) LOBFetchLen() int
```
returns a value of the lobFetchLen

#### func (StmtCfg) LobBufferSize

```go
func (c StmtCfg) LobBufferSize() int
```
LobBufferSize returns the LOB buffer size in bytes used to define the sql
select-column buffer size of an Oracle LOB type.

The default is 16,777,216 bytes.

The default is considered a moderate buffer where the 2GB max buffer may not be
feasible on all clients.

#### func (StmtCfg) LongBufferSize

```go
func (c StmtCfg) LongBufferSize() uint32
```
LongBufferSize returns the long buffer size in bytes used to define the sql
select-column buffer size of an Oracle LONG type.

The default is 16,777,216 bytes.

The default is considered a moderate buffer where the 2GB max buffer may not be
feasible on all clients.

#### func (StmtCfg) LongRawBufferSize

```go
func (c StmtCfg) LongRawBufferSize() uint32
```
LongRawBufferSize returns the LONG RAW buffer size in bytes used to define the
sql select-column buffer size of an Oracle LONG RAW type.

The default is 16,777,216 bytes.

The default is considered a moderate buffer where the 2GB max buffer may not be
feasible on all clients.

#### func (StmtCfg) PrefetchMemorySize

```go
func (c StmtCfg) PrefetchMemorySize() uint32
```
PrefetchMemorySize returns the prefetch memory size in bytes used during a SQL
select command.

The default is 134,217,728 bytes.

PrefetchMemorySize works in coordination with PrefetchRowCount. When
PrefetchRowCount is set to zero only PrefetchMemorySize is used; otherwise, the
minimum of PrefetchRowCount and PrefetchMemorySize is used.

#### func (StmtCfg) PrefetchRowCount

```go
func (c StmtCfg) PrefetchRowCount() uint32
```
PrefetchRowCount returns the number of rows to prefetch during a select query.

The default is 0.

PrefetchRowCount works in coordination with PrefetchMemorySize. When
PrefetchRowCount is set to zero only PrefetchMemorySize is used; otherwise, the
minimum of PrefetchRowCount and PrefetchMemorySize is used.

#### func (StmtCfg) SetBinaryDouble

```go
func (c StmtCfg) SetBinaryDouble(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetBinaryFloat

```go
func (c StmtCfg) SetBinaryFloat(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetBlob

```go
func (c StmtCfg) SetBlob(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetByteSlice

```go
func (c StmtCfg) SetByteSlice(gct GoColumnType) StmtCfg
```
SetByteSlice sets a GoColumnType associated to SQL statement []byte parameter.

Valid values are U8 and Bits.

Returns an error if U8 or Bits is not specified.

#### func (StmtCfg) SetChar

```go
func (c StmtCfg) SetChar(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetChar1

```go
func (c StmtCfg) SetChar1(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetClob

```go
func (c StmtCfg) SetClob(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetDate

```go
func (c StmtCfg) SetDate(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetFetchLen

```go
func (c StmtCfg) SetFetchLen(length int) StmtCfg
```
SetFetchLen overrides DefaultFetchLen for prefetch lengths.

#### func (StmtCfg) SetFloat

```go
func (c StmtCfg) SetFloat(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetLOBFetchLen

```go
func (c StmtCfg) SetLOBFetchLen(length int) StmtCfg
```
SetLOBFetchLen overrides DefaultLOBFetchLen for prefetch LOB lengths.

This affects result sets with any of the following column types: C.SQLT_LNG,
C.SQLT_BFILE, C.SQLT_BLOB, C.SQLT_CLOB, C.SQLT_LBI

Caution: the default buffer size for blob is 1MB. So, for example a single fetch
from the result set that contains just one blob will consume 128MB of RAM

#### func (StmtCfg) SetLobBufferSize

```go
func (c StmtCfg) SetLobBufferSize(size int) StmtCfg
```
SetLobBufferSize sets the LOB buffer size in bytes.

The maximum is 2,147,483,642 bytes.

Returns an error if the specified size is greater than 2,147,483,642.

#### func (StmtCfg) SetLong

```go
func (c StmtCfg) SetLong(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetLongBufferSize

```go
func (c StmtCfg) SetLongBufferSize(size uint32) StmtCfg
```
SetLongBufferSize sets the long buffer size in bytes.

The maximum is 2,147,483,642 bytes.

Returns an error if the specified size is less than 1 or greater than
2,147,483,642.

#### func (StmtCfg) SetLongRaw

```go
func (c StmtCfg) SetLongRaw(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetLongRawBufferSize

```go
func (c StmtCfg) SetLongRawBufferSize(size uint32) StmtCfg
```
SetLongRawBufferSize sets the LONG RAW buffer size in bytes.

The maximum is 2,147,483,642 bytes.

Returns an error if the specified size is greater than 2,147,483,642.

#### func (StmtCfg) SetNumberBigFloat

```go
func (c StmtCfg) SetNumberBigFloat(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetNumberBigInt

```go
func (c StmtCfg) SetNumberBigInt(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetNumberFloat

```go
func (c StmtCfg) SetNumberFloat(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetNumberInt

```go
func (c StmtCfg) SetNumberInt(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetPrefetchMemorySize

```go
func (c StmtCfg) SetPrefetchMemorySize(prefetchMemorySize uint32) StmtCfg
```
SetPrefetchMemorySize sets the prefetch memory size in bytes used during a SQL
select command.

#### func (StmtCfg) SetPrefetchRowCount

```go
func (c StmtCfg) SetPrefetchRowCount(prefetchRowCount uint32) StmtCfg
```
SetPrefetchRowCount sets the number of rows to prefetch during a select query.

#### func (StmtCfg) SetRaw

```go
func (c StmtCfg) SetRaw(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetStringPtrBufferSize

```go
func (c StmtCfg) SetStringPtrBufferSize(size int) StmtCfg
```
SetStringPtrBufferSize sets the size of a buffer used to store a string during
*string parameter binding and []*string parameter binding in a SQL statement.

#### func (StmtCfg) SetTimestamp

```go
func (c StmtCfg) SetTimestamp(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetTimestampLtz

```go
func (c StmtCfg) SetTimestampLtz(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetTimestampTz

```go
func (c StmtCfg) SetTimestampTz(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) SetVarchar

```go
func (c StmtCfg) SetVarchar(gct GoColumnType) StmtCfg
```

#### func (StmtCfg) StringPtrBufferSize

```go
func (c StmtCfg) StringPtrBufferSize() int
```
StringPtrBufferSize returns the size of a buffer in bytes used to store a string
during *string parameter binding and []*string parameter binding in a SQL
statement.

The default is 4000 bytes.

For a *string parameter binding, you may wish to increase the size of
StringPtrBufferSize depending on the Oracle column type. For VARCHAR2,
NVARCHAR2, and RAW oracle columns the Oracle MAX_STRING_SIZE is usually 4000 but
may be set up to 32767.

#### type String

```go
type String struct {
	IsNull bool
	Value  string
}
```

String is a nullable string.

#### func (String) Equals

```go
func (this String) Equals(other String) bool
```
Equals returns true when the receiver and specified String are both null, or
when the receiver and specified String are both not null and Values are equal.

#### func (String) MarshalJSON

```go
func (this String) MarshalJSON() ([]byte, error)
```

#### func (String) String

```go
func (this String) String() string
```

#### func (*String) UnmarshalJSON

```go
func (this *String) UnmarshalJSON(p []byte) error
```

#### type Time

```go
type Time struct {
	IsNull bool
	Value  time.Time
}
```

Time is a nullable time.Time.

#### func (Time) Equals

```go
func (this Time) Equals(other Time) bool
```
Equals returns true when the receiver and specified Time are both null, or when
the receiver and specified Time are both not null and Values are equal.

#### func (Time) MarshalJSON

```go
func (this Time) MarshalJSON() ([]byte, error)
```

#### func (*Time) UnmarshalJSON

```go
func (this *Time) UnmarshalJSON(p []byte) error
```

#### type Tx

```go
type Tx struct {
	sync.RWMutex
}
```

Tx represents an Oracle transaction associated with a session.

Implements the driver.Tx interface.

#### func (*Tx) Commit

```go
func (tx *Tx) Commit() (err error)
```
Commit commits the transaction.

Commit is a member of the driver.Tx interface.

#### func (*Tx) Rollback

```go
func (tx *Tx) Rollback() (err error)
```
Rollback rolls back a transaction.

Rollback is a member of the driver.Tx interface.

#### type TxOption

```go
type TxOption func(*txOption)
```


#### func  TxFlags

```go
func TxFlags(flags uint32) TxOption
```

#### func  TxTimeout

```go
func TxTimeout(timeout time.Duration) TxOption
```

#### type Uint16

```go
type Uint16 struct {
	IsNull bool
	Value  uint16
}
```

Uint16 is a nullable uint16.

#### func (Uint16) Equals

```go
func (this Uint16) Equals(other Uint16) bool
```
Equals returns true when the receiver and specified Uint16 are both null, or
when the receiver and specified Uint16 are both not null and Values are equal.

#### func (Uint16) MarshalJSON

```go
func (this Uint16) MarshalJSON() ([]byte, error)
```

#### func (*Uint16) UnmarshalJSON

```go
func (this *Uint16) UnmarshalJSON(p []byte) error
```

#### type Uint32

```go
type Uint32 struct {
	IsNull bool
	Value  uint32
}
```

Uint32 is a nullable uint32.

#### func (Uint32) Equals

```go
func (this Uint32) Equals(other Uint32) bool
```
Equals returns true when the receiver and specified Uint32 are both null, or
when the receiver and specified Uint32 are both not null and Values are equal.

#### func (Uint32) MarshalJSON

```go
func (this Uint32) MarshalJSON() ([]byte, error)
```

#### func (*Uint32) UnmarshalJSON

```go
func (this *Uint32) UnmarshalJSON(p []byte) error
```

#### type Uint64

```go
type Uint64 struct {
	IsNull bool
	Value  uint64
}
```

Uint64 is a nullable uint64.

#### func (Uint64) Equals

```go
func (this Uint64) Equals(other Uint64) bool
```
Equals returns true when the receiver and specified Uint64 are both null, or
when the receiver and specified Uint64 are both not null and Values are equal.

#### func (Uint64) MarshalJSON

```go
func (this Uint64) MarshalJSON() ([]byte, error)
```

#### func (*Uint64) UnmarshalJSON

```go
func (this *Uint64) UnmarshalJSON(p []byte) error
```

#### type Uint8

```go
type Uint8 struct {
	IsNull bool
	Value  uint8
}
```

Uint8 is a nullable uint8.

#### func (Uint8) Equals

```go
func (this Uint8) Equals(other Uint8) bool
```
Equals returns true when the receiver and specified Uint8 are both null, or when
the receiver and specified Uint8 are both not null and Values are equal.

#### func (Uint8) MarshalJSON

```go
func (this Uint8) MarshalJSON() ([]byte, error)
```

#### func (*Uint8) UnmarshalJSON

```go
func (this *Uint8) UnmarshalJSON(p []byte) error
```
