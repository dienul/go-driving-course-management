package services

import (
	"context"
	"strings"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

func (s *TrainingSessionService) ListTrainer(ctx context.Context, trainerID int64) ([]dto.TrainingSessionData, error) {
	sessions, err := s.records.ListTrainer(ctx, trainerID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return trainingSessionResponses(sessions), nil
}

func (s *TrainingSessionService) GetTrainer(ctx context.Context, trainerID, id int64) (*dto.TrainingSessionData, error) {
	session, err := s.records.GetTrainer(ctx, trainerID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) StartTrainer(ctx context.Context, trainerID, id int64) (*dto.TrainingSessionData, error) {
	session, err := s.records.StartTrainer(ctx, trainerID, id, s.now().UTC())
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) UpsertTrainerAssessments(
	ctx context.Context, trainerID, id int64, request dto.UpsertSessionAssessmentsRequest,
) ([]dto.SessionAssessmentData, error) {
	assessments, err := validateSessionAssessments(request)
	if err != nil {
		return nil, err
	}
	records, err := s.records.UpsertTrainerAssessments(ctx, trainerID, id, assessments)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return assessmentResponses(records), nil
}

func (s *TrainingSessionService) TrainerAssessments(ctx context.Context, trainerID, id int64) ([]dto.SessionAssessmentData, error) {
	records, err := s.records.TrainerAssessments(ctx, trainerID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return assessmentResponses(records), nil
}

func (s *TrainingSessionService) UpsertTrainerEvaluation(
	ctx context.Context, trainerID, id int64, request dto.UpsertSessionEvaluationRequest,
) (*dto.SessionEvaluationData, error) {
	evaluation, err := validateSessionEvaluation(request)
	if err != nil {
		return nil, err
	}
	record, err := s.records.UpsertTrainerEvaluation(ctx, trainerID, id, evaluation)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := evaluationResponse(record)
	return &result, nil
}

func (s *TrainingSessionService) TrainerEvaluation(ctx context.Context, trainerID, id int64) (*dto.SessionEvaluationData, error) {
	record, err := s.records.TrainerEvaluation(ctx, trainerID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := evaluationResponse(record)
	return &result, nil
}

func (s *TrainingSessionService) CompleteTrainer(ctx context.Context, trainerID, id int64) (*dto.TrainingSessionData, error) {
	session, err := s.records.CompleteTrainer(ctx, trainerID, id, s.now().UTC())
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) TrainerStudentProgress(
	ctx context.Context, trainerID, id int64,
) (*dto.TrainerStudentProgressData, error) {
	record, err := s.records.TrainerStudentProgress(ctx, trainerID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := dto.TrainerStudentProgressData{
		StudentID: record.StudentID, TrainingSessionID: record.TrainingSessionID,
		PreviousSessions: make([]dto.PreviousTrainingSessionData, 0, len(record.PreviousSessions)),
	}
	for index := range record.PreviousSessions {
		previous := record.PreviousSessions[index]
		item := dto.PreviousTrainingSessionData{
			Session: trainingSessionResponse(&previous.Session), Assessments: assessmentResponses(previous.Assessments),
		}
		if previous.Evaluation != nil {
			evaluation := evaluationResponse(previous.Evaluation)
			item.Evaluation = &evaluation
		}
		result.PreviousSessions = append(result.PreviousSessions, item)
	}
	return &result, nil
}

func validateSessionAssessments(request dto.UpsertSessionAssessmentsRequest) ([]models.SessionSkillAssessment, error) {
	if len(request.Assessments) == 0 || len(request.Assessments) > 100 {
		return nil, ErrInvalidInput
	}
	result := make([]models.SessionSkillAssessment, 0, len(request.Assessments))
	seen := make(map[int64]struct{}, len(request.Assessments))
	for _, item := range request.Assessments {
		if item.SubMaterialID <= 0 {
			return nil, ErrInvalidInput
		}
		if _, exists := seen[item.SubMaterialID]; exists {
			return nil, ErrInvalidInput
		}
		switch item.SkillStatus {
		case models.SkillNotStarted, models.SkillPracticed, models.SkillMastered:
		default:
			return nil, ErrInvalidInput
		}
		seen[item.SubMaterialID] = struct{}{}
		result = append(result, models.SessionSkillAssessment{
			SubMaterialID: item.SubMaterialID, SkillStatus: item.SkillStatus,
		})
	}
	return result, nil
}

func validateSessionEvaluation(request dto.UpsertSessionEvaluationRequest) (models.SessionEvaluation, error) {
	switch request.Predicate {
	case models.PredicateKurang, models.PredicateCukup, models.PredicateBaik, models.PredicateSangatBaik:
	default:
		return models.SessionEvaluation{}, ErrInvalidInput
	}
	notes, recommendation := strings.TrimSpace(request.Notes), strings.TrimSpace(request.Recommendation)
	if notes == "" || recommendation == "" || len(notes) > 5000 || len(recommendation) > 5000 {
		return models.SessionEvaluation{}, ErrInvalidInput
	}
	return models.SessionEvaluation{
		Predicate: request.Predicate, Notes: notes, Recommendation: recommendation,
	}, nil
}

func assessmentResponses(records []repositories.TrainingAssessmentRecord) []dto.SessionAssessmentData {
	result := make([]dto.SessionAssessmentData, 0, len(records))
	for _, item := range records {
		result = append(result, dto.SessionAssessmentData{
			ID: item.ID, TrainingSessionID: item.TrainingSessionID,
			MaterialID: item.MaterialID, MaterialName: item.MaterialName,
			SubMaterialID: item.SubMaterialID, SubMaterialName: item.SubMaterialName,
			SkillStatus: item.SkillStatus, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	return result
}

func evaluationResponse(record *models.SessionEvaluation) dto.SessionEvaluationData {
	return dto.SessionEvaluationData{
		ID: record.ID, TrainingSessionID: record.TrainingSessionID,
		Predicate: record.Predicate, Notes: record.Notes, Recommendation: record.Recommendation,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}
