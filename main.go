package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

type operation struct {
	success     bool
	err         error
	description string
}

var result operation
var log = logrus.New()

func init() {

	// Setup Logging
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	//log.SetLevel(logrus.InfoLevel)
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"validate"},
			Value:   false,
			Usage:   "dry-run",
			EnvVars: []string{"STACKER_VALIDATE", "STACKER_DRY_RUN"},
		},
	}

	// Setup the App
	// Kraken Exchange
	krakenCmd, err := krakenCliCommand()
	if err != nil {
		log.Fatal(err)
	}

	// Binance Exchange
	binanceCmd, err := binanceCliCommand()
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Name:     "sats-stacker",
		Version:  "0.0.1",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Francesco Ciocchetti",
				Email: "primeroznl@gmail.com",
			},
		},
		Copyright: "GPL",
		HelpName:  "SATs Stacker",
		Usage:     "demonstrate available API",
		UsageText: "sats-stacker - demonstrating the available API",
		Flags:     flags,
		// https://stackoverflow.com/questions/16248241/concatenate-two-slices-in-go
		Commands: append(krakenCmd, binanceCmd...),
		After: func(c *cli.Context) error {
			fmt.Printf("AFTER: %#v", result)
			return nil
		},
	}
	//app.UseShortOptionHandling = true

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
