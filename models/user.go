package models

import "time"

type User struct {
	ID           int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string       `gorm:"type:varchar(150);not null" json:"name"`
	Email        string       `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string       `gorm:"type:varchar(255);not null" json:"-"`
	Role         UserRole     `gorm:"type:varchar(20);not null" json:"role"`
	Status       RecordStatus `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`
	CreatedAt    time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"not null" json:"updated_at"`
}

func (User) TableName() string { return "users" }

type StudentProfile struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex" json:"user_id"`
	Phone     *string   `gorm:"type:varchar(30)" json:"phone"`
	Address   *string   `gorm:"type:text" json:"address"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (StudentProfile) TableName() string { return "student_profiles" }

type TrainerProfile struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex" json:"user_id"`
	Phone     *string   `gorm:"type:varchar(30)" json:"phone"`
	Address   *string   `gorm:"type:text" json:"address"`
	Bio       *string   `gorm:"type:text" json:"bio"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (TrainerProfile) TableName() string { return "trainer_profiles" }
