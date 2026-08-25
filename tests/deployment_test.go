package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestFinalDocumentationCoversProjectRequirements(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatalf("read project README: %v", err)
	}
	readme := strings.ReplaceAll(string(content), "\r\n", "\n")
	for _, heading := range []string{
		"## Project overview",
		"## Business flow",
		"## Roles",
		"## Technology stack",
		"## Architecture",
		"## Database",
		"### Entity relationship diagram",
		"## Environment variables",
		"## Local setup",
		"## Migration and seed",
		"## Authentication",
		"## REST API overview",
		"## Swagger and OpenAPI",
		"## Postman collection",
		"## Testing",
		"## Railway deployment",
	} {
		if !strings.Contains(readme, heading+"\n") {
			t.Errorf("required documentation section %q is missing", heading)
		}
	}
	diagram := regexp.MustCompile("(?s)```mermaid\\s+erDiagram\\s+(.*?)```").
		FindStringSubmatch(readme)
	if len(diagram) != 2 {
		t.Fatal("Mermaid entity relationship diagram is missing")
	}
	for _, table := range []string{
		"USERS", "STUDENT_PROFILES", "TRAINER_PROFILES", "COURSE_PACKAGES", "MATERIALS",
		"SUB_MATERIALS", "STUDENT_ENROLLMENTS", "PAYMENTS", "INVOICES",
		"TRAINER_AVAILABILITIES", "TRAINING_SESSIONS", "SESSION_SKILL_ASSESSMENTS",
		"SESSION_EVALUATIONS", "TRAINER_REVIEWS", "CERTIFICATES",
	} {
		entity := regexp.MustCompile(`(?m)^\s*` + table + `\s+\{`)
		if !entity.MatchString(diagram[1]) {
			t.Errorf("database entity %q is missing from Mermaid ERD", table)
		}
	}
	for _, value := range []string{
		"postman/Driving-Course-Management.postman_collection.json",
		"124 requests",
		"85 documented operations",
		"https://go-driving-course-management-production.up.railway.app/health",
		"https://go-driving-course-management-production.up.railway.app/swagger/index.html",
		"TEST_DATABASE_URL",
		"ADMIN_NAME",
		"ADMIN_EMAIL",
		"ADMIN_PASSWORD",
	} {
		if !strings.Contains(readme, value) {
			t.Errorf("required project documentation %q is missing", value)
		}
	}
}

func TestProductionDockerStartupAndSecretExclusions(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "Dockerfile"))
	if err != nil {
		t.Fatalf("read production Dockerfile: %v", err)
	}
	dockerfile := string(content)
	for _, instruction := range []string{
		"FROM golang:1.24-alpine AS builder",
		"FROM alpine:3.20",
		"go build -o /app/server ./cmd",
		"go build -o /app/migrate ./cmd/migrate",
		"go build -o /app/seed ./cmd/seed",
		"COPY --from=builder /app/migrations /app/migrations",
		"export APP_PORT=${PORT:-8080}",
		"/app/migrate up && /app/seed && exec /app/server",
	} {
		if !strings.Contains(dockerfile, instruction) {
			t.Errorf("required Docker production instruction %q is missing", instruction)
		}
	}
	runtime := strings.SplitN(dockerfile, "FROM alpine:3.20", 2)
	if len(runtime) != 2 || strings.Contains(runtime[1], "go run") {
		t.Fatal("production runtime must use compiled binaries without go run")
	}

	ignoreContent, err := os.ReadFile(filepath.Join("..", ".dockerignore"))
	if err != nil {
		t.Fatalf("read Docker build exclusions: %v", err)
	}
	ignore := make(map[string]bool)
	for _, line := range strings.Split(string(ignoreContent), "\n") {
		ignore[strings.TrimSpace(line)] = true
	}
	for _, excluded := range []string{".git", ".env", ".env.*", ".idea", ".vscode", "*.log"} {
		if !ignore[excluded] {
			t.Errorf("sensitive or unnecessary Docker context entry %q is not excluded", excluded)
		}
	}
	if !ignore["!.env.example"] {
		t.Error("safe example environment file must remain explicitly allowed")
	}
}
