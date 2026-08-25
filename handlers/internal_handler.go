package handlers

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InternalHandler struct {
	db *gorm.DB
}

func NewInternalHandler(db *gorm.DB) *InternalHandler {
	return &InternalHandler{db: db}
}

// Health godoc
// @Summary Check internal service and database health
// @Description Reports API and PostgreSQL health. Requires environment-configured HTTP Basic Auth.
// @Tags Internal
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Security BasicAuth
// @Router /api/v1/internal/health [get]
func (h *InternalHandler) Health(c *gin.Context) {
	NewHealthHandler(h.db).Health(c)
}

// Stats godoc
// @Summary Get protected operational statistics
// @Description Returns user, enrollment, training-session, payment, certificate, and review counts. Requires environment-configured HTTP Basic Auth.
// @Tags Internal
// @Produce json
// @Success 200 {object} dto.InternalStatsAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Security BasicAuth
// @Router /api/v1/internal/stats [get]
func (h *InternalHandler) Stats(c *gin.Context) {
	var result dto.InternalStatsData
	query := `SELECT
		(SELECT COUNT(*) FROM users WHERE role = 'STUDENT') AS total_students,
		(SELECT COUNT(*) FROM users WHERE role = 'TRAINER') AS total_trainers,
		(SELECT COUNT(*) FROM users WHERE role = 'ADMIN') AS total_admins,
		(SELECT COUNT(*) FROM student_enrollments) AS total_enrollments,
		(SELECT COUNT(*) FROM student_enrollments WHERE status = 'ACTIVE') AS active_enrollments,
		(SELECT COUNT(*) FROM training_sessions) AS total_training_sessions,
		(SELECT COUNT(*) FROM training_sessions WHERE status = 'SCHEDULED') AS scheduled_training_sessions,
		(SELECT COUNT(*) FROM training_sessions WHERE status = 'IN_PROGRESS') AS in_progress_training_sessions,
		(SELECT COUNT(*) FROM training_sessions WHERE status = 'COMPLETED') AS completed_training_sessions,
		(SELECT COUNT(*) FROM payments WHERE status = 'PAID') AS paid_payments,
		(SELECT COUNT(*) FROM certificates) AS total_certificates,
		(SELECT COUNT(*) FROM trainer_reviews) AS total_trainer_reviews`
	if err := h.db.WithContext(c.Request.Context()).Raw(query).Scan(&result).Error; err != nil {
		utils.Error(c, http.StatusServiceUnavailable, "database is unavailable")
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}
