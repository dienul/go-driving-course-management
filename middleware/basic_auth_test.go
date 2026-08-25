package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBasicAuthRejectsInvalidAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/internal", BasicAuth("phase10-user", "phase10-secret"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name          string
		authorization string
		username      string
		password      string
	}{
		{name: "missing authorization"},
		{name: "bearer token", authorization: "Bearer administrator-jwt"},
		{name: "malformed basic credentials", authorization: "Basic not-base64"},
		{name: "wrong username", username: "different-user", password: "phase10-secret"},
		{name: "wrong password", username: "phase10-user", password: "different-secret"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/internal", nil)
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			} else if test.username != "" {
				request.SetBasicAuth(test.username, test.password)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
			if challenge := recorder.Header().Get("WWW-Authenticate"); challenge != `Basic realm="internal"` {
				t.Fatalf("WWW-Authenticate = %q", challenge)
			}
			if strings.Contains(recorder.Body.String(), "phase10-secret") {
				t.Fatal("response exposes configured credentials")
			}
		})
	}
}

func TestBasicAuthFailsClosedWithoutConfiguredCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "missing username", password: "phase10-secret"},
		{name: "missing password", username: "phase10-user"},
		{name: "missing username and password"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/internal", BasicAuth(test.username, test.password), func(c *gin.Context) {
				t.Fatal("handler ran without configured credentials")
			})

			request := httptest.NewRequest(http.MethodGet, "/internal", nil)
			request.SetBasicAuth("phase10-user", "phase10-secret")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}
			if !strings.Contains(recorder.Body.String(), "internal server error") {
				t.Fatalf("unexpected response: %s", recorder.Body.String())
			}
			if strings.Contains(recorder.Body.String(), "phase10-secret") {
				t.Fatal("response exposes configured credentials")
			}
		})
	}
}
