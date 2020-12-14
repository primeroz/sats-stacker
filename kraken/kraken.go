package kraken

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/beldur/kraken-go-api-client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const crypto = "XBT"

var thisLog = log.WithFields(log.Fields{"exchange": "kraken"})

func GenerateCliCommand() ([]*cli.Command, error) {

	cmd := []*cli.Command{
		{
			Name:        "kraken",
			Usage:       "Kraken exchange",
			Description: "Stack some sats on Kraken full description",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "api-key",
					Usage:    "Kraken Api Key",
					EnvVars:  []string{"STACKER_KRAKEN_API_KEY"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "secret-key",
					Usage:    "Kraken Api Secret",
					EnvVars:  []string{"STACKER_KRAKEN_API_SECRET", "STACKER_KRAKEN_API_SECRET"},
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
					Usage:  "stack some sats on Kraken",
					Action: stack,
				},
				{
					Name:   "withdraw",
					Usage:  "withdraw sats from Kraken",
					Action: withdraw,
				},
			},
		},
	}

	return cmd, nil
}

func stack(c *cli.Context) error {
	stacklog := log.WithFields(log.Fields{"action": "stack"})
	stacklog.Info("Stacking some sats")

	// Pair to work on - kraken XXBTZ<FIAT>
	pair := "X" + crypto + "Z" + c.String("fiat")

	// Get API Object , Balance and Ticker from Kraken
	api := krakenapi.New(c.String("api-key"), c.String("secret-key"))

	// Get the current balance for api-key/secret-key
	balance, err := api.Balance()
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get Balance: %+v", err), 2)
	}

	// Get the current ticker for the given PAIR
	ticker, err := api.Ticker(pair)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get Ticker for pair %s: %+v", pair, err), 2)
	}

	// Extract Values from Kraken Responses
	r := reflect.ValueOf(balance)
	balanceCrypto := reflect.Indirect(r).FieldByName("X" + crypto)
	balanceFiat := reflect.Indirect(r).FieldByName("Z" + c.String("fiat"))

	stacklog.WithFields(log.Fields{
		"crypto":        crypto,
		"cryptoBalance": balanceCrypto,
		"fiat":          c.String("fiat"),
		"fiatBalance":   balanceFiat,
	}).Debug("BALANCE")

	// Define params for the Order request
	ask := ticker.GetPairTickerInfo(pair).Ask[0]
	//bid := ticker.GetPairTickerInfo(pair).Bid[0]
	price, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get Ask price for pair %s: %+v", pair, err), 2)
	}

	volume := strconv.FormatFloat((c.Float64("amount") / price), 'f', 8, 64)
	// TODO: If volume < 0.001 then error -this is the minimum kraken order volume
	if fVolume, _ := strconv.ParseFloat(volume, 64); fVolume < 0.001 {
		return cli.Exit(fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume), 1)
	}

	// TODO support for limit order ?
	orderType := "market"

	args := make(map[string]string)

	//args["validate"] = strconv.FormatBool(c.Bool("validate"))
	args["validate"] = "true" // Testing
	args["oflags"] = "fciq"   // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	stacklog.WithFields(log.Fields{
		"pair":       pair,
		"type":       "buy",
		"orderType":  orderType,
		"volume":     volume,
		"price":      price,
		"validate":   args["validate"],
		"orderFlags": args["oflags"],
	}).Debug("ORDER to execute")

	// Place the Order
	order, err := api.AddOrder(pair, "buy", orderType, volume, args)

	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to place Order: %+v", err), 2)
	}

	stacklog.WithFields(log.Fields{
		"description":  order.Description,
		"transactions": order.TransactionIds,
	}).Debug("ORDER Placed")

	return nil

	//ticker, err := api.Ticker(krakenapi.XXBTZEUR)
	//if err != nil {
	//log.Fatal(err)
	//}
}

func withdraw(c *cli.Context) error {
	wlog := thisLog.WithFields(log.Fields{"action": "withdraw"})
	wlog.Info("Withdrawing some sats")

	return nil
}
