package services

import (
	"errors"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

func TestValidPackage(t *testing.T) {
	valid := dto.UpsertPackageRequest{Name: "Pemula 6 Jam", Level: models.PackageLevelPemula, TotalHours: 6, Price: 900000}
	for _, hours := range []int16{6, 8, 10, 12} {
		request := valid
		request.TotalHours = hours
		if !validPackage(request) {
			t.Errorf("expected %d hours to be valid", hours)
		}
	}
	invalid := []dto.UpsertPackageRequest{
		{Name: " ", Level: valid.Level, TotalHours: 6, Price: 1},
		{Name: valid.Name, Level: models.PackageLevel("OTHER"), TotalHours: 6, Price: 1},
		{Name: valid.Name, Level: valid.Level, TotalHours: 7, Price: 1},
		{Name: valid.Name, Level: valid.Level, TotalHours: 6, Price: 0},
	}
	for index, request := range invalid {
		if validPackage(request) {
			t.Errorf("invalid package %d was accepted", index)
		}
	}
}

func TestAdminValidationAndErrorMapping(t *testing.T) {
	if !validRecordStatus(models.StatusActive) || !validRecordStatus(models.StatusInactive) || validRecordStatus(models.RecordStatus("DELETED")) {
		t.Fatal("record-status validation is incorrect")
	}
	if !validMaterial(" Steering ", 1) || validMaterial(" ", 1) || validMaterial("Steering", 0) {
		t.Fatal("material validation is incorrect")
	}
	if !errors.Is(adminError(repositories.ErrRecordNotFound), ErrResourceNotFound) {
		t.Fatal("missing records should map to a resource-not-found error")
	}
	if !errors.Is(adminError(repositories.ErrDuplicateRecord), ErrResourceConflict) {
		t.Fatal("duplicates should map to a resource-conflict error")
	}
}
