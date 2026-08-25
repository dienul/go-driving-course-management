package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoActiveStudentEnrollment      = errors.New("student has no active enrollment")
	ErrTrainerSessionOverlaps         = errors.New("trainer training session overlaps")
	ErrSessionCapacityExceeded        = errors.New("purchased session capacity exhausted")
	ErrStudentSessionOpen             = errors.New("student has an unfinished training session")
	ErrUnavailableTrainerAvailability = errors.New("trainer availability is unavailable")
	ErrInvalidTrainingSessionState    = errors.New("invalid training session state")
	ErrInvalidTrainingSlot            = errors.New("invalid training session slot")
)

type TrainingSessionRepository struct{ db *gorm.DB }

func NewTrainingSessionRepository(db *gorm.DB) *TrainingSessionRepository {
	return &TrainingSessionRepository{db: db}
}

func (r *TrainingSessionRepository) Book(ctx context.Context, studentID, availabilityID int64, start time.Time) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var enrollment models.StudentEnrollment
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("student_id = ? AND status = ?", studentID, models.EnrollmentActive).
			First(&enrollment).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoActiveStudentEnrollment
		}
		if err != nil {
			return mapPhase5Error(err)
		}

		var open int64
		if err := tx.Model(&models.TrainingSession{}).
			Where("enrollment_id = ? AND status IN ?", enrollment.ID,
				[]models.TrainingSessionStatus{models.SessionScheduled, models.SessionInProgress}).
			Count(&open).Error; err != nil {
			return err
		}
		if open > 0 {
			return ErrStudentSessionOpen
		}

		var completed int64
		if err := tx.Model(&models.TrainingSession{}).
			Where("enrollment_id = ? AND status = ?", enrollment.ID, models.SessionCompleted).
			Count(&completed).Error; err != nil {
			return err
		}
		if completed >= int64(enrollment.RequiredSessions()) {
			return ErrSessionCapacityExceeded
		}

		end := start.Add(2 * time.Hour)
		availability, err := lockBookableAvailability(tx, availabilityID, start, end, 0)
		if err != nil {
			return err
		}
		session = models.TrainingSession{
			EnrollmentID: enrollment.ID, TrainerID: availability.TrainerID,
			TrainerAvailabilityID: availability.ID, SessionNumber: int(completed) + 1,
			ScheduledDate: availability.AvailableDate, StartTime: models.NewClockTime(start),
			EndTime: models.NewClockTime(end), Status: models.SessionScheduled,
		}
		return mapPhase5Error(tx.Create(&session).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) ListStudent(ctx context.Context, studentID int64) ([]models.TrainingSession, error) {
	results := make([]models.TrainingSession, 0)
	err := r.db.WithContext(ctx).Model(&models.TrainingSession{}).
		Joins("JOIN student_enrollments ON student_enrollments.id = training_sessions.enrollment_id").
		Where("student_enrollments.student_id = ?", studentID).
		Order("training_sessions.scheduled_date DESC, training_sessions.start_time DESC, training_sessions.id DESC").
		Find(&results).Error
	return results, err
}

func (r *TrainingSessionRepository) GetStudent(ctx context.Context, studentID, id int64) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := r.db.WithContext(ctx).Model(&models.TrainingSession{}).
		Joins("JOIN student_enrollments ON student_enrollments.id = training_sessions.enrollment_id").
		Where("training_sessions.id = ? AND student_enrollments.student_id = ?", id, studentID).
		First(&session).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) CancelStudent(ctx context.Context, studentID, id int64, reason string, now time.Time) (*models.TrainingSession, error) {
	return r.cancel(ctx, id, studentID, &studentID, reason, now)
}

func (r *TrainingSessionRepository) RescheduleStudent(ctx context.Context, studentID, id, availabilityID int64, start time.Time) (*models.TrainingSession, error) {
	return r.reschedule(ctx, id, &studentID, availabilityID, start)
}

func (r *TrainingSessionRepository) List(ctx context.Context) ([]models.TrainingSession, error) {
	results := make([]models.TrainingSession, 0)
	err := r.db.WithContext(ctx).
		Order("scheduled_date DESC, start_time DESC, id DESC").Find(&results).Error
	return results, err
}

func (r *TrainingSessionRepository) Get(ctx context.Context, id int64) (*models.TrainingSession, error) {
	var session models.TrainingSession
	if err := r.db.WithContext(ctx).First(&session, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) Cancel(ctx context.Context, adminID, id int64, reason string, now time.Time) (*models.TrainingSession, error) {
	return r.cancel(ctx, id, adminID, nil, reason, now)
}

func (r *TrainingSessionRepository) Reschedule(ctx context.Context, id, availabilityID int64, start time.Time) (*models.TrainingSession, error) {
	return r.reschedule(ctx, id, nil, availabilityID, start)
}

func (r *TrainingSessionRepository) cancel(ctx context.Context, id, actorID int64, studentID *int64, reason string, now time.Time) (*models.TrainingSession, error) {
	var session models.TrainingSession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, _, err := lockOwnedTrainingSession(tx, id, studentID, false)
		if err != nil {
			return err
		}
		session = *item
		if session.Status != models.SessionScheduled {
			return ErrInvalidTrainingSessionState
		}
		session.Status, session.CancelledBy, session.CancellationReason, session.CancelledAt =
			models.SessionCancelled, &actorID, &reason, &now
		return mapPhase5Error(tx.Model(&session).Updates(map[string]any{
			"status": session.Status, "cancelled_by": session.CancelledBy,
			"cancellation_reason": session.CancellationReason, "cancelled_at": session.CancelledAt,
		}).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &session, nil
}

func (r *TrainingSessionRepository) reschedule(ctx context.Context, id int64, studentID *int64, availabilityID int64, start time.Time) (*models.TrainingSession, error) {
	var replacement models.TrainingSession
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		previous, enrollment, err := lockOwnedTrainingSession(tx, id, studentID, true)
		if err != nil {
			return err
		}
		if previous.Status != models.SessionScheduled {
			return ErrInvalidTrainingSessionState
		}
		if enrollment.Status != models.EnrollmentActive {
			return ErrNoActiveStudentEnrollment
		}
		end := start.Add(2 * time.Hour)
		availability, err := lockBookableAvailability(tx, availabilityID, start, end, previous.ID)
		if err != nil {
			return err
		}
		if err := tx.Model(previous).Update("status", models.SessionRescheduled).Error; err != nil {
			return mapPhase5Error(err)
		}
		oldID := previous.ID
		replacement = models.TrainingSession{
			EnrollmentID: previous.EnrollmentID, TrainerID: availability.TrainerID,
			TrainerAvailabilityID: availability.ID, SessionNumber: previous.SessionNumber,
			ScheduledDate: availability.AvailableDate, StartTime: models.NewClockTime(start),
			EndTime: models.NewClockTime(end), Status: models.SessionScheduled, RescheduledFromID: &oldID,
		}
		return mapPhase5Error(tx.Create(&replacement).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &replacement, nil
}

func lockOwnedTrainingSession(tx *gorm.DB, id int64, studentID *int64, lockEnrollment bool) (*models.TrainingSession, *models.StudentEnrollment, error) {
	var session models.TrainingSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, id).Error; err != nil {
		return nil, nil, mapPhase5Error(err)
	}
	query := tx.Where("id = ?", session.EnrollmentID)
	if lockEnrollment {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if studentID != nil {
		query = query.Where("student_id = ?", *studentID)
	}
	var enrollment models.StudentEnrollment
	if err := query.First(&enrollment).Error; err != nil {
		return nil, nil, mapPhase5Error(err)
	}
	return &session, &enrollment, nil
}

func lockBookableAvailability(tx *gorm.DB, id int64, start, end time.Time, excludeSessionID int64) (*models.TrainerAvailability, error) {
	var availability models.TrainerAvailability
	if err := tx.First(&availability, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	if availability.Status != models.AvailabilityPublished {
		return nil, ErrUnavailableTrainerAvailability
	}
	if err := lockActiveTrainer(tx, availability.TrainerID); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrUnavailableTrainerAvailability
		}
		return nil, err
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&availability, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	if availability.Status != models.AvailabilityPublished {
		return nil, ErrUnavailableTrainerAvailability
	}
	if availability.AvailableDate.Weekday() == time.Saturday || availability.AvailableDate.Weekday() == time.Sunday ||
		start.Minute() != 0 || end.Minute() != 0 || start.Hour() < 8 || end.Hour() > 17 ||
		start.Hour() < availability.StartTime.Hour() || end.Hour() > availability.EndTime.Hour() {
		return nil, ErrInvalidTrainingSlot
	}
	query := tx.Model(&models.TrainingSession{}).
		Where("trainer_id = ? AND scheduled_date = ? AND status IN ?", availability.TrainerID,
			availability.AvailableDate.Format("2006-01-02"), []models.TrainingSessionStatus{
				models.SessionScheduled, models.SessionInProgress, models.SessionCompleted,
			}).
		Where("start_time < ? AND end_time > ?", end.Format("15:04:05"), start.Format("15:04:05"))
	if excludeSessionID > 0 {
		query = query.Where("id <> ?", excludeSessionID)
	}
	var overlaps int64
	if err := query.Count(&overlaps).Error; err != nil {
		return nil, err
	}
	if overlaps > 0 {
		return nil, ErrTrainerSessionOverlaps
	}
	return &availability, nil
}
