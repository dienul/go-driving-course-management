package services

import (
	"context"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"strings"
)

func (s *TrainingSessionService) StudentSkills(ctx context.Context, studentID int64) (*dto.StudentSkillsData, error) {
	records, err := s.records.StudentSkills(ctx, studentID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	score, level := repositories.CalculateSkillScore(records)
	result := &dto.StudentSkillsData{StudentID: studentID, SkillScore: score, SkillLevel: level, TotalSubMaterials: len(records), Skills: make([]dto.StudentSkillData, 0, len(records))}
	for _, item := range records {
		result.Skills = append(result.Skills, dto.StudentSkillData{MaterialID: item.MaterialID, MaterialName: item.MaterialName, SubMaterialID: item.SubMaterialID, SubMaterialName: item.SubMaterialName, SkillStatus: item.SkillStatus, TrainingSessionID: item.TrainingSessionID, AssessedAt: item.AssessedAt})
	}
	return result, nil
}
func (s *TrainingSessionService) StudentSkillHistory(ctx context.Context, studentID int64) ([]dto.StudentSkillHistoryData, error) {
	records, err := s.records.StudentSkillHistory(ctx, studentID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := make([]dto.StudentSkillHistoryData, 0, len(records))
	for _, item := range records {
		result = append(result, dto.StudentSkillHistoryData{ID: item.ID, TrainingSessionID: item.TrainingSessionID, SessionNumber: item.SessionNumber, EnrollmentID: item.EnrollmentID, TrainerID: item.TrainerID, MaterialID: item.MaterialID, MaterialName: item.MaterialName, SubMaterialID: item.SubMaterialID, SubMaterialName: item.SubMaterialName, SkillStatus: item.SkillStatus, CompletedAt: item.CompletedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt})
	}
	return result, nil
}
func validateTrainerReview(request dto.UpsertTrainerReviewRequest) (models.TrainerReview, error) {
	if request.Rating < 1 || request.Rating > 5 {
		return models.TrainerReview{}, ErrInvalidInput
	}
	review := models.TrainerReview{Rating: request.Rating}
	if request.Feedback != nil {
		feedback := strings.TrimSpace(*request.Feedback)
		if len(feedback) > 2000 {
			return models.TrainerReview{}, ErrInvalidInput
		}
		if feedback != "" {
			review.Feedback = &feedback
		}
	}
	return review, nil
}
func (s *TrainingSessionService) CreateStudentReview(ctx context.Context, studentID, id int64, request dto.UpsertTrainerReviewRequest) (*dto.TrainerReviewData, error) {
	review, err := validateTrainerReview(request)
	if err != nil {
		return nil, err
	}
	record, err := s.records.CreateStudentReview(ctx, studentID, id, review)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := reviewResponse(record)
	return &result, nil
}
func (s *TrainingSessionService) UpdateStudentReview(ctx context.Context, studentID, id int64, request dto.UpsertTrainerReviewRequest) (*dto.TrainerReviewData, error) {
	review, err := validateTrainerReview(request)
	if err != nil {
		return nil, err
	}
	record, err := s.records.UpdateStudentReview(ctx, studentID, id, review)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := reviewResponse(record)
	return &result, nil
}
func (s *TrainingSessionService) StudentReview(ctx context.Context, studentID, id int64) (*dto.TrainerReviewData, error) {
	record, err := s.records.StudentReview(ctx, studentID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := reviewResponse(record)
	return &result, nil
}
func (s *TrainingSessionService) TrainerReviews(ctx context.Context, trainerID int64) ([]dto.TrainerReviewData, error) {
	records, err := s.records.TrainerReviews(ctx, trainerID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return reviewResponses(records), nil
}
func (s *TrainingSessionService) TrainerReviewSummary(ctx context.Context, trainerID int64) (*dto.TrainerReviewSummaryData, error) {
	record, err := s.records.TrainerReviewSummary(ctx, trainerID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return &dto.TrainerReviewSummaryData{TrainerID: trainerID, TotalReviews: record.TotalReviews, AverageRating: record.AverageRating}, nil
}
func (s *TrainingSessionService) AdminReviews(ctx context.Context) ([]dto.TrainerReviewData, error) {
	records, err := s.records.AdminReviews(ctx)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return reviewResponses(records), nil
}
func (s *TrainingSessionService) AdminTrainerReviews(ctx context.Context, trainerID int64) ([]dto.TrainerReviewData, error) {
	records, err := s.records.AdminTrainerReviews(ctx, trainerID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return reviewResponses(records), nil
}
func (s *TrainingSessionService) StudentCertificates(ctx context.Context, studentID int64) ([]dto.CertificateData, error) {
	records, err := s.records.StudentCertificates(ctx, studentID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return certificateResponses(records), nil
}
func (s *TrainingSessionService) StudentCertificate(ctx context.Context, studentID, id int64) (*dto.CertificateData, error) {
	record, err := s.records.StudentCertificate(ctx, studentID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := certificateResponse(record)
	return &result, nil
}
func (s *TrainingSessionService) AdminCertificates(ctx context.Context) ([]dto.CertificateData, error) {
	records, err := s.records.AdminCertificates(ctx)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return certificateResponses(records), nil
}
func (s *TrainingSessionService) AdminCertificate(ctx context.Context, id int64) (*dto.CertificateData, error) {
	record, err := s.records.AdminCertificate(ctx, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := certificateResponse(record)
	return &result, nil
}
func reviewResponses(records []repositories.TrainerReviewRecord) []dto.TrainerReviewData {
	result := make([]dto.TrainerReviewData, 0, len(records))
	for i := range records {
		result = append(result, reviewResponse(&records[i]))
	}
	return result
}
func reviewResponse(item *repositories.TrainerReviewRecord) dto.TrainerReviewData {
	return dto.TrainerReviewData{ID: item.ID, TrainingSessionID: item.TrainingSessionID, TrainerID: item.TrainerID, TrainerName: item.TrainerName, StudentID: item.StudentID, StudentName: item.StudentName, Rating: item.Rating, Feedback: item.Feedback, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func certificateResponses(records []repositories.CertificateRecord) []dto.CertificateData {
	result := make([]dto.CertificateData, 0, len(records))
	for i := range records {
		result = append(result, certificateResponse(&records[i]))
	}
	return result
}
func certificateResponse(record *repositories.CertificateRecord) dto.CertificateData {
	value := record.Certificate
	return dto.CertificateData{ID: value.ID, EnrollmentID: value.EnrollmentID, CertificateNumber: value.CertificateNumber, Student: userResponse(&record.Student), Enrollment: record.Enrollment, SkillScore: value.SkillScore, SkillLevel: value.SkillLevel, IssuedAt: value.IssuedAt, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
