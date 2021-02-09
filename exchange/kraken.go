package exchange

import (
	"errors"
	"fmt"
	"github.com/primeroz/kraken-go-api-client"
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

const MIN_BTC_AMOUNT = 0.0002

//func getTimeInfo() (time.Time, time.Time, time.Duration) {
//	// Always use the local timezone
//	loc, _ := time.LoadLocation("Local")
//
//	now := time.Now().In(loc)
//	year, month, day := now.Date()
//
//	// Start is TODAY at 00:00
//	start := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
//
//	// END is now
//	end := now
//
//	return start, end, end.Sub(start)
//}

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

//func (k *Kraken) closedOrderTodayForUserRef(orders *krakenapi.ClosedOrdersResponse, c *cli.Context) (string, krakenapi.Order, error) {
//
//	for id, v := range orders.Closed {
//		if v.Status == "closed" {
//			return id, v, nil
//		}
//	}
//	return "", krakenapi.Order{}, nil
//}

func (k *Kraken) priceModifierBasedOnGapFromHighPrice(c *cli.Context) (float64, error) {

	// 15 interval will give a week worth of data
	ohlcs, err := k.Api.OHLCWithInterval(k.Pair, "15")
	if err != nil {
		return 0.0, fmt.Errorf("Failed to get OHLC Data for pair %s: %s", k.Pair, err)
	}

	// Find highest price in the range of OHLC
	var highest float64
	for _, o := range ohlcs.OHLC {
		if o.High > highest {
			highest = o.High
		}
	}

	// max modifier is 40% ( applied to the discount price ) when the gap is >= 25%
	maxDiscountModifier := 40.0
	var discountModifier float64
	// Is the highest price from the last week more than GAP PERCENTAGE over the current ask price ?
	gapPrice := highest - k.AskFloat
	if gapPrice > 0 {
		gapPercentage := gapPrice / highest * 100
		if gapPercentage > c.Float64("high-price-gap-percentage") {

			// calculate modifier
			discountModifier = gapPercentage / 25.0 * maxDiscountModifier
			if discountModifier > maxDiscountModifier {
				discountModifier = maxDiscountModifier
			}

			log.WithFields(logrus.Fields{
				"action":             k.Action,
				"pair":               k.Pair,
				"arg-gap-interval":   "7d",
				"arg-gap-percentage": c.Float64("high-price-gap-percentage"),
				"highest":            highest,
				"ask":                k.AskFloat,
				"gap":                gapPrice,
				"gap-percentage":     gapPercentage,
				"discount-modifier":  discountModifier,
			}).Debug("Price Gap calculator")
		}
	}

	return discountModifier, nil
}

func (k *Kraken) createOrderArgs(c *cli.Context, volume float64, price string, longshot bool) (map[string]string, error) {

	args := make(map[string]string)

	// Simple DCA
	if k.Action == "stack" {
		args["orderType"] = "market"
	} else if k.Action == "btd" {
		args["orderType"] = "limit"
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

	// If volume < MIN_BTC_AMOUNT then error - this is the minimum kraken order volume
	if volume < MIN_BTC_AMOUNT {
		return args, fmt.Errorf("Minimum volume for BTC Order on Kraken is %f got %s. Consider increasing the amount of Fiat", MIN_BTC_AMOUNT, args["volume"])
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
	// TODO: Handle cancel only mode and do not cancel orders in that case so we might get the orders ?
	// TODO: Support Dollar Dip Average mode - only buy if not bought today
	// TODO: Add modifier support for weekly gap from high price
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

	totalOrders := c.Int64("n-orders")
	for totalOrders != 0 {
		totalOrderUnits += totalOrders
		totalOrders -= 1
	}

	fiatValueUnit = c.Float64("budget") / float64(totalOrderUnits)

	//var dipOrders []map[string]string

	log.WithFields(logrus.Fields{
		"action":          "btd",
		"budget":          c.Float64("budget"),
		"total-sats":      fmt.Sprintf("%.8f", c.Float64("budget")/k.AskFloat),
		"dip-percentage":  c.Int64("dip-percentage"),
		"dip-increments":  c.Int64("dip-increments"),
		"n-orders":        c.Int64("n-orders"),
		"total-units":     totalOrderUnits,
		"fiat-value-unit": fiatValueUnit,
	}).Debug("Calculating orders")

	var dipOrders []map[string]string

	var orderNumber int64
	//Calculate DIP Discount for this order
	modifier, err := k.priceModifierBasedOnGapFromHighPrice(c)
	if err != nil {
		modifier = 0.0
	}

	for orderNumber != c.Int64("n-orders") {
		// Discount based on order number
		discount := float64(c.Int64("dip-percentage") + (orderNumber * c.Int64("dip-increments")))
		// Calculate modifier to apply to discount based on the gap from the Highest Weekly price
		discountModifier := (float64(discount) * modifier) / float64(100.0)
		dipDiscountedPrice := (k.AskFloat / float64(100)) * (float64(100.0) - discount + discountModifier)
		dipVolume := (fiatValueUnit * float64(orderNumber+1)) / dipDiscountedPrice

		log.WithFields(logrus.Fields{
			"action":       "btd",
			"order-number": orderNumber + 1,
			"ask-price":    k.Ask,
			"dip-discount": discount - discountModifier,
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
			"askPrice":     k.Ask,
			"price":        thisOrder["price"],
			"orderFlags":   thisOrder["oflags"],
			"userref":      thisOrder["userref"],
		}).Info(fmt.Sprintf("Order Placed %d", orderNumber+1))

		orderNumber += 1
	}

	return "", nil
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
