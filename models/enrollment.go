package models

import "time"

type StudentEnrollment struct {
	ID           int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID    int64            `gorm:"not null;index:idx_student_enrollments_student_status,priority:1" json:"student_id"`
	PackageID    int64            `gorm:"not null;index" json:"package_id"`
	PackageName  string           `gorm:"type:varchar(150);not null" json:"package_name"`
	PackagePrice int64            `gorm:"not null" json:"package_price"`
	TotalHours   int16            `gorm:"not null" json:"total_hours"`
	Status       EnrollmentStatus `gorm:"type:varchar(30);not null;default:PENDING_PAYMENT;index:idx_student_enrollments_student_status,priority:2" json:"status"`
	StartedAt    *time.Time       `json:"started_at"`
	CompletedAt  *time.Time       `json:"completed_at"`
	CreatedAt    time.Time        `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"not null" json:"updated_at"`
}

func (StudentEnrollment) TableName() string { return "student_enrollments" }

func (e StudentEnrollment) RequiredSessions() int {
	return int(e.TotalHours) / 2
}
