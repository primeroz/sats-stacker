package exchange

import "github.com/urfave/cli/v2"

// Exchange is an interface that would allow for different implementations of exchange to be used
type Exchange interface {
	Config(c *cli.Context) (err error)
	Init(c *cli.Context) (err error)
	Stack(c *cli.Context) (result string, err error)
	DollarDipAverage(c *cli.Context) (result string, err error)
	BuyTheDips(c *cli.Context) (result string, err error)
	Withdraw(c *cli.Context) (result string, err error)
}
