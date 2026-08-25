package dto

import (
	"github.com/dienulhaq/go-driving-course-management/models"
	"time"
)

type UpsertTrainerReviewRequest struct {
	Rating   int16   `json:"rating" binding:"required,gte=1,lte=5" example:"5"`
	Feedback *string `json:"feedback" binding:"omitempty,max=2000" example:"Clear and patient instruction."`
}
type StudentSkillData struct {
	MaterialID        int64              `json:"material_id" example:"1"`
	MaterialName      string             `json:"material_name" example:"Vehicle Control"`
	SubMaterialID     int64              `json:"sub_material_id" example:"2"`
	SubMaterialName   string             `json:"sub_material_name" example:"Parking"`
	SkillStatus       models.SkillStatus `json:"skill_status" example:"PRACTICED"`
	TrainingSessionID *int64             `json:"training_session_id"`
	AssessedAt        *time.Time         `json:"assessed_at"`
}
type StudentSkillsData struct {
	StudentID         int64              `json:"student_id" example:"3"`
	SkillScore        int16              `json:"skill_score" example:"75"`
	SkillLevel        models.SkillLevel  `json:"skill_level" example:"CAPABLE"`
	TotalSubMaterials int                `json:"total_sub_materials" example:"12"`
	Skills            []StudentSkillData `json:"skills"`
}
type StudentSkillHistoryData struct {
	ID                int64              `json:"id" example:"1"`
	TrainingSessionID int64              `json:"training_session_id" example:"2"`
	SessionNumber     int                `json:"session_number" example:"1"`
	EnrollmentID      int64              `json:"enrollment_id" example:"1"`
	TrainerID         int64              `json:"trainer_id" example:"2"`
	MaterialID        int64              `json:"material_id" example:"1"`
	MaterialName      string             `json:"material_name" example:"Vehicle Control"`
	SubMaterialID     int64              `json:"sub_material_id" example:"2"`
	SubMaterialName   string             `json:"sub_material_name" example:"Parking"`
	SkillStatus       models.SkillStatus `json:"skill_status" example:"PRACTICED"`
	CompletedAt       time.Time          `json:"completed_at"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}
type TrainerReviewData struct {
	ID                int64     `json:"id" example:"1"`
	TrainingSessionID int64     `json:"training_session_id" example:"1"`
	TrainerID         int64     `json:"trainer_id" example:"2"`
	TrainerName       string    `json:"trainer_name" example:"Trainer Name"`
	StudentID         int64     `json:"student_id" example:"3"`
	StudentName       string    `json:"student_name" example:"Student Name"`
	Rating            int16     `json:"rating" example:"5"`
	Feedback          *string   `json:"feedback"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type TrainerReviewSummaryData struct {
	TrainerID     int64   `json:"trainer_id" example:"2"`
	TotalReviews  int64   `json:"total_reviews" example:"4"`
	AverageRating float64 `json:"average_rating" example:"4.75"`
}
type CertificateData struct {
	ID                int64                    `json:"id" example:"1"`
	EnrollmentID      int64                    `json:"enrollment_id" example:"1"`
	CertificateNumber string                   `json:"certificate_number" example:"CERT-20260825-0001"`
	Student           UserResponse             `json:"student"`
	Enrollment        models.StudentEnrollment `json:"enrollment"`
	SkillScore        int16                    `json:"skill_score" example:"75"`
	SkillLevel        models.SkillLevel        `json:"skill_level" example:"CAPABLE"`
	IssuedAt          time.Time                `json:"issued_at"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}
type StudentSkillsAPIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    StudentSkillsData `json:"data"`
}
type StudentSkillHistoryAPIResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    []StudentSkillHistoryData `json:"data"`
}
type TrainerReviewAPIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    TrainerReviewData `json:"data"`
}
type TrainerReviewListAPIResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []TrainerReviewData `json:"data"`
}
type TrainerReviewSummaryAPIResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    TrainerReviewSummaryData `json:"data"`
}
type CertificateAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    CertificateData `json:"data"`
}
type CertificateListAPIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    []CertificateData `json:"data"`
}
