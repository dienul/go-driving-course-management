package models

import "time"

type SessionSkillAssessment struct {
	ID                int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainingSessionID int64       `gorm:"not null;uniqueIndex:uq_session_skill_assessments_session_sub_material" json:"training_session_id"`
	SubMaterialID     int64       `gorm:"not null;uniqueIndex:uq_session_skill_assessments_session_sub_material;index" json:"sub_material_id"`
	SkillStatus       SkillStatus `gorm:"type:varchar(30);not null" json:"skill_status"`
	CreatedAt         time.Time   `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time   `gorm:"not null" json:"updated_at"`
}

func (SessionSkillAssessment) TableName() string { return "session_skill_assessments" }

type SessionEvaluation struct {
	ID                int64               `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainingSessionID int64               `gorm:"not null;uniqueIndex" json:"training_session_id"`
	Predicate         EvaluationPredicate `gorm:"type:varchar(30);not null" json:"predicate"`
	Notes             string              `gorm:"type:text;not null" json:"notes"`
	Recommendation    string              `gorm:"type:text;not null" json:"recommendation"`
	CreatedAt         time.Time           `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time           `gorm:"not null" json:"updated_at"`
}

func (SessionEvaluation) TableName() string { return "session_evaluations" }
