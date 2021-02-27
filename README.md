# Sats Stacker

A simple tool to stack, create orders to try and buy dips, withdraw sats from exchanges. At the moment only the [Kraken](https://www.kraken.com) plugin is implemented.

_sats-stacker_ is intented to be run through a scheduler like _Systemd Timer_ or _Crontab_ and is is provided as [Docker Images](https://hub.docker.com/r/primeroz/sats-stacker) and pre-compiled binaries.

**Use this at your own risk and decide for yourself whether or not you want to run this tool** - access to your exchange account , and so your funds, is required so you are expected to validate the Code before running it. This software is provided as it is.

## Supported Exchanges

#### Kraken

You will need to get your Kraken API Key and Secret Key from the [Api Settings page](https://www.kraken.com/u/settings/api).

Required permissions are:
* Query Funds
* Modify Orders
* Withdraw Funds ( Only if you plan to automate the withdrawal )

**Note** sats-stacker, on Kraken exchange, can only withdraw to a pre-configured withdrawal address referenced by name. You will need to create it using the UI before you can automate withdrawals. This is to avoid allowing this tool to create new withdrawal addresses on kraken and potentially loose you funds.

## Supported Notifiers

_sats-stacker_ support sending notifications for each _stack_ or _withdraw_ event so you always know how your stacking is going.

The infrastructure for notifications is pluggable but somewhat limited at the moment to `STDOUT` and [`SIMPLEPUSH.IO`](https://simplepush.io/)

#### STDOUT

The simplest of notifications. Just redirect `stdout` to whatever/wherever you like

#### SIMPLEPUSH

Send notification using [`SIMPLEPUSH.IO`](https://simplepush.io/) notification infrastructure with optional support for encrypting your message.

## Help

#### Global Options
```
GLOBAL OPTIONS:
   --debug, -d            debug logging (default: false) [$STACKER_DEBUG]
   --dry-run, --validate  dry-run (default: true) [$STACKER_VALIDATE, $STACKER_DRY_RUN]
   --exchange value       Exchange ['kraken', 'binance'] (default: "kraken") [$STACKER_EXCHANGE]
   --api-key value        Exchange Api Key [$STACKER_API_KEY]
   --secret-key value     Exchange Api Secret [$STACKER_SECRET_KEY, $STACKER_API_SECRET]
   --notifier value       What notifier to use ['stdout','simplepush'] (default: "stdout") [$STACKER_NOTIFIER]
   --sp-encrypt           Simplepush: If set, the message will be sent end-to-end encrypted with the provided Password/Salt. If false, the message is sent unencrypted. (default: true) [$STACKER_SP_ENCRYPT]
   --sp-key value         Simplepush: Your simplepush.io Key [$STACKER_SP_KEY]
   --sp-event value       Simplepush: The event the message should be associated with [$STACKER_SP_EVENT]
   --sp-password value    Simplepush: Encryption Password [$STACKER_SP_PASSWORD]
   --sp-salt value        Simplepush: The salt for the encrypted message [$STACKER_SP_SALT]
   --help, -h             show help (default: false)
   --version, -v          print the version (default: false)
```

#### Stack command options
the `stack` command will buy the specified _amount_ of _fiat_ in _btc_ using a market value order

This is the default `Dollar Cost Averaging` mode and will buy you some btc every time you run it.

```
OPTIONS:
   --amount value                    Amount of fiat to exchange (default: 0) [$STACKER_STACK_AMOUNT]
   --fiat value                      Fiat to exchange [$STACKER_STACK_FIAT]
   --order-type value, --type value  Order type (default: "limit") [$STACKER_STACK_ORDER_TYPE]
   --help, -h                        show help (default: false)
```

#### BTD command options
the `btd` command will place a number of orders at progressively more discounted prices and progressively higher amount of fiat, trying to catch a DIP in price.

this mode will _first delete_ any previous _btd_ order, will then place the new orders based on the budget it is allowed to spend.\
The frequency at which you run this command will define the window of time in which you will try and catch a DIP.

The first order will be placed at a discount of `dip-percentage` from the current ASK price for a given amount of fiat. Every successive order will apply a bigger discount, based on `dip-increments` for an amount of fiat higher the previous one.

```
OPTIONS:
   --budget value                     Budget to allocate for the DIPs, set to 0 to allocate all of the available budget (default: 0) [$STACKER_BTD_BUDGET]
   --dip-percentage value             Initial percentage of the firt dip, the other values will be calculated (default: 10) [$STACKER_BTD_DIP_PERCENTAGE]
   --dip-increments value             Increment of dip percentage for each order (default: 5) [$STACKER_BTD_DIP_INCREMENTS_PERCENTAGE]
   --n-orders value                   Number of DIPS orders to place (default: 5) [$STACKER_BTD_DIP_N_ORDERS]
   --high-price-gap-percentage value  Gap between current price and high price to trigger modifier (default: 5) [$STACKER_BTD_HIGH_PRICE_GAP_PERCENTAGE]
   --fiat value                       Fiat to exchange [$STACKER_BTD_FIAT]
   --help, -h                         show help (default: false)
```

##### Example BTD orders
* Budget: 500
* Fiat: Eur
* Dip-Percentage: 5%
* Dip-Ingrements: 5%
* N of Orders: 4
* Current Ask Price: 10000EUR

Given this conditions the following orders would be placed:
* 1st Order: FIAT: 50EUR - PRICE: 10000 - ( 5% of 10000) = 9500 - Volume: 0.00526315789 BTC
* 2nd Order: FIAT: 100EUR - PRICE: 10000 - ( 10% of 10000) = 9000 - Volume: 0.01111111111 BTC
* 3rd Order: FIAT: 150EUR - PRICE: 10000 - ( 15% of 10000) = 8500 - Volume: 0.01764705882 BTC
* 4th Order: FIAT: 200EUR - PRICE: 10000 - ( 20% of 10000) = 8000 - Volume: 0.025 BTC

##### Gap from High price discount
A modifier to the discount will be applied based on the gap between the _current Ask Price_ and the _Highest price_ in the last week. ( This is currently not configurable)\
The bigger the Gap between the current Ask price and the highest price in the last week the bigger the modifier is up to 40% of the actualy `dip-discount` specified.

This cannot be disabled right now but by using the `--high-price-gap-percentage` flag you can trick _sats-stacker_ to never initiate this modifier by setting it to 100 since it will only apply if the gap\
between _current Ask Price_ and _Highest price_ in the last week is > than that value.

##### Frequency of run
In order for the tool to try and catch the Dips you need to run the tool on a regular basis, the frequency at which you run it will determine the _window of time_ in which you will try to catch a dip. At every run the old orders will be deleted and new ones, based on the current Ask Price, will be created

I run the tool every 2 hour with an initial dip value of 5% , so i try to catch any dip of 5% in 2 hours of time.

The right values are entirely up to you depending on how big and how fast of a dip you are trying to catch

#### Withdraw command options
```
OPTIONS:
   --max-fee value  Max fee in percentage, only withdraw if the relative fee does not exceed this limit (default: 0) [$STACKER_WITHDRAW_MAX_FEE]
   --address value  Address to withdraw to, the actual value will depend on the exchange selected [$STACKER_WITHDRAW_ADDRESS]
   --help, -h       show help (default: false)
```

## Configuration

_sats-stacker_ support being configured either via Environment variables or cli arguments.

#### Running _sats-stacker_ using systemd timers

Example `systemd` units and timers , and environment configuration files, can be found in the [contrib/systemd](./contrib/systemd) directory

## Example Run of _sats-stacker_

```
export STACKER_DEBUG=true
export STACKER_DRY_RUN=false
export STACKER_EXCHANGE=kraken
export STACKER_NOTIFIER=stdout
export STACKER_API_KEY=YOUR_KRAKEN_KEY
export STACKER_API_SECRET=YOUR_KRAKEN_SECRET_KEY
```

`stack some sats`
```
# Stack with a too small amount of fiat

./sats-stacker stack --amount 10 --fiat eur
INFO[2020-12-30T10:46:06+01:00] Stacking some sats on KRAKEN                  action=stack exchange=kraken
DEBU[2020-12-30T10:46:07+01:00] Balance before placing the Order              action=stack crypto=XBT cryptoBalance=0.0040813 exchange=kraken fiat=EUR fiatBalance=291.1935
Minimum volume for BTC Order on Kraken is 0.001 got 0.00044237. Consider increasing the amount of Fiat
```

```
# Increase amount of fiat so that we can stack
./sats-stacker stack --amount 25 --fiat eur 2>/dev/null

Kraken - Stack Sats

🙌 Limit order successful

💰 Balance Before Order
   Crypto  XBT: 0.004081
   Fiat EUR: 291.193500

📈 Ask Price: 22608.00000

💸: buy 0.00110580 XBTEUR @ limit 22608.0
📎 Transatcion: DryRun
```

```
# Stack at market price rather than ask price ( In this case the ASK Price is printed for reference only )
./sats-stacker stack --amount 25 --fiat eur 2>/dev/null

Kraken - Stack Sats

🙌 Market order successful

💰 Balance Before Order
   Crypto  XBT: 0.004081
   Fiat EUR: 291.193500

📈 Ask Price: 22652.20000

💸: buy 0.00110365 XBTEUR @ market
📎 Transatcion: DryRun
```

---
`withdraw sats to pre-existing kraken address`
```
# Set relative fee to 0.5%  - Won't withdraw since the relative fee of withdrawal ( Kraken fee / XBT Amount ) = 3.55% > 0.5%

./sats-stacker withdraw --address "address" --max-fee 0.5 2>/dev/null

Kraken - Withdraw Sats

💡 Relative fee of withdrawal: 3.55%
❌ Fees are too high for withdrawal

👛 Kraken Address: address
💰 Withdraw Amount XBT: 0.01408130
🏦 Kraken Fees: 0.00050000

📎 Transatcion: DRY-RUN
```

### Example Run of BTD Command

```
export STACKER_DEBUG=true
export STACKER_DRY_RUN=false
export STACKER_EXCHANGE=kraken
export STACKER_API_KEY=YOUR_KRAKEN_KEY
export STACKER_API_SECRET=YOUR_KRAKEN_SECRET_KEY
```

`Create some Orders`
```
./sats-stacker btd --amount 500 --fiat eur --dip-percentage 5 --dip-increments 4 --n-orders 4

Feb 27 18:16:07 DietPi docker[18792]: time="2021-02-27T18:16:07Z" level=info msg="4 Open Orders Canceled" action=btd exchange=KRAKEN userref=300

Feb 27 18:16:07 DietPi docker[18792]: time="2021-02-27T18:16:07Z" level=info msg="Order Placed 1" action=btd askPrice=39074.00000 dryrun= exchange=KRAKEN order="buy 0.00132620 XBTEUR @ limit 37701.6" order-number=1 orderFlags=fciq orderId=AAAAAA-AAAAA-AAAAAA orderType=limit price=37701.6 userref=300 volume=0.00132620
Feb 27 18:16:07 DietPi docker[18792]: time="2021-02-27T18:16:07Z" level=info msg="Order Placed 2" action=btd askPrice=39074.00000 dryrun= exchange=KRAKEN order="buy 0.00273197 XBTEUR @ limit 36603.7" order-number=2 orderFlags=fciq orderId=BBBBBB-BBBBB-BBBBBB orderType=limit price=36603.7 userref=300 volume=0.00273197
Feb 27 18:16:08 DietPi docker[18792]: time="2021-02-27T18:16:08Z" level=info msg="Order Placed 3" action=btd askPrice=39074.00000 dryrun= exchange=KRAKEN order="buy 0.00422467 XBTEUR @ limit 35505.7" order-number=3 orderFlags=fciq orderId=CCCCCC-CCCCC-CCCCCC orderType=limit price=35505.7 userref=300 volume=0.00422467
Feb 27 18:16:08 DietPi docker[18792]: time="2021-02-27T18:16:08Z" level=info msg="Order Placed 4" action=btd askPrice=39074.00000 dryrun= exchange=KRAKEN order="buy 0.00581263 XBTEUR @ limit 34407.8" order-number=4 orderFlags=fciq orderId=DDDDDD-DDDDD-DDDDDD orderType=limit price=34407.8 userref=300 volume=0.00581263
```
