package handlers

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

// TrainerCreateAvailability godoc
// @Summary Create own pending availability
// @Description Create own pending availability. Requires an ACTIVE TRAINER account.
// @Tags Trainer Availability
// @Accept json
// @Produce json
// @Param request body dto.UpsertAvailabilityRequest true "Availability range"
// @Success 201 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/availabilities [post]
func (h *AvailabilityHandler) TrainerCreateAvailability(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	var request dto.UpsertAvailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.Create(c.Request.Context(), userID, request)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "availability created successfully", result)
}

// TrainerListAvailabilities godoc
// @Summary List own availability ranges
// @Description List own availability ranges. Requires an ACTIVE TRAINER account.
// @Tags Trainer Availability
// @Produce json
// @Success 200 {object} dto.AvailabilityListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/availabilities [get]
func (h *AvailabilityHandler) TrainerListAvailabilities(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.ListTrainer(c.Request.Context(), userID)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerGetAvailability godoc
// @Summary Get own availability range
// @Description Get own availability range. Requires an ACTIVE TRAINER account.
// @Tags Trainer Availability
// @Produce json
// @Param id path int true "Availability ID"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/availabilities/{id} [get]
func (h *AvailabilityHandler) TrainerGetAvailability(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.GetTrainer(c.Request.Context(), userID, id)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerUpdateAvailability godoc
// @Summary Update own pending availability
// @Description Update own pending availability. Requires an ACTIVE TRAINER account.
// @Tags Trainer Availability
// @Accept json
// @Produce json
// @Param id path int true "Availability ID"
// @Param request body dto.UpsertAvailabilityRequest true "Availability range"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/availabilities/{id} [put]
func (h *AvailabilityHandler) TrainerUpdateAvailability(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	var request dto.UpsertAvailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.UpdateTrainer(c.Request.Context(), userID, id, request)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "availability updated successfully", result)
}

// TrainerCancelAvailability godoc
// @Summary Cancel own availability
// @Description Cancel own availability. Requires an ACTIVE TRAINER account.
// @Tags Trainer Availability
// @Produce json
// @Param id path int true "Availability ID"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/availabilities/{id}/cancel [patch]
func (h *AvailabilityHandler) TrainerCancelAvailability(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.CancelTrainer(c.Request.Context(), userID, id)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "availability cancelled successfully", result)
}

// AdminListAvailabilities godoc
// @Summary List all trainer availability ranges
// @Description List all trainer availability ranges. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainer Availability
// @Produce json
// @Success 200 {object} dto.AvailabilityListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainer-availabilities [get]
func (h *AvailabilityHandler) AdminListAvailabilities(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminGetAvailability godoc
// @Summary Get trainer availability range
// @Description Get trainer availability range. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainer Availability
// @Produce json
// @Param id path int true "Availability ID"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainer-availabilities/{id} [get]
func (h *AvailabilityHandler) AdminGetAvailability(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminPublishAvailability godoc
// @Summary Publish pending trainer availability
// @Description Publish pending trainer availability. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainer Availability
// @Produce json
// @Param id path int true "Availability ID"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainer-availabilities/{id}/publish [post]
func (h *AvailabilityHandler) AdminPublishAvailability(c *gin.Context) {
	userID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.Publish(c.Request.Context(), userID, id)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "availability published successfully", result)
}

// AdminCancelAvailability godoc
// @Summary Cancel trainer availability
// @Description Cancel trainer availability. Requires an ACTIVE ADMIN account.
// @Tags Admin Trainer Availability
// @Produce json
// @Param id path int true "Availability ID"
// @Success 200 {object} dto.AvailabilityAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainer-availabilities/{id}/cancel [patch]
func (h *AvailabilityHandler) AdminCancelAvailability(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Cancel(c.Request.Context(), id)
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "availability cancelled successfully", result)
}

// StudentListSchedules godoc
// @Summary List published trainer schedules and available two-hour slots
// @Description List published trainer schedules and available two-hour slots. Requires an ACTIVE STUDENT account.
// @Tags Student Schedules
// @Produce json
// @Param date query string false "Filter by YYYY-MM-DD"
// @Success 200 {object} dto.StudentScheduleListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/schedules [get]
func (h *AvailabilityHandler) StudentListSchedules(c *gin.Context) {
	result, err := h.service.StudentSchedules(c.Request.Context(), c.Query("date"))
	if err != nil {
		availabilityFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}
