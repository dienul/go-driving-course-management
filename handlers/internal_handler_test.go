package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestInternalStatsReturnsOperationalCounts(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("create GORM database: %v", err)
	}
	columns := []string{
		"total_students", "total_trainers", "total_admins", "total_enrollments",
		"active_enrollments", "total_training_sessions", "scheduled_training_sessions",
		"in_progress_training_sessions", "completed_training_sessions", "paid_payments",
		"total_certificates", "total_trainer_reviews",
	}
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(columns).AddRow(
		int64(21), int64(8), int64(2), int64(30), int64(11), int64(90),
		int64(12), int64(3), int64(70), int64(25), int64(15), int64(44),
	))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/internal/stats", handlers.NewInternalHandler(db).Stats)
	request := httptest.NewRequest(http.MethodGet, "/internal/stats", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response dto.InternalStatsAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	expected := dto.InternalStatsData{
		TotalStudents: 21, TotalTrainers: 8, TotalAdmins: 2, TotalEnrollments: 30,
		ActiveEnrollments: 11, TotalTrainingSessions: 90, ScheduledTrainingSessions: 12,
		InProgressTrainingSessions: 3, CompletedTrainingSessions: 70, PaidPayments: 25,
		TotalCertificates: 15, TotalTrainerReviews: 44,
	}
	if !response.Success || response.Message != "success" || response.Data != expected {
		t.Fatalf("unexpected response: %+v; expected counts: %+v", response, expected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations: %v", err)
	}
}

func TestInternalStatsHidesDatabaseFailures(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("create GORM database: %v", err)
	}
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("private connection details"))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/internal/stats", handlers.NewInternalHandler(db).Stats)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/internal/stats", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Success || response.Message != "database is unavailable" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if strings.Contains(recorder.Body.String(), "private connection details") {
		t.Fatal("response exposes private database error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations: %v", err)
	}
}

func TestInternalHealthChecksDatabaseAvailability(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "available", status: http.StatusOK},
		{name: "unavailable", err: errors.New("private connection details"), status: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("create SQL mock: %v", err)
			}
			defer sqlDB.Close()
			db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				t.Fatalf("create GORM database: %v", err)
			}
			expectation := mock.ExpectPing()
			if test.err != nil {
				expectation.WillReturnError(test.err)
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/internal/health", handlers.NewInternalHandler(db).Health)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/internal/health", nil))

			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			if strings.Contains(recorder.Body.String(), "private connection details") {
				t.Fatal("response exposes private database error")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("database expectations: %v", err)
			}
		})
	}
}
