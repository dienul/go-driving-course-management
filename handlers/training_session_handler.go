package handlers

import (
	"errors"
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

type TrainingSessionHandler struct {
	service *services.TrainingSessionService
}

func NewTrainingSessionHandler(service *services.TrainingSessionService) *TrainingSessionHandler {
	return &TrainingSessionHandler{service: service}
}

func trainingSessionFailure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrResourceNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrResourceConflict), errors.Is(err, services.ErrNoActiveEnrollment),
		errors.Is(err, services.ErrTrainerSlotConflict), errors.Is(err, services.ErrNoRemainingSessions),
		errors.Is(err, services.ErrSessionAlreadyOpen), errors.Is(err, services.ErrAvailabilityUnavailable),
		errors.Is(err, services.ErrInvalidState), errors.Is(err, services.ErrAssessmentRequired),
		errors.Is(err, services.ErrEvaluationRequired):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidInput):
		utils.Error(c, http.StatusUnprocessableEntity, err.Error())
	default:
		utils.InternalServerError(c)
	}
}
