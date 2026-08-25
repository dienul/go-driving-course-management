package handlers

import (
	"errors"
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/middleware"
	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

type EnrollmentHandler struct{ service *services.EnrollmentService }

func NewEnrollmentHandler(service *services.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{service: service}
}

func currentStudentID(c *gin.Context) (int64, bool) {
	student, ok := middleware.AuthenticatedUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "authentication required")
		return 0, false
	}
	return student.ID, true
}

func enrollmentFailure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrResourceNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrResourceConflict), errors.Is(err, services.ErrActiveEnrollment), errors.Is(err, services.ErrPaymentProcessed), errors.Is(err, services.ErrInvalidState):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidInput):
		utils.Error(c, http.StatusUnprocessableEntity, err.Error())
	default:
		utils.InternalServerError(c)
	}
}
