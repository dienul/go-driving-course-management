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
	ErrAvailabilityOverlaps     = errors.New("trainer availability overlaps")
	ErrInvalidAvailabilityState = errors.New("invalid trainer availability state")
)

type ScheduleRecord struct {
	Availability models.TrainerAvailability
	Trainer      models.User
	Sessions     []models.TrainingSession
}

type AvailabilityRepository struct{ db *gorm.DB }

func NewAvailabilityRepository(db *gorm.DB) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

func (r *AvailabilityRepository) Create(ctx context.Context, trainerID int64, date, start, end time.Time) (*models.TrainerAvailability, error) {
	item := models.TrainerAvailability{
		TrainerID: trainerID, AvailableDate: date, StartTime: models.NewClockTime(start), EndTime: models.NewClockTime(end),
		Status: models.AvailabilityPending,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveTrainer(tx, trainerID); err != nil {
			return err
		}
		if err := ensureNoAvailabilityOverlap(tx, trainerID, date, start, end, 0); err != nil {
			return err
		}
		return mapPhase5Error(tx.Create(&item).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) ListTrainer(ctx context.Context, trainerID int64) ([]models.TrainerAvailability, error) {
	results := make([]models.TrainerAvailability, 0)
	err := r.db.WithContext(ctx).Where("trainer_id = ?", trainerID).
		Order("available_date ASC, start_time ASC, id ASC").Find(&results).Error
	return results, err
}

func (r *AvailabilityRepository) GetTrainer(ctx context.Context, trainerID, id int64) (*models.TrainerAvailability, error) {
	var item models.TrainerAvailability
	err := r.db.WithContext(ctx).Where("id = ? AND trainer_id = ?", id, trainerID).First(&item).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) UpdateTrainer(ctx context.Context, trainerID, id int64, date, start, end time.Time) (*models.TrainerAvailability, error) {
	var item models.TrainerAvailability
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveTrainer(tx, trainerID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND trainer_id = ?", id, trainerID).First(&item).Error; err != nil {
			return mapPhase5Error(err)
		}
		if item.Status != models.AvailabilityPending {
			return ErrInvalidAvailabilityState
		}
		if err := ensureNoAvailabilityOverlap(tx, trainerID, date, start, end, id); err != nil {
			return err
		}
		item.AvailableDate, item.StartTime, item.EndTime = date, models.NewClockTime(start), models.NewClockTime(end)
		return mapPhase5Error(tx.Model(&item).Updates(map[string]any{
			"available_date": date, "start_time": item.StartTime, "end_time": item.EndTime,
		}).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) CancelTrainer(ctx context.Context, trainerID, id int64) (*models.TrainerAvailability, error) {
	return r.cancel(ctx, id, &trainerID)
}

func (r *AvailabilityRepository) List(ctx context.Context) ([]models.TrainerAvailability, error) {
	results := make([]models.TrainerAvailability, 0)
	err := r.db.WithContext(ctx).Order("available_date ASC, start_time ASC, id ASC").Find(&results).Error
	return results, err
}

func (r *AvailabilityRepository) Get(ctx context.Context, id int64) (*models.TrainerAvailability, error) {
	var item models.TrainerAvailability
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) Publish(ctx context.Context, adminID, id int64, now time.Time) (*models.TrainerAvailability, error) {
	var item models.TrainerAvailability
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&item, id).Error; err != nil {
			return mapPhase5Error(err)
		}
		if err := lockActiveTrainer(tx, item.TrainerID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, id).Error; err != nil {
			return mapPhase5Error(err)
		}
		if item.Status != models.AvailabilityPending {
			return ErrInvalidAvailabilityState
		}
		var admin models.User
		if err := tx.Where("id = ? AND role = ? AND status = ?", adminID, models.RoleAdmin, models.StatusActive).
			First(&admin).Error; err != nil {
			return mapPhase5Error(err)
		}
		item.Status, item.PublishedBy, item.PublishedAt = models.AvailabilityPublished, &adminID, &now
		return mapPhase5Error(tx.Model(&item).Updates(map[string]any{
			"status": item.Status, "published_by": item.PublishedBy, "published_at": item.PublishedAt,
		}).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) Cancel(ctx context.Context, id int64) (*models.TrainerAvailability, error) {
	return r.cancel(ctx, id, nil)
}

func (r *AvailabilityRepository) cancel(ctx context.Context, id int64, trainerID *int64) (*models.TrainerAvailability, error) {
	var item models.TrainerAvailability
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id)
		if trainerID != nil {
			query = query.Where("trainer_id = ?", *trainerID)
		}
		if err := query.First(&item).Error; err != nil {
			return mapPhase5Error(err)
		}
		if item.Status == models.AvailabilityCancelled {
			return ErrInvalidAvailabilityState
		}
		var booked int64
		if err := tx.Model(&models.TrainingSession{}).
			Where("trainer_availability_id = ? AND status IN ?", item.ID, []models.TrainingSessionStatus{
				models.SessionScheduled, models.SessionInProgress,
			}).Count(&booked).Error; err != nil {
			return err
		}
		if booked > 0 {
			return ErrInvalidAvailabilityState
		}
		item.Status = models.AvailabilityCancelled
		return mapPhase5Error(tx.Model(&item).Update("status", item.Status).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &item, nil
}

func (r *AvailabilityRepository) ListPublishedSchedules(ctx context.Context, date *time.Time) ([]ScheduleRecord, error) {
	query := r.db.WithContext(ctx).Model(&models.TrainerAvailability{}).
		Joins("JOIN users ON users.id = trainer_availabilities.trainer_id").
		Where("trainer_availabilities.status = ? AND users.role = ? AND users.status = ?",
			models.AvailabilityPublished, models.RoleTrainer, models.StatusActive)
	if date != nil {
		query = query.Where("trainer_availabilities.available_date = ?", date.Format("2006-01-02"))
	}
	var items []models.TrainerAvailability
	if err := query.Order("trainer_availabilities.available_date ASC, trainer_availabilities.start_time ASC, trainer_availabilities.id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	results := make([]ScheduleRecord, 0, len(items))
	for i := range items {
		record := ScheduleRecord{Availability: items[i], Sessions: make([]models.TrainingSession, 0)}
		if err := r.db.WithContext(ctx).First(&record.Trainer, items[i].TrainerID).Error; err != nil {
			return nil, mapPhase5Error(err)
		}
		if err := r.db.WithContext(ctx).
			Where("trainer_id = ? AND scheduled_date = ? AND status IN ?", items[i].TrainerID,
				items[i].AvailableDate.Format("2006-01-02"), []models.TrainingSessionStatus{
					models.SessionScheduled, models.SessionInProgress, models.SessionCompleted,
				}).Find(&record.Sessions).Error; err != nil {
			return nil, err
		}
		results = append(results, record)
	}
	return results, nil
}

func lockActiveTrainer(tx *gorm.DB, trainerID int64) error {
	var trainer models.User
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND role = ? AND status = ?", trainerID, models.RoleTrainer, models.StatusActive).
		First(&trainer).Error
	return mapPhase5Error(err)
}

func ensureNoAvailabilityOverlap(tx *gorm.DB, trainerID int64, date, start, end time.Time, excludeID int64) error {
	query := tx.Model(&models.TrainerAvailability{}).
		Where("trainer_id = ? AND available_date = ? AND status <> ?", trainerID,
			date.Format("2006-01-02"), models.AvailabilityCancelled).
		Where("start_time < ? AND end_time > ?", end.Format("15:04:05"), start.Format("15:04:05"))
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrAvailabilityOverlaps
	}
	return nil
}
