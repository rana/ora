// +build never

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sync/errgroup"
)

var issues = []string{"z_issue131.go", "z_issue133.go"}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	grp, ctx := errgroup.WithContext(ctx)
	for _, nm := range issues {
		nm := nm
		grp.Go(func() error {
			cmd := exec.Command("go", "run", "-race", nm)
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			return cmd.Run()
		})
	}
	log.Fatal(grp.Wait())
}
