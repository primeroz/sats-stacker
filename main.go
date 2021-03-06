package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sats-stacker/exchange"
	"sats-stacker/notifier"
)

// Global Variables
var log = logrus.New()
var ex exchange.Exchange
var nf notifier.Notifier

// TODO Use a struct for result and set action inside of it rather than use 2 string variables
var result string
var action string

// Set version at compile time
var Version string

func init() {
	// Setup Logging
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.InfoLevel)

	// Set Version if not passed by compiler
	if len(Version) == 0 {
		out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
		if err != nil {
			Version = "master"
		} else {
			Version = string(out)
		}
	}
}

func main() {
	usage := `
		a cli-tool to stack sats on exchanges.
`

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "debug logging",
			EnvVars: []string{"STACKER_DEBUG"},
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"validate"},
			Value:   true,
			Usage:   "dry-run",
			EnvVars: []string{"STACKER_VALIDATE", "STACKER_DRY_RUN"},
		},
		&cli.StringFlag{
			Name:    "exchange",
			Usage:   "Exchange ['kraken', 'binance']",
			EnvVars: []string{"STACKER_EXCHANGE"},
			Value:   "kraken",
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

	notifierFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "notifier",
			Usage:   "What notifier to use ['stdout','simplepush']",
			Value:   "stdout",
			EnvVars: []string{"STACKER_NOTIFIER"},
		},
		&cli.BoolFlag{
			Name:    "sp-encrypt",
			Value:   true,
			Usage:   "Simplepush: If set, the message will be sent end-to-end encrypted with the provided Password/Salt. If false, the message is sent unencrypted.",
			EnvVars: []string{"STACKER_SP_ENCRYPT"},
		},
		&cli.StringFlag{
			Name:    "sp-key",
			Usage:   "Simplepush: Your simplepush.io Key",
			Value:   "",
			EnvVars: []string{"STACKER_SP_KEY"},
		},
		&cli.StringFlag{
			Name:    "sp-event",
			Usage:   "Simplepush: The event the message should be associated with",
			Value:   "",
			EnvVars: []string{"STACKER_SP_EVENT"},
		},
		&cli.StringFlag{
			Name:    "sp-password",
			Usage:   "Simplepush: Encryption Password",
			Value:   "",
			EnvVars: []string{"STACKER_SP_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    "sp-salt",
			Usage:   "Simplepush: The salt for the encrypted message",
			Value:   "",
			EnvVars: []string{"STACKER_SP_SALT"},
		},
	}

	stackCommand := cli.Command{
		Name:        "stack",
		Usage:       "Stack some sats",
		Description: "Stack some sats at market value, best used for DCA (Dollar Cost Averaging)",
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
		},
		Action: func(c *cli.Context) error {
			action = "stack"

			var err error

			// Initialize the exchange plugin
			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// Stack sats on the exchange selected
			result, err = ex.Stack(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	buyTheDipsCommand := cli.Command{
		Name:  "btd",
		Usage: "Buy the DIPs",
		Description: `Places a series of orders to buy the DIPs at different discounted prices
`,
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:     "budget",
				Usage:    "Budget to allocate for the DIPs, set to 0 to allocate all of the available budget",
				EnvVars:  []string{"STACKER_BTD_BUDGET"},
				Required: true,
			},
			&cli.Int64Flag{
				Name:    "dip-percentage",
				Value:   10,
				Usage:   "Initial percentage of the firt dip, the other values will be calculated",
				EnvVars: []string{"STACKER_BTD_DIP_PERCENTAGE"},
			},
			&cli.Int64Flag{
				Name:    "dip-increments",
				Value:   5,
				Usage:   "Increment of dip percentage for each order",
				EnvVars: []string{"STACKER_BTD_DIP_INCREMENTS_PERCENTAGE"},
			},
			&cli.Int64Flag{
				Name:    "n-orders",
				Value:   5,
				Usage:   "Number of DIPS orders to place",
				EnvVars: []string{"STACKER_BTD_DIP_N_ORDERS"},
			},
			&cli.Int64Flag{
				Name:    "high-price-days-modifier",
				Value:   7,
				Hidden:  true,
				Usage:   "Days behind to use to detect high-price, used to calculate a modifier to the discount percentage, the higher the gap from the high price the bigger the modifier will be.",
				EnvVars: []string{"STACKER_BTD_HIGH_PRICE_DAYS"},
			},
			&cli.Int64Flag{
				Name:    "high-price-gap-percentage",
				Value:   5,
				Usage:   "Gap between current price and high price to trigger modifier",
				EnvVars: []string{"STACKER_BTD_HIGH_PRICE_GAP_PERCENTAGE"},
			},
			&cli.StringFlag{
				Name:     "fiat",
				Usage:    "Fiat to exchange",
				EnvVars:  []string{"STACKER_BTD_FIAT"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			action = "btd"

			var err error

			// Initialize the exchange plugin
			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// Place orders to try and buy DIPS on the exchange selected
			result, err = ex.BuyTheDips(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	withdrawCommand := cli.Command{
		Name:        "withdraw",
		Usage:       "Withdraw some sats",
		Description: "Withdraw some sats from the exchange.",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:     "max-fee",
				Usage:    "Max fee in percentage, only withdraw if the relative fee does not exceed this limit",
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
		Action: func(c *cli.Context) error {
			action = "withdraw"
			var err error

			// Initialize the exchange plugin
			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			// Withdraw Funds from the exchange selected
			result, err = ex.Withdraw(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	// Group all commands together
	allCommands := []*cli.Command{
		&stackCommand,
		&buyTheDipsCommand,
		&withdrawCommand,
	}

	// Initialize the cli app
	app := &cli.App{
		Name:     "sats-stacker",
		Version:  Version,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Francesco Ciocchetti",
				Email: "primeroznl@gmail.com",
			},
		},
		Copyright: "GPL",
		HelpName:  "SATs Stacker",
		Usage:     "stack and withdraw sats",
		UsageText: usage,
		Flags:     append(flags, notifierFlags...),
		Commands:  allCommands,
		Before: func(c *cli.Context) error {

			if c.Bool("debug") {
				log.SetLevel(logrus.DebugLevel)
			}

			// Load Exchange Plugin
			switch c.String("exchange") {
			case "kraken":
				exchange.UseLogger(log, "kraken")
				ex = &exchange.Kraken{}
			case "binance":
				return cli.Exit("Binance Exchange not implemented yet", 4)
			default:
				return cli.Exit("Only supported exchange are ['kraken', 'binance']", 1)
			}

			// Configure the exchange plugin
			err := ex.Config(c)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error Configuring the Exchange Plugin: %s", err), 1)
			}

			// Load Notification Plugin
			switch c.String("notifier") {
			case "simplepush":
				notifier.UseLogger(log, "simplepush")
				nf = &notifier.SimplePush{}
			case "stdout":
				notifier.UseLogger(log, "stdout")
				nf = &notifier.Stdout{}
			default:
				return cli.Exit("Only supported notifiers are ['stdout', 'simplepush']", 1)
			}

			// Configure the notification plugin
			err = nf.Config(c)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error Configuring the Notification Plugin: %s", err), 1)
			}

			return nil
		},
		After: func(c *cli.Context) error {
			// Handle notification at the end of the CLI app run
			title := fmt.Sprintf("%s - %s Sats",
				strings.Title(c.String("exchange")),
				strings.Title(action),
			)

			// Do not notify if result is not set ( for example if the required args where not specified )
			// TODO Add support for notification on errors
			if result != "" {
				err := nf.Notify(title, result)

				if err != nil {
					return cli.Exit(fmt.Sprintf("Notification Error: %s", err), 1)
				}
			}

			return nil
		},
	}

	// Run the cli App
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
