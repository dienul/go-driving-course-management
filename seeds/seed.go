package seeds

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dienulhaq/go-driving-course-management/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Config struct {
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

type materialSeed struct {
	Name         string
	Description  string
	Sequence     int
	SubMaterials []subMaterialSeed
}

type subMaterialSeed struct {
	Name        string
	Description string
	Sequence    int
}

func Run(db *gorm.DB, cfg Config) error {
	if db == nil {
		return errors.New("database is required")
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedAdmin(tx, cfg); err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}
		if err := seedPackages(tx); err != nil {
			return fmt.Errorf("seed course packages: %w", err)
		}
		if err := seedCurriculum(tx); err != nil {
			return fmt.Errorf("seed curriculum: %w", err)
		}
		return nil
	})
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.AdminName) == "" {
		return errors.New("ADMIN_NAME is required")
	}
	if strings.TrimSpace(cfg.AdminEmail) == "" {
		return errors.New("ADMIN_EMAIL is required")
	}
	if strings.TrimSpace(cfg.AdminPassword) == "" {
		return errors.New("ADMIN_PASSWORD is required")
	}
	if len([]byte(cfg.AdminPassword)) > 72 {
		return errors.New("ADMIN_PASSWORD must be at most 72 bytes")
	}
	return nil
}

func seedAdmin(tx *gorm.DB, cfg Config) error {
	email := strings.ToLower(strings.TrimSpace(cfg.AdminEmail))

	var existing models.User
	err := tx.Where("email = ?", email).First(&existing).Error
	if err == nil {
		if existing.Role != models.RoleAdmin {
			return fmt.Errorf("email %q already belongs to role %s", email, existing.Role)
		}
		return tx.Model(&existing).Updates(map[string]any{
			"name":   strings.TrimSpace(cfg.AdminName),
			"status": models.StatusActive,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(cfg.AdminPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := models.User{
		Name:         strings.TrimSpace(cfg.AdminName),
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
	}
	return tx.Create(&admin).Error
}

func seedPackages(tx *gorm.DB) error {
	packages := []models.CoursePackage{
		{
			Name:        "Pemula 6 Jam",
			Level:       models.PackageLevelPemula,
			TotalHours:  6,
			Price:       900_000,
			Description: stringPointer("Paket latihan pemula selama 6 jam atau 3 sesi."),
			Status:      models.StatusActive,
		},
		{
			Name:        "Pemula 8 Jam",
			Level:       models.PackageLevelPemula,
			TotalHours:  8,
			Price:       1_100_000,
			Description: stringPointer("Paket latihan pemula selama 8 jam atau 4 sesi."),
			Status:      models.StatusActive,
		},
		{
			Name:        "Dasar 10 Jam",
			Level:       models.PackageLevelDasar,
			TotalHours:  10,
			Price:       1_400_000,
			Description: stringPointer("Paket latihan dasar selama 10 jam atau 5 sesi."),
			Status:      models.StatusActive,
		},
		{
			Name:        "Dasar 12 Jam",
			Level:       models.PackageLevelDasar,
			TotalHours:  12,
			Price:       1_600_000,
			Description: stringPointer("Paket latihan dasar selama 12 jam atau 6 sesi."),
			Status:      models.StatusActive,
		},
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"level", "total_hours", "price", "description", "status", "updated_at",
		}),
	}).Create(&packages).Error
}

func seedCurriculum(tx *gorm.DB) error {
	for _, item := range curriculum() {
		material := models.Material{
			Name:        item.Name,
			Description: stringPointer(item.Description),
			Sequence:    item.Sequence,
			Status:      models.StatusActive,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"description", "sequence", "status", "updated_at",
			}),
		}).Create(&material).Error; err != nil {
			return err
		}
		if err := tx.Where("name = ?", item.Name).First(&material).Error; err != nil {
			return err
		}

		for _, child := range item.SubMaterials {
			subMaterial := models.SubMaterial{
				MaterialID:  material.ID,
				Name:        child.Name,
				Description: stringPointer(child.Description),
				Sequence:    child.Sequence,
				Status:      models.StatusActive,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "material_id"},
					{Name: "name"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"description", "sequence", "status", "updated_at",
				}),
			}).Create(&subMaterial).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func curriculum() []materialSeed {
	return []materialSeed{
		{
			Name:        "Pengenalan Kendaraan",
			Description: "Mengenal kendaraan, fungsi kontrol, dan persiapan berkendara.",
			Sequence:    1,
			SubMaterials: []subMaterialSeed{
				{Name: "Pemeriksaan Kendaraan", Description: "Pemeriksaan dasar sebelum berkendara.", Sequence: 1},
				{Name: "Instrumen dan Dashboard", Description: "Mengenal indikator dan kontrol kendaraan.", Sequence: 2},
				{Name: "Posisi Mengemudi", Description: "Mengatur kursi, spion, dan posisi tangan.", Sequence: 3},
			},
		},
		{
			Name:        "Menjalankan Kendaraan",
			Description: "Teknik dasar menghidupkan, menjalankan, dan menghentikan kendaraan.",
			Sequence:    2,
			SubMaterials: []subMaterialSeed{
				{Name: "Menghidupkan dan Mematikan Mesin", Description: "Prosedur aman mengoperasikan mesin.", Sequence: 1},
				{Name: "Pengoperasian Kopling", Description: "Mengontrol kopling secara halus.", Sequence: 2},
				{Name: "Pemindahan Gigi", Description: "Memilih dan memindahkan gigi sesuai kondisi.", Sequence: 3},
				{Name: "Mulai dan Berhenti", Description: "Menjalankan dan menghentikan kendaraan dengan aman.", Sequence: 4},
			},
		},
		{
			Name:        "Mengendalikan Kendaraan",
			Description: "Mengendalikan arah, kecepatan, dan pengereman kendaraan.",
			Sequence:    3,
			SubMaterials: []subMaterialSeed{
				{Name: "Kontrol Kemudi", Description: "Mengendalikan kemudi dan menjaga arah kendaraan.", Sequence: 1},
				{Name: "Pengereman", Description: "Melakukan pengereman halus dan aman.", Sequence: 2},
				{Name: "Kontrol Kecepatan", Description: "Menyesuaikan kecepatan dengan kondisi.", Sequence: 3},
				{Name: "Berbelok", Description: "Melakukan manuver belok dengan observasi yang benar.", Sequence: 4},
			},
		},
		{
			Name:        "Menjalankan Kendaraan di Jalan Raya",
			Description: "Menerapkan teknik berkendara aman di lalu lintas.",
			Sequence:    4,
			SubMaterials: []subMaterialSeed{
				{Name: "Observasi Lalu Lintas", Description: "Membaca situasi dan potensi bahaya.", Sequence: 1},
				{Name: "Penggunaan Spion dan Lampu Sein", Description: "Berkomunikasi dengan pengguna jalan lain.", Sequence: 2},
				{Name: "Pindah Lajur", Description: "Berpindah lajur dengan observasi dan jarak aman.", Sequence: 3},
				{Name: "Rambu dan Etika Berkendara", Description: "Mematuhi rambu dan menerapkan etika jalan raya.", Sequence: 4},
			},
		},
		{
			Name:        "Memundurkan Kendaraan dan Parkir",
			Description: "Menguasai kontrol kendaraan saat mundur dan parkir.",
			Sequence:    5,
			SubMaterials: []subMaterialSeed{
				{Name: "Mundur Lurus", Description: "Mengendalikan kendaraan saat bergerak mundur.", Sequence: 1},
				{Name: "Parkir Paralel", Description: "Memarkir kendaraan sejajar dengan jalan.", Sequence: 2},
				{Name: "Parkir Seri", Description: "Memarkir kendaraan tegak lurus atau menyudut.", Sequence: 3},
				{Name: "Parkir Mundur", Description: "Memasuki ruang parkir dengan gerakan mundur.", Sequence: 4},
			},
		},
	}
}

func stringPointer(value string) *string {
	return &value
}
