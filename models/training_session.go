package models

import "time"

type TrainingSession struct {
	ID                    int64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	EnrollmentID          int64                 `gorm:"not null;index:idx_training_sessions_enrollment_status,priority:1" json:"enrollment_id"`
	TrainerID             int64                 `gorm:"not null;index:idx_training_sessions_trainer_schedule,priority:1" json:"trainer_id"`
	TrainerAvailabilityID int64                 `gorm:"not null;index" json:"trainer_availability_id"`
	SessionNumber         int                   `gorm:"not null;index:idx_training_sessions_enrollment_status,priority:3" json:"session_number"`
	ScheduledDate         time.Time             `gorm:"type:date;not null;index:idx_training_sessions_trainer_schedule,priority:2" json:"scheduled_date"`
	StartTime             ClockTime             `gorm:"type:time;not null;index:idx_training_sessions_trainer_schedule,priority:3" json:"start_time"`
	EndTime               ClockTime             `gorm:"type:time;not null;index:idx_training_sessions_trainer_schedule,priority:4" json:"end_time"`
	Status                TrainingSessionStatus `gorm:"type:varchar(30);not null;default:SCHEDULED;index:idx_training_sessions_enrollment_status,priority:2;index:idx_training_sessions_trainer_schedule,priority:5" json:"status"`
	ActualStartedAt       *time.Time            `json:"actual_started_at"`
	ActualCompletedAt     *time.Time            `json:"actual_completed_at"`
	RescheduledFromID     *int64                `gorm:"uniqueIndex" json:"rescheduled_from_id"`
	CancelledBy           *int64                `json:"cancelled_by"`
	CancellationReason    *string               `gorm:"type:text" json:"cancellation_reason"`
	CancelledAt           *time.Time            `json:"cancelled_at"`
	CreatedAt             time.Time             `gorm:"not null" json:"created_at"`
	UpdatedAt             time.Time             `gorm:"not null" json:"updated_at"`
}

func (TrainingSession) TableName() string { return "training_sessions" }
