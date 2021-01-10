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
	"time"
)

type Kraken struct {
	Name          string
	Action        string
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

func getTimeInfo() (time.Time, time.Time, time.Duration) {
	// Always use the local timezone
	loc, _ := time.LoadLocation("Local")

	now := time.Now().In(loc)
	year, month, day := now.Date()

	// Start is TODAY at 00:00
	start := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	// END is now
	end := now

	return start, end, end.Sub(start)
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

	k.Action = c.Command.FullName()

	if k.Action != "withdraw" {
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

func (k *Kraken) closedOrderTodayForUserRef(orders *krakenapi.ClosedOrdersResponse, c *cli.Context) (string, krakenapi.Order, error) {

	for id, v := range orders.Closed {
		if v.Status == "closed" {
			return id, v, nil
		}
	}
	return "", krakenapi.Order{}, nil
}

func (k *Kraken) createOrderArgs(c *cli.Context, volume float64, price string, longshot bool) (map[string]string, error) {

	args := make(map[string]string)

	// Simple DCA
	if k.Action == "stack" {
		args["orderType"] = "market"
	} else if k.Action == "btd" {
		args["orderType"] = "limit"
	} else if k.Action == "dda" {
		//if !longshot {
		//	//discountPercentage := c.Float64("dip-percentage")
		//	//_, _, timeDiff := getTimeInfo()

		//	// Calculate percentage modifier based on gap from Max High over last 7 days
		//	// Calculate percentage modifier based on time from 00:00 of today
		//	// numberOfSeconds in 23 timeDiff.Seconds()

		//} else {

		//}
	} else {
		return args, fmt.Errorf("Unknown Action: %s", k.Action)
	}

	validate := "false"
	if c.Bool("dry-run") {
		validate = "true"
		args["validate"] = "true"
	}

	args["userref"] = fmt.Sprintf("%d", k.UserRef)
	args["volume"] = strconv.FormatFloat(volume, 'f', 8, 64)
	args["price"] = price
	args["oflags"] = "fciq" // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	// If volume < 0.001 then error - this is the minimum kraken order volume
	if volume < 0.001 {
		return args, fmt.Errorf("Minimum volume for BTC Order on Kraken is 0.001 got %s. Consider increasing the amount of Fiat", args["volume"])
	}

	log.WithFields(logrus.Fields{
		"action":     k.Action,
		"pair":       k.Pair,
		"type":       "buy",
		"orderType":  args["orderType"],
		"volume":     args["volume"],
		"price":      args["price"],
		"dryrun":     validate,
		"orderFlags": args["oflags"],
		"userref":    args["userref"],
	}).Debug("Order to execute")

	return args, nil
}

func (k *Kraken) BuyTheDips(c *cli.Context) (result string, e error) {
	k.UserRef = 300

	log.WithFields(logrus.Fields{
		"action":  "btd",
		"userRef": k.UserRef,
	}).Info("Buying the DIPs on " + k.Name)

	log.WithFields(logrus.Fields{
		"action":        "btd",
		"crypto":        k.Crypto,
		"cryptoBalance": k.BalanceCrypto,
		"fiat":          k.Fiat,
		"fiatBalance":   k.BalanceFiat,
		"ask":           k.Ask,
		"budget":        c.Float64("budget"),
		"n-orders":      c.Int64("n-orders"),
	}).Debug("Balance before any action is taken")

	// Calculate order values from budget
	// Each _Unit_ will have the double the value of the unit before
	var totalOrderUnits int64
	var fiatValueUnit float64
	var cryptoVolumeUnit float64

	totalOrders := c.Int64("n-orders")
	for totalOrders != 0 {
		totalOrderUnits += totalOrders
		totalOrders -= 1
	}

	fiatValueUnit = c.Float64("budget") / float64(totalOrderUnits)
	cryptoVolumeUnit = fiatValueUnit / k.AskFloat

	//var dipOrders []map[string]string

	log.WithFields(logrus.Fields{
		"action":             "btd",
		"budget":             c.Float64("budget"),
		"total-sats":         fmt.Sprintf("%.8f", c.Float64("budget")/k.AskFloat),
		"dip-percentage":     c.Int64("dip-percentage"),
		"dip-increments":     c.Int64("dip-increments"),
		"n-orders":           c.Int64("n-orders"),
		"total-units":        totalOrderUnits,
		"fiat-value-unit":    fiatValueUnit,
		"crypto-volume-unit": cryptoVolumeUnit,
	}).Debug("Calculating orders")

	var dipOrders []map[string]string

	var orderNumber int64
	for orderNumber != c.Int64("n-orders") {
		//Calculate DIP Discount for this order
		discount := c.Int64("dip-percentage") + (orderNumber * c.Int64("dip-increments"))
		dipDiscountedPrice := (k.AskFloat / float64(100)) * float64(int64(100)-discount)
		dipVolume := cryptoVolumeUnit * float64(orderNumber+1)

		log.WithFields(logrus.Fields{
			"action":       "btd",
			"order-number": orderNumber + 1,
			"ask-price":    k.Ask,
			"dip-discount": discount,
			"dip-price":    dipDiscountedPrice,
			"dip-volume":   dipVolume,
		}).Debug(fmt.Sprintf("Creating discounted order %d", orderNumber+1))

		// Create Order and add to list
		dipOrderArgs, _ := k.createOrderArgs(c, dipVolume, fmt.Sprintf("%.1f", dipDiscountedPrice), false)

		// If volume < 0.001 then do not add to the list, skip to next iteration
		if dipVolume < 0.001 {
			orderNumber += 1
			continue
		}

		dipOrders = append(dipOrders, dipOrderArgs)
		log.WithFields(logrus.Fields{
			"action":       "btd",
			"order-number": orderNumber + 1,
		}).Debug("Added Order to list")

		orderNumber += 1
	}

	if len(dipOrders) == 0 {
		return "", fmt.Errorf("No Orders were added to the list")
	}

	//Cancel any open order with our UserRef
	if !c.Bool("dry-run") {
		openordersArgs := make(map[string]string)
		//openordersArgs["trades"] = "true"
		openordersArgs["userref"] = fmt.Sprintf("%d", k.UserRef)

		resp, _ := k.Api.OpenOrders(openordersArgs)
		if len(resp.Open) > 0 {

			_, err := k.Api.CancelOrder(fmt.Sprintf("%d", k.UserRef))
			fmt.Printf("\n%s\n", err)
			if err != nil {
				return "", fmt.Errorf("Failed to Cancel Orders for UserRef: %d - %s", k.UserRef, err)
			}

			log.WithFields(logrus.Fields{
				"action":  "btd",
				"userref": k.UserRef,
			}).Info(fmt.Sprintf("%d Open Orders Canceled", len(resp.Open)))
		}
	}

	//Place Orders
	orderNumber = 0
	for orderNumber != int64(len(dipOrders)) {
		thisOrder := dipOrders[orderNumber]
		order, err := k.Api.AddOrder(k.Pair, "buy", thisOrder["orderType"], thisOrder["volume"], thisOrder)
		if err != nil {
			log.WithFields(logrus.Fields{
				"action":       "btd",
				"userref":      thisOrder["userref"],
				"order-number": orderNumber + 1,
			}).Error(fmt.Sprintf("Error Creating orderNumber %d: %s", orderNumber+1, err))

			orderNumber += 1
			continue
		}

		var orderId string
		if c.Bool("dry-run") {
			orderId = "DRY-RUN"
		} else {
			orderId = strings.Join(order.TransactionIds, ",")
		}

		log.WithFields(logrus.Fields{
			"action":       "btd",
			"order":        order.Description.Order,
			"orderId":      orderId,
			"order-number": orderNumber + 1,
			"dryrun":       thisOrder["validate"],
			"orderType":    thisOrder["orderType"],
			"volume":       thisOrder["volume"],
			"price":        thisOrder["price"],
			"orderFlags":   thisOrder["oflags"],
			"userref":      thisOrder["userref"],
		}).Info(fmt.Sprintf("Order Placed %d", orderNumber+1))

		orderNumber += 1
	}

	return "", nil
}

func (k *Kraken) DollarDipAverage(c *cli.Context) (result string, e error) {

	// Define a user refernce to use to identify the orders placed by us
	// k.UserRef is the DDA order
	// k.UserRef+1 is the Long-Shot DDA order
	k.UserRef = 200

	log.WithFields(logrus.Fields{
		"action":          "dda",
		"userRef":         k.UserRef,
		"userRefLongShot": k.UserRef + 1,
	}).Info("Trying to buy the Next DIP on " + k.Name)

	log.WithFields(logrus.Fields{
		"action":        "dda",
		"crypto":        k.Crypto,
		"cryptoBalance": k.BalanceCrypto,
		"fiat":          k.Fiat,
		"fiatBalance":   k.BalanceFiat,
		"ask":           k.Ask,
	}).Debug("Balance before any action is taken")

	start, end, _ := getTimeInfo()

	closedOrdersArgs := make(map[string]string)
	closedOrdersArgs["trades"] = "false"
	closedOrdersArgs["start"] = fmt.Sprintf("%d", start.Unix())
	closedOrdersArgs["end"] = fmt.Sprintf("%d", end.Unix())
	closedOrdersArgs["closetime"] = "open"
	//closedOrdersArgs["userref"] = fmt.Sprintf("%d", k.UserRef)

	// GET Main DDA Closed Orders
	log.WithFields(logrus.Fields{
		"action":    "dda",
		"interval":  c.String("interval"),
		"trades":    closedOrdersArgs["trades"],
		"userref":   closedOrdersArgs["userref"],
		"start":     start.Format(time.RFC822),
		"end":       end.Format(time.RFC822),
		"startUnix": closedOrdersArgs["start"],
		"endUnix":   closedOrdersArgs["end"],
		"closetime": closedOrdersArgs["closetime"],
	}).Debug("Getting Main DDA closed orders")

	ddaClosedOrders, err := k.Api.ClosedOrders(closedOrdersArgs)
	if err != nil {
		return "", fmt.Errorf("Failed to get closed Orders: %s", err)
	}

	ddaOrderId, _, err := k.closedOrderTodayForUserRef(ddaClosedOrders, c)
	if err != nil {
		return "", fmt.Errorf("Failed to check closed orders: %s", err)
	}

	if ddaOrderId == "" {
		fmt.Printf("\nPlace new DDA Order")
	} else {
		log.WithFields(logrus.Fields{
			"action":          "dda",
			"closedOrderType": "DDA",
			"OrderId":         ddaOrderId,
			"UserRef":         k.UserRef,
			"start":           start.Format(time.RFC822),
			"end":             end.Format(time.RFC822),
		}).Debug("Closed order of type DDA found. Skipping")
	}

	// GET LongShot DDA Closed Orders
	//closedOrdersArgs["userref"] = fmt.Sprintf("%d", k.UserRef+1)
	log.WithFields(logrus.Fields{
		"action":    "dda",
		"interval":  c.String("interval"),
		"trades":    closedOrdersArgs["trades"],
		"userref":   closedOrdersArgs["userref"],
		"start":     start.Format(time.RFC822),
		"end":       end.Format(time.RFC822),
		"startUnix": closedOrdersArgs["start"],
		"endUnix":   closedOrdersArgs["end"],
		"closetime": closedOrdersArgs["closetime"],
	}).Debug("Getting Long-Shot DDA closed orders")

	lsDdaClosedOrders, err := k.Api.ClosedOrders(closedOrdersArgs)
	if err != nil {
		return "", fmt.Errorf("Failed to get closed Orders: %s", err)
	}

	lsDdaOrderId, _, err := k.closedOrderTodayForUserRef(lsDdaClosedOrders, c)
	if err != nil {
		return "", fmt.Errorf("Failed to check closed orders: %s", err)
	}

	if lsDdaOrderId == "" {
		fmt.Printf("\nPlace new Long-Shot DDA Order")
	} else {
		log.WithFields(logrus.Fields{
			"action":          "dda",
			"closedOrderType": "Long-Shot DDA",
			"OrderId":         lsDdaOrderId,
			"UserRef":         k.UserRef + 1,
			"start":           start.Format(time.RFC822),
			"end":             end.Format(time.RFC822),
		}).Debug("Closed order of type Long-Shot DDA found. Skipping")
	}

	return "", fmt.Errorf("\nNot Implemented Yet")
}

func (k *Kraken) Stack(c *cli.Context) (result string, e error) {

	k.UserRef = 100

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
	orderArgs, err := k.createOrderArgs(c, volume, k.Ask, false)
	if err != nil {
		return "", fmt.Errorf("Failed to create args to place order: %s", err)
	}

	// Place the Order
	order, err := k.Api.AddOrder(k.Pair, "buy", orderArgs["orderType"], orderArgs["volume"], orderArgs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action":  "btd",
			"userref": k.UserRef,
		}).Error(fmt.Sprintf("Error Creating order: %s", err))
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
		"dryrun":     orderArgs["validate"],
		"orderType":  orderArgs["orderType"],
		"volume":     orderArgs["volume"],
		"price":      orderArgs["price"],
		"orderFlags": orderArgs["oflags"],
		"userref":    orderArgs["userref"],
	}).Info("Order Placed")

	orderResultMessage := fmt.Sprintf(`🙌 %s order successful

💰 Balance Before Order
   Crypto  %s: %f
   Fiat %s: %f

📈 Ask Price: %s

💸: %s
📎 Transatcion: %s`,
		orderArgs["orderType"],
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
