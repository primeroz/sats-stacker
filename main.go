package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"sats-stacker/binance"
	"sats-stacker/kraken"
)

//TODO: struct for return document - so i can push to notifier

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

	// Setup Logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)

	// Setup the App
	// Kraken Exchange
	krakenCmd, err := kraken.GenerateCliCommand()
	if err != nil {
		log.Fatal(err)
	}

	// Binance Exchange
	binanceCmd, _ := binance.GenerateCliCommand()
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
	}
	//app.UseShortOptionHandling = true

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
