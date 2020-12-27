package exchange

// Exchange is an interface that would allow for different implementations of exchange to be used
type Exchange interface {
	Config(apiKey string, secretKey string) (err error)
	Stack(amount float64, fiat string, orderType string, dryRun bool) (result string, err error)
	Withdraw(address string, maxFee float64, dryRun bool) (result string, err error)
}
