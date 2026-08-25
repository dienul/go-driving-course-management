package services

import (
	"errors"
	"testing"
	"time"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

func TestAvailabilityOperationalValidation(t *testing.T) {
	valid := dto.UpsertAvailabilityRequest{AvailableDate: "2026-08-31", StartTime: "08:00", EndTime: "17:00"}
	if _, _, _, err := parseAvailability(valid); err != nil {
		t.Fatalf("valid Monday range rejected: %v", err)
	}
	invalid := []dto.UpsertAvailabilityRequest{
		{AvailableDate: "2026-08-29", StartTime: "08:00", EndTime: "12:00"},
		{AvailableDate: "2026-08-30", StartTime: "08:00", EndTime: "12:00"},
		{AvailableDate: "2026-08-31", StartTime: "07:00", EndTime: "12:00"},
		{AvailableDate: "2026-08-31", StartTime: "08:00", EndTime: "18:00"},
		{AvailableDate: "2026-08-31", StartTime: "08:30", EndTime: "12:00"},
		{AvailableDate: "2026-08-31", StartTime: "08:00", EndTime: "12:30"},
		{AvailableDate: "2026-08-31", StartTime: "12:00", EndTime: "12:00"},
		{AvailableDate: "2026-08-31", StartTime: "13:00", EndTime: "12:00"},
		{AvailableDate: "31-08-2026", StartTime: "08:00", EndTime: "12:00"},
	}
	for index, request := range invalid {
		if _, _, _, err := parseAvailability(request); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("invalid availability %d was not rejected: err=%v", index, err)
		}
	}
}

func TestStudentTwoHourScheduleSlots(t *testing.T) {
	at := func(hour int) time.Time { return time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC) }
	record := repositories.ScheduleRecord{
		Availability: models.TrainerAvailability{StartTime: models.NewClockTime(at(8)), EndTime: models.NewClockTime(at(13))},
	}
	slots := scheduleSlots(&record)
	if len(slots) != 4 || slots[0].StartTime != "08:00" || slots[3].EndTime != "13:00" {
		t.Fatalf("incorrect full-hour two-hour slots: %+v", slots)
	}
	record.Sessions = []models.TrainingSession{
		{StartTime: models.NewClockTime(at(9)), EndTime: models.NewClockTime(at(11)), Status: models.SessionScheduled},
	}
	slots = scheduleSlots(&record)
	if len(slots) != 1 || slots[0].StartTime != "11:00" || slots[0].EndTime != "13:00" {
		t.Fatalf("occupied trainer slots were not excluded: %+v", slots)
	}
	record.Sessions[0].Status = models.SessionCancelled
	if got := len(scheduleSlots(&record)); got != 4 {
		t.Fatalf("cancelled sessions blocked slots: %d", got)
	}
}

func TestAvailabilityBusinessErrorMapping(t *testing.T) {
	cases := []struct{ source, expected error }{
		{repositories.ErrRecordNotFound, ErrResourceNotFound},
		{repositories.ErrAvailabilityOverlaps, ErrAvailabilityOverlap},
		{repositories.ErrInvalidAvailabilityState, ErrInvalidState},
		{repositories.ErrDuplicateRecord, ErrResourceConflict},
	}
	for _, item := range cases {
		if !errors.Is(availabilityError(item.source), item.expected) {
			t.Errorf("error %v mapped to %v, want %v", item.source, availabilityError(item.source), item.expected)
		}
	}
}
