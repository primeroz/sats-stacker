package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

type operation struct {
	exchange    string
	success     bool
	err         error
	description string
}

var result = operation{exchange: "None", success: false, err: nil, description: ""}
var log = logrus.New()

func init() {

	// Setup Logging
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stderr)
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
		Commands:  append(krakenCmd, binanceCmd...),
		After: func(c *cli.Context) error {
			fmt.Printf("Fake Notifier for %s: %s\n", result.exchange, result.description)

			if result.err != nil {
				return cli.Exit(fmt.Sprintf("%s - %s: %s", result.exchange, result.description, result.err), 1)
			}
			return nil
		},
	}
	//app.UseShortOptionHandling = true

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
