package ora

/*
#cgo CFLAGS: -Iodpi/src -Iodpi/include
#cgo LDFLAGS: -Lodpi/lib -lodpic -ldl

#include "dpiImpl.h"
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	DpiMajorVersion = 2
	DpiMinorVersion = 0
)

func init() {
	var d drv
	err := &oraErr{}
	if C.dpiContext_create(C.uint(DpiMajorVersion), C.uint(DpiMinorVersion),
		(**C.dpiContext)(unsafe.Pointer(&d.dpiContext)), &err.errInfo,
	) == C.DPI_FAILURE {
		panic(err)
	}

	sql.Register("ora", &drv{})
}

var _ = driver.Driver((*drv)(nil))

type drv struct {
	dpiContext *C.dpiContext
}

// Open returns a new connection to the database.
// The name is a string in a driver-specific format.
func (d *drv) Open(connString string) (driver.Conn, error) {
	var username, password, sid, connClass string
	var isSysDBA, isSysOper bool
	if strings.HasPrefix(connString, "ora://") {
		u, err := url.Parse(connString)
		if err != nil {
			return nil, err
		}
		if usr := u.User; usr != nil {
			username = usr.Username()
			password, _ = usr.Password()
		}
		sid = u.Hostname()
		if u.Port() != "" {
			sid += ":" + u.Port()
		}
		q := u.Query()
		if isSysDBA = q.Get("sysdba") == "1"; !isSysDBA {
			isSysOper = q.Get("sysoper") == "1"
		}
	} else {
		i := strings.IndexByte(connString, '/')
		if i < 0 {
			return nil, errors.Errorf("no / in %q", connString)
		}
		username, connString = connString[:i], connString[i+1:]
		if i = strings.IndexByte(connString, '@'); i < 0 {
			return nil, errors.Errorf("no @ in %q", connString)
		}
		password, sid = connString[:i], connString[i+1:]
		uSid := strings.ToUpper(sid)
		if isSysDBA = strings.HasSuffix(uSid, " AS SYSDBA"); isSysDBA {
			sid = sid[:len(sid)-10]
		} else if isSysOper = strings.HasSuffix(uSid, " AS SYSOPER"); isSysOper {
			sid = sid[:len(sid)-11]
		}
		if strings.HasSuffix(sid, ":POOLED") {
			connClass, sid = "POOLED", sid[:len(sid)-7]
		}
	}

	authMode := C.dpiAuthMode(C.DPI_MODE_AUTH_DEFAULT)
	if isSysDBA {
		authMode |= C.DPI_MODE_AUTH_SYSDBA
	} else if isSysOper {
		authMode |= C.DPI_MODE_AUTH_SYSOPER
	}

	var c conn
	cUserName, cPassword, cSid := C.CString(username), C.CString(password), C.CString(sid)
	cUTF8, cConnClass := C.CString("AL32UTF8"), C.CString(connClass)
	defer func() {
		C.free(unsafe.Pointer(cUserName))
		C.free(unsafe.Pointer(cPassword))
		C.free(unsafe.Pointer(cSid))
		C.free(unsafe.Pointer(cUTF8))
		C.free(unsafe.Pointer(cConnClass))
	}()
	var extAuth C.int
	if username == "" && password == "" {
		extAuth = 1
	}
	if C.dpiConn_create(
		d.dpiContext,
		cUserName, C.uint32_t(len(username)),
		cPassword, C.uint32_t(len(password)),
		cSid, C.uint32_t(len(sid)),
		&C.dpiCommonCreateParams{
			createMode: C.DPI_MODE_CREATE_DEFAULT | C.DPI_MODE_CREATE_THREADED | C.DPI_MODE_CREATE_EVENTS,
			encoding:   cUTF8, nencoding: cUTF8,
		},
		&C.dpiConnCreateParams{
			authMode:        authMode,
			connectionClass: cConnClass, connectionClassLength: C.uint32_t(len(connClass)),
			externalAuth: extAuth,
		},
		(**C.dpiConn)(unsafe.Pointer(&c.dpiConn)),
	) == C.DPI_FAILURE {
		return nil, d.getError()
	}
	return &c, nil
}

var _ = driver.Conn((*conn)(nil))
var _ = driver.ConnBeginTx((*conn)(nil))
var _ = driver.ConnPrepareContext((*conn)(nil))

type conn struct {
	dpiConn *C.dpiConn
}

// Prepare returns a prepared statement, bound to this connection.
func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return nil, nil
}

// Close invalidates and potentially stops any current
// prepared statements and transactions, marking this
// connection as no longer in use.
//
// Because the sql package maintains a free pool of
// connections and only calls Close when there's a surplus of
// idle connections, it shouldn't be necessary for drivers to
// do their own connection caching.
func (c *conn) Close() error {
	return nil
}

// Begin starts and returns a new transaction.
//
// Deprecated: Drivers should implement ConnBeginTx instead (or additionally).
func (c *conn) Begin() (driver.Tx, error) {
	return nil, nil
}

// BeginTx starts and returns a new transaction.
// If the context is canceled by the user the sql package will
// call Tx.Rollback before discarding and closing the connection.
//
// This must check opts.Isolation to determine if there is a set
// isolation level. If the driver does not support a non-default
// level and one is set or if there is a non-default isolation level
// that is not supported, an error must be returned.
//
// This must also check opts.ReadOnly to determine if the read-only
// value is true to either set the read-only transaction property if supported
// or return an error if it is not supported.
func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, nil
}

// PrepareContext returns a prepared statement, bound to this connection.
// context is for the preparation of the statement,
// it must not store the context within the statement itself.
func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return nil, nil
}

type oraErr struct {
	errInfo C.dpiErrorInfo
}

func (oe *oraErr) Code() int       { return int(oe.errInfo.code) }
func (oe *oraErr) Message() string { return C.GoString(oe.errInfo.message) }
func (oe *oraErr) Error() string {
	if oe.errInfo.code == 0 {
		return ""
	}
	return fmt.Sprintf("ORA-%05d: %s", oe.Code(), oe.Message())
}

func (d *drv) getError() *oraErr {
	var oe oraErr
	C.dpiContext_getError(d.dpiContext, &oe.errInfo)
	return &oe
}
