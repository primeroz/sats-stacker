package main

import (
  log "github.com/sirupsen/logrus"
  "os"
  "fmt"
  "time"
  "strconv"
  "reflect"

  "github.com/urfave/cli/v2"
  "github.com/beldur/kraken-go-api-client"
)

const crypto = "XBT"

func stack(c *cli.Context) error {
  stacklog := log.WithFields(log.Fields{"action":"stack"})
  stacklog.Info("Stacking some sats")

  pair := "X"+crypto+"Z"+c.String("fiat")

  // Get API Object , Balance and Ticker from Kraken
  api := krakenapi.New(c.String("api-key"), c.String("secret-key"))

  balance, err := api.Balance()
  if err != nil {
    return cli.Exit(fmt.Sprintf("Failed to get Balance: %+v", err),2)
  }

  ticker, err := api.Ticker(pair)
  if err != nil {
    return cli.Exit(fmt.Sprintf("Failed to get Ticker for pair %s: %+v", pair,err),2)
  }

  // Extract Values from Kraken Responses
  r := reflect.ValueOf(balance)
  balanceCrypto := reflect.Indirect(r).FieldByName("X"+crypto)
  balanceFiat := reflect.Indirect(r).FieldByName("Z"+c.String("fiat"))

  stacklog.WithFields(log.Fields{
    "crypto": crypto,
    "cryptoBalance": balanceCrypto,
    "fiat": c.String("fiat"),
    "fiatBalance": balanceFiat,
  }).Debug("BALANCE")

  // Define Order params
  ask := ticker.GetPairTickerInfo(pair).Ask[0]
  //bid := ticker.GetPairTickerInfo(pair).Bid[0]
  price, err := strconv.ParseFloat(ask, 64)
  if err != nil {
    return cli.Exit(fmt.Sprintf("Failed to get Ask price for pair %s: %+v", pair ,err),2)
  }
  volume := strconv.FormatFloat((c.Float64("amount") / price), 'f', 8, 64)
  // TODO: If volume < 0.001 then error -this is the minimum kraken order volume

  // TODO support for limit order ?
  orderType := "market"

  args := make(map[string]string)
  //args["validate"] = strconv.FormatBool(c.Bool("validate"))
  args["validate"] = "true"
  args["oflags"] = "fciq" // "buy" button will actually sell the quote currency in exchange for the base currency 

  stacklog.WithFields(log.Fields{
    "pair": pair,
    "type": "buy",
    "orderType": orderType,
    "volume": volume,
    "price": price,
    "validate": args["validate"],
    "orderFlags": args["oflags"],
  }).Debug("ORDER to execute")

  order, err := api.AddOrder( pair, "buy", orderType, volume, args )

  if err != nil {
    return cli.Exit(fmt.Sprintf("Failed to place Order: %+v", err),2)
  }

  stacklog.WithFields(log.Fields{
    "description": order.Description,
    "transactions": order.TransactionIds,
  }).Debug("ORDER Placed")


  return nil

  //ticker, err := api.Ticker(krakenapi.XXBTZEUR)
  //if err != nil {
    //log.Fatal(err)
  //}
}

func main() {
  flags := []cli.Flag {
      &cli.BoolFlag{
        Name:    "validate",
        Aliases: []string{"dry-run"},
        Value:   false,
        Usage:   "dry-run",
        EnvVars: []string{"STACKER_VALIDATE","STACKER_DRY_RUN"},
      },
      &cli.StringFlag{
        Name:    "api-key",
        Aliases: []string{"a"},
        Usage:   "Kraken Api Key",
        EnvVars: []string{"STACKER_KRAKEN_API_KEY"},
        Required: true,
      },
      &cli.StringFlag{
        Name:    "secret-key",
        Aliases: []string{"s"},
        Usage:   "Kraken Api Secret",
        EnvVars: []string{"STACKER_KRAKEN_API_SECRET", "STACKER_KRAKEN_API_SECRET"},
        Required: true,
      },
    }

  commands :=  []*cli.Command{
    {
      Name:  "stack",
      Usage: "Stack some sats on Kraken",
      Description: "Stack some sats on Kraken full description",
      Flags: []cli.Flag{
        &cli.Float64Flag{
          Name:    "amount",
          Usage:   "Amount of fiat to exchange",
          EnvVars: []string{"STACKER_AMOUNT"},
          Required: true,
        },
        &cli.StringFlag{
          Name:    "fiat",
          Usage:   "Fiat to exchange",
          EnvVars: []string{"STACKER_FIAT"},
          Required: true,
        },
      },
      Action: stack,
    },
  }

  app := &cli.App{
    Name: "kraken-stacker",
    Version: "0.0.1",
    Compiled: time.Now(),
    Authors: []*cli.Author{
      &cli.Author{
        Name:  "Francesco Ciocchetti",
        Email: "primeroznl@gmail.com",
      },
    },
    Copyright: "GPL",
    HelpName: "contrive",
    Usage: "demonstrate available API",
    UsageText: "contrive - demonstrating the available API",
    Flags: flags,
    Commands: commands,
  }

  //app.UseShortOptionHandling = true

  // Setup Logging
  log.SetFormatter(&log.TextFormatter{
    FullTimestamp: true,
  })
  log.SetOutput(os.Stdout)
  log.SetLevel(log.InfoLevel)
  log.SetLevel(log.DebugLevel)


  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
