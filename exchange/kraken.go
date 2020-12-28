package exchange

import (
	"errors"
	"fmt"
	"github.com/beldur/kraken-go-api-client"
	"github.com/sirupsen/logrus"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

type Kraken struct {
	Name      string
	ApiKey    string
	SecretKey string
	Crypto    string
}

//Config the Kraken object

func (k *Kraken) Config(apiKey string, secretKey string) error {
	k.Name = strings.ToTitle("kraken")
	k.ApiKey = apiKey
	k.SecretKey = secretKey
	k.Crypto = "XBT"

	return nil
}

func (k *Kraken) Stack(amount float64, fiat string, orderType string, dryRun bool) (result string, e error) {

	log.WithFields(logrus.Fields{
		"action": "stack",
	}).Info("Stacking some sats on " + k.Name)

	// Initialize stack action
	// Pair to work on - kraken XXBTZ<FIAT>
	pair := strings.ToUpper("X" + k.Crypto + "Z" + fiat)

	// Get API Object , Balance and Ticker from Kraken
	api := krakenapi.New(k.ApiKey, k.SecretKey)

	// Get the current balance before any stacking is done
	balance, err := api.Balance()
	if err != nil {
		return "", errors.New("Failed to get Balance. Check API and SECRET Keys")
	}
	// Extract Values from Kraken Responses
	refBalance := reflect.ValueOf(balance)
	balanceCryptoPreOrder := reflect.Indirect(refBalance).FieldByName("X" + k.Crypto).Interface().(float64)
	balanceFiatPreOrder := reflect.Indirect(refBalance).FieldByName("Z" + strings.ToUpper(fiat)).Interface().(float64)

	// Get the current ticker for the given PAIR
	ticker, err := api.Ticker(pair)
	if err != nil {
		return "", fmt.Errorf("Failed to get ticker for pair %s: %s", pair, err)
	}

	log.WithFields(logrus.Fields{
		"action":        "stack",
		"crypto":        k.Crypto,
		"cryptoBalance": balanceCryptoPreOrder,
		"fiat":          strings.ToUpper(fiat),
		"fiatBalance":   balanceFiatPreOrder,
	}).Debug("Balance before placing the Order")

	// Define params for the Order request
	ask := ticker.GetPairTickerInfo(pair).Ask[0]
	price, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		return "", fmt.Errorf("Failed to get Ask price for pair %s: %s", pair, err)
	}

	volume := (amount / price)
	volumeString := strconv.FormatFloat(volume, 'f', 8, 64)
	// If volume < 0.001 then error - this is the minimum kraken order volume
	if volume < 0.001 {
		return "", fmt.Errorf("Minimum volume for BTC Order on Kraken is 0.001 got %s. Consider increasing the amount of Fiat", volumeString)
	}

	switch orderType {
	case "market", "limit":
		break
	default:
		return "", fmt.Errorf("Unsupporter order type %s , only ['limit', 'market']", orderType)
	}

	args := make(map[string]string)

	validate := "false"
	if dryRun {
		validate = strconv.FormatBool(dryRun)
		args["validate"] = strconv.FormatBool(dryRun)
	}

	args["price"] = fmt.Sprintf("%f", price) // for Market order this is not used
	args["oflags"] = "fciq"                  // "buy" button will actually sell the quote currency in exchange for the base currency, pay fee in the the quote currenty ( fiat )

	log.WithFields(logrus.Fields{
		"action":     "stack",
		"pair":       pair,
		"type":       "buy",
		"orderType":  orderType,
		"volume":     volumeString,
		"price":      args["price"],
		"dryrun":     validate,
		"orderFlags": args["oflags"],
	}).Debug("Order to execute")

	// Place the Order
	order, err := api.AddOrder(pair, "buy", orderType, volumeString, args)

	if err != nil {
		return "", fmt.Errorf("Failed to place order: %s", err)
	}

	var orderId string
	if dryRun {
		orderId = "DryRun"
	} else {
		orderId = strings.Join(order.TransactionIds, ",")
	}

	log.WithFields(logrus.Fields{
		"action":     "stack",
		"order":      order.Description.Order,
		"orderId":    orderId,
		"dryrun":     validate,
		"orderType":  orderType,
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
		strings.Title(orderType),
		strings.ToUpper(k.Crypto),
		balanceCryptoPreOrder,
		strings.ToUpper(fiat),
		balanceFiatPreOrder,
		ask,
		order.Description.Order,
		orderId,
	)

	return orderResultMessage, nil
}

func (k *Kraken) Withdraw(address string, maxFee float64, dryRun bool) (result string, e error) {
	log.WithFields(logrus.Fields{
		"action": "withdraw",
	}).Info("Whitdrawing sats from " + k.Name)

	// Get API Object
	api := krakenapi.New(k.ApiKey, k.SecretKey)

	withdrawInfo, err := api.WithdrawInfo(k.Crypto, address, new(big.Float).SetFloat64(0))

	if err != nil {
		return "", err
	}

	// Calculate relative fee
	limitWithdraw := withdrawInfo.Limit
	feeWithdraw := withdrawInfo.Fee
	relativeFee := new(big.Float).Quo(&feeWithdraw, &limitWithdraw)
	relativeFeeHumanReadable := new(big.Float).Mul(new(big.Float).SetFloat64(100), relativeFee)

	// Place Withdrawal when fee is low enough ( relatively )
	checkMaxFee := relativeFee.Cmp(new(big.Float).SetFloat64(maxFee / 100.0))
	// relativeFee < maxFee/100
	if checkMaxFee < 0 {
		log.WithFields(logrus.Fields{
			"action":      "withdraw",
			"amount":      fmt.Sprintf("%.8f", &limitWithdraw),
			"krakenFee":   fmt.Sprintf("%.8f", &feeWithdraw),
			"key":         address,
			"relativeFee": fmt.Sprintf("%.2f%%", relativeFeeHumanReadable),
			"maxFee":      fmt.Sprintf("%.2f%%", maxFee),
		}).Debug(fmt.Sprintf("Initiating Withdrawal of %.8f %s", &limitWithdraw, k.Crypto))
	} else {
		log.WithFields(logrus.Fields{
			"action":      "withdraw",
			"amount":      fmt.Sprintf("%.8f", &limitWithdraw),
			"krakenFee":   fmt.Sprintf("%.8f", &feeWithdraw),
			"key":         address,
			"relativeFee": fmt.Sprintf("%.2f%%", relativeFeeHumanReadable),
			"maxFee":      fmt.Sprintf("%.2f%%", maxFee),
		}).Debug(fmt.Sprintf("Fees are too high for withdrawal: %.2f%%", relativeFeeHumanReadable))

		return fmt.Sprintf(`💡 Relative fee of withdrawal: %.2f%%
❌ Fees are too high for withdrawal

👛 Kraken Address: %s
💰 Withdraw Amount %s: %.8f
🏦 Kraken Fees: %.8f`,
			relativeFeeHumanReadable,
			address,
			k.Crypto,
			&limitWithdraw,
			&feeWithdraw,
		), nil
	}

	withdrawResp, err := api.Withdraw(k.Crypto, address, &limitWithdraw)

	if err != nil {
		return "", err
	}

	withdrawReferenceId := withdrawResp.RefID

	return fmt.Sprintf(`💡 Relative fee of withdrawal: %.2f%%
✔️ Initiating Withdrawal

👛 Kraken Address: %s
💰 Withdraw Amount %s: %.8f
🏦 Kraken Fees: %.8f

📎 Transatcion: %s`,

		relativeFeeHumanReadable,
		address,
		k.Crypto,
		&limitWithdraw,
		&feeWithdraw,
		withdrawReferenceId,
	), nil
}
