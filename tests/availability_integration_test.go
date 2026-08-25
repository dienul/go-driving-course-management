package tests

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAvailabilityLifecyclePostgreSQL(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash fixture password: %v", err)
	}
	users := []models.User{
		{Name: "Phase 6 Admin", Email: "phase6-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive},
		{Name: "Phase 6 Trainer One", Email: "phase6-trainer-one@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 6 Trainer Two", Email: "phase6-trainer-two@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user fixture: %v", err)
		}
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h"})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
	register := performJSON(t, router, http.MethodPost, "/api/users/register", map[string]any{
		"name": "Phase 6 Student", "email": "phase6-student@example.com", "password": "strong-password",
	}, "")
	requireAdminStatus(t, register, http.StatusCreated)
	var student struct {
		Data dto.RegisterData `json:"data"`
	}
	decodeResponse(t, register, &student)
	login := func(email string) string {
		t.Helper()
		response := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
			"email": email, "password": "strong-password",
		}, "")
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.LoginData `json:"data"`
		}
		decodeResponse(t, response, &result)
		return "Bearer " + result.Data.Token
	}
	adminToken := login(users[0].Email)
	trainerToken := login(users[1].Email)
	otherTrainerToken := login(users[2].Email)
	studentToken := login("phase6-student@example.com")
	endpoint := "/api/v1/trainer/availabilities"
	date := "2026-08-31"
	createBody := func(day, start, end string) map[string]any {
		return map[string]any{"available_date": day, "start_time": start, "end_time": end}
	}

	t.Run("role authorization and operating-hour validation", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, studentToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, adminToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainer-availabilities", nil, trainerToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/schedules", nil, trainerToken), http.StatusForbidden)
		invalid := []map[string]any{
			createBody("2026-08-29", "08:00", "12:00"),
			createBody("2026-08-30", "08:00", "12:00"),
			createBody(date, "07:00", "12:00"),
			createBody(date, "08:00", "18:00"),
			createBody(date, "08:30", "12:00"),
			createBody(date, "08:00", "12:30"),
			createBody(date, "12:00", "12:00"),
		}
		for _, body := range invalid {
			requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, body, trainerToken), http.StatusUnprocessableEntity)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{}, trainerToken), http.StatusBadRequest)
	})

	var first, second, other dto.AvailabilityData
	t.Run("ownership non-overlap and pending lifecycle", func(t *testing.T) {
		body := createBody(date, "08:00", "12:00")
		body["trainer_id"] = users[2].ID
		body["status"] = "PUBLISHED"
		created := performJSON(t, router, http.MethodPost, endpoint, body, trainerToken)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data dto.AvailabilityData `json:"data"`
		}
		decodeResponse(t, created, &result)
		first = result.Data
		if first.TrainerID != users[1].ID || first.Status != models.AvailabilityPending || first.PublishedBy != nil {
			t.Fatalf("client overrode trainer ownership/publication: %+v", first)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, trainerToken), http.StatusOK)
		ownPath := fmt.Sprintf("%s/%d", endpoint, first.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, trainerToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, otherTrainerToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, ownPath, createBody(date, "09:00", "12:00"), otherTrainerToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, ownPath+"/cancel", nil, otherTrainerToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, createBody(date, "11:00", "14:00"), trainerToken), http.StatusConflict)
		adjacent := performJSON(t, router, http.MethodPost, endpoint, createBody(date, "12:00", "14:00"), trainerToken)
		requireAdminStatus(t, adjacent, http.StatusCreated)
		decodeResponse(t, adjacent, &result)
		second = result.Data
		updatePath := fmt.Sprintf("%s/%d", endpoint, second.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, updatePath, createBody(date, "11:00", "15:00"), trainerToken), http.StatusConflict)
		updated := performJSON(t, router, http.MethodPut, updatePath, createBody(date, "13:00", "17:00"), trainerToken)
		requireAdminStatus(t, updated, http.StatusOK)
		decodeResponse(t, updated, &result)
		second = result.Data
		parallel := performJSON(t, router, http.MethodPost, endpoint, createBody(date, "08:00", "12:00"), otherTrainerToken)
		requireAdminStatus(t, parallel, http.StatusCreated)
		decodeResponse(t, parallel, &result)
		other = result.Data
		unpublished := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, unpublished, http.StatusOK)
		var schedules struct {
			Data []dto.StudentScheduleData `json:"data"`
		}
		decodeResponse(t, unpublished, &schedules)
		if len(schedules.Data) != 0 {
			t.Fatalf("pending schedules were visible to students: %+v", schedules.Data)
		}
	})

	t.Run("admin publication and student slot calculation", func(t *testing.T) {
		adminPath := "/api/v1/admin/trainer-availabilities"
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, adminPath, nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("%s/%d", adminPath, first.ID), nil, adminToken), http.StatusOK)
		publish := performJSON(t, router, http.MethodPost, fmt.Sprintf("%s/%d/publish", adminPath, first.ID), nil, adminToken)
		requireAdminStatus(t, publish, http.StatusOK)
		var published struct {
			Data dto.AvailabilityData `json:"data"`
		}
		decodeResponse(t, publish, &published)
		if published.Data.Status != models.AvailabilityPublished || published.Data.PublishedBy == nil ||
			*published.Data.PublishedBy != users[0].ID || published.Data.PublishedAt == nil {
			t.Fatalf("publication audit metadata missing: %+v", published.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, fmt.Sprintf("%s/%d/publish", adminPath, first.ID), nil, adminToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, fmt.Sprintf("%s/%d", endpoint, first.ID), createBody(date, "08:00", "11:00"), trainerToken), http.StatusConflict)
		filtered := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, filtered, http.StatusOK)
		var schedules struct {
			Data []dto.StudentScheduleData `json:"data"`
		}
		decodeResponse(t, filtered, &schedules)
		if len(schedules.Data) != 1 || len(schedules.Data[0].Slots) != 3 ||
			schedules.Data[0].Slots[0].StartTime != "08:00" || schedules.Data[0].Slots[2].EndTime != "12:00" {
			t.Fatalf("incorrect published two-hour slots: %+v", schedules.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date=31-08-2026", nil, studentToken), http.StatusUnprocessableEntity)
		wrongDate := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date=2026-09-01", nil, studentToken)
		requireAdminStatus(t, wrongDate, http.StatusOK)
		decodeResponse(t, wrongDate, &schedules)
		if len(schedules.Data) != 0 {
			t.Fatalf("date filter included another day: %+v", schedules.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, fmt.Sprintf("%s/%d/publish", adminPath, second.ID), nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, fmt.Sprintf("%s/%d/publish", adminPath, other.ID), nil, adminToken), http.StatusOK)
		multiple := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, multiple, http.StatusOK)
		decodeResponse(t, multiple, &schedules)
		if len(schedules.Data) != 3 {
			t.Fatalf("parallel trainer or adjacent published ranges missing: %+v", schedules.Data)
		}
		if err := db.Model(&models.User{}).Where("id = ?", users[2].ID).Update("status", models.StatusInactive).Error; err != nil {
			t.Fatalf("deactivate trainer: %v", err)
		}
		hidden := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, hidden, http.StatusOK)
		decodeResponse(t, hidden, &schedules)
		if len(schedules.Data) != 2 {
			t.Fatalf("inactive trainer remained visible: %+v", schedules.Data)
		}
		if err := db.Model(&models.User{}).Where("id = ?", users[2].ID).Update("status", models.StatusActive).Error; err != nil {
			t.Fatalf("reactivate trainer: %v", err)
		}
	})

	t.Run("booked slot exclusion and cancellation safety", func(t *testing.T) {
		pkg := models.CoursePackage{Name: "Phase 6 Schedule Package", Level: models.PackageLevelPemula, TotalHours: 6, Price: 900000, Status: models.StatusActive}
		if err := db.Create(&pkg).Error; err != nil {
			t.Fatalf("create package fixture: %v", err)
		}
		started := time.Now().UTC()
		enrollment := models.StudentEnrollment{
			StudentID: student.Data.User.ID, PackageID: pkg.ID, PackageName: pkg.Name,
			PackagePrice: pkg.Price, TotalHours: pkg.TotalHours, Status: models.EnrollmentActive, StartedAt: &started,
		}
		if err := db.Create(&enrollment).Error; err != nil {
			t.Fatalf("create active enrollment fixture: %v", err)
		}
		day, _ := time.Parse("2006-01-02", date)
		start, _ := time.Parse("15:04", "09:00")
		end, _ := time.Parse("15:04", "11:00")
		session := models.TrainingSession{
			EnrollmentID: enrollment.ID, TrainerID: users[1].ID, TrainerAvailabilityID: first.ID,
			SessionNumber: 1, ScheduledDate: day, StartTime: models.NewClockTime(start), EndTime: models.NewClockTime(end), Status: models.SessionScheduled,
		}
		if err := db.Create(&session).Error; err != nil {
			t.Fatalf("create scheduled-session fixture: %v", err)
		}
		response := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var schedules struct {
			Data []dto.StudentScheduleData `json:"data"`
		}
		decodeResponse(t, response, &schedules)
		for _, item := range schedules.Data {
			if item.AvailabilityID == first.ID {
				t.Fatalf("occupied availability exposed overlapping slots: %+v", item)
			}
		}
		ownCancel := fmt.Sprintf("%s/%d/cancel", endpoint, first.ID)
		adminCancel := fmt.Sprintf("/api/v1/admin/trainer-availabilities/%d/cancel", first.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, ownCancel, nil, trainerToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, adminCancel, nil, adminToken), http.StatusConflict)
		reason := "Fixture cancellation"
		cancelledAt := time.Now().UTC()
		if err := db.Model(&session).Updates(map[string]any{
			"status": models.SessionCancelled, "cancelled_by": users[0].ID,
			"cancellation_reason": reason, "cancelled_at": cancelledAt,
		}).Error; err != nil {
			t.Fatalf("cancel fixture session: %v", err)
		}
		reopened := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date="+date, nil, studentToken)
		requireAdminStatus(t, reopened, http.StatusOK)
		decodeResponse(t, reopened, &schedules)
		found := false
		for _, item := range schedules.Data {
			if item.AvailabilityID == first.ID && len(item.Slots) == 3 {
				found = true
			}
		}
		if !found {
			t.Fatalf("cancelled training session did not restore availability slots: %+v", schedules.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, ownCancel, nil, trainerToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, ownCancel, nil, trainerToken), http.StatusConflict)
		reused := performJSON(t, router, http.MethodPost, endpoint, createBody(date, "08:00", "12:00"), trainerToken)
		requireAdminStatus(t, reused, http.StatusCreated)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch,
			fmt.Sprintf("/api/v1/admin/trainer-availabilities/%d/cancel", second.ID), nil, adminToken), http.StatusOK)
	})

	t.Run("concurrent overlapping requests are serialized", func(t *testing.T) {
		statuses := make(chan int, 2)
		var group sync.WaitGroup
		for i := 0; i < 2; i++ {
			group.Add(1)
			go func() {
				defer group.Done()
				response := performJSON(t, router, http.MethodPost, endpoint,
					createBody("2026-09-01", "08:00", "12:00"), trainerToken)
				statuses <- response.Code
			}()
		}
		group.Wait()
		close(statuses)
		created, conflicts := 0, 0
		for status := range statuses {
			if status == http.StatusCreated {
				created++
			}
			if status == http.StatusConflict {
				conflicts++
			}
		}
		if created != 1 || conflicts != 1 {
			t.Fatalf("concurrent overlap protection failed: created=%d conflicts=%d", created, conflicts)
		}
	})
}
