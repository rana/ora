// +build never

package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/rana/ora.v4"
)

const insellpid = "0000000048A16C23433210AC068C"

func main() {
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}

	srvCfg := ora.SrvCfg{Dblink: os.Getenv("GO_ORA_DRV_TEST_DB")}
	SrvPool := env.NewSrvPool(srvCfg, 40)
	defer SrvPool.Close()
	sesCfg := ora.SesCfg{
		Username: os.Getenv("GO_ORA_DRV_TEST_USERNAME"),
		Password: os.Getenv("GO_ORA_DRV_TEST_PASSWORD"),
	}
	srv, err := SrvPool.Get()
	if err != nil {
		log.Fatal(err)
	}
	defer SrvPool.Put(srv)
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer ses.Close()
	qry := `CREATE OR REPLACE PROCEDURE test_p1(insellpid IN VARCHAR2, procRset OUT SYS_REFCURSOR) IS
	    BEGIN
	         OPEN procRset FOR SELECT * FROM all_objects WHERE ROWNUM < 100;
	       END;`
	if _, err := ses.PrepAndExe(qry); err != nil {
		log.Fatal(errors.Wrap(err, qry))
	}

	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		if err := work(SrvPool, sesCfg); err != nil {
			log.Println(err)
		}
		//os.Stdout.Write([]byte{'.'})
	}
}

func work(SrvPool *ora.SrvPool, sesCfg ora.SesCfg) error {
	srv, err := SrvPool.Get()
	if err != nil {
		return err
	}
	defer SrvPool.Put(srv)
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		return err
	}
	defer ses.Close()

	qry := "CALL test_P1(:1,:2)"

	procRset := &ora.Rset{}
	stmtProcCall, err := ses.Prep(qry)
	if err != nil {
		return errors.Wrap(err, qry)
	}
	defer stmtProcCall.Close()

	if _, err = stmtProcCall.Exe(insellpid, procRset); err != nil {
		return err
	}
	if !procRset.IsOpen() {
		return nil
	}

	rmapArr := make([]map[string]interface{}, 0)
	for procRset.Next() {
		rmap := make(map[string]interface{})
		cols := procRset.Columns
		row := procRset.Row
		for j := 0; j < len(row); j++ {
			clo := cols[j].Name
			switch x := row[j].(type) {
			case ora.OCINum:
				va_n := x.String()
				if "" == va_n {
					rmap[clo] = nil
				} else {
					fl_64, err_ := strconv.ParseFloat(va_n, 64)
					if err_ != nil {
						return errors.Wrapf(err, "strconv.ParseFloat(%q)", va_n)
						panic(err_)
					}
					rmap[clo] = fl_64
				}

			case time.Time:
				rmap[clo] = x.Format("20060102 15:04:05")
			case string:
				if "" == row[j] {
					rmap[clo] = nil
				} else {
					rmap[clo] = row[j]
				}
			default:
				rmap[clo] = row[j]
			}

		}
		rmapArr = append(rmapArr, rmap)
	}
	log.Printf("%d rows, each with %d columns", len(rmapArr), len(rmapArr[0]))

	return nil
}
