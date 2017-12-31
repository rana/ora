// +build never

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sync/errgroup"
)

var issues = []string{"z_issue131.go"}

func main() {
	flagDur := flag.Duration("duration", 24*time.Hour, "test duration")
	flag.Parse()

	cmd := exec.Command("go", "install", "-race")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	os.Setenv("DURATION", flagDur.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	grp, ctx := errgroup.WithContext(ctx)
	for _, nm := range issues {
		nm := nm
		grp.Go(func() error {
			cmd := exec.CommandContext(ctx, "go", "run", "-race", nm)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			return cmd.Run()
		})
	}
	log.Fatal(grp.Wait())
}
