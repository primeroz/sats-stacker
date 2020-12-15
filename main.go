package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"plugin"
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
	apiKey := c.String("api-key")
	secretKey := c.String("secret-key")
	amount := c.Float64("amount")
	fiat := c.String("fiat")

	result, err := stackP.(func(string, string, float64, string) (string, error))(apiKey, secretKey, amount, fiat)
	fmt.Println(result)
	fmt.Println(err)
	return nil
}

func withdraw(c *cli.Context) error {
	fmt.Println("Withdrawing Sats")
	return nil
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
		&cli.StringFlag{
			Name:    "exchange",
			Usage:   "Exchange to use: ['kraken','binance']",
			Value:   "kraken",
			EnvVars: []string{"STACKER_EXCHANGE"},
		},
		&cli.StringFlag{
			Name:     "api-key",
			Usage:    "Exchange Api Key",
			EnvVars:  []string{"STACKER_KRAKEN_API_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "secret-key",
			Usage:    "Exchange Api Secret",
			EnvVars:  []string{"STACKER_API_SECRET", "STACKER_API_SECRET"},
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
					EnvVars:  []string{"STACKER_MAX_FEE"},
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
				pluginPath = "./kraken/kraken.so"
			case "binance":
				pluginPath = "./binance/binance.so"
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

			_, ok := stackP.(func(string, string, float64, string) (string, error))
			if !ok {
				return cli.Exit(fmt.Sprintf("Failed to Assert Stack(string,string,float64,string) function from exchange: %s", exc), 1)
			}
			_, ok = withdrawP.(func(string, string, float64, string) (string, error))
			if !ok {
				return cli.Exit(fmt.Sprintf("Failed to Assert Withdraw(string,string,float64,string) function from exchange: %s", exc), 1)
			}

			return nil
		},
		After: func(c *cli.Context) error {
			return nil
		},
	}
	//app.UseShortOptionHandling = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
