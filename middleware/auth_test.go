package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/gin-gonic/gin"
)

const middlewareTestSecret = "middleware-test-secret-with-at-least-32-bytes"

type fakeUserReader struct {
	user *models.User
	err  error
}

func (f *fakeUserReader) FindByID(_ context.Context, _ int64) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.user == nil {
		return nil, repositories.ErrUserNotFound
	}
	return f.user, nil
}

func TestJWTAuthAndRoleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := services.NewJWTManager(middlewareTestSecret, "1h")
	if err != nil {
		t.Fatalf("new JWT manager: %v", err)
	}
	user := &models.User{
		ID: 7, Role: models.RoleStudent, Status: models.StatusActive,
	}
	token, _, err := tokens.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	reader := &fakeUserReader{user: user}
	router := gin.New()
	router.GET(
		"/student",
		JWTAuth(tokens, reader),
		RoleAuth(models.RoleStudent),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)
	router.GET(
		"/admin",
		JWTAuth(tokens, reader),
		RoleAuth(models.RoleAdmin),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	assertRequestStatus(t, router, "/student", "", http.StatusUnauthorized)
	assertRequestStatus(t, router, "/student", "Bearer invalid", http.StatusUnauthorized)
	assertRequestStatus(t, router, "/student", "Bearer "+token, http.StatusNoContent)
	assertRequestStatus(t, router, "/admin", "Bearer "+token, http.StatusForbidden)

	user.Status = models.StatusInactive
	assertRequestStatus(t, router, "/student", "Bearer "+token, http.StatusForbidden)
}

func TestJWTAuthRejectsDeletedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := services.NewJWTManager(middlewareTestSecret, "1h")
	if err != nil {
		t.Fatalf("new JWT manager: %v", err)
	}
	user := &models.User{
		ID: 9, Role: models.RoleTrainer, Status: models.StatusActive,
	}
	token, _, err := tokens.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := gin.New()
	router.GET(
		"/protected",
		JWTAuth(tokens, &fakeUserReader{err: repositories.ErrUserNotFound}),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)
	assertRequestStatus(t, router, "/protected", "Bearer "+token, http.StatusUnauthorized)
}

func TestBasicAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(
		"/internal",
		BasicAuth("internal-user", "internal-password"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	request := httptest.NewRequest(http.MethodGet, "/internal", nil)
	request.SetBasicAuth("internal-user", "internal-password")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("valid basic auth status = %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/internal", nil)
	request.SetBasicAuth("internal-user", "wrong-password")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("invalid basic auth status = %d", recorder.Code)
	}
	if recorder.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("WWW-Authenticate header is missing")
	}
}

func assertRequestStatus(
	t *testing.T,
	router http.Handler,
	path string,
	authorization string,
	expected int,
) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != expected {
		t.Fatalf("%s status = %d, want %d; body=%s", path, recorder.Code, expected, recorder.Body.String())
	}
}
