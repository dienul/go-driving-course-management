package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type SessionAssessmentRequest struct {
	SubMaterialID int64              `json:"sub_material_id" binding:"required,gt=0" example:"1"`
	SkillStatus   models.SkillStatus `json:"skill_status" binding:"required,oneof=NOT_STARTED PRACTICED MASTERED" example:"MASTERED"`
}

type UpsertSessionAssessmentsRequest struct {
	Assessments []SessionAssessmentRequest `json:"assessments" binding:"required,min=1,max=100,dive"`
}

type UpsertSessionEvaluationRequest struct {
	Predicate      models.EvaluationPredicate `json:"predicate" binding:"required,oneof=KURANG CUKUP BAIK SANGAT_BAIK" example:"BAIK"`
	Notes          string                     `json:"notes" binding:"required,max=5000" example:"Kontrol kendaraan sudah cukup baik."`
	Recommendation string                     `json:"recommendation" binding:"required,max=5000" example:"Fokus latihan parkir."`
}

type SessionAssessmentData struct {
	ID                int64              `json:"id" example:"1"`
	TrainingSessionID int64              `json:"training_session_id" example:"1"`
	MaterialID        int64              `json:"material_id" example:"1"`
	MaterialName      string             `json:"material_name" example:"Kontrol Kendaraan"`
	SubMaterialID     int64              `json:"sub_material_id" example:"2"`
	SubMaterialName   string             `json:"sub_material_name" example:"Kemudi"`
	SkillStatus       models.SkillStatus `json:"skill_status" example:"MASTERED"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type SessionEvaluationData struct {
	ID                int64                      `json:"id" example:"1"`
	TrainingSessionID int64                      `json:"training_session_id" example:"1"`
	Predicate         models.EvaluationPredicate `json:"predicate" example:"BAIK"`
	Notes             string                     `json:"notes" example:"Kontrol kendaraan sudah cukup baik."`
	Recommendation    string                     `json:"recommendation" example:"Fokus latihan parkir."`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

type PreviousTrainingSessionData struct {
	Session     TrainingSessionData     `json:"session"`
	Assessments []SessionAssessmentData `json:"assessments"`
	Evaluation  *SessionEvaluationData  `json:"evaluation"`
}

type TrainerStudentProgressData struct {
	StudentID         int64                         `json:"student_id" example:"3"`
	TrainingSessionID int64                         `json:"training_session_id" example:"2"`
	PreviousSessions  []PreviousTrainingSessionData `json:"previous_sessions"`
}

type SessionAssessmentListAPIResponse struct {
	Success bool                    `json:"success" example:"true"`
	Message string                  `json:"message" example:"success"`
	Data    []SessionAssessmentData `json:"data"`
}

type SessionEvaluationAPIResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"success"`
	Data    SessionEvaluationData `json:"data"`
}

type TrainerStudentProgressAPIResponse struct {
	Success bool                       `json:"success" example:"true"`
	Message string                     `json:"message" example:"success"`
	Data    TrainerStudentProgressData `json:"data"`
}
