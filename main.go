package main

import (
	"context"
	"log"
	"os"

	"github.com/imtanmoy/openax/cmd"
)

// Build information, set by GoReleaser ldflags
//
//nolint:unused // These variables are set by build-time ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cmd.NewApp()

	// Set version information
	app.Version = version

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
