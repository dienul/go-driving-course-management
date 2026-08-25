package dto

import (
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
)

type CreateTrainerRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=150" example:"Pandu Pratama"`
	Email    string  `json:"email" binding:"required,email,max=255" example:"pandu@example.com"`
	Password string  `json:"password" binding:"required,min=8,max=72" example:"strong-password"`
	Phone    *string `json:"phone" binding:"omitempty,max=30" example:"081234567890"`
	Address  *string `json:"address" binding:"omitempty,max=2000" example:"Jakarta"`
	Bio      *string `json:"bio" binding:"omitempty,max=2000" example:"Trainer berpengalaman."`
}

type UpdateTrainerRequest struct {
	Name    string  `json:"name" binding:"required,min=2,max=150" example:"Pandu Pratama"`
	Email   string  `json:"email" binding:"required,email,max=255" example:"pandu@example.com"`
	Phone   *string `json:"phone" binding:"omitempty,max=30" example:"081234567890"`
	Address *string `json:"address" binding:"omitempty,max=2000" example:"Jakarta"`
	Bio     *string `json:"bio" binding:"omitempty,max=2000" example:"Trainer berpengalaman."`
}

type UpdateStatusRequest struct {
	Status models.RecordStatus `json:"status" binding:"required,oneof=ACTIVE INACTIVE" example:"ACTIVE"`
}

type UpsertPackageRequest struct {
	Name        string              `json:"name" binding:"required,min=2,max=150" example:"Pemula 6 Jam"`
	Level       models.PackageLevel `json:"level" binding:"required,oneof=PEMULA DASAR" example:"PEMULA"`
	TotalHours  int16               `json:"total_hours" binding:"required,oneof=6 8 10 12" example:"6"`
	Price       int64               `json:"price" binding:"required,gt=0" example:"900000"`
	Description *string             `json:"description" binding:"omitempty,max=2000" example:"Latihan selama 6 jam."`
}

type UpsertMaterialRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=200" example:"Pengenalan Kendaraan"`
	Description *string `json:"description" binding:"omitempty,max=2000" example:"Mengenal kendaraan."`
	Sequence    int     `json:"sequence" binding:"required,gt=0" example:"1"`
}

type UpsertSubMaterialRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=200" example:"Kontrol Kemudi"`
	Description *string `json:"description" binding:"omitempty,max=2000" example:"Mengendalikan kemudi."`
	Sequence    int     `json:"sequence" binding:"required,gt=0" example:"1"`
}

type TrainerProfileData struct {
	ID        int64     `json:"id" example:"1"`
	UserID    int64     `json:"user_id" example:"1"`
	Phone     *string   `json:"phone"`
	Address   *string   `json:"address"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StudentProfileData struct {
	ID        int64     `json:"id" example:"1"`
	UserID    int64     `json:"user_id" example:"1"`
	Phone     *string   `json:"phone"`
	Address   *string   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TrainerData struct {
	User    UserResponse       `json:"user"`
	Profile TrainerProfileData `json:"profile"`
}

type StudentData struct {
	User    UserResponse       `json:"user"`
	Profile StudentProfileData `json:"profile"`
}

type TrainerAPIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    TrainerData `json:"data"`
}

type TrainerListAPIResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"success"`
	Data    []TrainerData `json:"data"`
}

type StudentAPIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    StudentData `json:"data"`
}

type StudentListAPIResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"success"`
	Data    []StudentData `json:"data"`
}

type PackageAPIResponse struct {
	Success bool                 `json:"success" example:"true"`
	Message string               `json:"message" example:"success"`
	Data    models.CoursePackage `json:"data"`
}

type PackageListAPIResponse struct {
	Success bool                   `json:"success" example:"true"`
	Message string                 `json:"message" example:"success"`
	Data    []models.CoursePackage `json:"data"`
}

type MaterialAPIResponse struct {
	Success bool            `json:"success" example:"true"`
	Message string          `json:"message" example:"success"`
	Data    models.Material `json:"data"`
}

type MaterialListAPIResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"success"`
	Data    []models.Material `json:"data"`
}

type SubMaterialAPIResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"success"`
	Data    models.SubMaterial `json:"data"`
}

type SubMaterialListAPIResponse struct {
	Success bool                 `json:"success" example:"true"`
	Message string               `json:"message" example:"success"`
	Data    []models.SubMaterial `json:"data"`
}
