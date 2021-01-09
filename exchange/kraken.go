package exchange

import (
	"errors"
	"fmt"
	"github.com/beldur/kraken-go-api-client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	//"time"
)

type Kraken struct {
	Name          string
	ApiKey        string
	SecretKey     string
	Crypto        string
	Fiat          string
	Api           *krakenapi.KrakenApi
	Pair          string
	BalanceCrypto float64
	BalanceFiat   float64
	Ticker        krakenapi.PairTickerInfo
	Ask           string
	AskFloat      float64
	UserRef       int32
}

func (k *Kraken) Config(c *cli.Context) error {
	k.Name = strings.ToTitle("kraken")
	k.ApiKey = c.String("api-key")
	k.SecretKey = c.String("secret-key")
	k.Crypto = "XBT"

	return nil
}

func (k *Kraken) Init(c *cli.Context) error {
	k.Fiat = strings.ToUpper(c.String("fiat"))
	k.Pair = "X" + k.Crypto + "Z" + k.Fiat

	k.Api = krakenapi.New(k.ApiKey, k.SecretKey)

	if c.Command.FullName() != "withdraw" {
		// Initialize the current Balance
		balance, err := k.Api.Balance()
		if err != nil {
			return errors.New("Failed to get Balance. Check API and SECRET Keys")
		}

		// Extract Values from Kraken Responses
		refBalance := reflect.ValueOf(balance)
		k.BalanceCrypto = reflect.Indirect(refBalance).FieldByName("X" + k.Crypto).Interface().(float64)
		k.BalanceFiat = reflect.Indirect(refBalance).FieldByName("Z" + k.Fiat).Interface().(float64)

		// Get the current ticker for the given PAIR
		ticker, err := k.Api.Ticker(k.Pair)
		if err != nil {
			return fmt.Errorf("Failed to get ticker for pair %s: %s", k.Pair, err)
		}

		k.Ticker = ticker.GetPairTickerInfo(k.Pair)
		k.Ask = ticker.GetPairTickerInfo(k.Pair).Ask[0]
		k.AskFloat, err = strconv.ParseFloat(k.Ask, 64)
		if err != nil {
			return fmt.Errorf("Failed to get Ask price for pair %s: %s", k.Pair, err)
		}
	}

	return nil
}

func (k *Kraken) BuyTheDip(c *cli.Context) (result string, e error) {

	// Define a user refernce to use to identify the orders placed by us
	k.UserRef = 2021876122

	log.WithFields(logrus.Fields{
		"action":  "btd",
		"userRef": k.UserRef,
	}).Info("Trying to buy the Next DIP on " + k.Name)

	log.WithFields(logrus.Fields{
		"action":        "btd",
		"crypto":        k.Crypto,
		"cryptoBalance": k.BalanceCrypto,
		"fiat":          k.Fiat,
		"fiatBalance":   k.BalanceFiat,
		"ask":           k.Ask,
	}).Debug("Balance before any action is taken")

	// Get Closed orders for the Interval specified
	//loc, err := time.LoadLocation("Local")
	//if err != nil {
	//	return "", err
	//}
	//t := time.Now().In(loc)
	//year, month, day := t.Date()

	//var start int64
	//end := t.Unix()

	//switch interval {
	//case "daily":
	//	start = time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Unix()
	//case "weekly":
	//	start = time.Date(year, month, day-7, 0, 0, 0, 0, t.Location()).Unix()
	//case "monthly":
	//	start = time.Date(year, month-1, day, 0, 0, 0, 0, t.Location()).Unix()
	//}

	//args := make(map[string]string)
	////args["trades"] = "any position"
	//args["userref"] = userRef

	//closedOrders, err := api.TradesHistory(start, end, args)
	//if err != nil {
	//	return "", fmt.Errorf("Failed to get closed Orders: %s", err)
	//}

	//fmt.Printf("%#v", closedOrders)
	return "", fmt.Errorf("\nNot Implemented Yet")
}

func (k *Kraken) Stack(c *cli.Context) (result string, e error) {

	k.UserRef = 2021876123

	log.WithFields(logrus.Fields{
		"action":  "stack",
		"userRef": k.UserRef,
	}).Info("Stacking some sats on " + k.Name)

	log.WithFields(logrus.Fields{
		"action":        "stack",
		"crypto":        k.Crypto,
		"cryptoBalance": k.BalanceCrypto,
		"fiat":          k.Fiat,
		"fiatBalance":   k.BalanceFiat,
		"ask":           k.Ask,
	}).Debug("Balance before placing the Order")

	volume := (c.Float64("amount") / k.AskFloat)
	volumeString := strconv.FormatFloat(volume, 'f', 8, 64)
	// If volume < 0.001 then error - this is the minimum kraken order volume
	if volume < 0.001 {
		return "", fmt.Errorf("Minimum volume for BTC Order on Kraken is 0.001 got %s. Consider increasing the amount of Fiat", volumeString)
	}

	switch c.String("order-type") {
	case "market", "limit":
		break
	default:
		return "", fmt.Errorf("Unsupporter order type %s , only ['limit', 'market']", c.String("order-type"))
	}

	args := make(map[string]string)

	validate := "false"
	if c.Bool("dry-run") {
		validate = "true"
		args["validate"] = "true"
	}

	args["price"] = k.Ask   // for Market order this is not used
	args["oflags"] = "fciq" // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	log.WithFields(logrus.Fields{
		"action":     "stack",
		"pair":       k.Pair,
		"type":       "buy",
		"orderType":  c.String("order-type"),
		"volume":     volumeString,
		"price":      args["price"],
		"dryrun":     validate,
		"orderFlags": args["oflags"],
	}).Debug("Order to execute")

	// Place the Order
	order, err := k.Api.AddOrder(k.Pair, "buy", c.String("order-type"), volumeString, args)

	if err != nil {
		return "", fmt.Errorf("Failed to place order: %s", err)
	}

	var orderId string
	if c.Bool("dry-run") {
		orderId = "DRY-RUN"
	} else {
		orderId = strings.Join(order.TransactionIds, ",")
	}

	log.WithFields(logrus.Fields{
		"action":     "stack",
		"order":      order.Description.Order,
		"orderId":    orderId,
		"dryrun":     validate,
		"orderType":  c.String("order-type"),
		"volume":     volumeString,
		"price":      args["price"],
		"orderFlags": args["oflags"],
	}).Info("Order Placed")

	orderResultMessage := fmt.Sprintf(`🙌 %s order successful

💰 Balance Before Order
   Crypto  %s: %f
   Fiat %s: %f

📈 Ask Price: %s

💸: %s
📎 Transatcion: %s`,
		strings.Title(c.String("order-type")),
		k.Crypto,
		k.BalanceCrypto,
		k.Fiat,
		k.BalanceFiat,
		k.Ask,
		order.Description.Order,
		orderId,
	)

	return orderResultMessage, nil
}

func (k *Kraken) Withdraw(c *cli.Context) (result string, e error) {
	log.WithFields(logrus.Fields{
		"action": "withdraw",
	}).Info("Whitdrawing sats from " + k.Name)

	withdrawInfo, err := k.Api.WithdrawInfo(k.Crypto, c.String("address"), new(big.Float).SetFloat64(0))

	if err != nil {
		return "", err
	}

	// Calculate relative fee
	limitWithdraw := withdrawInfo.Limit
	feeWithdraw := withdrawInfo.Fee
	relativeFee := new(big.Float).Quo(&feeWithdraw, &limitWithdraw)
	relativeFeeHumanReadable := new(big.Float).Mul(new(big.Float).SetFloat64(100), relativeFee)

	// Place Withdrawal when fee is low enough ( relatively )
	checkMaxFee := relativeFee.Cmp(new(big.Float).SetFloat64(c.Float64("max-fee") / 100.0))

	withdrawLog := log.WithFields(logrus.Fields{
		"action":      "withdraw",
		"amount":      fmt.Sprintf("%.8f", &limitWithdraw),
		"krakenFee":   fmt.Sprintf("%.8f", &feeWithdraw),
		"key":         c.String("address"),
		"dryRun":      strconv.FormatBool(c.Bool("dry-run")),
		"relativeFee": fmt.Sprintf("%.2f%%", relativeFeeHumanReadable),
		"maxFee":      fmt.Sprintf("%.2f%%", c.Float64("max-fee")),
	})

	// Build slice of strings for message to return
	returnMsg := []string{fmt.Sprintf("💡 Relative fee of withdrawal: %.2f%%", relativeFeeHumanReadable)}

	// relativeFee < maxFee/100
	if checkMaxFee < 0 {
		withdrawLog.Info(fmt.Sprintf("Initiating Withdrawal of %.8f %s", &limitWithdraw, k.Crypto))
		returnMsg = append(returnMsg, "✔️ Initiating Withdrawall\n")
	} else {
		withdrawLog.Info(fmt.Sprintf("Fees are too high for withdrawal: %.2f%%", relativeFeeHumanReadable))
		returnMsg = append(returnMsg, "❌ Fees are too high for withdrawal\n")
	}

	returnMsg = append(returnMsg, fmt.Sprintf("👛 Kraken Address: %s", c.String("address")))
	returnMsg = append(returnMsg, fmt.Sprintf("💰 Withdraw Amount %s: %.8f", k.Crypto, &limitWithdraw))
	returnMsg = append(returnMsg, fmt.Sprintf("🏦 Kraken Fees: %.8f", &feeWithdraw))

	if c.Bool("dry-run") {
		returnMsg = append(returnMsg, fmt.Sprintf("\n📎 Transatcion: %s", "DRY-RUN"))
	} else {
		// do perform Withdraw
		if checkMaxFee < 0 {
			withdrawResp, err := k.Api.Withdraw(k.Crypto, c.String("address"), &limitWithdraw)

			if err != nil {
				return "", err
			}

			returnMsg = append(returnMsg, fmt.Sprintf("\n📎 Transatcion: %s", withdrawResp.RefID))
		} else {
			returnMsg = append(returnMsg, fmt.Sprintf("\n📎 Transatcion: %s", "Not ready to withdraw with those fees"))
		}
	}

	return strings.Join(returnMsg, "\n"), nil
}
