package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type UpsertAvailabilityRequest struct {
	AvailableDate string `json:"available_date" binding:"required" example:"2026-08-31"`
	StartTime     string `json:"start_time" binding:"required" example:"08:00"`
	EndTime       string `json:"end_time" binding:"required" example:"12:00"`
}

type AvailabilityData struct {
	ID            int64                     `json:"id" example:"1"`
	TrainerID     int64                     `json:"trainer_id" example:"1"`
	AvailableDate string                    `json:"available_date" example:"2026-08-31"`
	StartTime     string                    `json:"start_time" example:"08:00"`
	EndTime       string                    `json:"end_time" example:"12:00"`
	Status        models.AvailabilityStatus `json:"status" example:"PENDING"`
	PublishedBy   *int64                    `json:"published_by"`
	PublishedAt   *time.Time                `json:"published_at"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

type ScheduleTrainerData struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Pandu Pratama"`
}

type ScheduleSlotData struct {
	StartTime string `json:"start_time" example:"08:00"`
	EndTime   string `json:"end_time" example:"10:00"`
}

type StudentScheduleData struct {
	AvailabilityID int64               `json:"availability_id" example:"1"`
	Trainer        ScheduleTrainerData `json:"trainer"`
	AvailableDate  string              `json:"available_date" example:"2026-08-31"`
	Slots          []ScheduleSlotData  `json:"slots"`
}

type AvailabilityAPIResponse struct {
	Success bool             `json:"success" example:"true"`
	Message string           `json:"message" example:"success"`
	Data    AvailabilityData `json:"data"`
}

type AvailabilityListAPIResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"success"`
	Data    []AvailabilityData `json:"data"`
}

type StudentScheduleListAPIResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"success"`
	Data    []StudentScheduleData `json:"data"`
}
