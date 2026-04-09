package http

import (
	"time"

	"order-service/internal/domain"
)

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

type orderResponse struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ItemName   string    `json:"item_name"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func toOrderResponse(order *domain.Order) orderResponse {
	return orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	}
}
