package services

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	users  *repositories.AdminUserRepository
	master *repositories.MasterDataRepository
}

func NewAdminService(users *repositories.AdminUserRepository, master *repositories.MasterDataRepository) *AdminService {
	return &AdminService{users: users, master: master}
}

func (s *AdminService) CreateTrainer(ctx context.Context, request dto.CreateTrainerRequest) (*dto.TrainerData, error) {
	name, email := strings.TrimSpace(request.Name), strings.ToLower(strings.TrimSpace(request.Email))
	if name == "" || email == "" || len(request.Password) < 8 || len([]byte(request.Password)) > 72 || !utf8.ValidString(request.Password) {
		return nil, ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{Name: name, Email: email, PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive}
	profile := &models.TrainerProfile{Phone: normalizedOptional(request.Phone), Address: normalizedOptional(request.Address), Bio: normalizedOptional(request.Bio)}
	if err := s.users.CreateTrainer(ctx, user, profile); err != nil {
		return nil, adminError(err)
	}
	result := trainerResponse(&repositories.TrainerRecord{User: *user, Profile: *profile})
	return &result, nil
}

func (s *AdminService) ListTrainers(ctx context.Context) ([]dto.TrainerData, error) {
	records, err := s.users.ListTrainers(ctx)
	if err != nil {
		return nil, adminError(err)
	}
	results := make([]dto.TrainerData, 0, len(records))
	for i := range records {
		results = append(results, trainerResponse(&records[i]))
	}
	return results, nil
}

func (s *AdminService) GetTrainer(ctx context.Context, id int64) (*dto.TrainerData, error) {
	record, err := s.users.GetTrainer(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	result := trainerResponse(record)
	return &result, nil
}

func (s *AdminService) UpdateTrainer(ctx context.Context, id int64, request dto.UpdateTrainerRequest) (*dto.TrainerData, error) {
	name, email := strings.TrimSpace(request.Name), strings.ToLower(strings.TrimSpace(request.Email))
	if name == "" || email == "" {
		return nil, ErrInvalidInput
	}
	record, err := s.users.GetTrainer(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	record.User.Name, record.User.Email = name, email
	record.Profile.Phone, record.Profile.Address, record.Profile.Bio = normalizedOptional(request.Phone), normalizedOptional(request.Address), normalizedOptional(request.Bio)
	if err := s.users.UpdateTrainer(ctx, &record.User, &record.Profile); err != nil {
		return nil, adminError(err)
	}
	return s.GetTrainer(ctx, id)
}

func (s *AdminService) UpdateTrainerStatus(ctx context.Context, id int64, status models.RecordStatus) (*dto.TrainerData, error) {
	if !validRecordStatus(status) {
		return nil, ErrInvalidInput
	}
	if err := s.users.UpdateUserStatus(ctx, id, models.RoleTrainer, status); err != nil {
		return nil, adminError(err)
	}
	return s.GetTrainer(ctx, id)
}

func (s *AdminService) ListStudents(ctx context.Context) ([]dto.StudentData, error) {
	records, err := s.users.ListStudents(ctx)
	if err != nil {
		return nil, adminError(err)
	}
	results := make([]dto.StudentData, 0, len(records))
	for i := range records {
		results = append(results, studentResponse(&records[i]))
	}
	return results, nil
}

func (s *AdminService) GetStudent(ctx context.Context, id int64) (*dto.StudentData, error) {
	record, err := s.users.GetStudent(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	result := studentResponse(record)
	return &result, nil
}

func (s *AdminService) UpdateStudentStatus(ctx context.Context, id int64, status models.RecordStatus) (*dto.StudentData, error) {
	if !validRecordStatus(status) {
		return nil, ErrInvalidInput
	}
	if err := s.users.UpdateUserStatus(ctx, id, models.RoleStudent, status); err != nil {
		return nil, adminError(err)
	}
	return s.GetStudent(ctx, id)
}

func (s *AdminService) CreatePackage(ctx context.Context, request dto.UpsertPackageRequest) (*models.CoursePackage, error) {
	if !validPackage(request) {
		return nil, ErrInvalidInput
	}
	pkg := &models.CoursePackage{Name: strings.TrimSpace(request.Name), Level: request.Level, TotalHours: request.TotalHours, Price: request.Price, Description: normalizedOptional(request.Description), Status: models.StatusActive}
	if err := s.master.CreatePackage(ctx, pkg); err != nil {
		return nil, adminError(err)
	}
	return pkg, nil
}

func (s *AdminService) ListPackages(ctx context.Context) ([]models.CoursePackage, error) {
	results, err := s.master.ListPackages(ctx)
	return results, adminError(err)
}

func (s *AdminService) GetPackage(ctx context.Context, id int64) (*models.CoursePackage, error) {
	result, err := s.master.GetPackage(ctx, id)
	return result, adminError(err)
}

func (s *AdminService) UpdatePackage(ctx context.Context, id int64, request dto.UpsertPackageRequest) (*models.CoursePackage, error) {
	if !validPackage(request) {
		return nil, ErrInvalidInput
	}
	pkg, err := s.master.GetPackage(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	pkg.Name, pkg.Level, pkg.TotalHours, pkg.Price, pkg.Description = strings.TrimSpace(request.Name), request.Level, request.TotalHours, request.Price, normalizedOptional(request.Description)
	if err := s.master.SavePackage(ctx, pkg); err != nil {
		return nil, adminError(err)
	}
	return pkg, nil
}

func (s *AdminService) UpdatePackageStatus(ctx context.Context, id int64, status models.RecordStatus) (*models.CoursePackage, error) {
	if !validRecordStatus(status) {
		return nil, ErrInvalidInput
	}
	pkg, err := s.master.GetPackage(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	pkg.Status = status
	if err := s.master.SavePackage(ctx, pkg); err != nil {
		return nil, adminError(err)
	}
	return pkg, nil
}

func (s *AdminService) CreateMaterial(ctx context.Context, request dto.UpsertMaterialRequest) (*models.Material, error) {
	if !validMaterial(request.Name, request.Sequence) {
		return nil, ErrInvalidInput
	}
	material := &models.Material{Name: strings.TrimSpace(request.Name), Description: normalizedOptional(request.Description), Sequence: request.Sequence, Status: models.StatusActive}
	if err := s.master.CreateMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func (s *AdminService) ListMaterials(ctx context.Context) ([]models.Material, error) {
	results, err := s.master.ListMaterials(ctx)
	return results, adminError(err)
}

func (s *AdminService) GetMaterial(ctx context.Context, id int64) (*models.Material, error) {
	result, err := s.master.GetMaterial(ctx, id)
	return result, adminError(err)
}

func (s *AdminService) UpdateMaterial(ctx context.Context, id int64, request dto.UpsertMaterialRequest) (*models.Material, error) {
	if !validMaterial(request.Name, request.Sequence) {
		return nil, ErrInvalidInput
	}
	material, err := s.master.GetMaterial(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	material.Name, material.Description, material.Sequence = strings.TrimSpace(request.Name), normalizedOptional(request.Description), request.Sequence
	if err := s.master.SaveMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func (s *AdminService) UpdateMaterialStatus(ctx context.Context, id int64, status models.RecordStatus) (*models.Material, error) {
	if !validRecordStatus(status) {
		return nil, ErrInvalidInput
	}
	material, err := s.master.GetMaterial(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	material.Status = status
	if err := s.master.SaveMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func (s *AdminService) CreateSubMaterial(ctx context.Context, materialID int64, request dto.UpsertSubMaterialRequest) (*models.SubMaterial, error) {
	if !validMaterial(request.Name, request.Sequence) {
		return nil, ErrInvalidInput
	}
	if _, err := s.master.GetMaterial(ctx, materialID); err != nil {
		return nil, adminError(err)
	}
	material := &models.SubMaterial{MaterialID: materialID, Name: strings.TrimSpace(request.Name), Description: normalizedOptional(request.Description), Sequence: request.Sequence, Status: models.StatusActive}
	if err := s.master.CreateSubMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func (s *AdminService) ListSubMaterials(ctx context.Context, materialID int64) ([]models.SubMaterial, error) {
	if _, err := s.master.GetMaterial(ctx, materialID); err != nil {
		return nil, adminError(err)
	}
	results, err := s.master.ListSubMaterials(ctx, materialID)
	return results, adminError(err)
}

func (s *AdminService) GetSubMaterial(ctx context.Context, id int64) (*models.SubMaterial, error) {
	result, err := s.master.GetSubMaterial(ctx, id)
	return result, adminError(err)
}

func (s *AdminService) UpdateSubMaterial(ctx context.Context, id int64, request dto.UpsertSubMaterialRequest) (*models.SubMaterial, error) {
	if !validMaterial(request.Name, request.Sequence) {
		return nil, ErrInvalidInput
	}
	material, err := s.master.GetSubMaterial(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	material.Name, material.Description, material.Sequence = strings.TrimSpace(request.Name), normalizedOptional(request.Description), request.Sequence
	if err := s.master.SaveSubMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func (s *AdminService) UpdateSubMaterialStatus(ctx context.Context, id int64, status models.RecordStatus) (*models.SubMaterial, error) {
	if !validRecordStatus(status) {
		return nil, ErrInvalidInput
	}
	material, err := s.master.GetSubMaterial(ctx, id)
	if err != nil {
		return nil, adminError(err)
	}
	material.Status = status
	if err := s.master.SaveSubMaterial(ctx, material); err != nil {
		return nil, adminError(err)
	}
	return material, nil
}

func validRecordStatus(status models.RecordStatus) bool {
	return status == models.StatusActive || status == models.StatusInactive
}

func validPackage(request dto.UpsertPackageRequest) bool {
	if strings.TrimSpace(request.Name) == "" || request.Price <= 0 {
		return false
	}
	if request.Level != models.PackageLevelPemula && request.Level != models.PackageLevelDasar {
		return false
	}
	switch request.TotalHours {
	case 6, 8, 10, 12:
		return true
	}
	return false
}

func validMaterial(name string, sequence int) bool {
	return strings.TrimSpace(name) != "" && sequence > 0
}

func adminError(err error) error {
	if errors.Is(err, repositories.ErrRecordNotFound) {
		return ErrResourceNotFound
	}
	if errors.Is(err, repositories.ErrDuplicateRecord) {
		return ErrResourceConflict
	}
	return err
}

func trainerResponse(record *repositories.TrainerRecord) dto.TrainerData {
	profile := record.Profile
	return dto.TrainerData{User: userResponse(&record.User), Profile: dto.TrainerProfileData{ID: profile.ID, UserID: profile.UserID, Phone: profile.Phone, Address: profile.Address, Bio: profile.Bio, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt}}
}

func studentResponse(record *repositories.StudentRecord) dto.StudentData {
	profile := record.Profile
	return dto.StudentData{User: userResponse(&record.User), Profile: dto.StudentProfileData{ID: profile.ID, UserID: profile.UserID, Phone: profile.Phone, Address: profile.Address, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt}}
}
