package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "phase-3-test-secret-with-at-least-32-bytes"

type fakeUserRepository struct {
	user    *models.User
	profile *models.StudentProfile
}

func (f *fakeUserRepository) FindByEmail(_ context.Context, email string) (*models.User, error) {
	if f.user == nil || f.user.Email != email {
		return nil, repositories.ErrUserNotFound
	}
	copy := *f.user
	return &copy, nil
}

func (f *fakeUserRepository) FindByID(_ context.Context, id int64) (*models.User, error) {
	if f.user == nil || f.user.ID != id {
		return nil, repositories.ErrUserNotFound
	}
	copy := *f.user
	return &copy, nil
}

func (f *fakeUserRepository) CreateStudent(
	_ context.Context,
	user *models.User,
	profile *models.StudentProfile,
) error {
	if f.user != nil && f.user.Email == user.Email {
		return repositories.ErrEmailExists
	}
	user.ID = 10
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	profile.ID = 20
	profile.UserID = user.ID
	f.user = user
	f.profile = profile
	return nil
}

func TestRegisterCreatesActiveStudentAndHashesPassword(t *testing.T) {
	users := &fakeUserRepository{}
	tokens := mustJWTManager(t)
	auth := NewAuthService(users, tokens)
	phone := " 081234567890 "
	address := " Jakarta "

	data, err := auth.Register(context.Background(), dto.RegisterRequest{
		Name:     " Dienul Haq ",
		Email:    " DIENUL@EXAMPLE.COM ",
		Password: "strong-password",
		Phone:    &phone,
		Address:  &address,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if users.user.Role != models.RoleStudent || users.user.Status != models.StatusActive {
		t.Fatalf("unexpected role/status: %s/%s", users.user.Role, users.user.Status)
	}
	if users.user.Email != "dienul@example.com" {
		t.Fatalf("email was not normalized: %q", users.user.Email)
	}
	if users.user.PasswordHash == "strong-password" {
		t.Fatal("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(users.user.PasswordHash),
		[]byte("strong-password"),
	); err != nil {
		t.Fatalf("password hash does not match: %v", err)
	}
	if users.profile == nil || users.profile.UserID != users.user.ID {
		t.Fatal("student profile was not created with the user")
	}
	if users.profile.Phone == nil || *users.profile.Phone != "081234567890" {
		t.Fatalf("phone was not normalized: %#v", users.profile.Phone)
	}
	if data.User.Role != models.RoleStudent {
		t.Fatalf("response role = %s, want STUDENT", data.User.Role)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	users := &fakeUserRepository{user: &models.User{Email: "student@example.com"}}
	auth := NewAuthService(users, mustJWTManager(t))

	_, err := auth.Register(context.Background(), dto.RegisterRequest{
		Name:     "Student",
		Email:    "student@example.com",
		Password: "strong-password",
	})
	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("error = %v, want ErrEmailExists", err)
	}
}

func TestLoginReturnsTokenForDatabaseRole(t *testing.T) {
	user := userWithPassword(t, models.RoleTrainer, models.StatusActive)
	users := &fakeUserRepository{user: user}
	tokens := mustJWTManager(t)
	auth := NewAuthService(users, tokens)

	data, err := auth.Login(context.Background(), dto.LoginRequest{
		Email:    user.Email,
		Password: "strong-password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if data.User.Role != models.RoleTrainer {
		t.Fatalf("role = %s, want TRAINER", data.User.Role)
	}
	claims, err := tokens.Parse(data.Token)
	if err != nil {
		t.Fatalf("parse generated token: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != models.RoleTrainer {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestLoginRejectsIncorrectPasswordAndInactiveUser(t *testing.T) {
	user := userWithPassword(t, models.RoleStudent, models.StatusActive)
	users := &fakeUserRepository{user: user}
	auth := NewAuthService(users, mustJWTManager(t))

	if _, err := auth.Login(context.Background(), dto.LoginRequest{
		Email: user.Email, Password: "wrong-password",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong-password error = %v", err)
	}

	users.user.Status = models.StatusInactive
	if _, err := auth.Login(context.Background(), dto.LoginRequest{
		Email: user.Email, Password: "strong-password",
	}); !errors.Is(err, ErrInactiveUser) {
		t.Fatalf("inactive-user error = %v", err)
	}
}

func TestJWTManagerRejectsInvalidConfigurationAndTamperedToken(t *testing.T) {
	if _, err := NewJWTManager("short", "24h"); err == nil {
		t.Fatal("short secret was accepted")
	}
	if _, err := NewJWTManager(testJWTSecret, "invalid"); err == nil {
		t.Fatal("invalid expiry was accepted")
	}

	manager := mustJWTManager(t)
	user := userWithPassword(t, models.RoleAdmin, models.StatusActive)
	token, _, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := manager.Parse(token + "tampered"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered-token error = %v", err)
	}
}

func mustJWTManager(t *testing.T) *JWTManager {
	t.Helper()
	manager, err := NewJWTManager(testJWTSecret, "1h")
	if err != nil {
		t.Fatalf("new JWT manager: %v", err)
	}
	return manager
}

func userWithPassword(
	t *testing.T,
	role models.UserRole,
	status models.RecordStatus,
) *models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return &models.User{
		ID:           42,
		Name:         "Test User",
		Email:        "user@example.com",
		PasswordHash: string(hash),
		Role:         role,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}
