package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const railwayFrontendOrigin = "https://drive-academy.up.railway.app"

func newCORSRouter(t *testing.T) *gin.Engine {
	t.Helper()

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatalf("create GORM database: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router, err := New(db, config.Config{
		JWTSecret:    "cors-test-secret-with-at-least-32-bytes",
		JWTExpiresIn: "1h",
		CORSAllowedOrigins: []string{
			railwayFrontendOrigin,
			"http://localhost:5173",
		},
	})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
	return router
}

func TestCORSPreflightAllowsProductionAndLocalOriginsWithoutAuthentication(t *testing.T) {
	router := newCORSRouter(t)

	tests := []struct {
		name   string
		origin string
		path   string
		method string
	}{
		{
			name:   "production login",
			origin: railwayFrontendOrigin,
			path:   "/api/users/login",
			method: http.MethodPost,
		},
		{
			name:   "local registration",
			origin: "http://localhost:5173",
			path:   "/api/users/register",
			method: http.MethodPost,
		},
		{
			name:   "protected student route without a bearer token",
			origin: railwayFrontendOrigin,
			path:   "/api/v1/student/enrollments",
			method: http.MethodPost,
		},
		{
			name:   "protected admin route without a bearer token",
			origin: railwayFrontendOrigin,
			path:   "/api/v1/admin/packages/1",
			method: http.MethodPut,
		},
		{
			name:   "protected internal route without basic authentication",
			origin: railwayFrontendOrigin,
			path:   "/api/v1/internal/stats",
			method: http.MethodGet,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, test.path, nil)
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Access-Control-Request-Method", test.method)
			request.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("preflight status = %d, want %d; body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != test.origin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, test.origin)
			}
			assertHeaderContainsValues(t, recorder.Header().Get("Access-Control-Allow-Methods"), []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			})
			assertHeaderContainsValues(t, recorder.Header().Get("Access-Control-Allow-Headers"), []string{
				"Origin",
				"Content-Type",
				"Accept",
				"Authorization",
			})
			if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
				t.Fatalf("credential/cookie access was unexpectedly enabled: %q", got)
			}
		})
	}
}

func TestCORSRejectsOriginsOutsideTheExplicitWhitelist(t *testing.T) {
	router := newCORSRouter(t)
	request := httptest.NewRequest(http.MethodOptions, "/api/users/login", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("preflight status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("untrusted origin was allowed: %q", got)
	}
}

func TestCORSAllowsPublicAuthRequestsWithoutJWTAndPreservesProtectedAuth(t *testing.T) {
	router := newCORSRouter(t)

	for _, path := range []string{"/api/users/login", "/api/users/register"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
			request.Header.Set("Origin", railwayFrontendOrigin)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("public request status = %d, want validation status %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != railwayFrontendOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, railwayFrontendOrigin)
			}
		})
	}

	t.Run("protected requests still require bearer tokens", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/student/packages", nil)
		request.Header.Set("Origin", railwayFrontendOrigin)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("protected request status = %d, want %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != railwayFrontendOrigin {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, railwayFrontendOrigin)
		}
	})
}

func assertHeaderContainsValues(t *testing.T, header string, want []string) {
	t.Helper()

	values := make(map[string]bool)
	for _, value := range strings.Split(header, ",") {
		values[strings.ToLower(strings.TrimSpace(value))] = true
	}
	for _, value := range want {
		if !values[strings.ToLower(value)] {
			t.Errorf("header %q does not contain %q", header, value)
		}
	}
}
