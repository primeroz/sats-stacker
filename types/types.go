package types

type result struct {
	Exchange    string
	Success     bool
	Description string
}

type OrderResult struct {
	result
	OrderId      string
	OrderType    string
	AmountCrypto string
	AmountFiat   float64
	Price        float64
}

type WithdrawResult struct {
	result
	Address string
}

func (r *result) SetExchange(exchange string) {
	r.Exchange = exchange
}

func (r *result) SetFailed(description string) {
	r.Success = false
	r.Description = description
}

func (r *OrderResult) SetSuccess(description string, orderId string, orderType string, amountCrypto string, amountFiat float64, price float64) {
	r.Success = true
	r.Description = description
	r.OrderId = orderId
	r.OrderType = orderType
	r.AmountCrypto = amountCrypto
	r.AmountFiat = amountFiat
	r.Price = price
}
