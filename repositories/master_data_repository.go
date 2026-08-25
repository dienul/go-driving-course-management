package repositories

import (
	"context"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
)

type MasterDataRepository struct {
	db *gorm.DB
}

func NewMasterDataRepository(db *gorm.DB) *MasterDataRepository {
	return &MasterDataRepository{db: db}
}

func (r *MasterDataRepository) CreatePackage(
	ctx context.Context,
	pkg *models.CoursePackage,
) error {
	return mapAdminError(r.db.WithContext(ctx).Create(pkg).Error)
}

func (r *MasterDataRepository) ListPackages(
	ctx context.Context,
) ([]models.CoursePackage, error) {
	packages := make([]models.CoursePackage, 0)
	err := r.db.WithContext(ctx).Order("id ASC").Find(&packages).Error
	return packages, err
}

func (r *MasterDataRepository) GetPackage(
	ctx context.Context,
	id int64,
) (*models.CoursePackage, error) {
	var pkg models.CoursePackage
	if err := r.db.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &pkg, nil
}

func (r *MasterDataRepository) SavePackage(
	ctx context.Context,
	pkg *models.CoursePackage,
) error {
	return mapAdminError(r.db.WithContext(ctx).Save(pkg).Error)
}

func (r *MasterDataRepository) CreateMaterial(
	ctx context.Context,
	material *models.Material,
) error {
	return mapAdminError(r.db.WithContext(ctx).Create(material).Error)
}

func (r *MasterDataRepository) ListMaterials(
	ctx context.Context,
) ([]models.Material, error) {
	materials := make([]models.Material, 0)
	err := r.db.WithContext(ctx).Order("sequence ASC, id ASC").Find(&materials).Error
	return materials, err
}

func (r *MasterDataRepository) GetMaterial(
	ctx context.Context,
	id int64,
) (*models.Material, error) {
	var material models.Material
	if err := r.db.WithContext(ctx).First(&material, id).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &material, nil
}

func (r *MasterDataRepository) SaveMaterial(
	ctx context.Context,
	material *models.Material,
) error {
	return mapAdminError(r.db.WithContext(ctx).Save(material).Error)
}

func (r *MasterDataRepository) CreateSubMaterial(
	ctx context.Context,
	subMaterial *models.SubMaterial,
) error {
	return mapAdminError(r.db.WithContext(ctx).Create(subMaterial).Error)
}

func (r *MasterDataRepository) ListSubMaterials(
	ctx context.Context,
	materialID int64,
) ([]models.SubMaterial, error) {
	subMaterials := make([]models.SubMaterial, 0)
	err := r.db.WithContext(ctx).
		Where("material_id = ?", materialID).
		Order("sequence ASC, id ASC").
		Find(&subMaterials).Error
	return subMaterials, err
}

func (r *MasterDataRepository) GetSubMaterial(
	ctx context.Context,
	id int64,
) (*models.SubMaterial, error) {
	var subMaterial models.SubMaterial
	if err := r.db.WithContext(ctx).First(&subMaterial, id).Error; err != nil {
		return nil, mapAdminError(err)
	}
	return &subMaterial, nil
}

func (r *MasterDataRepository) SaveSubMaterial(
	ctx context.Context,
	subMaterial *models.SubMaterial,
) error {
	return mapAdminError(r.db.WithContext(ctx).Save(subMaterial).Error)
}
