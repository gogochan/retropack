package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var commands []*cli.Command

func init() {
	commands = []*cli.Command{
		cmdDeploy,
	}
}

func main() {
	app := &cli.App{
		Usage:    "A tool for packing and deploying on outdated operating systems",
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
