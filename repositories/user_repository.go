package repositories

import (
	"context"
	"errors"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
	CreateStudent(ctx context.Context, user *models.User, profile *models.StudentProfile) error
}

type GORMUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GORMUserRepository {
	return &GORMUserRepository{db: db}
}

func (r *GORMUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GORMUserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GORMUserRepository) CreateStudent(
	ctx context.Context,
	user *models.User,
	profile *models.StudentProfile,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrEmailExists
			}
			return err
		}

		profile.UserID = user.ID
		return tx.Create(profile).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailExists
	}
	return err
}
