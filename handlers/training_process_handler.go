package handlers

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

// TrainerListSessions godoc
// @Summary List assigned training sessions
// @Description List only sessions assigned to the authenticated active trainer.
// @Tags Trainer Training Process
// @Produce json
// @Success 200 {object} dto.TrainingSessionListAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions [get]
func (h *TrainingSessionHandler) TrainerListSessions(c *gin.Context) {
	trainerID, ok := availabilityUserID(c)
	if !ok {
		return
	}
	result, err := h.service.ListTrainer(c.Request.Context(), trainerID)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerGetSession godoc
// @Summary Get an assigned training session
// @Description Get a session assigned to the authenticated active trainer.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id} [get]
func (h *TrainingSessionHandler) TrainerGetSession(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.GetTrainer(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerStartSession godoc
// @Summary Start an assigned scheduled training session
// @Description Transition an assigned SCHEDULED session to IN_PROGRESS.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/start [post]
func (h *TrainingSessionHandler) TrainerStartSession(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.StartTrainer(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session started successfully", result)
}

// TrainerUpsertAssessments godoc
// @Summary Record skill assessments for an in-progress session
// @Description Batch-create or update assessments for active sub-materials while the assigned session remains IN_PROGRESS.
// @Tags Trainer Training Process
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.UpsertSessionAssessmentsRequest true "Assessment batch"
// @Success 200 {object} dto.SessionAssessmentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/assessments [put]
func (h *TrainingSessionHandler) TrainerUpsertAssessments(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	var request dto.UpsertSessionAssessmentsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.UpsertTrainerAssessments(c.Request.Context(), trainerID, id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "skill assessments saved successfully", result)
}

// TrainerGetAssessments godoc
// @Summary List an assigned session's skill assessments
// @Description List recorded assessments for an assigned training session.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.SessionAssessmentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/assessments [get]
func (h *TrainingSessionHandler) TrainerGetAssessments(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.TrainerAssessments(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerUpsertEvaluation godoc
// @Summary Record an in-progress session's general evaluation
// @Description Create or update the predicate, notes, and recommendation while the assigned session remains IN_PROGRESS.
// @Tags Trainer Training Process
// @Accept json
// @Produce json
// @Param id path int true "Training session ID"
// @Param request body dto.UpsertSessionEvaluationRequest true "Session evaluation"
// @Success 200 {object} dto.SessionEvaluationAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/evaluation [put]
func (h *TrainingSessionHandler) TrainerUpsertEvaluation(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	var request dto.UpsertSessionEvaluationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.UpsertTrainerEvaluation(c.Request.Context(), trainerID, id, request)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "session evaluation saved successfully", result)
}

// TrainerGetEvaluation godoc
// @Summary Get an assigned session's evaluation
// @Description Get the predicate, notes, and recommendation for an assigned session.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.SessionEvaluationAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/evaluation [get]
func (h *TrainingSessionHandler) TrainerGetEvaluation(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.TrainerEvaluation(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerStudentProgress godoc
// @Summary View a student's previous completed training sessions
// @Description View previous completed sessions, assessments, and evaluations across all enrollments for the student assigned to this trainer session.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainerStudentProgressAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/student-progress [get]
func (h *TrainingSessionHandler) TrainerStudentProgress(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.TrainerStudentProgress(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// TrainerCompleteSession godoc
// @Summary Complete an assigned in-progress training session
// @Description Complete an IN_PROGRESS session only after at least one skill assessment and a complete evaluation exist.
// @Tags Trainer Training Process
// @Produce json
// @Param id path int true "Training session ID"
// @Success 200 {object} dto.TrainingSessionAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/trainer/training-sessions/{id}/complete [post]
func (h *TrainingSessionHandler) TrainerCompleteSession(c *gin.Context) {
	trainerID, id, ok := trainerSessionIDs(c)
	if !ok {
		return
	}
	result, err := h.service.CompleteTrainer(c.Request.Context(), trainerID, id)
	if err != nil {
		trainingSessionFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "training session completed successfully", result)
}

func trainerSessionIDs(c *gin.Context) (int64, int64, bool) {
	trainerID, ok := availabilityUserID(c)
	if !ok {
		return 0, 0, false
	}
	id, ok := adminID(c, "id")
	if !ok {
		return 0, 0, false
	}
	return trainerID, id, true
}
