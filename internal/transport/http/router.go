package http

import "github.com/gin-gonic/gin"

func NewRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	router.POST("/orders", handler.CreateOrder)
	router.GET("/orders/:id", handler.GetOrder)
	router.PATCH("/orders/:id/cancel", handler.CancelOrder)

	return router
}
