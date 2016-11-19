// +build never

package main

import (
	"context"
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

func dbRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf(" %v. finished.", ctx.Value("id"))
			return
		default:
		}
		var temp int
		db.QueryRow("select 1 from dual").Scan(&temp)
		if rand.Int()%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func main() {
	dur, _ := time.ParseDuration(os.Getenv("DURATION"))
	if dur == 0 {
		dur = 24 * time.Hour
	}

	log.SetPrefix("#131 ")
	log.Println("starting")
	startDB(fmt.Sprintf("%s/%s@%s", os.Getenv("GO_ORA_DRV_TEST_USERNAME"), os.Getenv("GO_ORA_DRV_TEST_PASSWORD"), os.Getenv("GO_ORA_DRV_TEST_DB")))

	go func() {
		http.ListenAndServe(":8889", nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		ctx := context.WithValue(ctx, "id", i)
		go func() {
			defer wg.Done()
			dbRoutine(ctx)
		}()
	}
	wg.Wait()
	log.Println("finished.")
}

// vim: set fileencoding=utf-8 noet:
