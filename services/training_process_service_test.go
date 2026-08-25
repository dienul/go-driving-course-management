package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
)

func TestSessionAssessmentValidation(t *testing.T) {
	for _, status := range []models.SkillStatus{models.SkillNotStarted, models.SkillPracticed, models.SkillMastered} {
		request := dto.UpsertSessionAssessmentsRequest{
			Assessments: []dto.SessionAssessmentRequest{{SubMaterialID: 1, SkillStatus: status}},
		}
		result, err := validateSessionAssessments(request)
		if err != nil || len(result) != 1 || result[0].SkillStatus != status {
			t.Errorf("valid assessment status %q rejected: result=%+v err=%v", status, result, err)
		}
	}
	cases := []dto.UpsertSessionAssessmentsRequest{
		{},
		{Assessments: []dto.SessionAssessmentRequest{{SubMaterialID: 0, SkillStatus: models.SkillPracticed}}},
		{Assessments: []dto.SessionAssessmentRequest{{SubMaterialID: 1, SkillStatus: models.SkillStatus("INVALID")}}},
		{Assessments: []dto.SessionAssessmentRequest{
			{SubMaterialID: 1, SkillStatus: models.SkillPracticed},
			{SubMaterialID: 1, SkillStatus: models.SkillMastered},
		}},
	}
	tooMany := dto.UpsertSessionAssessmentsRequest{Assessments: make([]dto.SessionAssessmentRequest, 101)}
	for index := range tooMany.Assessments {
		tooMany.Assessments[index] = dto.SessionAssessmentRequest{
			SubMaterialID: int64(index + 1), SkillStatus: models.SkillPracticed,
		}
	}
	cases = append(cases, tooMany)
	for index, request := range cases {
		if _, err := validateSessionAssessments(request); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("invalid assessment case %d accepted: %v", index, err)
		}
	}
}

func TestSessionEvaluationValidation(t *testing.T) {
	predicates := []models.EvaluationPredicate{
		models.PredicateKurang, models.PredicateCukup, models.PredicateBaik, models.PredicateSangatBaik,
	}
	for _, predicate := range predicates {
		result, err := validateSessionEvaluation(dto.UpsertSessionEvaluationRequest{
			Predicate: predicate, Notes: "  Good progress  ", Recommendation: "  Practice parking  ",
		})
		if err != nil || result.Predicate != predicate || result.Notes != "Good progress" ||
			result.Recommendation != "Practice parking" {
			t.Errorf("valid evaluation predicate %q rejected: result=%+v err=%v", predicate, result, err)
		}
	}
	cases := []dto.UpsertSessionEvaluationRequest{
		{Predicate: models.EvaluationPredicate("INVALID"), Notes: "Notes", Recommendation: "Practice"},
		{Predicate: models.PredicateBaik, Notes: "   ", Recommendation: "Practice"},
		{Predicate: models.PredicateBaik, Notes: "Notes", Recommendation: "  "},
		{Predicate: models.PredicateBaik, Notes: strings.Repeat("a", 5001), Recommendation: "Practice"},
		{Predicate: models.PredicateBaik, Notes: "Notes", Recommendation: strings.Repeat("a", 5001)},
	}
	for index, request := range cases {
		if _, err := validateSessionEvaluation(request); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("invalid evaluation case %d accepted: %v", index, err)
		}
	}
}
