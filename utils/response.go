package utils

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, dto.ErrorResponse{
		Success: false,
		Message: message,
		Errors:  nil,
	})
}

func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "internal server error")
}
