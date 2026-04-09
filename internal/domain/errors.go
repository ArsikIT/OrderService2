package domain

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrOrderNotFound       = errors.New("order not found")
	ErrCancellationDenied  = errors.New("only pending orders can be cancelled")
	ErrPaymentUnavailable  = errors.New("payment service unavailable")
	ErrInvalidOrderPayload = errors.New("invalid order payload")
)
