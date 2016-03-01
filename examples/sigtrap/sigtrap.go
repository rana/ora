package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/rana/ora.v3"
)

var config = struct {
	TNS      string
	Username string
	Password string
}{
	TNS:      os.Getenv("GO_ORA_DRV_TEST_DB"),
	Username: os.Getenv("GO_ORA_DRV_TEST_USERNAME"),
	Password: os.Getenv("GO_ORA_DRV_TEST_PASSWORD"),
}

func procExec(proc string) (string, error) {

	// environment
	env, err := ora.OpenEnv(nil)
	if err != nil {
		return "", err
	}
	defer func() {
		env.Close()
		log.Printf("[debug] env closed")
	}()

	// server configuration
	srvCfg := ora.NewSrvCfg()
	srvCfg.Dblink = config.TNS

	// server
	srv, err := env.OpenSrv(srvCfg)
	if err != nil {
		return "", err
	}
	defer func() {
		srv.Close()
		log.Printf("[debug] srv closed")
	}()

	// session configuration
	sesCfg := ora.NewSesCfg()
	sesCfg.Username = config.Username
	sesCfg.Password = config.Password

	// session
	ses, err := srv.OpenSes(sesCfg)
	if err != nil {
		return "", err
	}
	defer func() {
		ses.Close()
		log.Printf("[debug] ses closed")
	}()

	// fetch records
	var x string
	stmt, err := ses.Prep(proc)
	if err != nil {
		return "", err
	}
	_, err = stmt.Exe(&x)
	if err != nil {
		return "", err
	}

	return x, nil
}

func main() {

	for {
		result, err := procExec(`
  DECLARE
    x varchar2(50);
  BEGIN
    SELECT to_char(sysdate, 'yyyy.mm.dd hh24:mi:ss')
      INTO x
      FROM dual;
    :result := x;
  END;
`)
		if err != nil {
			log.Printf("Error: %s", err)
		}

		log.Printf("Result: %s", result)

		time.Sleep(time.Second)
		log.Printf("INT self")
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			log.Fatal(err)
		}
		p.Signal(os.Interrupt)
		time.Sleep(30 * time.Second)
	}
}
