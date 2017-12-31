/*
   Copyright 2017 Tamás Gulácsi

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package connect

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gopkg.in/rana/ora.v4"

	"github.com/pkg/errors"

	"github.com/tgulacsi/go/dber"
	"github.com/tgulacsi/go/orahlp"
)

var (
	fDsn      = flag.String("db.dsn", "", "Oracle DSN (user/passw@sid)")
	fUsername = flag.String("db.username", "", "username to connect as (if you don't provide the dsn")
	fPassword = flag.String("db.password", "", "password to connect with (if you don't provide the dsn")
	//fHost        = flag.String("db.host", "", "Oracle DB's host (if you don't provide the dsn")
	//fPort        = flag.Int("db.port", 1521, "Oracle DB's port (if you don't provide the dsn) - defaults to 1521")
	fSid = flag.String("db.sid", "", "Oracle DB's SID (if you don't provide the dsn)")
	//fServiceName = flag.String("db.service", "", "Oracle DB's ServiceName (if you don't provide the dsn and the sid)")
)

// GetDSN returns a (command-line defined) connection string
func GetCfg(dsn string) (srvCfg ora.SrvCfg, sesCfg ora.SesCfg) {
	if dsn != "" {
		sesCfg.Username, sesCfg.Password, srvCfg.Dblink = orahlp.SplitDSN(*fDsn)
		return srvCfg, sesCfg
	}

	if !flag.Parsed() {
		flag.Parse()
		if *fDsn == "" {
			*fDsn = os.Getenv("DSN")
		}
	}

	if *fDsn != "" {
		sesCfg.Username, sesCfg.Password, srvCfg.Dblink = orahlp.SplitDSN(*fDsn)
	}
	if sesCfg.Username == "" {
		sesCfg.Username = *fUsername
		if sesCfg.Password == "" {
			sesCfg.Password = *fPassword
		}
	}
	if srvCfg.Dblink == "" {
		if *fSid != "" {
			srvCfg.Dblink = *fSid
		} else {
			if srvCfg.Dblink = os.Getenv("ORACLE_SID"); srvCfg.Dblink == "" {
				srvCfg.Dblink = os.Getenv("TWO_TASK")
			}
		}
	}
	return srvCfg, sesCfg
}

func GetDSN(srvCfg ora.SrvCfg, sesCfg ora.SesCfg) string {
	if srvCfg.Dblink == "" && sesCfg.Username == "" {
		srvCfg, sesCfg = GetCfg("")
	}
	return sesCfg.Username + "/" + sesCfg.Password + "@" + srvCfg.Dblink
}

// GetConnection returns a connection - using GetDSN if dsn is empty
func GetConnection(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = GetDSN(GetCfg(""))
	}
	log.Printf("GetConnection dsn=%v", dsn)
	conn, err := sql.Open("ora", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "dsn="+dsn)
	}
	return conn, nil
}

var (
	oraEnv  *ora.Env
	oraCxMu sync.Mutex
)

// GetRawConnection returns a raw (*ora.Ses) connection
// - using GetDSN if dsn is empty
func GetRawConnection(dsn string) (*ora.Ses, error) {
	oraCxMu.Lock()
	defer oraCxMu.Unlock()

	if oraEnv == nil {
		var err error
		if oraEnv, err = ora.OpenEnv(); err != nil {
			return nil, errors.Wrap(err, "OpenEnv")
		}
	}
	srvCfg, sesCfg := GetCfg(dsn)
	srv, err := oraEnv.OpenSrv(srvCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenSrv(%#v)", srvCfg)
	}
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		srv.Close()
		return nil, errors.Wrapf(err, "OpenSes(%#v)", sesCfg)
	}
	return ses, nil
}

// SplitDSN splits the username/password@sid string to its parts.
//
// Copied from github.com/tgulacsi/go/orahlp/orahlp.go
func SplitDSN(dsn string) (username, password, sid string) {
	if strings.HasPrefix(dsn, "/@") {
		return "", "", dsn[2:]
	}
	if i := strings.LastIndex(dsn, "@"); i >= 0 {
		sid, dsn = dsn[i+1:], dsn[:i]
	}
	if i := strings.IndexByte(dsn, '/'); i >= 0 {
		username, password = dsn[:i], dsn[i+1:]
	}
	return
}

type Column struct {
	Schema, Name                   string
	Type, Length, Precision, Scale int
	Nullable                       bool
	CharsetID, CharsetForm         int
}

// DescribeQuery describes the columns in the qry string,
// using DBMS_SQL.PARSE + DBMS_SQL.DESCRIBE_COLUMNS2.
//
// This can help using unknown-at-compile-time, a.k.a.
// dynamic queries.
func DescribeQuery(db dber.Execer, qry string) ([]Column, error) {
	//res := strings.Repeat("\x00", 32767)
	res := make([]byte, 32767)
	if _, err := db.Exec(`DECLARE
  c INTEGER;
  col_cnt INTEGER;
  rec_tab DBMS_SQL.DESC_TAB;
  a DBMS_SQL.DESC_REC;
  v_idx PLS_INTEGER;
  res VARCHAR2(32767);
BEGIN
  c := DBMS_SQL.OPEN_CURSOR;
  BEGIN
    DBMS_SQL.PARSE(c, :1, DBMS_SQL.NATIVE);
    DBMS_SQL.DESCRIBE_COLUMNS(c, col_cnt, rec_tab);
    v_idx := rec_tab.FIRST;
    WHILE v_idx IS NOT NULL LOOP
      a := rec_tab(v_idx);
      res := res||a.col_schema_name||' '||a.col_name||' '||a.col_type||' '||
                  a.col_max_len||' '||a.col_precision||' '||a.col_scale||' '||
                  (CASE WHEN a.col_null_ok THEN 1 ELSE 0 END)||' '||
                  a.col_charsetid||' '||a.col_charsetform||
                  CHR(10);
      v_idx := rec_tab.NEXT(v_idx);
    END LOOP;
  EXCEPTION WHEN OTHERS THEN NULL;
    DBMS_SQL.CLOSE_CURSOR(c);
	RAISE;
  END;
  :2 := UTL_RAW.CAST_TO_RAW(res);
END;`, qry, &res,
	); err != nil {
		return nil, err
	}
	if i := bytes.IndexByte(res, 0); i >= 0 {
		res = res[:i]
	}
	lines := bytes.Split(res, []byte{'\n'})
	cols := make([]Column, 0, len(lines))
	var nullable int
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var col Column
		switch j := bytes.IndexByte(line, ' '); j {
		case -1:
			continue
		case 0:
			line = line[1:]
		default:
			col.Schema, line = string(line[:j]), line[j+1:]
		}
		if n, err := fmt.Sscanf(string(line), "%s %d %d %d %d %d %d %d",
			&col.Name, &col.Type, &col.Length, &col.Precision, &col.Scale, &nullable, &col.CharsetID, &col.CharsetForm,
		); err != nil {
			return cols, errors.Wrapf(err, "parsing %q (parsed: %d)", line, n)
		}
		col.Nullable = nullable != 0
		cols = append(cols, col)
	}
	return cols, nil
}

type Version struct {
	// major.maintenance.application-server.component-specific.platform-specific
	Major, Maintenance, AppServer, Component, Platform int8
}

func GetVersion(db dber.Queryer) (Version, error) {
	var s sql.NullString
	if err := db.QueryRow("SELECT MIN(VERSION) FROM product_component_version " +
		" WHERE product LIKE 'Oracle Database%'").Scan(&s); err != nil {
		return Version{Major: -1}, err
	}
	var v Version
	if _, err := fmt.Sscanf(s.String, "%d.%d.%d.%d.%d",
		&v.Major, &v.Maintenance, &v.AppServer, &v.Component, &v.Platform); err != nil {
		return v, errors.Wrapf(err, "scan version number %q", s.String)
	}
	return v, nil
}

// MapToSlice modifies query for map (:paramname) to :%d placeholders + slice of params.
//
// Calls metParam for each parameter met, and returns the slice of their results.
func MapToSlice(qry string, metParam func(string) interface{}) (string, []interface{}) {
	if metParam == nil {
		metParam = func(string) interface{} { return nil }
	}
	arr := make([]interface{}, 0, 16)
	var buf bytes.Buffer
	state, p, last := 0, 0, 0
	for i, r := range qry {
		switch {
		case state == 0 && r == ':':
			state++
			p = i
			// An identifier consists of a letter optionally followed by more letters, numerals, dollar signs, underscores, and number signs.
			// http://docs.oracle.com/cd/B19306_01/appdev.102/b14261/fundamentals.htm#sthref309
		case state == 1 &&
			!('A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' ||
				(i-p > 1 && ('0' <= r && r <= '9' || r == '$' || r == '_' || r == '#'))):
			state = 0
			if i-p <= 1 { // :=
				continue
			}
			arr = append(arr, metParam(qry[p+1:i]))
			param := fmt.Sprintf(":%d", len(arr))
			buf.WriteString(qry[last:p])
			buf.WriteString(param)
			last = i
		}
	}
	if last < len(qry)-1 {
		buf.WriteString(qry[last:])
	}
	return buf.String(), arr
}
