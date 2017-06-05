package ora_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "gopkg.in/rana/ora.v5"
)

var testDb *sql.DB

func init() {
	var err error
	if testDb, err = sql.Open("ora", os.Getenv("GO_ORA_DRV_TEST_USERNAME")+"/"+os.Getenv("GO_ORA_DRV_TEST_PASSWORD")+"@"+os.Getenv("GO_ORA_DRV_TEST_DB")); err != nil {
		panic(err)
	}
}

func TestSelect(t *testing.T) {
	rows, err := testDb.Query("SELECT object_name, object_type, object_id, created FROM all_objects WHERE ROWNUM < 1000 ORDER BY object_id")
	if err != nil {
		t.Fatal(err)
	}
	n, oldOid := 0, int64(0)
	for rows.Next() {
		var tbl, typ string
		var oid int64
		var created time.Time
		if err := rows.Scan(&tbl, &typ, &oid, &created); err != nil {
			t.Fatal(err)
		}
		t.Log(tbl, typ, oid, created)
		n++
		if oldOid > oid {
			t.Errorf("got oid=%d, wanted sth < %d.", oid, oldOid)
		}
		oldOid = oid
	}
	if n != 999 {
		t.Errorf("got %d rows, wanted 999")
	}
}
