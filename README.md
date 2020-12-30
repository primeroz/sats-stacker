# Sats Stacker

A simple tool to stack, and withdraw, sats from exchanges. At the moment only the [Kraken](https://www.kraken.com) plugin is implemented.

_sats-stacker_ is intented to be run through a scheduler like _Systemd Timer_ or _Crontab_ and is is provided as [Docker Images](https://hub.docker.com/r/primeroz/sats-stacker) and pre-compiled binaries.

**Use this at your own risk and decide for yourself whether or not you want to run this tool** - access to your exchange account , and so your funds, is required so you are expected to validate the Code before running it. This software is provided as it is.

This tool is loosely based on [dennisreimann/stacking-sats-kraken](https://github.com/dennisreimann/stacking-sats-kraken) , implemented in Golang for easier portability

## Supported Exchanges

#### Kraken

You will need to get your Kraken API Key and Secret Key from the [Api Settings page](https://www.kraken.com/u/settings/api).

Required permissions are:
* Query Funds
* Modify Orders
* Withdraw Funds ( Only if you plan to automate the withdrawal )

**Note** that on Kraken you can only withdraw to a pre-configured withdrawal address referenced by name. You will need to create it using the UI before you can automate withdrawal using _sats-stacker_

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
```
OPTIONS:
   --amount value                    Amount of fiat to exchange (default: 0) [$STACKER_STACK_AMOUNT]
   --fiat value                      Fiat to exchange [$STACKER_STACK_FIAT]
   --order-type value, --type value  Order type (default: "limit") [$STACKER_STACK_ORDER_TYPE]
   --help, -h                        show help (default: false)
```

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

An environment file is required to configure the service. This file should have the minimum amount of permissions since it includes your exchange keys

---
`Environment file _/etc/sats/env_`
```
STACKER_DEBUG=true
STACKER_DRY_RUN=false

# Select the exchange you want to use
STACKER_EXCHANGE=kraken

# Keys used to authenticate with exchange
STACKER_API_KEY=
STACKER_API_SECRET=

# used for buying
STACKER_STACK_FIAT=EUR
STACKER_STACK_AMOUNT=30
STACKER_STACK_ORDER_TYPE=market

# used for withdrawal
STACKER_WITHDRAW_MAX_FEE=0.5
STACKER_WITHDRAW_ADDRESS="descriptionOfWithdrawalAddress"

## SimplePush Notifier
STACKER_NOTIFIER=simplepush
STACKER_SP_PASSWORD=password
STACKER_SP_SALT=salt
STACKER_SP_KEY=key
STACKER_SP_EVENT=stacksats
```

---
`/etc/systemd/system/sats-stacker.timer`
```
[Unit]
Description=Stack sats on Kraken using sats-stacker

[Timer]
#OnCalendar=Sun 10:00
OnCalendar=10:00:00
AccuracySec=1h
Persistent=true

[Install]
WantedBy=timers.target
```

---
`/etc/systemd/system/sats-stacker.service`
```
[Unit]
Description=Stack sats on Kraken using sats-stacker
After=docker.service
Requires=docker.service

[Service]
Type=simple
ExecStartPre=-/usr/bin/docker kill stacksats
ExecStartPre=-/usr/bin/docker rm stacksats
ExecStart=/usr/bin/docker run --rm --name stacksats --env-file /etc/sats/env  --entrypoint /sats-stacker/sats-stacker primeroz/sats-stacker:v0.1.1 stack

# Only enable if you want to automate withdrawls
#TimeoutStopSec=60
#ExecStop=/usr/bin/docker run --rm --name stacksats --env-file /etc/sats/env  --entrypoint /sats-stacker/sats-stacker primeroz/sats-stacker:v0.1.1 withdraw
```

---
`enable the systemd timer`
```
$ systemctl daemon-reload
$ systemctl enable --now sats-stacker.timer
$ systemctl list-timers
```

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
./sats-stacker stack --amount 25 --fiat eur --type market 2>/dev/null

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
