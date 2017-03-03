/*
   Package main in conntest represents a connection testing program.

   Copyright 2013 Tamás Gulácsi

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
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"

	"gopkg.in/rana/ora.v4/examples/connect"
)

var fWait = flag.Bool("wait", false, "wait for USR1 signal?")

func ListObjects(t *testing.T) {
	conn, err := connect.GetConnection("")
	if err != nil {
		t.Errorf("error connectiong: %s", err)
		t.FailNow()
		return
	}
	defer conn.Close()

	qry := "SELECT owner, object_name, object_id FROM all_objects WHERE ROWNUM < 20"
	log.Printf(`executing "%s"`, qry)
	rows, err := conn.Query(qry)
	if err != nil {
		t.Logf(`error with %q: %s`, qry, err)
		t.FailNow()
		return
	}
	var (
		owner, objectName string
		objectID          int
	)
	for rows.Next() {
		if err = rows.Scan(&owner, &objectName, &objectID); err != nil {
			t.Errorf("error fetching: %s", err)
			break
		}
		t.Logf(`row: "%s";"%s";%d`, owner, objectName, objectID)
	}
}

func main() {
	flag.Parse()
	t := new(testing.T)
	if *fWait {
		c := make(chan os.Signal)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			log.Printf("waiting for signal...")
			sig := <-c
			log.Printf("got signal %s", sig)
			ListObjects(t)
			wg.Done()
		}()
		signal.Notify(c, syscall.SIGUSR1)
		wg.Wait()
	} else {
		ListObjects(t)
	}
}
