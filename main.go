package main

import (
	"context"
	"log"
	"os"

	"github.com/imtanmoy/openax/cmd"
)

func main() {
	app := cmd.NewApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}