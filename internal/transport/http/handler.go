package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type Handler struct {
	orderUC *usecase.OrderUseCase
}

func NewHandler(orderUC *usecase.OrderUseCase) *Handler {
	return &Handler{orderUC: orderUC}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	output, err := h.orderUC.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	statusCode := nethttp.StatusCreated
	if !output.Created {
		statusCode = nethttp.StatusOK
	}

	c.JSON(statusCode, toOrderResponse(output.Order))
}

func (h *Handler) GetOrder(c *gin.Context) {
	order, err := h.orderUC.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, toOrderResponse(order))
}

func (h *Handler) CancelOrder(c *gin.Context) {
	order, err := h.orderUC.CancelOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(nethttp.StatusOK, toOrderResponse(order))
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidOrderPayload):
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrOrderNotFound):
		c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrCancellationDenied):
		c.JSON(nethttp.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrPaymentUnavailable):
		c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
	default:
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
