package grpc

import (
	"context"
	"errors"
	"time"

	orderv1 "github.com/ArsikIT/generated-proto-go/proto/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"order-service/internal/domain"
)

type orderQuery interface {
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
}

type orderUpdatesBroker interface {
	Subscribe(orderID string) (<-chan *domain.Order, func())
}

type Handler struct {
	orderv1.UnimplementedOrderServiceServer
	orderQuery orderQuery
	broker     orderUpdatesBroker
}

func NewHandler(orderQuery orderQuery, broker orderUpdatesBroker) *Handler {
	return &Handler{
		orderQuery: orderQuery,
		broker:     broker,
	}
}

func (h *Handler) SubscribeToOrderUpdates(req *orderv1.OrderRequest, stream orderv1.OrderService_SubscribeToOrderUpdatesServer) error {
	orderID := req.GetOrderId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := h.orderQuery.GetOrder(stream.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return status.Error(codes.NotFound, err.Error())
		}
		return status.Error(codes.Internal, "internal server error")
	}

	if err := stream.Send(toOrderStatusUpdate(order)); err != nil {
		return err
	}
	lastSentStatus := order.Status

	updates, cancel := h.broker.Subscribe(orderID)
	defer cancel()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case orderUpdate, ok := <-updates:
			if !ok {
				return nil
			}
			if orderUpdate.Status == lastSentStatus {
				continue
			}
			if err := stream.Send(toOrderStatusUpdate(orderUpdate)); err != nil {
				return err
			}
			lastSentStatus = orderUpdate.Status
		case <-ticker.C:
			currentOrder, err := h.orderQuery.GetOrder(stream.Context(), orderID)
			if err != nil {
				if errors.Is(err, domain.ErrOrderNotFound) {
					return status.Error(codes.NotFound, err.Error())
				}
				return status.Error(codes.Internal, "internal server error")
			}
			if currentOrder.Status == lastSentStatus {
				continue
			}
			if err := stream.Send(toOrderStatusUpdate(currentOrder)); err != nil {
				return err
			}
			lastSentStatus = currentOrder.Status
		}
	}
}

func toOrderStatusUpdate(order *domain.Order) *orderv1.OrderStatusUpdate {
	return &orderv1.OrderStatusUpdate{
		OrderId:   order.ID,
		Status:    order.Status,
		UpdatedAt: timestamppb.Now(),
	}
}
