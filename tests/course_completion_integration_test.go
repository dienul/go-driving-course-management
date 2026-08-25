package tests

import (
	"fmt"
	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestCourseCompletionPostgreSQL(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := []models.User{
		{Name: "Phase 9 Admin", Email: "phase9-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive},
		{Name: "Phase 9 Trainer", Email: "phase9-trainer@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 9 Other Trainer", Email: "phase9-other-trainer@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 9 Student", Email: "phase9-student@example.com", PasswordHash: string(hash), Role: models.RoleStudent, Status: models.StatusActive},
		{Name: "Phase 9 Other Student", Email: "phase9-other-student@example.com", PasswordHash: string(hash), Role: models.RoleStudent, Status: models.StatusActive},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{
		JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		BasicAuthUser: "phase10-internal", BasicAuthPass: "phase10-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	login := func(user models.User) string {
		t.Helper()
		response := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{"email": user.Email, "password": "strong-password"}, "")
		requireAdminStatus(t, response, http.StatusOK)
		var value struct {
			Data dto.LoginData `json:"data"`
		}
		decodeResponse(t, response, &value)
		return "Bearer " + value.Data.Token
	}
	adminToken, trainerToken, otherTrainerToken := login(users[0]), login(users[1]), login(users[2])
	studentToken, otherStudentToken := login(users[3]), login(users[4])
	now := time.Now().UTC()
	past1, past2 := now.Add(-48*time.Hour), now.Add(-24*time.Hour)
	pkg := models.CoursePackage{Name: "Phase 9 Course", Level: models.PackageLevelPemula, TotalHours: 6, Price: 900000, Status: models.StatusActive}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	enrollment := models.StudentEnrollment{StudentID: users[3].ID, PackageID: pkg.ID, PackageName: pkg.Name, PackagePrice: pkg.Price, TotalHours: 6, Status: models.EnrollmentActive, StartedAt: &past1}
	if err := db.Create(&enrollment).Error; err != nil {
		t.Fatal(err)
	}
	otherEnrollment := models.StudentEnrollment{StudentID: users[4].ID, PackageID: pkg.ID, PackageName: pkg.Name, PackagePrice: pkg.Price, TotalHours: 6, Status: models.EnrollmentActive, StartedAt: &past1}
	if err := db.Create(&otherEnrollment).Error; err != nil {
		t.Fatal(err)
	}
	material := models.Material{Name: "Phase 9 Skills", Sequence: 9001, Status: models.StatusActive}
	if err := db.Create(&material).Error; err != nil {
		t.Fatal(err)
	}
	subs := []models.SubMaterial{{MaterialID: material.ID, Name: "Phase 9 Steering", Sequence: 1, Status: models.StatusActive}, {MaterialID: material.ID, Name: "Phase 9 Parking", Sequence: 2, Status: models.StatusActive}, {MaterialID: material.ID, Name: "Phase 9 Highway", Sequence: 3, Status: models.StatusActive}, {MaterialID: material.ID, Name: "Phase 9 Hidden", Sequence: 4, Status: models.StatusInactive}}
	for i := range subs {
		if err := db.Create(&subs[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	clock := func(value string) models.ClockTime {
		parsed, err := time.Parse("15:04", value)
		if err != nil {
			t.Fatal(err)
		}
		return models.NewClockTime(parsed)
	}
	day, _ := time.Parse("2006-01-02", "2026-09-21")
	availability := models.TrainerAvailability{TrainerID: users[1].ID, AvailableDate: day, StartTime: clock("08:00"), EndTime: clock("17:00"), Status: models.AvailabilityPublished, PublishedBy: &users[0].ID, PublishedAt: &now}
	if err := db.Create(&availability).Error; err != nil {
		t.Fatal(err)
	}
	sessions := []models.TrainingSession{
		{EnrollmentID: enrollment.ID, TrainerID: users[1].ID, TrainerAvailabilityID: availability.ID, SessionNumber: 1, ScheduledDate: day, StartTime: clock("08:00"), EndTime: clock("10:00"), Status: models.SessionCompleted, ActualStartedAt: &past1, ActualCompletedAt: &past1},
		{EnrollmentID: enrollment.ID, TrainerID: users[1].ID, TrainerAvailabilityID: availability.ID, SessionNumber: 2, ScheduledDate: day, StartTime: clock("10:00"), EndTime: clock("12:00"), Status: models.SessionCompleted, ActualStartedAt: &past2, ActualCompletedAt: &past2},
		{EnrollmentID: enrollment.ID, TrainerID: users[1].ID, TrainerAvailabilityID: availability.ID, SessionNumber: 3, ScheduledDate: day, StartTime: clock("12:00"), EndTime: clock("14:00"), Status: models.SessionInProgress, ActualStartedAt: &now},
	}
	for i := range sessions {
		if err := db.Create(&sessions[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	assessments := []models.SessionSkillAssessment{
		{TrainingSessionID: sessions[0].ID, SubMaterialID: subs[0].ID, SkillStatus: models.SkillMastered},
		{TrainingSessionID: sessions[1].ID, SubMaterialID: subs[0].ID, SkillStatus: models.SkillPracticed},
		{TrainingSessionID: sessions[1].ID, SubMaterialID: subs[1].ID, SkillStatus: models.SkillMastered},
		{TrainingSessionID: sessions[2].ID, SubMaterialID: subs[0].ID, SkillStatus: models.SkillMastered},
	}
	for i := range assessments {
		if err := db.Create(&assessments[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	finalEvaluation := models.SessionEvaluation{TrainingSessionID: sessions[2].ID, Predicate: models.PredicateBaik, Notes: "Course time fulfilled.", Recommendation: "Continue parking practice."}
	if err := db.Create(&finalEvaluation).Error; err != nil {
		t.Fatal(err)
	}

	var before dto.StudentSkillsData
	t.Run("global latest completed skill and active curriculum scoring", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, trainerToken), http.StatusForbidden)
		response := performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.StudentSkillsData `json:"data"`
		}
		decodeResponse(t, response, &result)
		before = result.Data
		found := map[int64]models.SkillStatus{}
		for _, skill := range before.Skills {
			found[skill.SubMaterialID] = skill.SkillStatus
		}
		if found[subs[0].ID] != models.SkillPracticed || found[subs[1].ID] != models.SkillMastered || found[subs[2].ID] != models.SkillNotStarted {
			t.Fatalf("latest completed skill, decline, or default incorrect: %+v", found)
		}
		if _, exists := found[subs[3].ID]; exists {
			t.Fatal("inactive skill leaked into score")
		}
		history := performJSON(t, router, http.MethodGet, "/api/v1/student/skills/history", nil, studentToken)
		requireAdminStatus(t, history, http.StatusOK)
		var values struct {
			Data []dto.StudentSkillHistoryData `json:"data"`
		}
		decodeResponse(t, history, &values)
		if len(values.Data) != 3 {
			t.Fatalf("in-progress assessment leaked into completed history: %+v", values.Data)
		}
	})

	reviewPath := fmt.Sprintf("/api/v1/student/training-sessions/%d/review", sessions[0].ID)
	secondReviewPath := fmt.Sprintf("/api/v1/student/training-sessions/%d/review", sessions[1].ID)
	t.Run("review ownership uniqueness updates and calculated trainer averages", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, reviewPath, map[string]any{"rating": 6}, studentToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, reviewPath, map[string]any{"rating": 0}, studentToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, reviewPath, map[string]any{"rating": 5}, otherStudentToken), http.StatusNotFound)
		unfinished := fmt.Sprintf("/api/v1/student/training-sessions/%d/review", sessions[2].ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, unfinished, map[string]any{"rating": 5}, studentToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, reviewPath, nil, studentToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPost, reviewPath, map[string]any{"rating": 5, "feedback": "  Very patient  "}, studentToken)
		requireAdminStatus(t, response, http.StatusCreated)
		var review struct {
			Data dto.TrainerReviewData `json:"data"`
		}
		decodeResponse(t, response, &review)
		if review.Data.TrainerID != users[1].ID || review.Data.StudentID != users[3].ID || review.Data.Feedback == nil || *review.Data.Feedback != "Very patient" {
			t.Fatalf("review ownership or feedback incorrect: %+v", review.Data)
		}
		firstID := review.Data.ID
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, reviewPath, map[string]any{"rating": 4}, studentToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, secondReviewPath, map[string]any{"rating": 4}, studentToken), http.StatusNotFound)
		response = performJSON(t, router, http.MethodPut, reviewPath, map[string]any{"rating": 4, "feedback": "  Updated  "}, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &review)
		if review.Data.ID != firstID || review.Data.Rating != 4 {
			t.Fatalf("review update was not in place: %+v", review.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, secondReviewPath, map[string]any{"rating": 2}, studentToken), http.StatusCreated)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, reviewPath, nil, otherStudentToken), http.StatusNotFound)
		summary := performJSON(t, router, http.MethodGet, "/api/v1/trainer/reviews/summary", nil, trainerToken)
		requireAdminStatus(t, summary, http.StatusOK)
		var rating struct {
			Data dto.TrainerReviewSummaryData `json:"data"`
		}
		decodeResponse(t, summary, &rating)
		if rating.Data.TotalReviews != 2 || rating.Data.AverageRating != 3 {
			t.Fatalf("trainer average incorrect: %+v", rating.Data)
		}
		empty := performJSON(t, router, http.MethodGet, "/api/v1/trainer/reviews/summary", nil, otherTrainerToken)
		requireAdminStatus(t, empty, http.StatusOK)
		decodeResponse(t, empty, &rating)
		if rating.Data.TotalReviews != 0 || rating.Data.AverageRating != 0 {
			t.Fatalf("other trainer inherited ratings: %+v", rating.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/trainer/reviews", nil, trainerToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainer-reviews", nil, studentToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainer-reviews", nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("/api/v1/admin/trainers/%d/reviews", users[1].ID), nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/trainers/999999/reviews", nil, adminToken), http.StatusNotFound)
	})

	var certificate dto.CertificateData
	t.Run("final completion atomically completes enrollment and issues unique certificate", func(t *testing.T) {
		prior := performJSON(t, router, http.MethodGet, "/api/v1/student/certificates", nil, studentToken)
		requireAdminStatus(t, prior, http.StatusOK)
		var empty struct {
			Data []dto.CertificateData `json:"data"`
		}
		decodeResponse(t, prior, &empty)
		if len(empty.Data) != 0 {
			t.Fatalf("certificate issued early: %+v", empty.Data)
		}
		path := fmt.Sprintf("/api/v1/trainer/training-sessions/%d/complete", sessions[2].ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, path, nil, otherTrainerToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, path, nil, trainerToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, path, nil, trainerToken), http.StatusConflict)
		var updated models.StudentEnrollment
		if err := db.First(&updated, enrollment.ID).Error; err != nil || updated.Status != models.EnrollmentCompleted || updated.CompletedAt == nil {
			t.Fatalf("enrollment not atomically completed: %+v %v", updated, err)
		}
		var count int64
		if err := db.Model(&models.Certificate{}).Where("enrollment_id = ?", enrollment.ID).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("certificate uniqueness failed: %d %v", count, err)
		}
		response := performJSON(t, router, http.MethodGet, "/api/v1/student/certificates", nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &empty)
		if len(empty.Data) != 1 {
			t.Fatalf("missing student certificate: %+v", empty.Data)
		}
		certificate = empty.Data[0]
		if !regexp.MustCompile("^CERT-[0-9]{8}-[0-9]{4}$").MatchString(certificate.CertificateNumber) || certificate.Student.ID != users[3].ID || certificate.Enrollment.Status != models.EnrollmentCompleted {
			t.Fatalf("certificate data invalid: %+v", certificate)
		}
		current := performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, studentToken)
		var skills struct {
			Data dto.StudentSkillsData `json:"data"`
		}
		decodeResponse(t, current, &skills)
		if certificate.SkillScore != skills.Data.SkillScore || certificate.SkillLevel != skills.Data.SkillLevel || certificate.SkillScore <= before.SkillScore {
			t.Fatalf("certificate did not capture final global skill: certificate=%+v current=%+v", certificate, skills.Data)
		}
		own := fmt.Sprintf("/api/v1/student/certificates/%d", certificate.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, own, nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, own, nil, otherStudentToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/certificates", nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("/api/v1/admin/certificates/%d", certificate.ID), nil, adminToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/certificates", nil, studentToken), http.StatusForbidden)
	})

	t.Run("global skill continues across enrollments while certificate snapshot stays immutable", func(t *testing.T) {
		next := models.StudentEnrollment{StudentID: users[3].ID, PackageID: pkg.ID, PackageName: pkg.Name, PackagePrice: pkg.Price, TotalHours: 6, Status: models.EnrollmentActive, StartedAt: &now}
		if err := db.Create(&next).Error; err != nil {
			t.Fatal(err)
		}
		later := now.Add(time.Hour)
		additional := models.TrainingSession{EnrollmentID: next.ID, TrainerID: users[1].ID, TrainerAvailabilityID: availability.ID, SessionNumber: 1, ScheduledDate: day, StartTime: clock("14:00"), EndTime: clock("16:00"), Status: models.SessionCompleted, ActualStartedAt: &later, ActualCompletedAt: &later}
		if err := db.Create(&additional).Error; err != nil {
			t.Fatal(err)
		}
		decline := models.SessionSkillAssessment{TrainingSessionID: additional.ID, SubMaterialID: subs[0].ID, SkillStatus: models.SkillNotStarted}
		if err := db.Create(&decline).Error; err != nil {
			t.Fatal(err)
		}
		response := performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var latest struct {
			Data dto.StudentSkillsData `json:"data"`
		}
		decodeResponse(t, response, &latest)
		if latest.Data.SkillScore >= certificate.SkillScore {
			t.Fatalf("global skill did not decline in later enrollment: current=%+v certificate=%+v", latest.Data, certificate)
		}
		response = performJSON(t, router, http.MethodGet, fmt.Sprintf("/api/v1/student/certificates/%d", certificate.ID), nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var saved struct {
			Data dto.CertificateData `json:"data"`
		}
		decodeResponse(t, response, &saved)
		if saved.Data.SkillScore != certificate.SkillScore || saved.Data.SkillLevel != certificate.SkillLevel {
			t.Fatalf("certificate snapshot changed: %+v", saved.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/internal/health", nil, adminToken), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/internal/stats", nil, adminToken), http.StatusUnauthorized)
	})
}
