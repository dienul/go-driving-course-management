package models

import "time"

type CoursePackage struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(150);not null;uniqueIndex" json:"name"`
	Level       PackageLevel `gorm:"type:varchar(20);not null" json:"level"`
	TotalHours  int16        `gorm:"not null" json:"total_hours"`
	Price       int64        `gorm:"not null" json:"price"`
	Description *string      `gorm:"type:text" json:"description"`
	Status      RecordStatus `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
}

func (CoursePackage) TableName() string { return "course_packages" }

func (p CoursePackage) TotalSessions() int {
	return int(p.TotalHours) / 2
}

type Material struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(200);not null;uniqueIndex" json:"name"`
	Description *string      `gorm:"type:text" json:"description"`
	Sequence    int          `gorm:"not null;uniqueIndex" json:"sequence"`
	Status      RecordStatus `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
}

func (Material) TableName() string { return "materials" }

type SubMaterial struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID  int64        `gorm:"not null;uniqueIndex:uq_sub_materials_material_sequence;uniqueIndex:uq_sub_materials_material_name" json:"material_id"`
	Name        string       `gorm:"type:varchar(200);not null;uniqueIndex:uq_sub_materials_material_name" json:"name"`
	Description *string      `gorm:"type:text" json:"description"`
	Sequence    int          `gorm:"not null;uniqueIndex:uq_sub_materials_material_sequence" json:"sequence"`
	Status      RecordStatus `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
}

func (SubMaterial) TableName() string { return "sub_materials" }
