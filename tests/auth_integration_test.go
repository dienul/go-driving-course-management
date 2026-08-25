package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAuthenticationFlowPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	const email = "phase3-student@example.com"
	db.Where("email = ?", email).Delete(&models.User{})

	cfg := config.Config{
		JWTSecret:    "integration-test-secret-with-at-least-32-bytes",
		JWTExpiresIn: "1h",
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, cfg)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	registerBody := map[string]any{
		"name":     "Phase 3 Student",
		"email":    email,
		"password": "strong-password",
		"phone":    "081234567890",
		"role":     "ADMIN",
	}
	register := performJSON(t, router, http.MethodPost, "/api/users/register", registerBody, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d; body=%s", register.Code, register.Body.String())
	}
	var registered struct {
		Data dto.RegisterData `json:"data"`
	}
	decodeResponse(t, register, &registered)
	if registered.Data.User.Role != models.RoleStudent ||
		registered.Data.User.Status != models.StatusActive {
		t.Fatalf("public registration chose role/status: %+v", registered.Data.User)
	}

	var stored models.User
	if err := db.Where("email = ?", email).First(&stored).Error; err != nil {
		t.Fatalf("find registered user: %v", err)
	}
	if stored.Role != models.RoleStudent || stored.PasswordHash == "strong-password" {
		t.Fatalf("unsafe stored user: %+v", stored)
	}
	var profileCount int64
	if err := db.Model(&models.StudentProfile{}).
		Where("user_id = ?", stored.ID).
		Count(&profileCount).Error; err != nil {
		t.Fatalf("count student profile: %v", err)
	}
	if profileCount != 1 {
		t.Fatalf("student profile count = %d, want 1", profileCount)
	}

	duplicate := performJSON(t, router, http.MethodPost, "/api/users/register", registerBody, "")
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d; body=%s", duplicate.Code, duplicate.Body.String())
	}

	wrongLogin := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
		"email": email, "password": "wrong-password",
	}, "")
	if wrongLogin.Code != http.StatusUnauthorized {
		t.Fatalf("wrong login status = %d; body=%s", wrongLogin.Code, wrongLogin.Body.String())
	}

	login := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
		"email": email, "password": "strong-password",
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d; body=%s", login.Code, login.Body.String())
	}
	var loggedIn struct {
		Data dto.LoginData `json:"data"`
	}
	decodeResponse(t, login, &loggedIn)
	if loggedIn.Data.Token == "" {
		t.Fatal("login token is empty")
	}

	me := performJSON(
		t,
		router,
		http.MethodGet,
		"/api/v1/auth/me",
		nil,
		"Bearer "+loggedIn.Data.Token,
	)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d; body=%s", me.Code, me.Body.String())
	}

	if err := db.Model(&models.User{}).
		Where("id = ?", stored.ID).
		Update("status", models.StatusInactive).Error; err != nil {
		t.Fatalf("deactivate user: %v", err)
	}

	inactiveMe := performJSON(
		t,
		router,
		http.MethodGet,
		"/api/v1/auth/me",
		nil,
		"Bearer "+loggedIn.Data.Token,
	)
	if inactiveMe.Code != http.StatusForbidden {
		t.Fatalf("inactive me status = %d; body=%s", inactiveMe.Code, inactiveMe.Body.String())
	}

	inactiveLogin := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
		"email": email, "password": "strong-password",
	}, "")
	if inactiveLogin.Code != http.StatusForbidden {
		t.Fatalf("inactive login status = %d; body=%s", inactiveLogin.Code, inactiveLogin.Body.String())
	}
}

func performJSON(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body any,
	authorization string,
) *httptest.ResponseRecorder {
	t.Helper()
	var content []byte
	var err error
	if body != nil {
		content, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(content))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
}
