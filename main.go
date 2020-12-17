package main

import (
	"fmt"
	"os"
	"plugin"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var sResult = orderResult{}
var wResult = withdrawResult{}
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

func stack(c *cli.Context) error {
	err := krakenStack(c, &sResult, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while stacking on Kraken : %s", err), 1)
	}

	fmt.Printf("Result\n")
	fmt.Printf("%#v", sResult)
	return nil
}

func withdraw(c *cli.Context) error {
	err := krakenWithdraw(c, &wResult, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while withdrawing from Kraken : %s", err), 1)
	}

	fmt.Printf("Result\n")
	fmt.Printf("%#v", wResult)
	return nil
}

func main() {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"validate"},
			//Value:   false,
			Value:   true,
			Usage:   "dry-run",
			EnvVars: []string{"STACKER_VALIDATE", "STACKER_DRY_RUN"},
		},
		&cli.StringFlag{
			Name:     "api-key",
			Usage:    "Kraken Api Key",
			EnvVars:  []string{"STACKER_API_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "secret-key",
			Usage:    "Kraken Api Secret",
			EnvVars:  []string{"STACKER_SECRET_KEY", "STACKER_API_SECRET"},
			Required: true,
		},
	}

	commands := []*cli.Command{
		{
			Name:        "stack",
			Usage:       "Stack some sats on Kraken",
			Description: "Stack some sats on Kraken full description",
			Flags: []cli.Flag{
				&cli.Float64Flag{
					Name:     "amount",
					Usage:    "Amount of fiat to exchange",
					EnvVars:  []string{"STACKER_STACK_AMOUNT"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "fiat",
					Usage:    "Fiat to exchange",
					EnvVars:  []string{"STACKER_STACK_FIAT"},
					Required: true,
				},
				&cli.StringFlag{
					Name:    "order-type",
					Aliases: []string{"type"},
					Value:   "limit",
					Usage:   "Order type",
					EnvVars: []string{"STACKER_STACK_ORDER_TYPE"},
				},
			},
			Action: stack,
		},
		{
			Name:        "withdraw",
			Usage:       "Withdraw some sats from Kraken",
			Description: "Withdraw some sats from Kraken full description",
			Flags: []cli.Flag{
				&cli.Float64Flag{
					Name:     "max-fee",
					Usage:    "Max fee in percentage",
					EnvVars:  []string{"STACKER_WITHDRAW_MAX_FEE"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "address",
					Usage:    "Address to withdraw to, the actual value will depend on the exchange selected",
					EnvVars:  []string{"STACKER_WITHDRAW_ADDRESS"},
					Required: true,
				},
			},
			Action: withdraw,
		},
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
		HelpName:  "Kraken SATs Stacker",
		Usage:     "demonstrate available API",
		UsageText: "sats-stacker - stack and withdraw for Kraken exchange",
		Flags:     flags,
		Commands:  commands,
		Before: func(c *cli.Context) error {
			return nil
		},
		After: func(c *cli.Context) error {
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
