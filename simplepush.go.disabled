package main

import (
	"github.com/simplepush/simplepush-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sats-stacker/types"
)

func send() {
}

var Name string

func init() {
	Name = "Simplepush"
}

func Send(c *cli.Context, r *types.Notification, l *logrus.Logger) (err error) {
	// Define logging and Response values
	log := l.WithFields(logrus.Fields{"exchange": Name, "action": "stack"})
	log.Info("Stacking some sats on " + Name)

	r.SetExchange(Name)

	log.Debug("Not Implemented Yet")
	return nil
}
