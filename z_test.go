package ora_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "gopkg.in/rana/ora.v5"
)

var testDb *sql.DB

func init() {
	var err error
	if testDb, err = sql.Open("ora", os.Getenv("GO_ORA_DRV_TEST_USERNAME")+"/"+os.Getenv("GO_ORA_DRV_TEST_PASSWORD")+"@"+os.Getenv("GO_ORA_DRV_TEST_DB")); err != nil {
		fmt.Println("ERROR")
		panic(err)
	}
}

func TestSelect(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const num = 1000
	rows, err := testDb.QueryContext(ctx, "SELECT object_name, object_type, object_id, created FROM all_objects WHERE ROWNUM < NVL(:alpha, 2) ORDER BY object_id", sql.Named("alpha", num))
	//rows, err := testDb.QueryContext(ctx, "SELECT object_name, object_type, object_id, created FROM all_objects WHERE ROWNUM < 1000 ORDER BY object_id")
	if err != nil {
		t.Fatalf("%+v", err)
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
		if tbl == "" {
			t.Fatal("empty tbl")
		}
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
func TestExecuteMany(t *testing.T) {
	t.Parallel()
	testDb.Exec("CREATE TABLE test_em (i INTEGER)")
	defer testDb.Exec("DROP TABLE test_em")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const num = 1000
	nums := make([]int, num)
	for i := range nums {
		nums[i] = i << 1
	}
	res, err := testDb.ExecContext(ctx, "INSERT INTO test_em (i) VALUES (:1)", nums)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	t.Logf("result=%+v", res)
}
