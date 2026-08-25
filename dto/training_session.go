package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type BookTrainingSessionRequest struct {
	TrainerAvailabilityID int64  `json:"trainer_availability_id" binding:"required,gt=0" example:"10"`
	StartTime             string `json:"start_time" binding:"required" example:"08:00"`
}

type RescheduleTrainingSessionRequest struct {
	TrainerAvailabilityID int64  `json:"trainer_availability_id" binding:"required,gt=0" example:"11"`
	StartTime             string `json:"start_time" binding:"required" example:"10:00"`
}

type CancelTrainingSessionRequest struct {
	CancellationReason string `json:"cancellation_reason" binding:"required,min=2,max=2000" example:"Student has an urgent appointment."`
}

type TrainingSessionData struct {
	ID                    int64                        `json:"id" example:"1"`
	EnrollmentID          int64                        `json:"enrollment_id" example:"1"`
	TrainerID             int64                        `json:"trainer_id" example:"2"`
	TrainerAvailabilityID int64                        `json:"trainer_availability_id" example:"10"`
	SessionNumber         int                          `json:"session_number" example:"1"`
	ScheduledDate         string                       `json:"scheduled_date" example:"2026-08-31"`
	StartTime             string                       `json:"start_time" example:"08:00"`
	EndTime               string                       `json:"end_time" example:"10:00"`
	Status                models.TrainingSessionStatus `json:"status" example:"SCHEDULED"`
	ActualStartedAt       *time.Time                   `json:"actual_started_at"`
	ActualCompletedAt     *time.Time                   `json:"actual_completed_at"`
	RescheduledFromID     *int64                       `json:"rescheduled_from_id"`
	CancelledBy           *int64                       `json:"cancelled_by"`
	CancellationReason    *string                      `json:"cancellation_reason"`
	CancelledAt           *time.Time                   `json:"cancelled_at"`
	CreatedAt             time.Time                    `json:"created_at"`
	UpdatedAt             time.Time                    `json:"updated_at"`
}

type TrainingSessionAPIResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message" example:"success"`
	Data    TrainingSessionData `json:"data"`
}

type TrainingSessionListAPIResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"success"`
	Data    []TrainingSessionData `json:"data"`
}
