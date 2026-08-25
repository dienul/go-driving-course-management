package tests

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgreSQLMigrationsAndDatabaseConstraints(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}

	var migration struct {
		Version int64
		Dirty   bool
	}
	if err := db.Table("schema_migrations").Take(&migration).Error; err != nil {
		t.Fatalf("read PostgreSQL migration version: %v", err)
	}
	if migration.Version != 15 || migration.Dirty {
		t.Fatalf("migration version = %d, dirty = %t; expected clean version 15", migration.Version, migration.Dirty)
	}
	tables := []string{
		"users", "student_profiles", "trainer_profiles", "course_packages", "materials",
		"sub_materials", "student_enrollments", "payments", "invoices", "trainer_availabilities",
		"training_sessions", "session_skill_assessments", "session_evaluations", "trainer_reviews",
		"certificates",
	}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("migrated PostgreSQL table %q is missing", table)
		}
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin isolated constraint transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	insertID := func(query string, arguments ...any) int64 {
		t.Helper()
		var id int64
		if err := tx.Raw(query, arguments...).Scan(&id).Error; err != nil {
			t.Fatalf("insert PostgreSQL fixture: %v", err)
		}
		return id
	}
	studentID := insertID(
		"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, 'STUDENT') RETURNING id",
		"Phase 11 Constraint Student", "phase11-constraint-student@example.com", "hashed-password",
	)
	trainerID := insertID(
		"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, 'TRAINER') RETURNING id",
		"Phase 11 Constraint Trainer", "phase11-constraint-trainer@example.com", "hashed-password",
	)
	adminID := insertID(
		"INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, 'ADMIN') RETURNING id",
		"Phase 11 Constraint Admin", "phase11-constraint-admin@example.com", "hashed-password",
	)
	packageID := insertID(
		"INSERT INTO course_packages (name, level, total_hours, price) VALUES (?, 'PEMULA', 6, 900000) RETURNING id",
		"Phase 11 Constraint Package",
	)
	materialID := insertID(
		"INSERT INTO materials (name, sequence) VALUES (?, 11001) RETURNING id",
		"Phase 11 Constraint Material",
	)
	subMaterialID := insertID(
		"INSERT INTO sub_materials (material_id, name, sequence) VALUES (?, ?, 1) RETURNING id",
		materialID, "Phase 11 Constraint Skill",
	)
	enrollmentID := insertID(
		"INSERT INTO student_enrollments (student_id, package_id, package_name, package_price, total_hours, status, started_at) VALUES (?, ?, ?, 900000, 6, 'ACTIVE', CURRENT_TIMESTAMP) RETURNING id",
		studentID, packageID, "Phase 11 Constraint Package",
	)
	availabilityID := insertID(
		"INSERT INTO trainer_availabilities (trainer_id, available_date, start_time, end_time, status, published_by, published_at) VALUES (?, '2026-09-21', '08:00', '16:00', 'PUBLISHED', ?, CURRENT_TIMESTAMP) RETURNING id",
		trainerID, adminID,
	)
	sessionID := insertID(
		"INSERT INTO training_sessions (enrollment_id, trainer_id, trainer_availability_id, session_number, scheduled_date, start_time, end_time) VALUES (?, ?, ?, 1, '2026-09-21', '08:00', '10:00') RETURNING id",
		enrollmentID, trainerID, availabilityID,
	)
	insertID(
		"INSERT INTO session_skill_assessments (training_session_id, sub_material_id, skill_status) VALUES (?, ?, 'MASTERED') RETURNING id",
		sessionID, subMaterialID,
	)

	tests := []struct {
		name      string
		statement string
		arguments []any
		sqlState  string
	}{
		{
			name: "user role enum", sqlState: "23514",
			statement: "INSERT INTO users (name, email, password_hash, role) VALUES ('Invalid Role', 'phase11-invalid-role@example.com', 'hash', 'SUPERADMIN')",
		},
		{
			name: "duplicate user email", sqlState: "23505",
			statement: "INSERT INTO users (name, email, password_hash, role) VALUES ('Duplicate', ?, 'hash', 'STUDENT')",
			arguments: []any{"phase11-constraint-student@example.com"},
		},
		{
			name: "package purchased hours", sqlState: "23514",
			statement: "INSERT INTO course_packages (name, level, total_hours, price) VALUES ('Invalid Hours', 'PEMULA', 7, 900000)",
		},
		{
			name: "package positive price", sqlState: "23514",
			statement: "INSERT INTO course_packages (name, level, total_hours, price) VALUES ('Invalid Price', 'PEMULA', 6, 0)",
		},
		{
			name: "single active student enrollment", sqlState: "23505",
			statement: "INSERT INTO student_enrollments (student_id, package_id, package_name, package_price, total_hours, status, started_at) VALUES (?, ?, 'Duplicate Active', 900000, 6, 'ACTIVE', CURRENT_TIMESTAMP)",
			arguments: []any{studentID, packageID},
		},
		{
			name: "paid payment requires method and paid timestamp", sqlState: "23514",
			statement: "INSERT INTO payments (enrollment_id, payment_code, amount, status) VALUES (?, 'PHASE11-INVALID-PAID', 900000, 'PAID')",
			arguments: []any{enrollmentID},
		},
		{
			name: "trainer availability rejects weekends", sqlState: "23514",
			statement: "INSERT INTO trainer_availabilities (trainer_id, available_date, start_time, end_time) VALUES (?, '2026-09-26', '08:00', '10:00')",
			arguments: []any{trainerID},
		},
		{
			name: "trainer availability enforces operating hours", sqlState: "23514",
			statement: "INSERT INTO trainer_availabilities (trainer_id, available_date, start_time, end_time) VALUES (?, '2026-09-22', '07:00', '09:00')",
			arguments: []any{trainerID},
		},
		{
			name: "training sessions are exactly two hours", sqlState: "23514",
			statement: "INSERT INTO training_sessions (enrollment_id, trainer_id, trainer_availability_id, session_number, scheduled_date, start_time, end_time) VALUES (?, ?, ?, 2, '2026-09-21', '10:00', '11:00')",
			arguments: []any{enrollmentID, trainerID, availabilityID},
		},
		{
			name: "completed training sessions require lifecycle timestamps", sqlState: "23514",
			statement: "INSERT INTO training_sessions (enrollment_id, trainer_id, trainer_availability_id, session_number, scheduled_date, start_time, end_time, status) VALUES (?, ?, ?, 2, '2026-09-21', '10:00', '12:00', 'COMPLETED')",
			arguments: []any{enrollmentID, trainerID, availabilityID},
		},
		{
			name: "one assessment per session and sub-material", sqlState: "23505",
			statement: "INSERT INTO session_skill_assessments (training_session_id, sub_material_id, skill_status) VALUES (?, ?, 'PRACTICED')",
			arguments: []any{sessionID, subMaterialID},
		},
		{
			name: "trainer review rating range", sqlState: "23514",
			statement: "INSERT INTO trainer_reviews (training_session_id, rating) VALUES (?, 6)",
			arguments: []any{sessionID},
		},
		{
			name: "certificate score range", sqlState: "23514",
			statement: "INSERT INTO certificates (enrollment_id, certificate_number, skill_score, skill_level, issued_at) VALUES (?, 'PHASE11-INVALID-CERT', 101, 'PROFICIENT', CURRENT_TIMESTAMP)",
			arguments: []any{enrollmentID},
		},
		{
			name: "certificate enrollment foreign key", sqlState: "23503",
			statement: "INSERT INTO certificates (enrollment_id, certificate_number, skill_score, skill_level, issued_at) VALUES (9223372036854775807, 'PHASE11-FOREIGN-CERT', 80, 'PROFICIENT', CURRENT_TIMESTAMP)",
		},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			savepoint := fmt.Sprintf("phase11_constraint_%d", index)
			if err := tx.SavePoint(savepoint).Error; err != nil {
				t.Fatalf("create PostgreSQL savepoint: %v", err)
			}
			err := tx.Exec(test.statement, test.arguments...).Error
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != test.sqlState {
				t.Fatalf("PostgreSQL error = %v; expected SQLSTATE %s", err, test.sqlState)
			}
			if err := tx.RollbackTo(savepoint).Error; err != nil {
				t.Fatalf("rollback PostgreSQL savepoint: %v", err)
			}
		})
	}
}
