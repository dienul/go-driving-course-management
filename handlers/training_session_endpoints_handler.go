package handlers

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

// StudentBookSession godoc
// @Summary Book an available two-hour training session
// @Description Book an available two-hour training session. Requires an ACTIVE STUDENT account.
// @Tags Student Training Sessions
// @Accept json
// @Produce json
// @Param request body dto.BookTrainingSessionRequest true "Request body"
// @Success 201 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions [post]
func (h *TrainingSessionHandler) StudentBookSession(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	var request dto.BookTrainingSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.Book(c.Request.Context(), userID, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "training session booked successfully", result)
}

// StudentListSessions godoc
// @Summary List own training session history
// @Description List own training session history. Requires an ACTIVE STUDENT account.
// @Tags Student Training Sessions
// @Produce json
// @Success 200 {object} dto.TrainingSessionListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions [get]
func (h *TrainingSessionHandler) StudentListSessions(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.ListStudent(c.Request.Context(), userID)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentGetSession godoc
// @Summary Get own training session
// @Description Get own training session. Requires an ACTIVE STUDENT account.
// @Tags Student Training Sessions
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id} [get]
func (h *TrainingSessionHandler) StudentGetSession(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.GetStudent(c.Request.Context(), userID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentCancelSession godoc
// @Summary Cancel own scheduled training session
// @Description Cancel own scheduled training session. Requires an ACTIVE STUDENT account.
// @Tags Student Training Sessions
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.CancelTrainingSessionRequest true "Request body"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id}/cancel [patch]
func (h *TrainingSessionHandler) StudentCancelSession(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	var request dto.CancelTrainingSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.CancelStudent(c.Request.Context(), userID, id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session cancelled successfully", result)
}

// StudentRescheduleSession godoc
// @Summary Reschedule own scheduled training session
// @Description Reschedule own scheduled training session. Requires an ACTIVE STUDENT account.
// @Tags Student Training Sessions
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.RescheduleTrainingSessionRequest true "Request body"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id}/reschedule [post]
func (h *TrainingSessionHandler) StudentRescheduleSession(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	var request dto.RescheduleTrainingSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.RescheduleStudent(c.Request.Context(), userID, id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session rescheduled successfully", result)
}

// AdminListSessions godoc
// @Summary List all training session history
// @Description List all training session history. Requires an ACTIVE ADMIN account.
// @Tags Admin Training Sessions
// @Produce json
// @Success 200 {object} dto.TrainingSessionListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/training-sessions [get]
func (h *TrainingSessionHandler) AdminListSessions(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminGetSession godoc
// @Summary Get training session
// @Description Get training session. Requires an ACTIVE ADMIN account.
// @Tags Admin Training Sessions
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/training-sessions/{id} [get]
func (h *TrainingSessionHandler) AdminGetSession(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminCancelSession godoc
// @Summary Cancel a scheduled training session
// @Description Cancel a scheduled training session. Requires an ACTIVE ADMIN account.
// @Tags Admin Training Sessions
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.CancelTrainingSessionRequest true "Request body"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/training-sessions/{id}/cancel [patch]
func (h *TrainingSessionHandler) AdminCancelSession(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	var request dto.CancelTrainingSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.Cancel(c.Request.Context(), userID, id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session cancelled successfully", result)
}

// AdminRescheduleSession godoc
// @Summary Reschedule a scheduled training session
// @Description Reschedule a scheduled training session. Requires an ACTIVE ADMIN account.
// @Tags Admin Training Sessions
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.RescheduleTrainingSessionRequest true "Request body"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/training-sessions/{id}/reschedule [post]
func (h *TrainingSessionHandler) AdminRescheduleSession(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	var request dto.RescheduleTrainingSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.Reschedule(c.Request.Context(), id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session rescheduled successfully", result)
}
