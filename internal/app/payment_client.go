package app

import (
	"context"

	paymentv1 "github.com/ArsikIT/generated-proto-go/proto/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type PaymentGRPCClient struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentGRPCClient(client paymentv1.PaymentServiceClient) *PaymentGRPCClient {
	return &PaymentGRPCClient{client: client}
}

func (c *PaymentGRPCClient) Authorize(ctx context.Context, req usecase.AuthorizePaymentRequest) (*usecase.AuthorizePaymentResponse, error) {
	resp, err := c.client.ProcessPayment(ctx, &paymentv1.ProcessPaymentRequest{
		OrderId: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return nil, err
			default:
				return nil, domain.ErrPaymentUnavailable
			}
		}
		return nil, domain.ErrPaymentUnavailable
	}

	return &usecase.AuthorizePaymentResponse{
		TransactionID: resp.GetTransactionId(),
		Status:        resp.GetStatus(),
	}, nil
}
