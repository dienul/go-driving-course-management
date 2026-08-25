package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

type HealthResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"service is running"`
	Data    interface{} `json:"data" swaggertype:"object"`
}

type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"database is unavailable"`
	Errors  interface{} `json:"errors" swaggertype:"object"`
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary Check service health
// @Description Reports whether the API and PostgreSQL connection are healthy.
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "database is unavailable",
			Errors:  nil,
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Success: false,
			Message: "database is unavailable",
			Errors:  nil,
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Success: true,
		Message: "service is running",
		Data:    nil,
	})
}
