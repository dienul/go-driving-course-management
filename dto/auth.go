package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type RegisterRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=150" example:"Dienul Haq"`
	Email    string  `json:"email" binding:"required,email,max=255" example:"dienul@example.com"`
	Password string  `json:"password" binding:"required,min=8,max=72" example:"strong-password"`
	Phone    *string `json:"phone" binding:"omitempty,max=30" example:"081234567890"`
	Address  *string `json:"address" binding:"omitempty,max=2000" example:"Jakarta"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"dienul@example.com"`
	Password string `json:"password" binding:"required,max=72" example:"strong-password"`
}

type UserResponse struct {
	ID        int64               `json:"id" example:"1"`
	Name      string              `json:"name" example:"Dienul Haq"`
	Email     string              `json:"email" example:"dienul@example.com"`
	Role      models.UserRole     `json:"role" example:"STUDENT"`
	Status    models.RecordStatus `json:"status" example:"ACTIVE"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type RegisterData struct {
	User UserResponse `json:"user"`
}

type LoginData struct {
	User      UserResponse `json:"user"`
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType string       `json:"token_type" example:"Bearer"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"error message"`
	Errors  interface{} `json:"errors" swaggertype:"object"`
}

type RegisterResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"student registered successfully"`
	Data    RegisterData `json:"data"`
}

type LoginResponse struct {
	Success bool      `json:"success" example:"true"`
	Message string    `json:"message" example:"login successful"`
	Data    LoginData `json:"data"`
}

type UserAPIResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"success"`
	Data    UserResponse `json:"data"`
}
