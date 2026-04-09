package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"order-service/internal/usecase"
)

type PaymentHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentHTTPClient(baseURL string, client *http.Client) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (c *PaymentHTTPClient) Authorize(ctx context.Context, req usecase.AuthorizePaymentRequest) (*usecase.AuthorizePaymentResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("payment service error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentResp usecase.AuthorizePaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}
