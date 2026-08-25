package models

import "time"

type TrainerAvailability struct {
	ID            int64              `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainerID     int64              `gorm:"not null;index:idx_trainer_availabilities_trainer_date,priority:1" json:"trainer_id"`
	AvailableDate time.Time          `gorm:"type:date;not null;index:idx_trainer_availabilities_trainer_date,priority:2" json:"available_date"`
	StartTime     ClockTime          `gorm:"type:time;not null;index:idx_trainer_availabilities_trainer_date,priority:3" json:"start_time"`
	EndTime       ClockTime          `gorm:"type:time;not null;index:idx_trainer_availabilities_trainer_date,priority:4" json:"end_time"`
	Status        AvailabilityStatus `gorm:"type:varchar(20);not null;default:PENDING" json:"status"`
	PublishedBy   *int64             `json:"published_by"`
	PublishedAt   *time.Time         `json:"published_at"`
	CreatedAt     time.Time          `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time          `gorm:"not null" json:"updated_at"`
}

func (TrainerAvailability) TableName() string { return "trainer_availabilities" }
