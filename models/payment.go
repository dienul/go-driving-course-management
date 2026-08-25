package models

import "time"

type Payment struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	EnrollmentID  int64          `gorm:"not null;uniqueIndex" json:"enrollment_id"`
	PaymentCode   string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"payment_code"`
	Amount        int64          `gorm:"not null" json:"amount"`
	PaymentMethod *PaymentMethod `gorm:"type:varchar(30)" json:"payment_method"`
	Status        PaymentStatus  `gorm:"type:varchar(20);not null;default:UNPAID;index" json:"status"`
	PaidAt        *time.Time     `json:"paid_at"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }

type Invoice struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PaymentID     int64     `gorm:"not null;uniqueIndex" json:"payment_id"`
	InvoiceNumber string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"invoice_number"`
	IssuedAt      time.Time `gorm:"not null" json:"issued_at"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
}

func (Invoice) TableName() string { return "invoices" }
