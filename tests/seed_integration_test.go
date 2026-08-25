package tests

import (
	"os"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/seeds"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSeedCreatesIdempotentAdminAndMasterDataPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin isolated seed transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	cfg := seeds.Config{
		AdminName: "Phase 11 Administrator", AdminEmail: "phase11-seed-admin@example.com",
		AdminPassword: "phase11-strong-admin-password",
	}
	if err := seeds.Run(tx, cfg); err != nil {
		t.Fatalf("run first database seed: %v", err)
	}
	var admin models.User
	if err := tx.Where("email = ?", cfg.AdminEmail).First(&admin).Error; err != nil {
		t.Fatalf("find seeded admin: %v", err)
	}
	if admin.Role != models.RoleAdmin || admin.Status != models.StatusActive || admin.Name != cfg.AdminName {
		t.Fatalf("invalid seeded administrator: %+v", admin)
	}
	if admin.PasswordHash == cfg.AdminPassword || bcrypt.CompareHashAndPassword(
		[]byte(admin.PasswordHash), []byte(cfg.AdminPassword),
	) != nil {
		t.Fatal("admin password was not securely hashed with bcrypt")
	}
	initialPasswordHash := admin.PasswordHash

	packageNames := []string{"Pemula 6 Jam", "Pemula 8 Jam", "Dasar 10 Jam", "Dasar 12 Jam"}
	materialNames := []string{
		"Pengenalan Kendaraan", "Menjalankan Kendaraan", "Mengendalikan Kendaraan",
		"Menjalankan Kendaraan di Jalan Raya", "Memundurkan Kendaraan dan Parkir",
	}
	assertCount := func(model any, query string, argument any, expected int64) {
		t.Helper()
		var total int64
		if err := tx.Model(model).Where(query, argument).Count(&total).Error; err != nil {
			t.Fatalf("count seeded records: %v", err)
		}
		if total != expected {
			t.Fatalf("seeded records = %d, want %d", total, expected)
		}
	}
	assertCount(&models.CoursePackage{}, "name IN ?", packageNames, 4)
	assertCount(&models.Material{}, "name IN ?", materialNames, 5)

	var subMaterials int64
	if err := tx.Model(&models.SubMaterial{}).Joins(
		"JOIN materials ON materials.id = sub_materials.material_id",
	).Where("materials.name IN ?", materialNames).Count(&subMaterials).Error; err != nil {
		t.Fatalf("count seeded sub-materials: %v", err)
	}
	if subMaterials != 19 {
		t.Fatalf("seeded sub-materials = %d, want 19", subMaterials)
	}

	cfg.AdminName = "Updated Phase 11 Administrator"
	cfg.AdminPassword = "a-different-password-must-not-overwrite"
	if err := seeds.Run(tx, cfg); err != nil {
		t.Fatalf("run idempotent database seed: %v", err)
	}
	if err := tx.Where("email = ?", cfg.AdminEmail).First(&admin).Error; err != nil {
		t.Fatalf("find updated seeded admin: %v", err)
	}
	if admin.Name != cfg.AdminName || admin.PasswordHash != initialPasswordHash {
		t.Fatalf("idempotent admin seed changed unexpected fields: %+v", admin)
	}
	assertCount(&models.User{}, "email = ?", cfg.AdminEmail, 1)
	assertCount(&models.CoursePackage{}, "name IN ?", packageNames, 4)
	assertCount(&models.Material{}, "name IN ?", materialNames, 5)
}
