package tests

import (
	"fmt"
	"net/http"
	"os"
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

func TestTrainingProcessPostgreSQL(t *testing.T) {
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
		{Name: "Phase 8 Admin", Email: "phase8-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive},
		{Name: "Phase 8 Trainer One", Email: "phase8-trainer-one@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 8 Trainer Two", Email: "phase8-trainer-two@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
		{Name: "Phase 8 Student", Email: "phase8-student@example.com", PasswordHash: string(hash), Role: models.RoleStudent, Status: models.StatusActive},
		{Name: "Phase 8 Other Student", Email: "phase8-other-student@example.com", PasswordHash: string(hash), Role: models.RoleStudent, Status: models.StatusActive},
	}
	for index := range users {
		if err := db.Create(&users[index]).Error; err != nil {
			t.Fatalf("create user fixture: %v", err)
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
	login := func(user models.User) string {
		t.Helper()
		response := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
			"email": user.Email, "password": "strong-password",
		}, "")
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.LoginData `json:"data"`
		}
		decodeResponse(t, response, &result)
		return "Bearer " + result.Data.Token
	}
	adminToken, trainerOneToken := login(users[0]), login(users[1])
	trainerTwoToken, studentToken := login(users[2]), login(users[3])

	pkg := models.CoursePackage{
		Name: "Phase 8 Training Package", Level: models.PackageLevelPemula,
		TotalHours: 6, Price: 900000, Status: models.StatusActive,
	}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatalf("create package fixture: %v", err)
	}
	now := time.Now().UTC()
	oldStart, oldComplete := now.Add(-72*time.Hour), now.Add(-48*time.Hour)
	enrollments := []models.StudentEnrollment{
		{
			StudentID: users[3].ID, PackageID: pkg.ID, PackageName: pkg.Name,
			PackagePrice: pkg.Price, TotalHours: pkg.TotalHours, Status: models.EnrollmentCompleted,
			StartedAt: &oldStart, CompletedAt: &oldComplete,
		},
		{
			StudentID: users[3].ID, PackageID: pkg.ID, PackageName: pkg.Name,
			PackagePrice: pkg.Price, TotalHours: pkg.TotalHours, Status: models.EnrollmentActive,
			StartedAt: &now,
		},
		{
			StudentID: users[4].ID, PackageID: pkg.ID, PackageName: pkg.Name,
			PackagePrice: pkg.Price, TotalHours: pkg.TotalHours, Status: models.EnrollmentActive,
			StartedAt: &now,
		},
	}
	for index := range enrollments {
		if err := db.Create(&enrollments[index]).Error; err != nil {
			t.Fatalf("create enrollment fixture: %v", err)
		}
	}
	materials := []models.Material{
		{Name: "Phase 8 Vehicle Control", Sequence: 8001, Status: models.StatusActive},
		{Name: "Phase 8 Hidden Material", Sequence: 8002, Status: models.StatusInactive},
	}
	for index := range materials {
		if err := db.Create(&materials[index]).Error; err != nil {
			t.Fatalf("create material fixture: %v", err)
		}
	}
	subMaterials := []models.SubMaterial{
		{MaterialID: materials[0].ID, Name: "Phase 8 Steering", Sequence: 1, Status: models.StatusActive},
		{MaterialID: materials[0].ID, Name: "Phase 8 Parking", Sequence: 2, Status: models.StatusActive},
		{MaterialID: materials[0].ID, Name: "Phase 8 Hidden Skill", Sequence: 3, Status: models.StatusInactive},
		{MaterialID: materials[1].ID, Name: "Phase 8 Hidden Parent", Sequence: 1, Status: models.StatusActive},
	}
	for index := range subMaterials {
		if err := db.Create(&subMaterials[index]).Error; err != nil {
			t.Fatalf("create sub-material fixture: %v", err)
		}
	}
	clock := func(value string) models.ClockTime {
		parsed, err := time.Parse("15:04", value)
		if err != nil {
			t.Fatalf("parse clock fixture: %v", err)
		}
		return models.NewClockTime(parsed)
	}
	createAvailability := func(trainerID int64, value string) models.TrainerAvailability {
		t.Helper()
		day, err := time.Parse("2006-01-02", value)
		if err != nil {
			t.Fatalf("parse date fixture: %v", err)
		}
		availability := models.TrainerAvailability{
			TrainerID: trainerID, AvailableDate: day, StartTime: clock("08:00"), EndTime: clock("12:00"),
			Status: models.AvailabilityPublished, PublishedBy: &users[0].ID, PublishedAt: &now,
		}
		if err := db.Create(&availability).Error; err != nil {
			t.Fatalf("create availability fixture: %v", err)
		}
		return availability
	}
	historicalAvailability := createAvailability(users[2].ID, "2026-09-14")
	currentAvailability := createAvailability(users[1].ID, "2026-09-15")
	otherAvailability := createAvailability(users[2].ID, "2026-09-15")
	nextAvailability := createAvailability(users[2].ID, "2026-09-16")
	historical := models.TrainingSession{
		EnrollmentID: enrollments[0].ID, TrainerID: users[2].ID,
		TrainerAvailabilityID: historicalAvailability.ID, SessionNumber: 3,
		ScheduledDate: historicalAvailability.AvailableDate, StartTime: clock("08:00"), EndTime: clock("10:00"),
		Status: models.SessionCompleted, ActualStartedAt: &oldStart, ActualCompletedAt: &oldComplete,
	}
	current := models.TrainingSession{
		EnrollmentID: enrollments[1].ID, TrainerID: users[1].ID,
		TrainerAvailabilityID: currentAvailability.ID, SessionNumber: 1,
		ScheduledDate: currentAvailability.AvailableDate, StartTime: clock("08:00"), EndTime: clock("10:00"),
		Status: models.SessionScheduled,
	}
	other := models.TrainingSession{
		EnrollmentID: enrollments[2].ID, TrainerID: users[2].ID,
		TrainerAvailabilityID: otherAvailability.ID, SessionNumber: 1,
		ScheduledDate: otherAvailability.AvailableDate, StartTime: clock("08:00"), EndTime: clock("10:00"),
		Status: models.SessionScheduled,
	}
	for _, session := range []*models.TrainingSession{&historical, &current, &other} {
		if err := db.Create(session).Error; err != nil {
			t.Fatalf("create session fixture: %v", err)
		}
	}
	historicalAssessment := models.SessionSkillAssessment{
		TrainingSessionID: historical.ID, SubMaterialID: subMaterials[0].ID, SkillStatus: models.SkillMastered,
	}
	historicalEvaluation := models.SessionEvaluation{
		TrainingSessionID: historical.ID, Predicate: models.PredicateBaik,
		Notes: "Previous enrollment completed.", Recommendation: "Keep practicing steering.",
	}
	if err := db.Create(&historicalAssessment).Error; err != nil {
		t.Fatalf("create historical assessment: %v", err)
	}
	if err := db.Create(&historicalEvaluation).Error; err != nil {
		t.Fatalf("create historical evaluation: %v", err)
	}

	endpoint := "/api/v1/trainer/training-sessions"
	sessionPath := fmt.Sprintf("%s/%d", endpoint, current.ID)
	assessmentsPath, evaluationPath := sessionPath+"/assessments", sessionPath+"/evaluation"
	assessment := func(subMaterialID int64, status string) map[string]any {
		return map[string]any{"sub_material_id": subMaterialID, "skill_status": status}
	}
	assessmentBatch := func(items ...map[string]any) map[string]any {
		return map[string]any{"assessments": items}
	}
	evaluation := func(predicate, notes, recommendation string) map[string]any {
		return map[string]any{"predicate": predicate, "notes": notes, "recommendation": recommendation}
	}

	t.Run("role authorization and trainer session ownership", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, studentToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, adminToken), http.StatusForbidden)
		response := performJSON(t, router, http.MethodGet, endpoint, nil, trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data []dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if len(result.Data) != 1 || result.Data[0].ID != current.ID {
			t.Fatalf("trainer list leaked sessions: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, sessionPath, nil, trainerOneToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, sessionPath, nil, trainerTwoToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint+"/invalid", nil, trainerOneToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/start", nil, trainerTwoToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, assessmentsPath, nil, trainerTwoToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, evaluationPath, nil, trainerTwoToken), http.StatusNotFound)
	})

	t.Run("scheduled sessions cannot be assessed evaluated or completed", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED")), trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "Good progress", "Practice parking"), trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, evaluationPath, nil, trainerOneToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodGet, assessmentsPath, nil, trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data []dto.SessionAssessmentData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if len(result.Data) != 0 {
			t.Fatalf("scheduled session unexpectedly had assessments: %+v", result.Data)
		}
	})

	t.Run("assigned trainer starts exactly once", func(t *testing.T) {
		response := performJSON(t, router, http.MethodPost, sessionPath+"/start", nil, trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if result.Data.Status != models.SessionInProgress || result.Data.ActualStartedAt == nil ||
			result.Data.ActualCompletedAt != nil {
			t.Fatalf("session did not transition to IN_PROGRESS: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/start", nil, trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerOneToken), http.StatusConflict)
	})

	var firstAssessmentID int64
	t.Run("atomic batch assessment validates active sub-materials and updates in place", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(), trainerOneToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "INVALID")), trainerOneToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED"), assessment(subMaterials[0].ID, "MASTERED")),
			trainerOneToken), http.StatusUnprocessableEntity)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED"), assessment(999999, "MASTERED")),
			trainerOneToken), http.StatusNotFound)
		var count int64
		if err := db.Model(&models.SessionSkillAssessment{}).
			Where("training_session_id = ?", current.ID).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("invalid batch did not roll back: count=%d err=%v", count, err)
		}
		for _, subMaterial := range []models.SubMaterial{subMaterials[2], subMaterials[3]} {
			requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
				assessmentBatch(assessment(subMaterial.ID, "PRACTICED")), trainerOneToken), http.StatusNotFound)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED")), trainerTwoToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED"), assessment(subMaterials[1].ID, "NOT_STARTED")),
			trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data []dto.SessionAssessmentData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if len(result.Data) != 2 || result.Data[0].MaterialName != materials[0].Name ||
			result.Data[0].SubMaterialName != subMaterials[0].Name || result.Data[1].SkillStatus != models.SkillNotStarted {
			t.Fatalf("assessment batch lost curriculum details: %+v", result.Data)
		}
		firstAssessmentID = result.Data[0].ID
		response = performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "MASTERED")), trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if len(result.Data) != 2 || result.Data[0].ID != firstAssessmentID ||
			result.Data[0].SkillStatus != models.SkillMastered || result.Data[1].SkillStatus != models.SkillNotStarted {
			t.Fatalf("same-session assessment was duplicated or untouched skills changed: %+v", result.Data)
		}
		response = performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "PRACTICED")), trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data[0].ID != firstAssessmentID || result.Data[0].SkillStatus != models.SkillPracticed {
			t.Fatalf("skill decline was incorrectly rejected: %+v", result.Data)
		}
		if err := db.Model(&models.SessionSkillAssessment{}).
			Where("training_session_id = ?", current.ID).Count(&count).Error; err != nil || count != 2 {
			t.Fatalf("assessment upsert duplicated rows: count=%d err=%v", count, err)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, assessmentsPath, nil, trainerOneToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerOneToken), http.StatusConflict)
	})

	var evaluationID int64
	t.Run("complete evaluation is mandatory and updateable during training", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("INVALID", "Notes", "Practice"), trainerOneToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "", "Practice"), trainerOneToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "  ", "Practice"), trainerOneToken), http.StatusUnprocessableEntity)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "Notes", "   "), trainerOneToken), http.StatusUnprocessableEntity)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "Notes", "Practice"), trainerTwoToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("BAIK", "  Good vehicle control  ", "  Practice parking  "), trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.SessionEvaluationData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if result.Data.Predicate != models.PredicateBaik || result.Data.Notes != "Good vehicle control" ||
			result.Data.Recommendation != "Practice parking" {
			t.Fatalf("evaluation was not normalized: %+v", result.Data)
		}
		evaluationID = result.Data.ID
		response = performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("CUKUP", "Needs more steering practice.", "Focus on steering."), trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		decodeResponse(t, response, &result)
		if result.Data.ID != evaluationID || result.Data.Predicate != models.PredicateCukup {
			t.Fatalf("evaluation was not updated in place: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, evaluationPath, nil, trainerOneToken), http.StatusOK)
	})

	t.Run("completion locks records while unfinished enrollment remains active", func(t *testing.T) {
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerTwoToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerOneToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, response, &result)
		if result.Data.Status != models.SessionCompleted || result.Data.ActualStartedAt == nil ||
			result.Data.ActualCompletedAt == nil {
			t.Fatalf("session did not transition to COMPLETED: %+v", result.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/complete", nil, trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, sessionPath+"/start", nil, trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, assessmentsPath,
			assessmentBatch(assessment(subMaterials[0].ID, "MASTERED")), trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodPut, evaluationPath,
			evaluation("SANGAT_BAIK", "Changed", "Changed"), trainerOneToken), http.StatusConflict)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, assessmentsPath, nil, trainerOneToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, evaluationPath, nil, trainerOneToken), http.StatusOK)
		var enrollment models.StudentEnrollment
		if err := db.First(&enrollment, enrollments[1].ID).Error; err != nil ||
			enrollment.Status != models.EnrollmentActive || enrollment.CompletedAt != nil {
			t.Fatalf("unfinished enrollment unexpectedly completed: enrollment=%+v err=%v", enrollment, err)
		}
	})

	t.Run("next trainer sees completed history across all student enrollments", func(t *testing.T) {
		booking := performJSON(t, router, http.MethodPost, "/api/v1/student/training-sessions", map[string]any{
			"trainer_availability_id": nextAvailability.ID, "start_time": "08:00",
		}, studentToken)
		requireAdminStatus(t, booking, http.StatusCreated)
		var next struct {
			Data dto.TrainingSessionData `json:"data"`
		}
		decodeResponse(t, booking, &next)
		if next.Data.SessionNumber != 2 || next.Data.TrainerID != users[2].ID {
			t.Fatalf("completed session did not advance active enrollment session number: %+v", next.Data)
		}
		progressPath := fmt.Sprintf("%s/%d/student-progress", endpoint, next.Data.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, progressPath, nil, trainerOneToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, progressPath, nil, studentToken), http.StatusForbidden)
		response := performJSON(t, router, http.MethodGet, progressPath, nil, trainerTwoToken)
		requireAdminStatus(t, response, http.StatusOK)
		var progress struct {
			Data dto.TrainerStudentProgressData `json:"data"`
		}
		decodeResponse(t, response, &progress)
		if progress.Data.StudentID != users[3].ID || progress.Data.TrainingSessionID != next.Data.ID ||
			len(progress.Data.PreviousSessions) != 2 {
			t.Fatalf("cross-enrollment progress history is incomplete: %+v", progress.Data)
		}
		recent, previous := progress.Data.PreviousSessions[0], progress.Data.PreviousSessions[1]
		if recent.Session.ID != current.ID || recent.Session.EnrollmentID != enrollments[1].ID ||
			len(recent.Assessments) != 2 || recent.Assessments[0].SkillStatus != models.SkillPracticed ||
			recent.Evaluation == nil || recent.Evaluation.ID != evaluationID {
			t.Fatalf("current-enrollment completed history is incorrect: %+v", recent)
		}
		if previous.Session.ID != historical.ID || previous.Session.EnrollmentID != enrollments[0].ID ||
			len(previous.Assessments) != 1 || previous.Assessments[0].SkillStatus != models.SkillMastered ||
			previous.Evaluation == nil || previous.Evaluation.ID != historicalEvaluation.ID {
			t.Fatalf("previous enrollment history was not preserved: %+v", previous)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/skills", nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/certificates", nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/trainer/reviews", nil, trainerTwoToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/internal/health", nil, adminToken), http.StatusUnauthorized)
	})
}
