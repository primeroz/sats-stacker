package main

type orderResult struct {
	success       bool
	description   string
	dryrun        bool
	requestType   string
	orderId       string
	orderType     string
	crypto        string
	amountCrypto  string
	balanceCrypto string
	fiat          string
	amountFiat    float64
	balanceFiat   string
	price         float64
}

func (r *orderResult) setDryRun(dryrun bool) {
	r.dryrun = dryrun
}

func (r *orderResult) setRequestType(requestType string) {
	r.requestType = requestType
}

func (r *orderResult) setBalances(crypto string, fiat string) {
	r.balanceCrypto = crypto
	r.balanceFiat = fiat
}

func (r *orderResult) setFailed(description string) {
	r.success = false
	r.description = description
}

func (r *orderResult) setSuccess(description string, orderId string, orderType string, amountCrypto string, amountFiat float64, price float64, crypto string, fiat string) {
	r.success = true
	r.description = description
	r.orderId = orderId
	r.orderType = orderType
	r.crypto = crypto
	r.amountCrypto = amountCrypto
	r.fiat = fiat
	r.amountFiat = amountFiat
	r.price = price
}
