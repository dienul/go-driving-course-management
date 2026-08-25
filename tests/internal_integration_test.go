package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/handlers"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestInternalEndpointsPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{
		JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		BasicAuthUser: "phase10-observer", BasicAuthPass: "phase10-secret",
	})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	perform := func(path, username, password, authorization string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		if username != "" {
			request.SetBasicAuth(username, password)
		} else if authorization != "" {
			request.Header.Set("Authorization", authorization)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}

	for _, endpoint := range []string{"/api/v1/internal/health", "/api/v1/internal/stats"} {
		t.Run(endpoint+" rejects missing credentials", func(t *testing.T) {
			response := perform(endpoint, "", "", "")
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if challenge := response.Header().Get("WWW-Authenticate"); challenge != `Basic realm="internal"` {
				t.Fatalf("WWW-Authenticate = %q", challenge)
			}
		})
		t.Run(endpoint+" rejects invalid credentials", func(t *testing.T) {
			response := perform(endpoint, "phase10-observer", "wrong-password", "")
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
		t.Run(endpoint+" rejects JWT bearer tokens", func(t *testing.T) {
			response := perform(endpoint, "", "", "Bearer administrator-jwt")
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}

	t.Run("authenticated internal health", func(t *testing.T) {
		response := perform("/api/v1/internal/health", "phase10-observer", "phase10-secret", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
		}
		var health handlers.HealthResponse
		if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
			t.Fatalf("decode health: %v", err)
		}
		if !health.Success || health.Message != "service is running" || health.Data != nil {
			t.Fatalf("unexpected health response: %+v", health)
		}
	})

	t.Run("authenticated statistics match PostgreSQL", func(t *testing.T) {
		response := perform("/api/v1/internal/stats", "phase10-observer", "phase10-secret", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
		}
		var stats dto.InternalStatsAPIResponse
		if err := json.Unmarshal(response.Body.Bytes(), &stats); err != nil {
			t.Fatalf("decode statistics: %v", err)
		}
		count := func(table, condition string, arguments ...any) int64 {
			t.Helper()
			query := db.Table(table)
			if condition != "" {
				query = query.Where(condition, arguments...)
			}
			var result int64
			if err := query.Count(&result).Error; err != nil {
				t.Fatalf("count %s: %v", table, err)
			}
			return result
		}
		expected := dto.InternalStatsData{
			TotalStudents:              count("users", "role = ?", models.RoleStudent),
			TotalTrainers:              count("users", "role = ?", models.RoleTrainer),
			TotalAdmins:                count("users", "role = ?", models.RoleAdmin),
			TotalEnrollments:           count("student_enrollments", ""),
			ActiveEnrollments:          count("student_enrollments", "status = ?", models.EnrollmentActive),
			TotalTrainingSessions:      count("training_sessions", ""),
			ScheduledTrainingSessions:  count("training_sessions", "status = ?", models.SessionScheduled),
			InProgressTrainingSessions: count("training_sessions", "status = ?", models.SessionInProgress),
			CompletedTrainingSessions:  count("training_sessions", "status = ?", models.SessionCompleted),
			PaidPayments:               count("payments", "status = ?", models.PaymentPaid),
			TotalCertificates:          count("certificates", ""),
			TotalTrainerReviews:        count("trainer_reviews", ""),
		}
		if !stats.Success || stats.Message != "success" || stats.Data != expected {
			t.Fatalf("statistics = %+v; expected counts: %+v", stats, expected)
		}
	})

	t.Run("public health remains unauthenticated", func(t *testing.T) {
		response := perform("/health", "", "", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("basic credentials cannot access JWT routes", func(t *testing.T) {
		response := perform("/api/v1/student/skills", "phase10-observer", "phase10-secret", "")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("missing configured credentials fail closed", func(t *testing.T) {
		unconfigured, err := routes.New(db, config.Config{
			JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		})
		if err != nil {
			t.Fatalf("create router: %v", err)
		}
		for _, endpoint := range []string{"/api/v1/internal/health", "/api/v1/internal/stats"} {
			request := httptest.NewRequest(http.MethodGet, endpoint, nil)
			request.SetBasicAuth("phase10-observer", "phase10-secret")
			response := httptest.NewRecorder()
			unconfigured.ServeHTTP(response, request)
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("%s status = %d, want %d", endpoint, response.Code, http.StatusInternalServerError)
			}
		}
	})
}
