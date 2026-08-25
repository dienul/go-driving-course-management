package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationPairsAreCompleteAndOrdered(t *testing.T) {
	names := []string{
		"create_users",
		"create_student_profiles",
		"create_trainer_profiles",
		"create_course_packages",
		"create_materials",
		"create_sub_materials",
		"create_student_enrollments",
		"create_payments",
		"create_invoices",
		"create_trainer_availabilities",
		"create_training_sessions",
		"create_session_skill_assessments",
		"create_session_evaluations",
		"create_trainer_reviews",
		"create_certificates",
	}

	for index, name := range names {
		prefix := fmt.Sprintf("%06d_%s", index+1, name)
		up := readMigration(t, prefix+".up.sql")
		down := readMigration(t, prefix+".down.sql")

		if !strings.Contains(strings.ToUpper(up), "CREATE TABLE") {
			t.Errorf("%s.up.sql does not create a table", prefix)
		}
		if !strings.Contains(strings.ToUpper(down), "DROP TABLE") {
			t.Errorf("%s.down.sql does not drop a table", prefix)
		}
	}

	files, err := filepath.Glob(filepath.Join("..", "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("list migration files: %v", err)
	}
	if len(files) != len(names)*2 {
		t.Fatalf("expected %d SQL migration files, got %d", len(names)*2, len(files))
	}
}

func TestRequiredDatabaseRulesAreInMigrations(t *testing.T) {
	enrollment := readMigration(t, "000007_create_student_enrollments.up.sql")
	if !strings.Contains(enrollment, "uq_student_active_enrollment") ||
		!strings.Contains(enrollment, "WHERE status = 'ACTIVE'") {
		t.Fatal("active enrollment partial unique index is missing")
	}

	assessment := readMigration(t, "000012_create_session_skill_assessments.up.sql")
	if !strings.Contains(assessment, "UNIQUE (training_session_id, sub_material_id)") {
		t.Fatal("assessment composite unique constraint is missing")
	}

	session := readMigration(t, "000011_create_training_sessions.up.sql")
	if !strings.Contains(session, "end_time = start_time + INTERVAL '2 hours'") {
		t.Fatal("two-hour training session constraint is missing")
	}
}

func readMigration(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "migrations", name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		t.Fatalf("migration %s is empty", name)
	}
	return string(content)
}
