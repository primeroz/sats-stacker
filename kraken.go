package main

import (
	"errors"
	"fmt"
	"github.com/beldur/kraken-go-api-client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"reflect"
	"strconv"
	"strings"
)

const name = "Kraken"
const crypto = "XBT"

func krakenStack(c *cli.Context, r *orderResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": name, "action": "stack"})
	log.Info("Stacking some sats on " + name)

	// Pair to work on - kraken XXBTZ<FIAT>
	pair := strings.ToUpper("X" + crypto + "Z" + c.String("fiat"))

	// Get API Object , Balance and Ticker from Kraken
	api := krakenapi.New(c.String("api-key"), c.String("secret-key"))

	// Get the current balance for api-key/secret-key
	balance, err := api.Balance()
	if err != nil {
		r.setFailed("Failed to get Balance - Check API and SECRET Keys")
		return err
	}

	// Get the current ticker for the given PAIR
	ticker, err := api.Ticker(pair)
	if err != nil {
		r.setFailed(fmt.Sprintf("Failed to get Ticker for pair %s", pair))
		return err
	}

	// Extract Values from Kraken Responses
	ref := reflect.ValueOf(balance)
	balanceCrypto := reflect.Indirect(ref).FieldByName("X" + crypto)
	balanceFiat := reflect.Indirect(ref).FieldByName("Z" + strings.ToUpper(c.String("fiat")))

	log.WithFields(logrus.Fields{
		"crypto":        crypto,
		"cryptoBalance": balanceCrypto,
		"fiat":          c.String("fiat"),
		"fiatBalance":   balanceFiat,
	}).Debug("Balance before placing the Order")
	r.setBalances(fmt.Sprintf("%f", balanceCrypto), fmt.Sprintf("%f", balanceFiat))

	// Define params for the Order request
	ask := ticker.GetPairTickerInfo(pair).Ask[0]
	price, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		r.setFailed(fmt.Sprintf("Failed to get Ask price for pair %s", pair))
		return err
	}

	volume := strconv.FormatFloat((c.Float64("amount") / price), 'f', 8, 64)
	// TODO: If volume < 0.001 then error -this is the minimum kraken order volume
	if fVolume, _ := strconv.ParseFloat(volume, 64); fVolume < 0.001 {
		r.setFailed(fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume))
		return errors.New(fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume))
	}

	var orderType string
	switch otype := c.String("order-type"); strings.ToLower(otype) {
	case "market":
		orderType = "market"
	case "limit":
		orderType = "limit"
	default:
		r.setFailed(fmt.Sprintf("Unsupporter order type %s , only ['limit', 'market']", otype))
		return errors.New("Unsupported order type " + otype)
	}

	args := make(map[string]string)

	var validate string
	if c.Bool("validate") {
		validate = strconv.FormatBool(c.Bool("validate"))
		args["validate"] = strconv.FormatBool(c.Bool("validate"))
	}
	args["price"] = fmt.Sprintf("%f", price) // for Market order this is not used
	args["oflags"] = "fciq"                  // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	log.WithFields(logrus.Fields{
		"pair":       pair,
		"type":       "buy",
		"orderType":  orderType,
		"volume":     volume,
		"price":      args["price"],
		"dryrun":     validate,
		"orderFlags": args["oflags"],
	}).Debug("Order to execute")

	// Place the Order
	order, err := api.AddOrder(pair, "buy", orderType, volume, args)

	if err != nil {
		r.setFailed("Failed to place Order")
		return err
	}

	log.WithFields(logrus.Fields{
		"order":        order.Description.Order,
		"transactions": strings.Join(order.TransactionIds, ","),
		"dryrun":       validate,
		"orderFlags":   args["oflags"],
	}).Debug("Order Placed")
	r.setSuccess(order.Description.Order, strings.Join(order.TransactionIds, ","), orderType, volume, c.Float64("amount"), price, crypto, c.String("fiat"))

	return nil
}

func krakenWithdraw(c *cli.Context, r *orderResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": name, "action": "whitdraw"})
	log.Info("Whitdrawing some sats on " + name)

	log.Debug("Not Implemented Yet")

	return nil
}
