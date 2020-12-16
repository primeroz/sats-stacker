#VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

#GOPLUGINS = kraken binance
# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

#.PHONY: $(GOPLUGINS)
#$(GOPLUGINS):
#	@echo "  >  Building plugin... $(<)"
#	go build -buildmode=plugin -o $(<)/$(<).so $(<)/$(<).go

.PHONY: kraken
kraken:
	@echo "  >  Building plugin... kraken"
	go build $(LDFLAGS) -buildmode=plugin -o kraken/kraken.so kraken/kraken.go

.PHONY: binance
binance:
	@echo "  >  Building plugin... binance"
	go build $(LDFLAGS) -buildmode=plugin -o binance/binance.so binance/binance.go

#build: $(GOPLUGINS)
build: kraken binance
	@echo "  >  Building binary..."
	go build $(LDFLAGS) -o bin/$(PROJECTNAME) .
