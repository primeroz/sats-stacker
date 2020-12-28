BUILD := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags || echo $(BUILD))

#GOPLUGINS = kraken binance
# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-s -w -X=main.Version=$(VERSION)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.DEFAULT_GOAL := build
build:
	@echo "  >  Building binary..."
	go build $(LDFLAGS) -o sats-stacker .
