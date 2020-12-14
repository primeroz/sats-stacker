package binance

import (
	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

const crypto = "BTC"

var thisLog = log.WithFields(log.Fields{"exchange": "binance"})

func GenerateCliCommand() ([]*cli.Command, error) {

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
					Name:   "stack",
					Usage:  "stack some sats on Binance",
					Action: stack,
				},
				{
					Name:   "withdraw",
					Usage:  "withdraw sats from Binance",
					Action: withdraw,
				},
			},
		},
	}

	return cmd, nil
}

func stack(c *cli.Context) error {
	slog := thisLog.WithFields(log.Fields{"action": "stack"})
	slog.Info("Stacking some sats")

	return nil
}

func withdraw(c *cli.Context) error {
	wlog := thisLog.WithFields(log.Fields{"action": "withdraw"})
	wlog.Info("Withdrawing some sats")

	return nil
}
