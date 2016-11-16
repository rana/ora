// +build never

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "gopkg.in/rana/ora.v3"
)

var db *sql.DB

func startDB(dsn string) {
	var err error
	db, err = sql.Open("ora", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(32)
	db.SetConnMaxLifetime(10 * time.Second)
}

func dbRoutine() {
	tick := time.Tick(1000 * time.Second)
	for {
		select {
		case <-tick:
			log.Println("finish")
			return
		default:
			var temp int
			db.QueryRow("select popid from devicetable where devid=7008").Scan(&temp)
			if rand.Int()%10 == 0 {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

func main() {
	log.Println("starting")
	startDB(fmt.Sprintf("%s/%s@%s", os.Getenv("GO_ORA_DRV_TEST_USERNAME"), os.Getenv("GO_ORA_DRV_TEST_PASSWORD"), os.Getenv("GO_ORA_DRV_TEST_DB")))

	for i := 0; i < 40; i++ {
		go dbRoutine()
	}
	select {}
}
