package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrDuplicateRecord = errors.New("record already exists")
)

type TrainerRecord struct {
	User    models.User
	Profile models.TrainerProfile
}

type StudentRecord struct {
	User    models.User
	Profile models.StudentProfile
}

type AdminUserRepository struct {
	db *gorm.DB
}

func NewAdminUserRepository(db *gorm.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) CreateTrainer(
	ctx context.Context,
	user *models.User,
	profile *models.TrainerProfile,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return mapAdminError(err)
		}
		profile.UserID = user.ID
		return mapAdminError(tx.Create(profile).Error)
	})
	return mapAdminError(err)
}

func (r *AdminUserRepository) ListTrainers(ctx context.Context) ([]TrainerRecord, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Where("role = ?", models.RoleTrainer).
		Order("id ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}

	records := make([]TrainerRecord, 0, len(users))
	if len(users) == 0 {
		return records, nil
	}
	ids := userIDs(users)
	var profiles []models.TrainerProfile
	if err := r.db.WithContext(ctx).Where("user_id IN ?", ids).Find(&profiles).Error; err != nil {
		return nil, err
	}
	profileMap := make(map[int64]models.TrainerProfile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}
	for _, user := range users {
		profile, ok := profileMap[user.ID]
		if !ok {
			return nil, fmt.Errorf("trainer profile missing for user %d", user.ID)
		}
		records = append(records, TrainerRecord{User: user, Profile: profile})
	}
	return records, nil
}

func (r *AdminUserRepository) GetTrainer(ctx context.Context, id int64) (*TrainerRecord, error) {
	user, err := r.userByRole(ctx, id, models.RoleTrainer)
	if err != nil {
		return nil, err
	}
	var profile models.TrainerProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", id).First(&profile).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &TrainerRecord{User: *user, Profile: profile}, nil
}

func (r *AdminUserRepository) UpdateTrainer(
	ctx context.Context,
	user *models.User,
	profile *models.TrainerProfile,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.Where("id = ? AND role = ?", user.ID, models.RoleTrainer).
			First(&current).Error; err != nil {
			return mapAdminError(err)
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"name":  user.Name,
			"email": user.Email,
		}).Error; err != nil {
			return mapAdminError(err)
		}
		result := tx.Model(&models.TrainerProfile{}).
			Where("user_id = ?", user.ID).
			Updates(map[string]any{
				"phone":   profile.Phone,
				"address": profile.Address,
				"bio":     profile.Bio,
			})
		if result.Error != nil {
			return mapAdminError(result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrRecordNotFound
		}
		return nil
	})
	return mapAdminError(err)
}

func (r *AdminUserRepository) UpdateUserStatus(
	ctx context.Context,
	id int64,
	role models.UserRole,
	status models.RecordStatus,
) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ? AND role = ?", id, role).
		Update("status", status)
	if result.Error != nil {
		return mapAdminError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (r *AdminUserRepository) ListStudents(ctx context.Context) ([]StudentRecord, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Where("role = ?", models.RoleStudent).
		Order("id ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}

	records := make([]StudentRecord, 0, len(users))
	if len(users) == 0 {
		return records, nil
	}
	var profiles []models.StudentProfile
	if err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs(users)).
		Find(&profiles).Error; err != nil {
		return nil, err
	}
	profileMap := make(map[int64]models.StudentProfile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}
	for _, user := range users {
		profile, ok := profileMap[user.ID]
		if !ok {
			return nil, fmt.Errorf("student profile missing for user %d", user.ID)
		}
		records = append(records, StudentRecord{User: user, Profile: profile})
	}
	return records, nil
}

func (r *AdminUserRepository) GetStudent(ctx context.Context, id int64) (*StudentRecord, error) {
	user, err := r.userByRole(ctx, id, models.RoleStudent)
	if err != nil {
		return nil, err
	}
	var profile models.StudentProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", id).First(&profile).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &StudentRecord{User: *user, Profile: profile}, nil
}

func (r *AdminUserRepository) userByRole(
	ctx context.Context,
	id int64,
	role models.UserRole,
) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("id = ? AND role = ?", id, role).
		First(&user).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &user, nil
}

func userIDs(users []models.User) []int64 {
	ids := make([]int64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func mapAdminError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateRecord
	}
	return err
}
