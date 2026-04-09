package usecase

import (
	"context"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type PaymentClient interface {
	Authorize(ctx context.Context, req AuthorizePaymentRequest) (*AuthorizePaymentResponse, error)
}

type OrderUpdatesPublisher interface {
	Publish(order *domain.Order)
}

type AuthorizePaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type AuthorizePaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
