package main

type result struct {
	success     bool
	description string
}

type orderResult struct {
	result
	orderId      string
	orderType    string
	amountCrypto string
	amountFiat   float64
	price        float64
}

type withdrawResult struct {
	result
	address string
}

func (r *result) setFailed(description string) {
	r.success = false
	r.description = description
}

func (r *orderResult) setSuccess(description string, orderId string, orderType string, amountCrypto string, amountFiat float64, price float64) {
	r.success = true
	r.description = description
	r.orderId = orderId
	r.orderType = orderType
	r.amountCrypto = amountCrypto
	r.amountFiat = amountFiat
	r.price = price
}
