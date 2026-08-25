package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/dienulhaq/go-driving-course-management/seeds"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostmanEndToEndBusinessFlowPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin isolated end-to-end transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	seedConfig := seeds.Config{
		AdminName: "Phase 11 End-to-End Admin", AdminEmail: "phase11-flow-admin@example.com",
		AdminPassword: "phase11-flow-admin-password",
	}
	if err := seeds.Run(tx, seedConfig); err != nil {
		t.Fatalf("prepare seeded end-to-end database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(tx, config.Config{
		JWTSecret: "phase11-end-to-end-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		BasicAuthUser: "phase11-observer", BasicAuthPass: "phase11-observer-password",
	})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	availableDate := time.Now().UTC().AddDate(0, 0, 1)
	for availableDate.Weekday() == time.Saturday || availableDate.Weekday() == time.Sunday {
		availableDate = availableDate.AddDate(0, 0, 1)
	}
	variables := map[string]string{
		"base_url":         "http://phase11.test",
		"admin_email":      seedConfig.AdminEmail,
		"admin_password":   seedConfig.AdminPassword,
		"student_email":    "phase11-flow-student@example.com",
		"student_password": "phase11-flow-student-password",
		"trainer_email":    "phase11-flow-trainer@example.com",
		"trainer_password": "phase11-flow-trainer-password",
		"basic_username":   "phase11-observer",
		"basic_password":   "phase11-observer-password",
		"available_date":   availableDate.Format("2006-01-02"),
	}
	var collection postmanContract
	readJSONContract(t, filepath.Join("..", "postman", "Driving-Course-Management.postman_collection.json"), &collection)
	var flow []postmanContractItem
	for _, folder := range collection.Item {
		if strings.Contains(folder.Name, "End-to-End Business Flow") {
			flow = folder.Item
			break
		}
	}
	if len(flow) < 25 {
		t.Fatalf("Postman end-to-end flow is incomplete: %d steps", len(flow))
	}

	for _, step := range flow {
		t.Run(step.Name, func(t *testing.T) {
			path := renderPostmanVariables(step.Request.URL.Raw, variables)
			path = strings.TrimPrefix(path, variables["base_url"])
			body := renderPostmanVariables(step.Request.Body.Raw, variables)
			request := httptest.NewRequest(step.Request.Method, path, bytes.NewBufferString(body))
			if body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			switch {
			case step.Request.Auth.Type == "basic":
				request.SetBasicAuth(variables["basic_username"], variables["basic_password"])
			case strings.HasPrefix(path, "/api/v1/admin/"):
				request.Header.Set("Authorization", "Bearer "+variables["admin_token"])
			case strings.HasPrefix(path, "/api/v1/trainer/"):
				request.Header.Set("Authorization", "Bearer "+variables["trainer_token"])
			case strings.HasPrefix(path, "/api/v1/student/"):
				request.Header.Set("Authorization", "Bearer "+variables["student_token"])
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK && recorder.Code != http.StatusCreated {
				t.Fatalf("%s %s status = %d; body=%s", step.Request.Method, path, recorder.Code, recorder.Body.String())
			}
			var response struct {
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				var listResponse struct {
					Success bool             `json:"success"`
					Data    []map[string]any `json:"data"`
				}
				if listErr := json.Unmarshal(recorder.Body.Bytes(), &listResponse); listErr != nil {
					t.Fatalf("decode Postman flow response: %v", listErr)
				}
				if !listResponse.Success {
					t.Fatalf("unsuccessful list response: %s", recorder.Body.String())
				}
				capturePostmanList(t, step, path, listResponse.Data, variables)
				return
			}
			if !response.Success {
				t.Fatalf("unsuccessful response: %s", recorder.Body.String())
			}
			capturePostmanObject(t, step, path, response.Data, variables)
		})
	}
	for _, key := range []string{
		"admin_token", "student_token", "trainer_token", "trainer_id", "student_id",
		"package_id", "material_id", "sub_material_id", "enrollment_id", "payment_id",
		"invoice_id", "availability_id", "session_id", "certificate_id",
	} {
		if variables[key] == "" {
			t.Errorf("Postman end-to-end flow did not capture variable %q", key)
		}
	}
}

func renderPostmanVariables(value string, variables map[string]string) string {
	replacements := make([]string, 0, len(variables)*2)
	for key, current := range variables {
		replacements = append(replacements, "{{"+key+"}}", current)
	}
	return strings.NewReplacer(replacements...).Replace(value)
}

func capturePostmanObject(t *testing.T, step postmanContractItem, path string, data map[string]any, variables map[string]string) {
	t.Helper()
	capture := func(key string, value any) {
		t.Helper()
		if value == nil {
			t.Fatalf("response does not contain %q: %+v", key, data)
		}
		variables[key] = fmt.Sprint(value)
	}
	switch {
	case path == "/api/users/login":
		var role string
		switch {
		case strings.Contains(step.Name, "admin"):
			role = "admin"
		case strings.Contains(step.Name, "student"):
			role = "student"
		case strings.Contains(step.Name, "trainer"):
			role = "trainer"
		default:
			t.Fatalf("unknown login role in step %q", step.Name)
		}
		capture(role+"_token", data["token"])
		user, ok := data["user"].(map[string]any)
		if !ok {
			t.Fatalf("login response has no user: %+v", data)
		}
		capture(role+"_id", user["id"])
	case path == "/api/users/register":
		user, ok := data["user"].(map[string]any)
		if !ok {
			t.Fatalf("registration response has no user: %+v", data)
		}
		capture("student_id", user["id"])
	case path == "/api/v1/admin/trainers" && step.Request.Method == http.MethodPost:
		user, ok := data["user"].(map[string]any)
		if !ok {
			t.Fatalf("trainer response has no user: %+v", data)
		}
		capture("trainer_id", user["id"])
	case path == "/api/v1/trainer/availabilities" && step.Request.Method == http.MethodPost:
		capture("availability_id", data["id"])
	case path == "/api/v1/student/enrollments" && step.Request.Method == http.MethodPost:
		enrollment, ok := data["enrollment"].(map[string]any)
		if !ok {
			t.Fatalf("checkout response has no enrollment: %+v", data)
		}
		payment, ok := data["payment"].(map[string]any)
		if !ok {
			t.Fatalf("checkout response has no payment: %+v", data)
		}
		capture("enrollment_id", enrollment["id"])
		capture("payment_id", payment["id"])
	case strings.Contains(path, "/payments/") && strings.HasSuffix(path, "/pay"):
		invoice, ok := data["invoice"].(map[string]any)
		if !ok {
			t.Fatalf("payment response has no invoice: %+v", data)
		}
		capture("invoice_id", invoice["id"])
	case path == "/api/v1/student/training-sessions" && step.Request.Method == http.MethodPost:
		capture("session_id", data["id"])
	}
}

func capturePostmanList(t *testing.T, step postmanContractItem, path string, data []map[string]any, variables map[string]string) {
	t.Helper()
	if len(data) == 0 {
		t.Fatalf("Postman end-to-end step %q returned an empty list", step.Name)
	}
	switch {
	case path == "/api/v1/student/packages":
		for _, item := range data {
			if item["total_hours"] == float64(6) {
				variables["package_id"] = fmt.Sprint(item["id"])
				return
			}
		}
		t.Fatal("seeded six-hour package is missing")
	case path == "/api/v1/admin/materials":
		variables["material_id"] = fmt.Sprint(data[0]["id"])
	case strings.Contains(path, "/sub-materials"):
		variables["sub_material_id"] = fmt.Sprint(data[0]["id"])
	case path == "/api/v1/student/certificates":
		variables["certificate_id"] = fmt.Sprint(data[0]["id"])
	}
}
