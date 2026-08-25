package handlers

import (
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// StudentCurrentSkills godoc
// @Summary Get current global student skills and score
// @Description Calculate the latest completed-session assessment for every active sub-material across all student enrollments.
// @Tags Student Skills
// @Produce json
// @Success 200 {object} dto.StudentSkillsAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/skills [get]
func (h *TrainingSessionHandler) StudentCurrentSkills(c *gin.Context) {
	id, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.StudentSkills(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// StudentSkillHistory godoc
// @Summary Get global student skill assessment history
// @Description List assessments from completed sessions across every student enrollment.
// @Tags Student Skills
// @Produce json
// @Success 200 {object} dto.StudentSkillHistoryAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/skills/history [get]
func (h *TrainingSessionHandler) StudentSkillHistory(c *gin.Context) {
	id, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.StudentSkillHistory(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// StudentCreateReview godoc
// @Summary Review the trainer of a completed owned session
// @Tags Student Trainer Reviews
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.UpsertTrainerReviewRequest true "Trainer review"
// @Success 201 {object} dto.TrainerReviewAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id}/review [post]
func (h *TrainingSessionHandler) StudentCreateReview(c *gin.Context) {
	student, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	var request dto.UpsertTrainerReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.CreateStudentReview(c.Request.Context(), student, id, request)
	courseResult(c, result, err, http.StatusCreated, "trainer review created successfully")
}

// StudentUpdateReview godoc
// @Summary Update an owned completed-session trainer review
// @Tags Student Trainer Reviews
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.UpsertTrainerReviewRequest true "Trainer review"
// @Success 200 {object} dto.TrainerReviewAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id}/review [put]
func (h *TrainingSessionHandler) StudentUpdateReview(c *gin.Context) {
	student, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	var request dto.UpsertTrainerReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.UpdateStudentReview(c.Request.Context(), student, id, request)
	courseResult(c, result, err, http.StatusOK, "trainer review updated successfully")
}

// StudentGetReview godoc
// @Summary Get an owned session's trainer review
// @Tags Student Trainer Reviews
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainerReviewAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/training-sessions/{id}/review [get]
func (h *TrainingSessionHandler) StudentGetReview(c *gin.Context) {
	student, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.StudentReview(c.Request.Context(), student, id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// TrainerListReviews godoc
// @Summary List reviews for the authenticated trainer
// @Tags Trainer Reviews
// @Produce json
// @Success 200 {object} dto.TrainerReviewListAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/reviews [get]
func (h *TrainingSessionHandler) TrainerListReviews(c *gin.Context) {
	id, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.TrainerReviews(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// TrainerGetReviewSummary godoc
// @Summary Calculate the authenticated trainer's average rating
// @Tags Trainer Reviews
// @Produce json
// @Success 200 {object} dto.TrainerReviewSummaryAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/reviews/summary [get]
func (h *TrainingSessionHandler) TrainerGetReviewSummary(c *gin.Context) {
	id, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.TrainerReviewSummary(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// AdminListReviews godoc
// @Summary List every trainer review
// @Tags Admin Trainer Reviews
// @Produce json
// @Success 200 {object} dto.TrainerReviewListAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainer-reviews [get]
func (h *TrainingSessionHandler) AdminListReviews(c *gin.Context) {
	result, err := h.service.AdminReviews(c.Request.Context())
	courseResult(c, result, err, http.StatusOK, "success")
}

// AdminListTrainerReviews godoc
// @Summary List reviews for a selected trainer
// @Tags Admin Trainer Reviews
// @Produce json
// @Param id path int true "Trainer ID"
// @Success 200 {object} dto.TrainerReviewListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trainers/{id}/reviews [get]
func (h *TrainingSessionHandler) AdminListTrainerReviews(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.AdminTrainerReviews(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// StudentListCertificates godoc
// @Summary List certificates owned by the authenticated student
// @Tags Student Certificates
// @Produce json
// @Success 200 {object} dto.CertificateListAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/certificates [get]
func (h *TrainingSessionHandler) StudentListCertificates(c *gin.Context) {
	id, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.StudentCertificates(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// StudentGetCertificate godoc
// @Summary Get an owned certificate with its immutable skill snapshot
// @Tags Student Certificates
// @Produce json
// @Param id path int true "Certificate ID"
// @Success 200 {object} dto.CertificateAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/certificates/{id} [get]
func (h *TrainingSessionHandler) StudentGetCertificate(c *gin.Context) {
	student, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.StudentCertificate(c.Request.Context(), student, id)
	courseResult(c, result, err, http.StatusOK, "success")
}

// AdminListCertificates godoc
// @Summary List all issued student certificates
// @Tags Admin Certificates
// @Produce json
// @Success 200 {object} dto.CertificateListAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/certificates [get]
func (h *TrainingSessionHandler) AdminListCertificates(c *gin.Context) {
	result, err := h.service.AdminCertificates(c.Request.Context())
	courseResult(c, result, err, http.StatusOK, "success")
}

// AdminGetCertificate godoc
// @Summary Get an issued student certificate
// @Tags Admin Certificates
// @Produce json
// @Param id path int true "Certificate ID"
// @Success 200 {object} dto.CertificateAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/certificates/{id} [get]
func (h *TrainingSessionHandler) AdminGetCertificate(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.AdminCertificate(c.Request.Context(), id)
	courseResult(c, result, err, http.StatusOK, "success")
}

func courseResult(c *gin.Context, result any, err error, status int, message string) {
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, status, message, result)
}
