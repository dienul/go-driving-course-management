package models

import "time"

type TrainerReview struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainingSessionID int64     `gorm:"not null;uniqueIndex" json:"training_session_id"`
	Rating            int16     `gorm:"not null" json:"rating"`
	Feedback          *string   `gorm:"type:text" json:"feedback"`
	CreatedAt         time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time `gorm:"not null" json:"updated_at"`
}

func (TrainerReview) TableName() string { return "trainer_reviews" }

type Certificate struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	EnrollmentID      int64      `gorm:"not null;uniqueIndex" json:"enrollment_id"`
	CertificateNumber string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"certificate_number"`
	SkillScore        int16      `gorm:"not null" json:"skill_score"`
	SkillLevel        SkillLevel `gorm:"type:varchar(30);not null" json:"skill_level"`
	IssuedAt          time.Time  `gorm:"not null" json:"issued_at"`
	CreatedAt         time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"not null" json:"updated_at"`
}

func (Certificate) TableName() string { return "certificates" }
