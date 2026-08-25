package services

import (
	"context"
	"errors"
	"time"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

type AvailabilityService struct {
	records *repositories.AvailabilityRepository
	now     func() time.Time
}

func NewAvailabilityService(records *repositories.AvailabilityRepository) *AvailabilityService {
	return &AvailabilityService{records: records, now: time.Now}
}

func (s *AvailabilityService) Create(ctx context.Context, trainerID int64, request dto.UpsertAvailabilityRequest) (*dto.AvailabilityData, error) {
	date, start, end, err := parseAvailability(request)
	if err != nil {
		return nil, err
	}
	item, err := s.records.Create(ctx, trainerID, date, start, end)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) ListTrainer(ctx context.Context, trainerID int64) ([]dto.AvailabilityData, error) {
	items, err := s.records.ListTrainer(ctx, trainerID)
	if err != nil {
		return nil, availabilityError(err)
	}
	return availabilityResponses(items), nil
}

func (s *AvailabilityService) GetTrainer(ctx context.Context, trainerID, id int64) (*dto.AvailabilityData, error) {
	item, err := s.records.GetTrainer(ctx, trainerID, id)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) UpdateTrainer(ctx context.Context, trainerID, id int64, request dto.UpsertAvailabilityRequest) (*dto.AvailabilityData, error) {
	date, start, end, err := parseAvailability(request)
	if err != nil {
		return nil, err
	}
	item, err := s.records.UpdateTrainer(ctx, trainerID, id, date, start, end)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) CancelTrainer(ctx context.Context, trainerID, id int64) (*dto.AvailabilityData, error) {
	item, err := s.records.CancelTrainer(ctx, trainerID, id)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) List(ctx context.Context) ([]dto.AvailabilityData, error) {
	items, err := s.records.List(ctx)
	if err != nil {
		return nil, availabilityError(err)
	}
	return availabilityResponses(items), nil
}

func (s *AvailabilityService) Get(ctx context.Context, id int64) (*dto.AvailabilityData, error) {
	item, err := s.records.Get(ctx, id)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) Publish(ctx context.Context, adminID, id int64) (*dto.AvailabilityData, error) {
	item, err := s.records.Publish(ctx, adminID, id, s.now().UTC())
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) Cancel(ctx context.Context, id int64) (*dto.AvailabilityData, error) {
	item, err := s.records.Cancel(ctx, id)
	if err != nil {
		return nil, availabilityError(err)
	}
	result := availabilityResponse(item)
	return &result, nil
}

func (s *AvailabilityService) StudentSchedules(ctx context.Context, filter string) ([]dto.StudentScheduleData, error) {
	var date *time.Time
	if filter != "" {
		parsed, err := time.Parse("2006-01-02", filter)
		if err != nil || parsed.Format("2006-01-02") != filter {
			return nil, ErrInvalidInput
		}
		date = &parsed
	}
	records, err := s.records.ListPublishedSchedules(ctx, date)
	if err != nil {
		return nil, availabilityError(err)
	}
	results := make([]dto.StudentScheduleData, 0, len(records))
	for i := range records {
		slots := scheduleSlots(&records[i])
		if len(slots) == 0 {
			continue
		}
		results = append(results, dto.StudentScheduleData{
			AvailabilityID: records[i].Availability.ID,
			Trainer:        dto.ScheduleTrainerData{ID: records[i].Trainer.ID, Name: records[i].Trainer.Name},
			AvailableDate:  records[i].Availability.AvailableDate.Format("2006-01-02"),
			Slots:          slots,
		})
	}
	return results, nil
}

func parseAvailability(request dto.UpsertAvailabilityRequest) (time.Time, time.Time, time.Time, error) {
	date, err := time.Parse("2006-01-02", request.AvailableDate)
	if err != nil || date.Format("2006-01-02") != request.AvailableDate ||
		date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return time.Time{}, time.Time{}, time.Time{}, ErrInvalidInput
	}
	start, startErr := time.Parse("15:04", request.StartTime)
	end, endErr := time.Parse("15:04", request.EndTime)
	if startErr != nil || endErr != nil || start.Format("15:04") != request.StartTime ||
		end.Format("15:04") != request.EndTime || start.Minute() != 0 || end.Minute() != 0 ||
		start.Hour() < 8 || end.Hour() > 17 || !start.Before(end) {
		return time.Time{}, time.Time{}, time.Time{}, ErrInvalidInput
	}
	return date, start, end, nil
}

func scheduleSlots(record *repositories.ScheduleRecord) []dto.ScheduleSlotData {
	result := make([]dto.ScheduleSlotData, 0)
	for hour := record.Availability.StartTime.Hour(); hour+2 <= record.Availability.EndTime.Hour(); hour++ {
		candidateStart, candidateEnd := hour*60, (hour+2)*60
		occupied := false
		for _, session := range record.Sessions {
			if session.Status == models.SessionCancelled || session.Status == models.SessionRescheduled {
				continue
			}
			sessionStart := session.StartTime.Hour()*60 + session.StartTime.Minute()
			sessionEnd := session.EndTime.Hour()*60 + session.EndTime.Minute()
			if candidateStart < sessionEnd && candidateEnd > sessionStart {
				occupied = true
				break
			}
		}
		if occupied {
			continue
		}
		start := time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC)
		result = append(result, dto.ScheduleSlotData{
			StartTime: start.Format("15:04"), EndTime: start.Add(2 * time.Hour).Format("15:04"),
		})
	}
	return result
}

func availabilityResponses(items []models.TrainerAvailability) []dto.AvailabilityData {
	result := make([]dto.AvailabilityData, 0, len(items))
	for i := range items {
		result = append(result, availabilityResponse(&items[i]))
	}
	return result
}

func availabilityResponse(item *models.TrainerAvailability) dto.AvailabilityData {
	return dto.AvailabilityData{
		ID: item.ID, TrainerID: item.TrainerID,
		AvailableDate: item.AvailableDate.Format("2006-01-02"),
		StartTime:     item.StartTime.Format("15:04"), EndTime: item.EndTime.Format("15:04"),
		Status: item.Status, PublishedBy: item.PublishedBy, PublishedAt: item.PublishedAt,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func availabilityError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrRecordNotFound):
		return ErrResourceNotFound
	case errors.Is(err, repositories.ErrAvailabilityOverlaps):
		return ErrAvailabilityOverlap
	case errors.Is(err, repositories.ErrInvalidAvailabilityState):
		return ErrInvalidState
	case errors.Is(err, repositories.ErrDuplicateRecord):
		return ErrResourceConflict
	default:
		return err
	}
}
