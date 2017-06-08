package ora_test

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const num = 1000
	rows, err := testDb.QueryContext(ctx, "SELECT object_name, object_type, object_id, created FROM all_objects WHERE ROWNUM < :alpha ORDER BY object_id", sql.Named("alpha", num))
	if err != nil {
		t.Fatalf("%#v", err)
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
	if n != num-1 {
		t.Errorf("got %d rows, wanted %d", n, num-1)
	}
}
