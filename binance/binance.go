package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sats-stacker/types"
)

var Name string
var crypto string

func init() {
	Name = "Binance"
	crypto = "BTC"
}

func Stack(c *cli.Context, r *types.OrderResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": Name, "action": "stack"})
	log.Info("Stacking some sats on " + Name)

	r.SetExchange(Name)

	log.Debug("Not Implemented Yet")
	return nil
}

func Withdraw(c *cli.Context, r *types.WithdrawResult, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": Name, "action": "whitdraw"})
	log.Info("Whitdrawing some sats on " + Name)

	log.Debug("Not Implemented Yet")

	return nil
}
