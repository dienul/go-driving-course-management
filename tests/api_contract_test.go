package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dienulhaq/go-driving-course-management/config"
	_ "github.com/dienulhaq/go-driving-course-management/docs"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type swaggerContract struct {
	Paths map[string]map[string]struct {
		Security []map[string][]string `json:"security"`
	} `json:"paths"`
	SecurityDefinitions map[string]struct {
		Type string `json:"type"`
		Name string `json:"name"`
		In   string `json:"in"`
	} `json:"securityDefinitions"`
}

type postmanContract struct {
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Item     []postmanContractItem `json:"item"`
	Variable []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"variable"`
}

type postmanContractItem struct {
	Name    string                `json:"name"`
	Item    []postmanContractItem `json:"item"`
	Request struct {
		Method string `json:"method"`
		URL    struct {
			Raw string `json:"raw"`
		} `json:"url"`
		Auth struct {
			Type string `json:"type"`
		} `json:"auth"`
		Body struct {
			Raw string `json:"raw"`
		} `json:"body"`
	} `json:"request"`
	Event []struct {
		Listen string `json:"listen"`
	} `json:"event"`
}

func TestSwaggerAndPostmanCoverEveryAPIEndpoint(t *testing.T) {
	var swagger swaggerContract
	readJSONContract(t, filepath.Join("..", "docs", "swagger.json"), &swagger)
	var postman postmanContract
	readJSONContract(t, filepath.Join("..", "postman", "Driving-Course-Management.postman_collection.json"), &postman)

	if !strings.Contains(postman.Info.Schema, "/collection/v2.1.0/collection.json") {
		t.Fatalf("unsupported Postman collection schema: %q", postman.Info.Schema)
	}
	if swagger.SecurityDefinitions["BasicAuth"].Type != "basic" {
		t.Fatal("Swagger BasicAuth definition is missing or invalid")
	}
	bearer := swagger.SecurityDefinitions["BearerAuth"]
	if bearer.Type != "apiKey" || bearer.In != "header" || bearer.Name != "Authorization" {
		t.Fatalf("Swagger BearerAuth definition is invalid: %+v", bearer)
	}

	variableNames := make(map[string]bool)
	for _, variable := range postman.Variable {
		variableNames[variable.Key] = true
	}
	for _, key := range []string{
		"base_url", "admin_email", "admin_password", "student_email", "student_password",
		"trainer_email", "trainer_password", "admin_token", "student_token", "trainer_token",
		"basic_username", "basic_password", "package_id", "material_id", "sub_material_id",
		"enrollment_id", "payment_id", "invoice_id", "availability_id", "session_id",
		"certificate_id", "available_date",
	} {
		if !variableNames[key] {
			t.Errorf("Postman collection variable %q is missing", key)
		}
	}

	postmanEndpoints := make(map[string]map[string]bool)
	var requestCount int
	var sawEndToEnd bool
	var walk func([]postmanContractItem)
	walk = func(items []postmanContractItem) {
		for _, item := range items {
			if strings.Contains(item.Name, "End-to-End Business Flow") {
				sawEndToEnd = true
			}
			if item.Request.Method == "" {
				walk(item.Item)
				continue
			}
			requestCount++
			key := strings.ToUpper(item.Request.Method) + " " + normalizeContractPath(item.Request.URL.Raw)
			if postmanEndpoints[key] == nil {
				postmanEndpoints[key] = make(map[string]bool)
			}
			postmanEndpoints[key][item.Request.Auth.Type] = true
			if len(item.Event) == 0 {
				t.Errorf("Postman request %q does not contain test assertions", item.Name)
			}
			assertNumericPostmanIdentifiers(t, item)
		}
	}
	walk(postman.Item)
	if requestCount < 100 || !sawEndToEnd {
		t.Fatalf("incomplete Postman regression collection: requests=%d, end-to-end=%t", requestCount, sawEndToEnd)
	}

	counts := map[string]int{}
	for path, operations := range swagger.Paths {
		for method, operation := range operations {
			key := strings.ToUpper(method) + " " + normalizeContractPath(path)
			authTypes := postmanEndpoints[key]
			if len(authTypes) == 0 {
				t.Errorf("Swagger operation %s is missing from Postman", key)
				continue
			}

			switch {
			case strings.HasPrefix(path, "/api/v1/internal/"):
				counts["internal"]++
				if !swaggerRequires(operation.Security, "BasicAuth") || swaggerRequires(operation.Security, "BearerAuth") {
					t.Errorf("internal operation %s must require BasicAuth only", key)
				}
				if !authTypes["basic"] {
					t.Errorf("Postman operation %s has no Basic Auth request", key)
				}
			case strings.HasPrefix(path, "/api/v1/student/"):
				counts["student"]++
				assertBearerContract(t, key, operation.Security, authTypes)
			case strings.HasPrefix(path, "/api/v1/trainer/"):
				counts["trainer"]++
				assertBearerContract(t, key, operation.Security, authTypes)
			case strings.HasPrefix(path, "/api/v1/admin/"):
				counts["admin"]++
				assertBearerContract(t, key, operation.Security, authTypes)
			case path == "/api/v1/auth/me":
				counts["authenticated"]++
				assertBearerContract(t, key, operation.Security, authTypes)
			default:
				counts["public"]++
				if len(operation.Security) != 0 {
					t.Errorf("public operation %s unexpectedly declares security", key)
				}
				if !authTypes["noauth"] {
					t.Errorf("Postman operation %s has no unauthenticated request", key)
				}
			}
		}
	}
	expected := map[string]int{"student": 22, "trainer": 16, "admin": 41, "internal": 2, "authenticated": 1, "public": 3}
	for group, count := range expected {
		if counts[group] != count {
			t.Errorf("%s documented operations = %d, want %d", group, counts[group], count)
		}
	}
}

func TestSwaggerUIAndGeneratedDocumentAreServed(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("create GORM database: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{
		JWTSecret: "phase11-contract-secret-with-at-least-32-bytes", JWTExpiresIn: "1h",
		BasicAuthUser: "phase11-internal", BasicAuthPass: "phase11-password",
	})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
	for _, endpoint := range []string{"/swagger/index.html", "/swagger/doc.json"} {
		t.Run(endpoint, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, endpoint, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if endpoint == "/swagger/doc.json" {
				var swagger swaggerContract
				if err := json.Unmarshal(recorder.Body.Bytes(), &swagger); err != nil || len(swagger.Paths) == 0 {
					t.Fatalf("invalid generated Swagger document: %v", err)
				}
			} else if !strings.Contains(recorder.Body.String(), "Swagger UI") {
				t.Fatal("Swagger UI HTML was not served")
			}
		})
	}
}

func readJSONContract(t *testing.T, path string, value any) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(content, value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func normalizeContractPath(path string) string {
	path = strings.TrimPrefix(path, "{{base_url}}")
	path = regexp.MustCompile(`\{\{[^}]+_id\}\}`).ReplaceAllString(path, "{id}")
	return regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "{id}")
}

func swaggerRequires(security []map[string][]string, scheme string) bool {
	for _, requirement := range security {
		if _, exists := requirement[scheme]; exists {
			return true
		}
	}
	return false
}

func assertBearerContract(t *testing.T, key string, security []map[string][]string, authTypes map[string]bool) {
	t.Helper()
	if !swaggerRequires(security, "BearerAuth") || swaggerRequires(security, "BasicAuth") {
		t.Errorf("protected operation %s must require BearerAuth only", key)
	}
	if !authTypes["bearer"] {
		t.Errorf("Postman operation %s has no Bearer token request", key)
	}
}

func assertNumericPostmanIdentifiers(t *testing.T, item postmanContractItem) {
	t.Helper()
	if item.Request.Body.Raw == "" {
		return
	}
	rendered := regexp.MustCompile(`\{\{[^}]+\}\}`).ReplaceAllString(item.Request.Body.Raw, "1")
	var body any
	if err := json.Unmarshal([]byte(rendered), &body); err != nil {
		t.Errorf("Postman request %q has invalid JSON body after variable substitution: %v", item.Name, err)
		return
	}
	var validate func(any)
	validate = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				if strings.HasSuffix(key, "_id") {
					if _, ok := child.(float64); !ok {
						t.Errorf("Postman request %q sends numeric field %q as %T", item.Name, key, child)
					}
				}
				validate(child)
			}
		case []any:
			for _, child := range typed {
				validate(child)
			}
		}
	}
	validate(body)
}
