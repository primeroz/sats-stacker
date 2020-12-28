# Sats Stacker

A simple tool to stack, and withdraw, sats from exchanges. At the moment only the [Kraken](https://www.kraken.com) plugin is implemented.

This tool is intented to be run through a scheduler like _Systemd Timer_ or _Crontab_

Tool is provided as [Docker Images](https://hub.docker.com/r/primeroz/sats-stacker) and golang binaries

**Use this at your own risk and decide for yourself whether or not you want to run this tool**

This tool is loosely based on [dennisreimann/stacking-sats-kraken](https://github.com/dennisreimann/stacking-sats-kraken) , implemented in Golang for easier portability

## Configuration

## Help

* Global Options
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

* Stack command options
```
OPTIONS:
   --amount value                    Amount of fiat to exchange (default: 0) [$STACKER_STACK_AMOUNT]
   --fiat value                      Fiat to exchange [$STACKER_STACK_FIAT]
   --order-type value, --type value  Order type (default: "limit") [$STACKER_STACK_ORDER_TYPE]
   --help, -h                        show help (default: false)
```

* Withdraw command options
```
OPTIONS:
   --max-fee value  Max fee in percentage, only withdraw if the relative fee does not exceed this limit (default: 0) [$STACKER_WITHDRAW_MAX_FEE]
   --address value  Address to withdraw to, the actual value will depend on the exchange selected [$STACKER_WITHDRAW_ADDRESS]
   --help, -h       show help (default: false)
```

### Kraken

You will need to get your Kraken API Key and Secret Key from the [Api Settings page](https://www.kraken.com/u/settings/api).

Required permissions are:
* Query Funds
* Modify Orders
* Withdraw Funds ( Only if you plan to automate the withdrawal )

**Note** that on Kraken you can only withdraw to a pre-configured withdrawal address, referenced by name.

### Setup

the _sats-stacker_ tool support being configured either via Environment variables or cli arguments.

#### Example setup using systemd service and timer

* Environment file _/etc/sats/env
```
STACKER_DEBUG=true
STACKER_EXCHANGE=kraken

# used to authenticate with Kraken
STACKER_API_KEY=
STACKER_API_SECRET=

# used for buying
STACKER_STACK_FIAT=EUR
STACKER_STACK_AMOUNT=30
STACKER_STACK_ORDER_TYPE=market

# used for withdrawal
STACKER_WITHDRAW_MAX_FEE=0.5
STACKER_WITHDRAW_ADDRESS="descriptionOfWithdrawalAddress"

## SimplePush
STACKER_NOTIFIER=simplepush
STACKER_SP_PASSWORD=password
STACKER_SP_SALT=salt
STACKER_SP_KEY=key
STACKER_SP_EVENT=stacksats

STACKER_DRY_RUN=false
```

* timer.unit
```
[Unit]
Description=Stack sats on Kraken

[Timer]
#OnCalendar=Sun 10:00
OnCalendar=10:00:00
AccuracySec=1h
Persistent=true

[Install]
WantedBy=timers.target
```

* service.unit
```
[Unit]
Description=Stacksats on Kraken
After=docker.service
Requires=docker.service

[Service]
Type=simple
ExecStartPre=-/usr/bin/docker kill stacksats
ExecStartPre=-/usr/bin/docker rm stacksats
ExecStart=/usr/bin/docker run --rm --name stacksats --env-file /etc/sats/env  --entrypoint /sats-stacker/sats-stacker primeroz/sats-stacker:v0.1.0 stack

# Only enable if you want to automate withdrawls
#TimeoutStopSec=60
#ExecStop=/usr/bin/docker run --rm --name stacksats --env-file /etc/sats/env  --entrypoint /sats-stacker/sats-stacker primeroz/sats-stacker:v0.1.0 withdraw
```
