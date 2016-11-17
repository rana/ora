// +build never

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
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
	for i := 0; i < 3; i++ {
		var temp int
		db.QueryRow("select 1 from DUAL").Scan(&temp)
		log.Println("running")
		if rand.Int()%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func main() {
	log.SetPrefix("#133 ")
	log.Println("starting")

	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	startDB(fmt.Sprintf("%s/%s@%s", os.Getenv("GO_ORA_DRV_TEST_USERNAME"), os.Getenv("GO_ORA_DRV_TEST_PASSWORD"), os.Getenv("GO_ORA_DRV_TEST_DB")))

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbRoutine()
		}()
	}
	wg.Wait()
}
