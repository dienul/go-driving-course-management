package services

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (*dto.RegisterData, error)
	Login(ctx context.Context, request dto.LoginRequest) (*dto.LoginData, error)
}

type DefaultAuthService struct {
	users  repositories.UserRepository
	tokens *JWTManager
}

func NewAuthService(
	users repositories.UserRepository,
	tokens *JWTManager,
) *DefaultAuthService {
	return &DefaultAuthService{users: users, tokens: tokens}
}

func (s *DefaultAuthService) Register(
	ctx context.Context,
	request dto.RegisterRequest,
) (*dto.RegisterData, error) {
	name := strings.TrimSpace(request.Name)
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if name == "" || email == "" || len([]byte(request.Password)) > 72 ||
		!utf8.ValidString(request.Password) {
		return nil, ErrInvalidInput
	}

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         models.RoleStudent,
		Status:       models.StatusActive,
	}
	profile := &models.StudentProfile{
		Phone:   normalizedOptional(request.Phone),
		Address: normalizedOptional(request.Address),
	}
	if err := s.users.CreateStudent(ctx, user, profile); err != nil {
		if errors.Is(err, repositories.ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return &dto.RegisterData{User: userResponse(user)}, nil
}

func (s *DefaultAuthService) Login(
	ctx context.Context,
	request dto.LoginRequest,
) (*dto.LoginData, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(request.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Status != models.StatusActive {
		return nil, ErrInactiveUser
	}

	token, expiresAt, err := s.tokens.Generate(user)
	if err != nil {
		return nil, err
	}
	return &dto.LoginData{
		User:      userResponse(user),
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
	}, nil
}

func userResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func normalizedOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
