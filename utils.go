package main

func (r *operation) setExchange(exchange string) {
	r.exchange = exchange
}

func (r *operation) setFailed(err error, description string) {
	r.success = false
	r.err = err
	r.description = description
}

func (r *operation) setSuccess(description string) {
	r.success = true
	r.err = nil
	r.description = description
}
