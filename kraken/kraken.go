package main

import (
	//"errors"
	"fmt"
	//"reflect"
	//"strconv"
	//"github.com/beldur/kraken-go-api-client"
	//"github.com/sirupsen/logrus"
)

var crypto string
var Name string

func init() {
	Name = "Kraken"
	crypto = "XBT"
}

func Stack(apiKey string, secretKey string, amount float64, fiat string) (result string, err error) {
	fmt.Println("This is kraken stack")
	//thisLog := krakenLog.WithFields(logrus.Fields{"action": "stack"})
	//thisLog.Info("Stacking some sats")

	//// Pair to work on - kraken XXBTZ<FIAT>
	//pair := "X" + crypto + "Z" + c.String("fiat")

	//// Get API Object , Balance and Ticker from Kraken
	//api := krakenapi.New(c.String("api-key"), c.String("secret-key"))

	//// Get the current balance for api-key/secret-key
	//balance, err := api.Balance()
	//if err != nil {
	//	result.setFailed(err, "Failed to get Balance")
	//	return nil
	//}

	//// Get the current ticker for the given PAIR
	//ticker, err := api.Ticker(pair)
	//if err != nil {
	//	result.setFailed(err, fmt.Sprintf("Failed to get Ticker for pair %s", pair))
	//	return nil
	//}

	//// Extract Values from Kraken Responses
	//r := reflect.ValueOf(balance)
	//balanceCrypto := reflect.Indirect(r).FieldByName("X" + crypto)
	//balanceFiat := reflect.Indirect(r).FieldByName("Z" + c.String("fiat"))

	//thisLog.WithFields(logrus.Fields{
	//	"crypto":        crypto,
	//	"cryptoBalance": balanceCrypto,
	//	"fiat":          c.String("fiat"),
	//	"fiatBalance":   balanceFiat,
	//}).Debug("BALANCE")

	//// Define params for the Order request
	//ask := ticker.GetPairTickerInfo(pair).Ask[0]
	////bid := ticker.GetPairTickerInfo(pair).Bid[0]
	//price, err := strconv.ParseFloat(ask, 64)
	//if err != nil {
	//	result.setFailed(err, fmt.Sprintf("Failed to get Ask price for pair %s", pair))
	//	return nil
	//}

	//volume := strconv.FormatFloat((c.Float64("amount") / price), 'f', 8, 64)
	//// TODO: If volume < 0.001 then error -this is the minimum kraken order volume
	//if fVolume, _ := strconv.ParseFloat(volume, 64); fVolume < 0.001 {
	//	result.setFailed(errors.New("Minimum order volume too low"), fmt.Sprintf("Minimum volume for BTC Order is 0.001 got %s", volume))
	//	return nil
	//}

	//// TODO support for limit order ?
	//orderType := "market"

	//args := make(map[string]string)

	////args["validate"] = strconv.FormatBool(c.Bool("validate"))
	//args["validate"] = "true" // Testing
	//args["oflags"] = "fciq"   // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	//thisLog.WithFields(logrus.Fields{
	//	"pair":       pair,
	//	"type":       "buy",
	//	"orderType":  orderType,
	//	"volume":     volume,
	//	"price":      price,
	//	"validate":   args["validate"],
	//	"orderFlags": args["oflags"],
	//}).Debug("ORDER to execute")

	//// Place the Order
	//order, err := api.AddOrder(pair, "buy", orderType, volume, args)

	//if err != nil {
	//	result.setFailed(err, "Failed to place Order")
	//	return nil
	//}

	//thisLog.WithFields(logrus.Fields{
	//	"description":  order.Description,
	//	"transactions": order.TransactionIds,
	//}).Debug("ORDER Placed")

	//result.setSuccess(fmt.Sprintf("%#v %#v", order.Description, order.TransactionIds))

	return "blah", nil

	//ticker, err := api.Ticker(krakenapi.XXBTZEUR)
	//if err != nil {
	//log.Fatal(err)
	//}
}

func Withdraw(apiKey string, secretKey string, maxFee float64, address string) (result string, err error) {
	//thisLog := krakenLog.WithFields(logrus.Fields{"action": "whitdraw"})
	//thisLog.Info("Whitdrawing some sats")

	fmt.Println("This is kraken withdraw")
	return "", nil
}
