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

func TestTrainingSessionBookingPostgreSQL(t *testing.T) {
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
	staff := []models.User{
		{Name: "Phase 7 Admin", Email: "phase7-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive},
		{Name: "Phase 7 Trainer One", Email: "phase7-trainer-one@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 7 Trainer Two", Email: "phase7-trainer-two@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
	}
	for i := range staff {
		if err := db.Create(&staff[i]).Error; err != nil {
			t.Fatalf("create staff fixture: %v", err)
		}
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{
		JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		BasicAuthUser: "phase10-internal", BasicAuthPass: "phase10-password",
	})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
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
	adminToken := login(staff[0].Email)
	trainerOneToken := login(staff[1].Email)
	trainerTwoToken := login(staff[2].Email)
	students := make([]dto.UserResponse, 0, 5)
	tokens := make([]string, 0, 5)
	for index := 1; index <= 5; index++ {
		email := fmt.Sprintf("phase7-student-%d@example.com", index)
		response := performJSON(t, router, http.MethodPost, "/api/users/register", map[string]any{
			"name": fmt.Sprintf("Phase 7 Student %d", index), "email": email, "password": "strong-password",
		}, "")
		requireAdminStatus(t, response, http.StatusCreated)
		var result struct {
			Data dto.RegisterData `json:"data"`
		}
		decodeResponse(t, response, &result)
		students = append(students, result.Data.User)
		tokens = append(tokens, login(email))
	}
	pkg := models.CoursePackage{Name: "Phase 7 Booking Package", Level: models.PackageLevelPemula, TotalHours: 6, Price: 900000, Status: models.StatusActive}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatalf("create package fixture: %v", err)
	}
	for index := 0; index < 4; index++ {
		created := performJSON(t, router, http.MethodPost, "/api/v1/student/enrollments", map[string]any{"package_id": pkg.ID}, tokens[index])
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data dto.EnrollmentCheckoutData `json:"data"`
		}
		decodeResponse(t, created, &result)
		paid := performJSON(t, router, http.MethodPost, fmt.Sprintf("/api/v1/student/payments/%d/pay", result.Data.Payment.ID),
			map[string]any{"payment_method": "CASH"}, tokens[index])
		requireAdminStatus(t, paid, http.StatusOK)
	}
	createAvailability := func(token, day, start, end string, publish bool) dto.AvailabilityData {
		t.Helper()
		created := performJSON(t, router, http.MethodPost, "/api/v1/trainer/availabilities", map[string]any{
			"available_date": day, "start_time": start, "end_time": end,
		}, token)
		requireAdminStatus(t, created, http.StatusCreated)
		var result struct {
			Data dto.AvailabilityData `json:"data"`
		}
		decodeResponse(t, created, &result)
		if publish {
			published := performJSON(t, router, http.MethodPost,
				fmt.Sprintf("/api/v1/admin/trainer-availabilities/%d/publish", result.Data.ID), nil, adminToken)
			requireAdminStatus(t, published, http.StatusOK)
			decodeResponse(t, published, &result)
		}
		return result.Data
	}
	mondayMorning := createAvailability(trainerOneToken, "2026-09-07", "08:00", "12:00", true)
	mondayAfternoon := createAvailability(trainerOneToken, "2026-09-07", "13:00", "17:00", true)
	otherTrainerMonday := createAvailability(trainerTwoToken, "2026-09-07", "08:00", "12:00", true)
	tuesdayMorning := createAvailability(trainerOneToken, "2026-09-08", "08:00", "12:00", true)
	concurrentRange := createAvailability(trainerTwoToken, "2026-09-08", "08:00", "12:00", true)
	unpublished := createAvailability(trainerOneToken, "2026-09-09", "08:00", "12:00", false)
	endpoint := "/api/v1/student/training-sessions"
	request := func(availabilityID int64, start string) map[string]any {
		return map[string]any{"trainer_availability_id": availabilityID, "start_time": start}
	}

	t.Run("authorization active enrollment and slot validation", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, trainerOneToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, adminToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/training-sessions", nil, tokens[0]), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(mondayMorning.ID, "08:00"), tokens[4]), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(unpublished.ID, "08:00"), tokens[0]), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(999999, "08:00"), tokens[0]), http.StatusNotFound)
		for _, value := range []string{"07:00", "08:30", "16:00", "11:00"} {
			requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(mondayMorning.ID, value), tokens[0]), http.StatusUnprocessableEntity)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{}, tokens[0]), http.StatusBadRequest)
		if err := db.Model(&models.User{}).Where("id = ?", staff[1].ID).Update("status", models.StatusInactive).Error; err != nil {
			t.Fatalf("deactivate trainer: %v", err)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(mondayMorning.ID, "08:00"), tokens[0]), http.StatusConflict)
		if err := db.Model(&models.User{}).Where("id = ?", staff[1].ID).Update("status", models.StatusActive).Error; err != nil {
			t.Fatalf("reactivate trainer: %v", err)
		}
	})

	var original, otherSession dto.TrainingSessionData
	t.Run("authoritative booking ownership and trainer conflict", func(t *testing.T) {
		body := request(mondayMorning.ID, "08:00")
		body["session_number"] = 99
		body["end_time"] = "17:00"
		body["trainer_id"] = staff[2].ID
		booked := performJSON(t, router, http.MethodPost, endpoint, body, tokens[0])
		requireAdminStatus(t, booked, http.StatusCreated)
		var result struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, booked, &result)
		original = result.Data
		if original.SessionNumber != 1 || original.TrainerID != staff[1].ID ||
			original.StartTime != "08:00" || original.EndTime != "10:00" ||
			original.ScheduledDate != "2026-09-07" || original.Status != models.SessionScheduled {
			t.Fatalf("client overrode authoritative booking fields: %+v", original)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(mondayAfternoon.ID, "13:00"), tokens[0]), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, request(mondayMorning.ID, "09:00"), tokens[1]), http.StatusConflict)
		parallel := performJSON(t, router, http.MethodPost, endpoint, request(otherTrainerMonday.ID, "09:00"), tokens[1])
		requireAdminStatus(t, parallel, http.StatusCreated)
		decodeResponse(t, parallel, &result)
		otherSession = result.Data
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, tokens[0]), http.StatusOK)
		ownPath := fmt.Sprintf("%s/%d", endpoint, original.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, tokens[0]), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, tokens[1]), http.StatusNotFound)
		schedules := performJSON(t, router, http.MethodGet, "/api/v1/student/schedules?date=2026-09-07", nil, tokens[0])
		requireAdminStatus(t, schedules, http.StatusOK)
		var scheduleResult struct {
			Data []dto.StudentScheduleData `json:"data"`
		}
		decodeResponse(t, schedules, &scheduleResult)
		for _, item := range scheduleResult.Data {
			if item.AvailabilityID == mondayMorning.ID &&
				(len(item.Slots) != 1 || item.Slots[0].StartTime != "10:00") {
				t.Fatalf("booked slot remained visible in student schedules: %+v", item)
			}
		}
	})

	var moved dto.TrainingSessionData
	t.Run("atomic student and admin rescheduling history", func(t *testing.T) {
		ownPath := fmt.Sprintf("%s/%d/reschedule", endpoint, original.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, ownPath, request(unpublished.ID, "08:00"), tokens[0]), http.StatusConflict)
		var unchanged models.TrainingSession
		if err := db.First(&unchanged, original.ID).Error; err != nil || unchanged.Status != models.SessionScheduled {
			t.Fatalf("failed reschedule did not roll back original session: session=%+v err=%v", unchanged, err)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, ownPath, request(mondayAfternoon.ID, "13:00"), tokens[1]), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPost, ownPath, request(mondayAfternoon.ID, "13:00"), tokens[0])
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, response, &result)
		moved = result.Data
		if moved.SessionNumber != original.SessionNumber || moved.RescheduledFromID == nil ||
			*moved.RescheduledFromID != original.ID || moved.StartTime != "13:00" || moved.EndTime != "15:00" {
			t.Fatalf("student reschedule lost session history: %+v", moved)
		}
		if err := db.First(&unchanged, original.ID).Error; err != nil || unchanged.Status != models.SessionRescheduled {
			t.Fatalf("original session history not retained: session=%+v err=%v", unchanged, err)
		}
		adminPath := fmt.Sprintf("/api/v1/admin/training-sessions/%d/reschedule", moved.ID)
		response = performJSON(t, router, http.MethodPost, adminPath, request(tuesdayMorning.ID, "08:00"), adminToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data.SessionNumber != 1 || result.Data.ScheduledDate != "2026-09-08" ||
			result.Data.RescheduledFromID == nil || *result.Data.RescheduledFromID != moved.ID {
			t.Fatalf("admin reschedule did not preserve history/number: %+v", result.Data)
		}
		moved = result.Data
	})

	t.Run("audited cancellation and session-number reuse", func(t *testing.T) {
		path := fmt.Sprintf("%s/%d/cancel", endpoint, moved.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path, map[string]any{}, tokens[0]), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path, map[string]any{"cancellation_reason": "  "}, tokens[0]), http.StatusUnprocessableEntity)
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path, map[string]any{"cancellation_reason": "Not my session"}, tokens[1]), http.StatusNotFound)
		cancelled := performJSON(t, router, http.MethodPatch, path, map[string]any{"cancellation_reason": "  Student unavailable  "}, tokens[0])
		requireAdminStatus(t, cancelled, http.StatusOK)
		var result struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, cancelled, &result)
		if result.Data.Status != models.SessionCancelled || result.Data.CancelledBy == nil ||
			*result.Data.CancelledBy != students[0].ID || result.Data.CancellationReason == nil ||
			*result.Data.CancellationReason != "Student unavailable" || result.Data.CancelledAt == nil {
			t.Fatalf("student cancellation audit fields incorrect: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPatch, path, map[string]any{"cancellation_reason": "Again"}, tokens[0]), http.StatusConflict)
		firstAgain := performJSON(t, router, http.MethodPost, endpoint, request(mondayMorning.ID, "08:00"), tokens[0])
		requireAdminStatus(t, firstAgain, http.StatusCreated)
		decodeResponse(t, firstAgain, &result)
		if result.Data.SessionNumber != 1 {
			t.Fatalf("cancelled session consumed package hours: %+v", result.Data)
		}
		adminCancel := fmt.Sprintf("/api/v1/admin/training-sessions/%d/cancel", result.Data.ID)
		response := performJSON(t, router, http.MethodPatch, adminCancel, map[string]any{"cancellation_reason": "Admin scheduling change"}, adminToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data.CancelledBy == nil || *result.Data.CancelledBy != staff[0].ID {
			t.Fatalf("admin cancellation actor was not recorded: %+v", result.Data)
		}
	})

	t.Run("purchased-session capacity and automatic numbering", func(t *testing.T) {
		slots := []struct {
			availability int64
			start        string
		}{
			{mondayMorning.ID, "08:00"},
			{mondayMorning.ID, "10:00"},
			{tuesdayMorning.ID, "08:00"},
		}
		for index, slot := range slots {
			booked := performJSON(t, router, http.MethodPost, endpoint, request(slot.availability, slot.start), tokens[0])
			requireAdminStatus(t, booked, http.StatusCreated)
			var result struct {
				Data dto.TrainingSessionData `json:"data"`
			}
			decodeResponse(t, booked, &result)
			if result.Data.SessionNumber != index+1 {
				t.Fatalf("session number = %d, want %d", result.Data.SessionNumber, index+1)
			}
			now := time.Now().UTC()
			if err := db.Model(&models.TrainingSession{}).Where("id = ?", result.Data.ID).Updates(map[string]any{
				"status": models.SessionCompleted, "actual_started_at": now, "actual_completed_at": now,
			}).Error; err != nil {
				t.Fatalf("mark fixture session complete: %v", err)
			}
		}
		exhausted := performJSON(t, router, http.MethodPost, endpoint, request(tuesdayMorning.ID, "10:00"), tokens[0])
		requireAdminStatus(t, exhausted, http.StatusConflict)
	})

	t.Run("administrator history, trainer ownership, and internal auth isolation", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/training-sessions", nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet,
			fmt.Sprintf("/api/v1/admin/training-sessions/%d", otherSession.ID), nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/trainer/training-sessions", nil, trainerOneToken), http.StatusOK)
		trainerPath := fmt.Sprintf("/api/v1/trainer/training-sessions/%d", otherSession.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, trainerPath, nil, trainerOneToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, trainerPath, nil, trainerTwoToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, tokens[0]), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/trainer/reviews", nil, trainerOneToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/internal/stats", nil, adminToken), http.StatusUnauthorized)
	})

	t.Run("concurrent trainer double-booking protection", func(t *testing.T) {
		statuses := make(chan int, 2)
		var group sync.WaitGroup
		for _, token := range []string{tokens[2], tokens[3]} {
			bookingToken := token
			group.Add(1)
			go func() {
				defer group.Done()
				response := performJSON(t, router, http.MethodPost, endpoint,
					request(concurrentRange.ID, "08:00"), bookingToken)
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
			t.Fatalf("concurrent trainer double booking was not prevented: created=%d conflicts=%d", created, conflicts)
		}
		var count int64
		if err := db.Model(&models.TrainingSession{}).
			Where("trainer_id = ? AND scheduled_date = ? AND start_time = ? AND status = ?",
				staff[2].ID, "2026-09-08", "08:00:00", models.SessionScheduled).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("concurrent slot has %d scheduled sessions, err=%v", count, err)
		}
	})
}
