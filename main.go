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

// Variable to hold global
var log = logrus.New()
var ex exchange.Exchange
var nf notifier.Notifier

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
		a cli-tool to stack, and withdraw, sats on exchanges.

		more information on usage will follow
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
		Description: "Stack some sats full description",
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
		Action: func(c *cli.Context) error {
			action = "stack"

			var err error

			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			result, err = ex.Stack(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	buyTheDipCommand := cli.Command{
		Name:        "btd",
		Usage:       "Buy the dip",
		Description: "Stack some sats trying to buy the dip",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:     "amount",
				Usage:    "Amount of fiat to allocate to the BTD strategy",
				EnvVars:  []string{"STACKER_BTD_AMOUNT"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "interval",
				Value:   "daily",
				Usage:   "Interval for the BTD engine ['daily','weekly','monthly']",
				EnvVars: []string{"STACKER_BTD_INTERVAL"},
			},
			&cli.Int64Flag{
				Name:    "n-orders",
				Value:   2,
				Usage:   "Max number of orders to place and complete in the interval",
				EnvVars: []string{"STACKER_BTD_NUMBER_ORDERS"},
			},
			&cli.Int64Flag{
				Name:    "orders-discount-percentage",
				Value:   10,
				Usage:   "discount to use to place orders",
				EnvVars: []string{"STACKER_BTD_ORDERS_DISCOUNT_PERCENTAGE"},
			},
			&cli.Int64Flag{
				Name:    "high-price-days",
				Value:   7,
				Usage:   "Days behind to use to detect max-price",
				EnvVars: []string{"STACKER_BTD_HIGH_PRICE_DAYS"},
			},
			&cli.Int64Flag{
				Name:    "high-price-gap-percentage",
				Value:   5,
				Usage:   "Gap between current price and high price detected to trigger reducing the dip order percentage",
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

			// This is only supported on the kraken exchange
			if c.String("exchange") != "kraken" {
				return cli.Exit("The btd is only supported on the Kraken exchange", 1)
			}

			switch c.String("interval") {
			case "daily", "weekly", "monthly":
				break
			default:
				return cli.Exit("Invalid value for <interval> flag", 1)
			}

			var err error

			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			result, err = ex.BuyTheDip(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	withdrawCommand := cli.Command{
		Name:        "withdraw",
		Usage:       "Withdraw some sats",
		Description: "Withdraw some sats from full description",
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

			err = ex.Init(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			result, err = ex.Withdraw(c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	allCommands := []*cli.Command{
		&stackCommand,
		&buyTheDipCommand,
		&withdrawCommand,
	}

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

			err = nf.Config(c)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error Configuring the Notification Plugin: %s", err), 1)
			}

			return nil
		},
		After: func(c *cli.Context) error {
			// Notify at the end of the run
			title := fmt.Sprintf("%s - %s Sats",
				strings.Title(c.String("exchange")),
				strings.Title(action),
			)

			// Do not notify if result is not set ( for example if the required args where not specified )
			if result != "" {
				err := nf.Notify(title, result)

				if err != nil {
					return cli.Exit(fmt.Sprintf("Notification Error: %s", err), 1)
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
