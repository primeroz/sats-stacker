package main

import (
	"errors"
	"fmt"
	"github.com/beldur/kraken-go-api-client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"reflect"
	"sats-stacker/types"
	"strconv"
	"strings"
)

var Name string
var crypto string

func init() {
	Name = "Kraken"
	crypto = "XBT"
}

func Stack(c *cli.Context, r *types.OrderResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": Name, "action": "stack"})
	log.Info("Stacking some sats on " + Name)

	r.SetExchange(Name)

	// Pair to work on - kraken XXBTZ<FIAT>
	pair := strings.ToUpper("X" + crypto + "Z" + c.String("fiat"))

	// Get API Object , Balance and Ticker from Kraken
	api := krakenapi.New(c.String("api-key"), c.String("secret-key"))

	// Get the current balance for api-key/secret-key
	balance, err := api.Balance()
	if err != nil {
		r.SetFailed("Failed to get Balance - Check API and SECRET Keys")
		return err
	}

	// Get the current ticker for the given PAIR
	ticker, err := api.Ticker(pair)
	if err != nil {
		r.SetFailed(fmt.Sprintf("Failed to get Ticker for pair %s", pair))
		return err
	}

	// Extract Values from Kraken Responses
	ref := reflect.ValueOf(balance)
	balanceCrypto := reflect.Indirect(ref).FieldByName("X" + crypto)
	balanceFiat := reflect.Indirect(ref).FieldByName("Z" + c.String("fiat"))

	log.WithFields(logrus.Fields{
		"crypto":        crypto,
		"cryptoBalance": balanceCrypto,
		"fiat":          c.String("fiat"),
		"fiatBalance":   balanceFiat,
	}).Debug("Balance before placing the Order")

	// Define params for the Order request
	ask := ticker.GetPairTickerInfo(pair).Ask[0]
	price, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		r.SetFailed(fmt.Sprintf("Failed to get Ask price for pair %s", pair))
		return err
	}

	volume := strconv.FormatFloat((c.Float64("amount") / price), 'f', 8, 64)
	// TODO: If volume < 0.001 then error -this is the minimum kraken order volume
	if fVolume, _ := strconv.ParseFloat(volume, 64); fVolume < 0.001 {
		r.SetFailed(fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume))
		return errors.New(fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume))
	}

	var orderType string
	switch otype := c.String("order-type"); strings.ToLower(otype) {
	case "market":
		orderType = "market"
	case "limit":
		orderType = "limit"
	default:
		r.SetFailed(fmt.Sprintf("Unsupporter order type %s , only ['limit', 'market']", otype))
		return errors.New("Unsupported order type " + otype)
	}

	args := make(map[string]string)
	args["validate"] = strconv.FormatBool(c.Bool("validate"))
	args["price"] = fmt.Sprintf("%f", price) // for Market order this is not used
	args["oflags"] = "fciq"                  // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	log.WithFields(logrus.Fields{
		"pair":       pair,
		"type":       "buy",
		"orderType":  orderType,
		"volume":     volume,
		"price":      args["price"],
		"validate":   args["validate"],
		"orderFlags": args["oflags"],
	}).Debug("Order to execute")

	// Place the Order
	order, err := api.AddOrder(pair, "buy", orderType, volume, args)

	if err != nil {
		r.SetFailed("Failed to place Order")
		return err
	}

	if c.Bool("validate") {
		log.WithFields(logrus.Fields{
			"order":        order.Description.Order,
			"transactions": strings.Join(order.TransactionIds, ","),
		}).Debug("DRY-RUN Order Placed")
		r.SetSuccess("DRY-RUN "+order.Description.Order, strings.Join(order.TransactionIds, ","), orderType, volume, c.Float64("amount"), price)
	} else {
		log.WithFields(logrus.Fields{
			"order":        order.Description.Order,
			"transactions": strings.Join(order.TransactionIds, ","),
		}).Debug("Order Placed")
		r.SetSuccess(order.Description.Order, strings.Join(order.TransactionIds, ","), orderType, volume, c.Float64("amount"), price)
	}

	return nil
}

func Withdraw(c *cli.Context, r *types.WithdrawResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": Name, "action": "whitdraw"})
	log.Info("Whitdrawing some sats on " + Name)

	log.Debug("Not Implemented Yet")

	return nil
}
