package main

import (
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

func binanceCliCommand() ([]*cli.Command, error) {

	const crypto = "BTC"
	binanceLog := log.WithFields(logrus.Fields{"exchange": "binance"})

	cmd := []*cli.Command{
		{
			Name:        "binance",
			Usage:       "Binance exchange",
			Description: "Stack some sats on Binance full description",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "api-key",
					Usage:    "Binance Api Key",
					EnvVars:  []string{"STACKER_BINANCE_API_KEY"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "secret-key",
					Usage:    "Binance Api Secret",
					EnvVars:  []string{"STACKER_BINANCE_API_SECRET", "STACKER_BINANCE_API_SECRET"},
					Required: true,
				},
				&cli.Float64Flag{
					Name:     "amount",
					Usage:    "Amount of fiat to exchange",
					EnvVars:  []string{"STACKER_AMOUNT"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "fiat",
					Usage:    "Fiat to exchange",
					EnvVars:  []string{"STACKER_FIAT"},
					Required: true,
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:  "stack",
					Usage: "stack some sats on Binance",
					Action: func(c *cli.Context) error {
						thisLog := binanceLog.WithFields(logrus.Fields{"action": "stack"})
						thisLog.Info("Stacking some sats")

						return nil
					},
				},
				{
					Name:  "withdraw",
					Usage: "withdraw sats from Binance",
					Action: func(c *cli.Context) error {
						thisLog := binanceLog.WithFields(logrus.Fields{"action": "withdraw"})
						thisLog.Info("Withdrawing some sats")

						return nil
					},
				},
			},
		},
	}

	return cmd, nil
}
