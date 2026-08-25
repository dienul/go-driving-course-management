package handlers

import (
	"errors"
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/middleware"
	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

type AvailabilityHandler struct{ service *services.AvailabilityService }

func NewAvailabilityHandler(service *services.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{service: service}
}

func availabilityUserID(c *gin.Context) (int64, bool) {
	user, ok := middleware.AuthenticatedUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "authentication required")
		return 0, false
	}
	return user.ID, true
}

func availabilityFailure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrResourceNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrAvailabilityOverlap), errors.Is(err, services.ErrInvalidState), errors.Is(err, services.ErrResourceConflict):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidInput):
		utils.Error(c, http.StatusUnprocessableEntity, err.Error())
	default:
		utils.InternalServerError(c)
	}
}
