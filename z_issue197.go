// +build never

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/tgulacsi/go/dber"
	"github.com/tgulacsi/go/orahlp"

	ora "gopkg.in/rana/ora.v4"
)

func main() {
	go func() {
		path := "./tmp"

		for {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGUSR2)
			s := <-c
			fmt.Println("get signal:", s, ",pprof ...")
			fm, err := os.OpenFile(path+"/"+"mem.out", os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(fm)
			fm.Close()
		}
	}()

	insellpid := "0000000048E053433210AC068C"
	procRset := &ora.Rset{}
	env, err := ora.OpenEnv()
	if err != nil {
		panic(err)
	}
	defer env.Close()
	srvCfg := ora.SrvCfg{Dblink: os.Getenv("GO_ORA_DRV_TEST_DB")}
	srv, err := env.OpenSrv(srvCfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	sesCfg := ora.SesCfg{
		Username: os.Getenv("GO_ORA_DRV_TEST_USERNAME"),
		Password: os.Getenv("GO_ORA_DRV_TEST_PASSWORD"),
	}
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		panic(err)
	}
	defer ses.Close()
	qry := `CREATE OR REPLACE PROCEDURE test_p1(insellpid IN VARCHAR2, procRset OUT SYS_REFCURSOR) IS
    BEGIN
	  OPEN procRset FOR SELECT * FROM user_objects;
	END;`
	if _, err := ses.PrepAndExe(qry); err != nil {
		log.Fatal(errors.Wrap(err, qry))
	}

	qry = "CALL test_P1(:1,:2)"

	var rmapArr []interface{}
	for {
		stmtProcCall, err := ses.Prep(qry)
		if err != nil {
			log.Fatal(errors.Wrap(err, qry))
		}
		if _, err := stmtProcCall.Exe(insellpid, procRset); err != nil {
			db, _ := sql.Open("ora", sesCfg.Username+"/"+sesCfg.Password+"@"+srvCfg.Dblink)
			defer db.Close()
			log.Println(orahlp.GetCompileErrors(dber.SqlDBer{db}, false))

			log.Fatal(errors.Wrapf(err, "%q, [%v, %v]", qry, insellpid, procRset))
		}
		if !procRset.IsOpen() {
			log.Println("rset is closed!")
			continue
		}
		log.Println(procRset.Columns)
		rmapArr = rmapArr[:0]
		for procRset.Next() {
			rmap := make(map[string]interface{})
			cols := procRset.Columns
			row := procRset.Row
			for j, col := range cols {
				clo := col.Name
				switch x := row[j].(type) {
				case ora.OCINum:
					s := x.String()
					if "" == s {
						rmap[clo] = nil
					} else {
						if f, err := strconv.ParseFloat(s, 64); err != nil {
							log.Fatal(errors.Wrap(err, s))
						} else {
							rmap[clo] = f
						}
					}

				case time.Time:
					rmap[clo] = x.Format("20060102 15:04:05")
				case string:
					if "" == row[j] {
						rmap[clo] = nil
					} else {
						rmap[clo] = x
					}
				default:
					rmap[clo] = row[j]
				}

			}
			rmapArr = append(rmapArr, rmap)
			rmap = nil
		}
		fmt.Println("rmapArr=", rmapArr)
	}
}
