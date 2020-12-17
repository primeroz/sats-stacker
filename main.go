package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var result = orderResult{}
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
	result.setRequestType("stack")
	result.setDryRun(c.Bool("validate"))

	err := krakenStack(c, &result, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while stacking on Kraken : %s", err), 1)
	}

	notify(c)
	return nil
}

func withdraw(c *cli.Context) error {
	result.setRequestType("withdraw")
	result.setDryRun(c.Bool("validate"))

	err := krakenWithdraw(c, &result, log)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Something went wrong while withdrawing from Kraken : %s", err), 1)
	}

	fmt.Printf("Result\n")
	fmt.Printf("%#v", result)
	return nil
}

func notify(c *cli.Context) error {

	switch c.String("notifier") {
	case "stdout":
		fmt.Printf("Result\n")
		fmt.Printf("%#v", result)
	case "simplepush":
		sendMessageSP(c, &result, log)
	default:
		return nil
	}

	return nil
}

func main() {
	usage := `
		a cli-tool to stack, and withdraw, sats on Kraken exchange.

		more information on usage will follow	
`

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
		Usage:     "stack and withdraw sats for Kraken exchange",
		UsageText: usage,
		Flags:     append(flags, notifierFlags...),
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
