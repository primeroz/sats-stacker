# Sats Stacker

A simple tool to stack, and withdraw, sats from exchanges. At the moment only the [Kraken](https://www.kraken.com) plugin is implemented.

This tool is intented to be run through a scheduler like _Systemd Timer_ or _Crontab_

Tool is provided as [Docker Images](https://hub.docker.com/r/primeroz/sats-stacker) and golang binaries

**Use this at your own risk and decide for yourself whether or not you want to run this tool**

This tool is loosely based on [dennisreimann/stacking-sats-kraken](https://github.com/dennisreimann/stacking-sats-kraken) , implemented in Golang for easier portability

## Configuration

### Kraken

You will need to get your Kraken API Key and Secret Key from the [Api Settings page](https://www.kraken.com/u/settings/api).

Required permissions are:
* Query Funds
* Modify Orders
* Withdraw Funds ( Only if you plan to automate the withdrawal )

**Note** that on Kraken you can only withdraw to a pre-configured withdrawal address, referenced by name.

### Setup

the _sats-stacker_ tool support being configured either via Environment variables or cli arguments.
