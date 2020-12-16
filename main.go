package main

import (
	"fmt"
	"os"
	"plugin"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sats-stacker/types"
)

var stackResult = types.OrderResult{}
var withdrawResult = types.WithdrawResult{}
var log = logrus.New()

// Store the plugin loaded
var stackP plugin.Symbol
var withdrawP plugin.Symbol

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
	err := stackP.(func(*cli.Context, *types.OrderResult, *logrus.Logger) error)(c, &stackResult, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while stacking on %s : %s", stackResult.Exchange, err), 1)
	}

	fmt.Printf("Result\n")
	fmt.Printf("%#v", stackResult)
	return nil
}

func withdraw(c *cli.Context) error {
	err := withdrawP.(func(*cli.Context, *types.WithdrawResult, *logrus.Logger) error)(c, &withdrawResult, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while withdrawing from %s : %s", stackResult.Exchange, err), 1)
	}

	fmt.Printf("Result\n")
	fmt.Printf("%#v", withdrawResult)
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
			Name:    "exchange",
			Usage:   "Exchange to use: ['kraken','binance']",
			Value:   "kraken",
			EnvVars: []string{"STACKER_EXCHANGE"},
		},
		&cli.StringFlag{
			Name:     "api-key",
			Usage:    "Exchange Api Key",
			EnvVars:  []string{"STACKER_API_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "secret-key",
			Usage:    "Exchange Api Secret",
			EnvVars:  []string{"STACKER_SECRET_KEY", "STACKER_API_SECRET"},
			Required: true,
		},
	}

	commands := []*cli.Command{
		{
			Name:        "stack",
			Usage:       "Stack some sats on Exchange",
			Description: "Stack some sats on Exchange full description",
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
			Usage:       "Withdraw some sats from Exchange",
			Description: "Withdraw some sats from Exchange full description",
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
		HelpName:  "SATs Stacker",
		Usage:     "demonstrate available API",
		UsageText: "sats-stacker - demonstrating the available API",
		Flags:     flags,
		Commands:  commands,
		Before: func(c *cli.Context) error {
			// Validate Exchange selected
			exc := c.String("exchange")
			var pluginPath string
			switch exc {
			case "kraken":
				pluginPath = "./plugins/kraken.so"
			case "binance":
				pluginPath = "./plugins/binance.so"
			default:
				return cli.Exit(fmt.Sprintf("Unsupported Exchange: %s", exc), 2)
			}

			// Load Plugin
			plug, err := plugin.Open(pluginPath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to load Exchange plugin: %s", err), 1)
			}

			// Lookup the symbol
			stackP, err = plug.Lookup("Stack")
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to lookup Stack() function from Exchange plugin: %s", err), 1)
			}
			withdrawP, err = plug.Lookup("Withdraw")
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to lookup Withdraw() function from Exchange plugin: %s", err), 1)
			}

			// Validate the loaded Symbols
			_, ok := stackP.(func(*cli.Context, *types.OrderResult, *logrus.Logger) error)
			if !ok {
				return cli.Exit(fmt.Sprintf("Failed to Assert Stack(*cli.Context, *types.OrderResult, *logrus.Logger) function from exchange: %s", exc), 1)
			}
			_, ok = withdrawP.(func(*cli.Context, *types.WithdrawResult, *logrus.Logger) error)
			if !ok {
				return cli.Exit(fmt.Sprintf("Failed to Assert Withdraw(*cli.Context, *types.WithdrawResult, *logrus.Logger) function from exchange: %s", exc), 1)
			}

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
