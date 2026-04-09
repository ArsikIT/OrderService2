package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"order-service/internal/domain"
)

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

type CreateOrderOutput struct {
	Order   *domain.Order
	Created bool
}

type OrderUseCase struct {
	repo             OrderRepository
	paymentClient    PaymentClient
	updatesPublisher OrderUpdatesPublisher
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient, updatesPublisher OrderUpdatesPublisher) *OrderUseCase {
	return &OrderUseCase{
		repo:             repo,
		paymentClient:    paymentClient,
		updatesPublisher: updatesPublisher,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error) {
	if strings.TrimSpace(input.CustomerID) == "" || strings.TrimSpace(input.ItemName) == "" {
		return nil, domain.ErrInvalidOrderPayload
	}
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	if input.IdempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil {
			return &CreateOrderOutput{Order: existing, Created: false}, nil
		}
		if !errors.Is(err, domain.ErrOrderNotFound) {
			return nil, err
		}
	}

	order := &domain.Order{
		ID:         uuid.NewString(),
		CustomerID: input.CustomerID,
		ItemName:   input.ItemName,
		Amount:     input.Amount,
		Status:     domain.OrderStatusPending,
		CreatedAt:  time.Now().UTC(),
	}
	if input.IdempotencyKey != "" {
		order.IdempotencyKey = &input.IdempotencyKey
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	uc.publishOrderUpdate(order)

	paymentResp, err := uc.paymentClient.Authorize(ctx, AuthorizePaymentRequest{
		OrderID: order.ID,
		Amount:  order.Amount,
	})
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed)
		order.Status = domain.OrderStatusFailed
		return nil, domain.ErrPaymentUnavailable
	}

	if paymentResp.Status == "Authorized" {
		order.Status = domain.OrderStatusPaid
	} else {
		order.Status = domain.OrderStatusFailed
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}
	uc.publishOrderUpdate(order)

	return &CreateOrderOutput{Order: order, Created: true}, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.CanBeCancelled() {
		return nil, domain.ErrCancellationDenied
	}

	if err := uc.repo.UpdateStatus(ctx, id, domain.OrderStatusCancelled); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	uc.publishOrderUpdate(order)
	return order, nil
}

func (uc *OrderUseCase) publishOrderUpdate(order *domain.Order) {
	if uc.updatesPublisher == nil || order == nil {
		return
	}

	orderCopy := *order
	uc.updatesPublisher.Publish(&orderCopy)
}
