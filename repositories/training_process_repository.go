package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrSessionAssessmentRequired = errors.New("at least one skill assessment is required")
	ErrSessionEvaluationRequired = errors.New("a complete session evaluation is required")
)

type TrainingAssessmentRecord struct {
	ID                int64
	TrainingSessionID int64
	MaterialID        int64
	MaterialName      string
	SubMaterialID     int64
	SubMaterialName   string
	SkillStatus       models.SkillStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PreviousTrainingSessionRecord struct {
	Session     models.TrainingSession
	Assessments []TrainingAssessmentRecord
	Evaluation  *models.SessionEvaluation
}

type TrainerStudentProgressRecord struct {
	StudentID         int64
	TrainingSessionID int64
	PreviousSessions  []PreviousTrainingSessionRecord
}

func (r *TrainingSessionRepository) ListTrainer(ctx context.Context, trainerID int64) ([]models.TrainingSession, error) {
	sessions := make([]models.TrainingSession, 0)
	err := r.db.WithContext(ctx).Where("trainer_id = ?", trainerID).
		Order("scheduled_date DESC, start_time DESC, id DESC").Find(&sessions).Error
	return sessions, err
}

func (r *TrainingSessionRepository) GetTrainer(ctx context.Context, trainerID, id int64) (*models.TrainingSession, error) {
	var session models.TrainingSession
	if err := r.db.WithContext(ctx).Where("id = ? AND trainer_id = ?", id, trainerID).First(&session).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) StartTrainer(ctx context.Context, trainerID, id int64, now time.Time) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := lockTrainerSession(tx, trainerID, id)
		if err != nil {
			return err
		}
		if item.Status != models.SessionScheduled {
			return ErrInvalidTrainingSessionState
		}
		item.Status, item.ActualStartedAt = models.SessionInProgress, &now
		if err := tx.Model(item).Updates(map[string]any{
			"status": item.Status, "actual_started_at": item.ActualStartedAt,
		}).Error; err != nil {
			return mapPhase5Error(err)
		}
		session = *item
		return nil
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) UpsertTrainerAssessments(
	ctx context.Context, trainerID, id int64, assessments []models.SessionSkillAssessment,
) ([]TrainingAssessmentRecord, error) {
	var result []TrainingAssessmentRecord
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockTrainerSession(tx, trainerID, id)
		if err != nil {
			return err
		}
		if session.Status != models.SessionInProgress {
			return ErrInvalidTrainingSessionState
		}
		for index := range assessments {
			var count int64
			err := tx.Model(&models.SubMaterial{}).
				Joins("JOIN materials ON materials.id = sub_materials.material_id").
				Where("sub_materials.id = ? AND sub_materials.status = ? AND materials.status = ?",
					assessments[index].SubMaterialID, models.StatusActive, models.StatusActive).
				Count(&count).Error
			if err != nil {
				return err
			}
			if count == 0 {
				return ErrRecordNotFound
			}
			assessments[index].TrainingSessionID = session.ID
			err = tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "training_session_id"}, {Name: "sub_material_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"skill_status", "updated_at"}),
			}).Create(&assessments[index]).Error
			if err != nil {
				return mapPhase5Error(err)
			}
		}
		result, err = sessionAssessments(tx, session.ID)
		return err
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return result, nil
}

func (r *TrainingSessionRepository) TrainerAssessments(ctx context.Context, trainerID, id int64) ([]TrainingAssessmentRecord, error) {
	if _, err := r.GetTrainer(ctx, trainerID, id); err != nil {
		return nil, err
	}
	return sessionAssessments(r.db.WithContext(ctx), id)
}

func (r *TrainingSessionRepository) UpsertTrainerEvaluation(
	ctx context.Context, trainerID, id int64, evaluation models.SessionEvaluation,
) (*models.SessionEvaluation, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockTrainerSession(tx, trainerID, id)
		if err != nil {
			return err
		}
		if session.Status != models.SessionInProgress {
			return ErrInvalidTrainingSessionState
		}
		evaluation.TrainingSessionID = session.ID
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "training_session_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"predicate", "notes", "recommendation", "updated_at"}),
		}).Create(&evaluation).Error
		if err != nil {
			return mapPhase5Error(err)
		}
		return mapPhase5Error(tx.Where("training_session_id = ?", session.ID).First(&evaluation).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &evaluation, nil
}

func (r *TrainingSessionRepository) TrainerEvaluation(ctx context.Context, trainerID, id int64) (*models.SessionEvaluation, error) {
	if _, err := r.GetTrainer(ctx, trainerID, id); err != nil {
		return nil, err
	}
	var evaluation models.SessionEvaluation
	if err := r.db.WithContext(ctx).Where("training_session_id = ?", id).First(&evaluation).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &evaluation, nil
}

func (r *TrainingSessionRepository) CompleteTrainer(ctx context.Context, trainerID, id int64, now time.Time) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := lockTrainerSession(tx, trainerID, id)
		if err != nil {
			return err
		}
		if item.Status != models.SessionInProgress {
			return ErrInvalidTrainingSessionState
		}
		var count int64
		if err := tx.Model(&models.SessionSkillAssessment{}).
			Where("training_session_id = ?", item.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrSessionAssessmentRequired
		}
		var evaluation models.SessionEvaluation
		err = tx.Where("training_session_id = ?", item.ID).First(&evaluation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionEvaluationRequired
		}
		if err != nil {
			return mapPhase5Error(err)
		}
		if evaluation.Predicate == "" || strings.TrimSpace(evaluation.Notes) == "" ||
			strings.TrimSpace(evaluation.Recommendation) == "" {
			return ErrSessionEvaluationRequired
		}
		item.Status, item.ActualCompletedAt = models.SessionCompleted, &now
		if err := tx.Model(item).Updates(map[string]any{
			"status": item.Status, "actual_completed_at": item.ActualCompletedAt,
		}).Error; err != nil {
			return mapPhase5Error(err)
		}
		if err := completeEnrollmentIfEligible(tx, item, now); err != nil {
			return err
		}
		session = *item
		return nil
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) TrainerStudentProgress(
	ctx context.Context, trainerID, id int64,
) (*TrainerStudentProgressRecord, error) {
	result := TrainerStudentProgressRecord{
		TrainingSessionID: id, PreviousSessions: make([]PreviousTrainingSessionRecord, 0),
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.TrainingSession
		err := tx.Where("id = ? AND trainer_id = ?", id, trainerID).First(&current).Error
		if err != nil {
			return mapPhase5Error(err)
		}
		var enrollment models.StudentEnrollment
		if err := tx.First(&enrollment, current.EnrollmentID).Error; err != nil {
			return mapPhase5Error(err)
		}
		result.StudentID = enrollment.StudentID
		sessions := make([]models.TrainingSession, 0)
		err = tx.Model(&models.TrainingSession{}).
			Joins("JOIN student_enrollments ON student_enrollments.id = training_sessions.enrollment_id").
			Where("student_enrollments.student_id = ? AND training_sessions.status = ? AND training_sessions.id <> ?",
				enrollment.StudentID, models.SessionCompleted, current.ID).
			Order("training_sessions.actual_completed_at DESC, training_sessions.scheduled_date DESC, training_sessions.id DESC").
			Find(&sessions).Error
		if err != nil {
			return err
		}
		for index := range sessions {
			assessments, err := sessionAssessments(tx, sessions[index].ID)
			if err != nil {
				return err
			}
			previous := PreviousTrainingSessionRecord{Session: sessions[index], Assessments: assessments}
			var evaluation models.SessionEvaluation
			err = tx.Where("training_session_id = ?", sessions[index].ID).First(&evaluation).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return mapPhase5Error(err)
			}
			if err == nil {
				previous.Evaluation = &evaluation
			}
			result.PreviousSessions = append(result.PreviousSessions, previous)
		}
		return nil
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &result, nil
}

func lockTrainerSession(tx *gorm.DB, trainerID, id int64) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND trainer_id = ?", id, trainerID).First(&session).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func sessionAssessments(db *gorm.DB, sessionID int64) ([]TrainingAssessmentRecord, error) {
	result := make([]TrainingAssessmentRecord, 0)
	err := db.Model(&models.SessionSkillAssessment{}).
		Select("session_skill_assessments.*, materials.id AS material_id, materials.name AS material_name, sub_materials.name AS sub_material_name").
		Joins("JOIN sub_materials ON sub_materials.id = session_skill_assessments.sub_material_id").
		Joins("JOIN materials ON materials.id = sub_materials.material_id").
		Where("session_skill_assessments.training_session_id = ?", sessionID).
		Order("materials.sequence ASC, sub_materials.sequence ASC, session_skill_assessments.id ASC").
		Scan(&result).Error
	return result, err
}
