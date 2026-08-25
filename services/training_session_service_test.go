package services

import (
	"errors"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/repositories"
)

func TestTrainingSessionStartValidation(t *testing.T) {
	for _, value := range []string{"08:00", "09:00", "12:00", "15:00"} {
		start, err := parseSessionStart(value)
		if err != nil || start.Format("15:04") != value {
			t.Errorf("valid session start %q rejected: %v", value, err)
		}
	}
	for _, value := range []string{"07:00", "08:30", "15:30", "16:00", "17:00", "8:00", "invalid"} {
		if _, err := parseSessionStart(value); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("invalid session start %q accepted: %v", value, err)
		}
	}
}

func TestTrainingSessionBusinessErrorMapping(t *testing.T) {
	cases := []struct{ source, expected error }{
		{repositories.ErrRecordNotFound, ErrResourceNotFound},
		{repositories.ErrDuplicateRecord, ErrResourceConflict},
		{repositories.ErrNoActiveStudentEnrollment, ErrNoActiveEnrollment},
		{repositories.ErrTrainerSessionOverlaps, ErrTrainerSlotConflict},
		{repositories.ErrSessionCapacityExceeded, ErrNoRemainingSessions},
		{repositories.ErrStudentSessionOpen, ErrSessionAlreadyOpen},
		{repositories.ErrUnavailableTrainerAvailability, ErrAvailabilityUnavailable},
		{repositories.ErrInvalidTrainingSessionState, ErrInvalidState},
		{repositories.ErrInvalidTrainingSlot, ErrInvalidInput},
		{repositories.ErrSessionAssessmentRequired, ErrAssessmentRequired},
		{repositories.ErrSessionEvaluationRequired, ErrEvaluationRequired},
	}
	for _, item := range cases {
		if !errors.Is(trainingSessionError(item.source), item.expected) {
			t.Errorf("error %v mapped to %v, want %v", item.source, trainingSessionError(item.source), item.expected)
		}
	}
}
