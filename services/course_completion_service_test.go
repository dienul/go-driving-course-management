package services

import (
	"errors"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"strings"
	"testing"
)

func TestGlobalSkillScoreThresholds(t *testing.T) {
	cases := []struct {
		mastered int
		score    int16
		level    models.SkillLevel
	}{
		{0, 0, models.SkillLevelBeginner}, {39, 39, models.SkillLevelBeginner},
		{40, 40, models.SkillLevelDeveloping}, {59, 59, models.SkillLevelDeveloping},
		{60, 60, models.SkillLevelCapable}, {79, 79, models.SkillLevelCapable},
		{80, 80, models.SkillLevelProficient}, {100, 100, models.SkillLevelProficient},
	}
	for _, item := range cases {
		records := make([]repositories.StudentSkillRecord, 100)
		for i := range records {
			records[i].SkillStatus = models.SkillNotStarted
			if i < item.mastered {
				records[i].SkillStatus = models.SkillMastered
			}
		}
		score, level := repositories.CalculateSkillScore(records)
		if score != item.score || level != item.level {
			t.Errorf("mastered=%d: score=%d level=%s, want %d %s", item.mastered, score, level, item.score, item.level)
		}
	}
	if score, level := repositories.CalculateSkillScore(nil); score != 0 || level != models.SkillLevelBeginner {
		t.Errorf("empty curriculum: %d %s", score, level)
	}
	rounded := []repositories.StudentSkillRecord{{SkillStatus: models.SkillPracticed}, {SkillStatus: models.SkillNotStarted}, {SkillStatus: models.SkillNotStarted}}
	if score, _ := repositories.CalculateSkillScore(rounded); score != 17 {
		t.Errorf("fractional score was not rounded: %d", score)
	}
}
func TestTrainerReviewValidation(t *testing.T) {
	feedback := "  Patient and clear  "
	review, err := validateTrainerReview(dto.UpsertTrainerReviewRequest{Rating: 5, Feedback: &feedback})
	if err != nil || review.Feedback == nil || *review.Feedback != "Patient and clear" {
		t.Fatalf("valid review rejected: %+v %v", review, err)
	}
	blank := "   "
	review, err = validateTrainerReview(dto.UpsertTrainerReviewRequest{Rating: 1, Feedback: &blank})
	if err != nil || review.Feedback != nil {
		t.Errorf("blank optional feedback was not normalized: %+v %v", review, err)
	}
	lengthy := strings.Repeat("a", 2001)
	for _, item := range []dto.UpsertTrainerReviewRequest{{Rating: 0}, {Rating: 6}, {Rating: 5, Feedback: &lengthy}} {
		if _, err := validateTrainerReview(item); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("invalid review accepted: %+v %v", item, err)
		}
	}
}
