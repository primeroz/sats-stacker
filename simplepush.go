package main

import (
	"errors"
	"fmt"
	"github.com/simplepush/simplepush-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"strings"
)

func formatMessageSP(r *orderResult) string {
	if r.success {
		if r.requestType == "stack" {
			msg := fmt.Sprintf(`🙌 %s order successful

💰Balance %s: %s 
💰Balance %s: %s

💸 %s
📎 Transatcion %s`,
				strings.Title(r.orderType),
				strings.ToUpper(r.crypto),
				r.balanceCrypto,
				strings.ToUpper(r.fiat),
				r.balanceFiat,
				r.description,
				r.orderId)
			return msg
		}
	} else {
		return "Order Failed"
	}

	return "Failed to create Message body"
}

func formatTitleSP(r *orderResult) string {
	if r.success {
		if r.requestType == "stack" {
			msg := fmt.Sprintf("%s - Stacking Sats", name)
			if r.dryrun {
				msg = msg + " - DRYRUN"
			}
			return msg
		}
	} else {
		return fmt.Sprintf("%s Order Failed", name)
	}

	return "Failed to create Title"
}

func sendMessageSP(c *cli.Context, r *orderResult, l *logrus.Logger) error {

	if c.String("notifier") != "simplepush" {
		return errors.New("Weird, SimplePush notification not enabled but somehow `sendMessageSP` was called")
	}

	message := simplepush.Message{
		SimplePushKey: c.String("sp-key"),
		Password:      c.String("sp-password"),
		Title:         formatTitleSP(r),
		Message:       formatMessageSP(r),
		Event:         c.String("sp-event"),
		Encrypt:       c.Bool("sp-encrypt"),
		Salt:          c.String("sp-salt"),
	}

	err := simplepush.Send(message)

	if err != nil {
		return err
	}

	return nil
}
