package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type CreateEnrollmentRequest struct {
	PackageID int64 `json:"package_id" binding:"required,gt=0" example:"1"`
}

type PayRequest struct {
	PaymentMethod models.PaymentMethod `json:"payment_method" binding:"required,oneof=BANK_TRANSFER CASH" example:"BANK_TRANSFER"`
}

type EnrollmentCheckoutData struct {
	Enrollment models.StudentEnrollment `json:"enrollment"`
	Payment    models.Payment           `json:"payment"`
}

type PaymentCheckoutData struct {
	Enrollment models.StudentEnrollment `json:"enrollment"`
	Payment    models.Payment           `json:"payment"`
	Invoice    InvoiceData              `json:"invoice"`
}

type InvoiceData struct {
	ID            int64                    `json:"id" example:"1"`
	PaymentID     int64                    `json:"payment_id" example:"1"`
	InvoiceNumber string                   `json:"invoice_number" example:"INV-20260825-0001"`
	Student       UserResponse             `json:"student"`
	Enrollment    models.StudentEnrollment `json:"enrollment"`
	Payment       models.Payment           `json:"payment"`
	IssuedAt      time.Time                `json:"issued_at"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

type EnrollmentCheckoutAPIResponse struct {
	Success bool                   `json:"success" example:"true"`
	Message string                 `json:"message" example:"enrollment created successfully"`
	Data    EnrollmentCheckoutData `json:"data"`
}
type EnrollmentAPIResponse struct {
	Success bool                     `json:"success" example:"true"`
	Message string                   `json:"message" example:"success"`
	Data    models.StudentEnrollment `json:"data"`
}
type EnrollmentListAPIResponse struct {
	Success bool                       `json:"success" example:"true"`
	Message string                     `json:"message" example:"success"`
	Data    []models.StudentEnrollment `json:"data"`
}
type PaymentAPIResponse struct {
	Success bool           `json:"success" example:"true"`
	Message string         `json:"message" example:"success"`
	Data    models.Payment `json:"data"`
}
type PaymentListAPIResponse struct {
	Success bool             `json:"success" example:"true"`
	Message string           `json:"message" example:"success"`
	Data    []models.Payment `json:"data"`
}
type PaymentCheckoutAPIResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message" example:"payment completed successfully"`
	Data    PaymentCheckoutData `json:"data"`
}
type InvoiceAPIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    InvoiceData `json:"data"`
}
type InvoiceListAPIResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"success"`
	Data    []InvoiceData `json:"data"`
}
