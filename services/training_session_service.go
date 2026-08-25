package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

type TrainingSessionService struct {
	records *repositories.TrainingSessionRepository
	now     func() time.Time
}

func NewTrainingSessionService(records *repositories.TrainingSessionRepository) *TrainingSessionService {
	return &TrainingSessionService{records: records, now: time.Now}
}

func (s *TrainingSessionService) Book(ctx context.Context, studentID int64, request dto.BookTrainingSessionRequest) (*dto.TrainingSessionData, error) {
	start, err := parseSessionStart(request.StartTime)
	if err != nil || request.TrainerAvailabilityID <= 0 {
		return nil, ErrInvalidInput
	}
	session, err := s.records.Book(ctx, studentID, request.TrainerAvailabilityID, start)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) ListStudent(ctx context.Context, studentID int64) ([]dto.TrainingSessionData, error) {
	sessions, err := s.records.ListStudent(ctx, studentID)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return trainingSessionResponses(sessions), nil
}

func (s *TrainingSessionService) GetStudent(ctx context.Context, studentID, id int64) (*dto.TrainingSessionData, error) {
	session, err := s.records.GetStudent(ctx, studentID, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) CancelStudent(ctx context.Context, studentID, id int64, request dto.CancelTrainingSessionRequest) (*dto.TrainingSessionData, error) {
	reason := strings.TrimSpace(request.CancellationReason)
	if len(reason) < 2 {
		return nil, ErrInvalidInput
	}
	session, err := s.records.CancelStudent(ctx, studentID, id, reason, s.now().UTC())
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) RescheduleStudent(ctx context.Context, studentID, id int64, request dto.RescheduleTrainingSessionRequest) (*dto.TrainingSessionData, error) {
	start, err := parseSessionStart(request.StartTime)
	if err != nil || request.TrainerAvailabilityID <= 0 {
		return nil, ErrInvalidInput
	}
	session, err := s.records.RescheduleStudent(ctx, studentID, id, request.TrainerAvailabilityID, start)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) List(ctx context.Context) ([]dto.TrainingSessionData, error) {
	sessions, err := s.records.List(ctx)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	return trainingSessionResponses(sessions), nil
}

func (s *TrainingSessionService) Get(ctx context.Context, id int64) (*dto.TrainingSessionData, error) {
	session, err := s.records.Get(ctx, id)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) Cancel(ctx context.Context, adminID, id int64, request dto.CancelTrainingSessionRequest) (*dto.TrainingSessionData, error) {
	reason := strings.TrimSpace(request.CancellationReason)
	if len(reason) < 2 {
		return nil, ErrInvalidInput
	}
	session, err := s.records.Cancel(ctx, adminID, id, reason, s.now().UTC())
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func (s *TrainingSessionService) Reschedule(ctx context.Context, id int64, request dto.RescheduleTrainingSessionRequest) (*dto.TrainingSessionData, error) {
	start, err := parseSessionStart(request.StartTime)
	if err != nil || request.TrainerAvailabilityID <= 0 {
		return nil, ErrInvalidInput
	}
	session, err := s.records.Reschedule(ctx, id, request.TrainerAvailabilityID, start)
	if err != nil {
		return nil, trainingSessionError(err)
	}
	result := trainingSessionResponse(session)
	return &result, nil
}

func parseSessionStart(input string) (time.Time, error) {
	start, err := time.Parse("15:04", input)
	if err != nil || start.Format("15:04") != input || start.Minute() != 0 ||
		start.Hour() < 8 || start.Hour()+2 > 17 {
		return time.Time{}, ErrInvalidInput
	}
	return start, nil
}

func trainingSessionResponses(sessions []models.TrainingSession) []dto.TrainingSessionData {
	result := make([]dto.TrainingSessionData, 0, len(sessions))
	for i := range sessions {
		result = append(result, trainingSessionResponse(&sessions[i]))
	}
	return result
}

func trainingSessionResponse(session *models.TrainingSession) dto.TrainingSessionData {
	return dto.TrainingSessionData{
		ID: session.ID, EnrollmentID: session.EnrollmentID, TrainerID: session.TrainerID,
		TrainerAvailabilityID: session.TrainerAvailabilityID, SessionNumber: session.SessionNumber,
		ScheduledDate: session.ScheduledDate.Format("2006-01-02"),
		StartTime:     session.StartTime.Format("15:04"), EndTime: session.EndTime.Format("15:04"),
		Status: session.Status, ActualStartedAt: session.ActualStartedAt, ActualCompletedAt: session.ActualCompletedAt,
		RescheduledFromID: session.RescheduledFromID, CancelledBy: session.CancelledBy,
		CancellationReason: session.CancellationReason, CancelledAt: session.CancelledAt,
		CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt,
	}
}

func trainingSessionError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrRecordNotFound):
		return ErrResourceNotFound
	case errors.Is(err, repositories.ErrDuplicateRecord):
		return ErrResourceConflict
	case errors.Is(err, repositories.ErrNoActiveStudentEnrollment):
		return ErrNoActiveEnrollment
	case errors.Is(err, repositories.ErrTrainerSessionOverlaps):
		return ErrTrainerSlotConflict
	case errors.Is(err, repositories.ErrSessionCapacityExceeded):
		return ErrNoRemainingSessions
	case errors.Is(err, repositories.ErrStudentSessionOpen):
		return ErrSessionAlreadyOpen
	case errors.Is(err, repositories.ErrUnavailableTrainerAvailability):
		return ErrAvailabilityUnavailable
	case errors.Is(err, repositories.ErrInvalidTrainingSessionState):
		return ErrInvalidState
	case errors.Is(err, repositories.ErrInvalidTrainingSlot):
		return ErrInvalidInput
	case errors.Is(err, repositories.ErrSessionAssessmentRequired):
		return ErrAssessmentRequired
	case errors.Is(err, repositories.ErrSessionEvaluationRequired):
		return ErrEvaluationRequired
	default:
		return err
	}
}
