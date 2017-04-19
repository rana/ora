// +build ignore

// Copyright 2017 Tamás Gulácsi. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "gopkg.in/rana/ora.v4"
)

func main() {
	db, err := sql.Open("ora", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Printf("Current open connections: %d", db.Stats().OpenConnections)
		if err := executeSQL(db); err != nil {
			log.Printf("%s\n%#v %T", err, err, err)
			//panic(err)
			log.Println("code:", err.(interface {
				Code() int
			}).Code())
		}

		log.Println("Loop finish, wait for next.")
		time.Sleep(5 * time.Second)
	}

	//log.Printf("All collect finished\nCurrent open connections: %d", db.Stats().OpenConnections)
}

func executeSQL(db *sql.DB) error {
	var n int64
	return db.QueryRow("SELECT COUNT(0) FROM cat").Scan(&n)
}
