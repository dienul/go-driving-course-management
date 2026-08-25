package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAdminMasterDataPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	admin := models.User{Name: "Phase 4 Admin", Email: "phase4-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin fixture: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h"})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	loginToken := func(email string) string {
		t.Helper()
		response := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{"email": email, "password": "strong-password"}, "")
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.LoginData `json:"data"`
		}
		decodeResponse(t, response, &result)
		return "Bearer " + result.Data.Token
	}
	adminToken := loginToken(admin.Email)
	registerStudent := performJSON(t, router, http.MethodPost, "/api/users/register", map[string]any{
		"name": "Phase 4 Student", "email": "phase4-student@example.com", "password": "strong-password",
	}, "")
	requireAdminStatus(t, registerStudent, http.StatusCreated)
	var registered struct {
		Data dto.RegisterData `json:"data"`
	}
	decodeResponse(t, registerStudent, &registered)
	studentToken := loginToken("phase4-student@example.com")

	t.Run("authorization", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers", nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers", nil, studentToken), http.StatusForbidden)
	})

	var trainerID int64
	t.Run("trainers", func(t *testing.T) {
		body := map[string]any{"name": " Phase 4 Trainer ", "email": "PHASE4-TRAINER@EXAMPLE.COM", "password": "strong-password", "phone": "081234567890", "bio": "Driving instructor"}
		created := performJSON(t, router, http.MethodPost, "/api/v1/admin/trainers", body, adminToken)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data dto.TrainerData `json:"data"`
		}
		decodeResponse(t, created, &result)
		trainerID = result.Data.User.ID
		if result.Data.User.Role != models.RoleTrainer || result.Data.User.Email != "phase4-trainer@example.com" || result.Data.Profile.UserID != trainerID {
			t.Fatalf("unexpected trainer: %+v", result.Data)
		}
		var stored models.User
		if err := db.First(&stored, trainerID).Error; err != nil || stored.PasswordHash == "strong-password" {
			t.Fatalf("trainer password was not safely persisted: err=%v", err)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, "/api/v1/admin/trainers", body, adminToken), http.StatusConflict)
		path := fmt.Sprintf("/api/v1/admin/trainers/%d", trainerID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers", nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusOK)
		updated := performJSON(t, router, http.MethodPut, path, map[string]any{"name": "Updated Trainer", "email": "phase4-trainer-updated@example.com", "phone": "089999"}, adminToken)
		requireAdminStatus(t, updated, http.StatusOK)
		decodeResponse(t, updated, &result)
		if result.Data.User.Name != "Updated Trainer" || result.Data.User.Role != models.RoleTrainer {
			t.Fatalf("trainer update changed unexpected fields: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "INACTIVE"}, adminToken), http.StatusOK)
		inactiveLogin := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{"email": "phase4-trainer-updated@example.com", "password": "strong-password"}, "")
		requireAdminStatus(t, inactiveLogin, http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "ACTIVE"}, adminToken), http.StatusOK)
		_ = loginToken("phase4-trainer-updated@example.com")
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers/not-a-number", nil, adminToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers/999999", nil, adminToken), http.StatusNotFound)
	})

	t.Run("students", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/students", nil, adminToken), http.StatusOK)
		path := fmt.Sprintf("/api/v1/admin/students/%d", registered.Data.User.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "INACTIVE"}, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/auth/me", nil, studentToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "ACTIVE"}, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "DELETED"}, adminToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("/api/v1/admin/students/%d", trainerID), nil, adminToken), http.StatusNotFound)
	})

	t.Run("packages", func(t *testing.T) {
		body := map[string]any{"name": "Phase 4 Pemula 6 Jam", "level": "PEMULA", "total_hours": 6, "price": 900000}
		created := performJSON(t, router, http.MethodPost, "/api/v1/admin/packages", body, adminToken)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data models.CoursePackage `json:"data"`
		}
		decodeResponse(t, created, &result)
		if result.Data.TotalSessions() != 3 || result.Data.Status != models.StatusActive {
			t.Fatalf("unexpected package: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, "/api/v1/admin/packages", body, adminToken), http.StatusConflict)
		invalid := map[string]any{"name": "Invalid Hours", "level": "PEMULA", "total_hours": 7, "price": 100}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, "/api/v1/admin/packages", invalid, adminToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/packages", nil, adminToken), http.StatusOK)
		path := fmt.Sprintf("/api/v1/admin/packages/%d", result.Data.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusOK)
		updated := map[string]any{"name": "Phase 4 Dasar 10 Jam", "level": "DASAR", "total_hours": 10, "price": 1400000}
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, path, updated, adminToken), http.StatusOK)
		response := performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "INACTIVE"}, adminToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data.Status != models.StatusInactive {
			t.Fatalf("package was not deactivated: %+v", result.Data)
		}
	})

	var materialID int64
	t.Run("materials", func(t *testing.T) {
		body := map[string]any{"name": "Phase 4 Vehicle Introduction", "sequence": 90}
		created := performJSON(t, router, http.MethodPost, "/api/v1/admin/materials", body, adminToken)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data models.Material `json:"data"`
		}
		decodeResponse(t, created, &result)
		materialID = result.Data.ID
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, "/api/v1/admin/materials", map[string]any{"name": "Duplicate Sequence", "sequence": 90}, adminToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/materials", nil, adminToken), http.StatusOK)
		path := fmt.Sprintf("/api/v1/admin/materials/%d", materialID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, path, map[string]any{"name": "Phase 4 Updated Material", "sequence": 91}, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "INACTIVE"}, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "ACTIVE"}, adminToken), http.StatusOK)
	})

	t.Run("sub-materials", func(t *testing.T) {
		parent := fmt.Sprintf("/api/v1/admin/materials/%d/sub-materials", materialID)
		body := map[string]any{"name": "Phase 4 Steering", "sequence": 1}
		created := performJSON(t, router, http.MethodPost, parent, body, adminToken)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data models.SubMaterial `json:"data"`
		}
		decodeResponse(t, created, &result)
		if result.Data.MaterialID != materialID {
			t.Fatalf("wrong material relationship: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, parent, body, adminToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, parent, nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/materials/999999/sub-materials", nil, adminToken), http.StatusNotFound)
		path := fmt.Sprintf("/api/v1/admin/sub-materials/%d", result.Data.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, path, map[string]any{"name": "Phase 4 Steering Updated", "sequence": 2}, adminToken), http.StatusOK)
		response := performJSON(t, router, http.MethodPatch, path+"/status", map[string]any{"status": "INACTIVE"}, adminToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data.Status != models.StatusInactive {
			t.Fatalf("sub-material was not deactivated: %+v", result.Data)
		}
	})
}

func requireAdminStatus(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if response.Code != expected {
		t.Fatalf("response status = %d, want %d; body=%s", response.Code, expected, response.Body.String())
	}
}
