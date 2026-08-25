package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentSkillRecord struct {
	MaterialID        int64
	MaterialName      string
	SubMaterialID     int64
	SubMaterialName   string
	SkillStatus       models.SkillStatus
	TrainingSessionID *int64
	AssessedAt        *time.Time
}
type StudentSkillHistoryRecord struct {
	ID                int64
	TrainingSessionID int64
	SessionNumber     int
	EnrollmentID      int64
	TrainerID         int64
	MaterialID        int64
	MaterialName      string
	SubMaterialID     int64
	SubMaterialName   string
	SkillStatus       models.SkillStatus
	CompletedAt       time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
type TrainerReviewRecord struct {
	ID                int64
	TrainingSessionID int64
	TrainerID         int64
	TrainerName       string
	StudentID         int64
	StudentName       string
	Rating            int16
	Feedback          *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
type TrainerReviewSummaryRecord struct {
	TotalReviews  int64
	AverageRating float64
}
type CertificateRecord struct {
	Certificate models.Certificate
	Enrollment  models.StudentEnrollment
	Student     models.User
}

func (r *TrainingSessionRepository) StudentSkills(ctx context.Context, studentID int64) ([]StudentSkillRecord, error) {
	return studentSkillRecords(r.db.WithContext(ctx), studentID)
}
func studentSkillRecords(db *gorm.DB, studentID int64) ([]StudentSkillRecord, error) {
	result := make([]StudentSkillRecord, 0)
	query := `SELECT materials.id AS material_id, materials.name AS material_name,
 sub_materials.id AS sub_material_id, sub_materials.name AS sub_material_name,
 COALESCE(latest.skill_status, 'NOT_STARTED') AS skill_status,
 latest.training_session_id, latest.assessed_at
 FROM sub_materials JOIN materials ON materials.id = sub_materials.material_id
 LEFT JOIN LATERAL (
  SELECT assessments.skill_status, sessions.id AS training_session_id,
   sessions.actual_completed_at AS assessed_at
  FROM session_skill_assessments assessments
  JOIN training_sessions sessions ON sessions.id = assessments.training_session_id
  JOIN student_enrollments enrollments ON enrollments.id = sessions.enrollment_id
  WHERE assessments.sub_material_id = sub_materials.id AND enrollments.student_id = ?
    AND sessions.status = 'COMPLETED'
  ORDER BY sessions.actual_completed_at DESC, assessments.updated_at DESC, assessments.id DESC
  LIMIT 1
 ) latest ON TRUE
 WHERE materials.status = 'ACTIVE' AND sub_materials.status = 'ACTIVE'
 ORDER BY materials.sequence, sub_materials.sequence`
	err := db.Raw(query, studentID).Scan(&result).Error
	return result, err
}
func (r *TrainingSessionRepository) StudentSkillHistory(ctx context.Context, studentID int64) ([]StudentSkillHistoryRecord, error) {
	result := make([]StudentSkillHistoryRecord, 0)
	err := r.db.WithContext(ctx).Table("session_skill_assessments AS assessments").
		Select("assessments.id, assessments.training_session_id, sessions.session_number, sessions.enrollment_id, sessions.trainer_id, materials.id AS material_id, materials.name AS material_name, sub_materials.id AS sub_material_id, sub_materials.name AS sub_material_name, assessments.skill_status, sessions.actual_completed_at AS completed_at, assessments.created_at, assessments.updated_at").
		Joins("JOIN training_sessions sessions ON sessions.id = assessments.training_session_id").
		Joins("JOIN student_enrollments enrollments ON enrollments.id = sessions.enrollment_id").
		Joins("JOIN sub_materials ON sub_materials.id = assessments.sub_material_id").
		Joins("JOIN materials ON materials.id = sub_materials.material_id").
		Where("enrollments.student_id = ? AND sessions.status = ?", studentID, models.SessionCompleted).
		Order("sessions.actual_completed_at DESC, materials.sequence ASC, sub_materials.sequence ASC, assessments.id DESC").Scan(&result).Error
	return result, err
}
func CalculateSkillScore(skills []StudentSkillRecord) (int16, models.SkillLevel) {
	if len(skills) == 0 {
		return 0, models.SkillLevelBeginner
	}
	obtained := 0
	for _, skill := range skills {
		switch skill.SkillStatus {
		case models.SkillPracticed:
			obtained++
		case models.SkillMastered:
			obtained += 2
		}
	}
	maximum := len(skills) * 2
	score := int16((obtained*100 + maximum/2) / maximum)
	switch {
	case score >= 80:
		return score, models.SkillLevelProficient
	case score >= 60:
		return score, models.SkillLevelCapable
	case score >= 40:
		return score, models.SkillLevelDeveloping
	default:
		return score, models.SkillLevelBeginner
	}
}
func completeEnrollmentIfEligible(tx *gorm.DB, session *models.TrainingSession, now time.Time) error {
	var enrollment models.StudentEnrollment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&enrollment, session.EnrollmentID).Error; err != nil {
		return mapPhase5Error(err)
	}
	if enrollment.Status != models.EnrollmentActive {
		return ErrInvalidTrainingSessionState
	}
	var completed int64
	if err := tx.Model(&models.TrainingSession{}).Where("enrollment_id = ? AND status = ?", enrollment.ID, models.SessionCompleted).Count(&completed).Error; err != nil {
		return err
	}
	if completed < int64(enrollment.RequiredSessions()) {
		return nil
	}
	if completed > int64(enrollment.RequiredSessions()) {
		return ErrSessionCapacityExceeded
	}
	if err := tx.Model(&enrollment).Updates(map[string]any{"status": models.EnrollmentCompleted, "completed_at": now}).Error; err != nil {
		return mapPhase5Error(err)
	}
	skills, err := studentSkillRecords(tx, enrollment.StudentID)
	if err != nil {
		return err
	}
	score, level := CalculateSkillScore(skills)
	prefix := "CERT-" + now.Format("20060102") + "-"
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", prefix).Error; err != nil {
		return err
	}
	var existing int64
	if err := tx.Model(&models.Certificate{}).Where("certificate_number LIKE ?", prefix+"%").Count(&existing).Error; err != nil {
		return err
	}
	certificate := models.Certificate{EnrollmentID: enrollment.ID, CertificateNumber: fmt.Sprintf("%s%04d", prefix, existing+1), SkillScore: score, SkillLevel: level, IssuedAt: now}
	return mapPhase5Error(tx.Create(&certificate).Error)
}
func (r *TrainingSessionRepository) CreateStudentReview(ctx context.Context, studentID, sessionID int64, review models.TrainerReview) (*TrainerReviewRecord, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, _, err := lockOwnedTrainingSession(tx, sessionID, &studentID, false)
		if err != nil {
			return err
		}
		if session.Status != models.SessionCompleted {
			return ErrInvalidTrainingSessionState
		}
		review.TrainingSessionID = session.ID
		return mapPhase5Error(tx.Create(&review).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return r.StudentReview(ctx, studentID, sessionID)
}
func (r *TrainingSessionRepository) UpdateStudentReview(ctx context.Context, studentID, sessionID int64, review models.TrainerReview) (*TrainerReviewRecord, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, _, err := lockOwnedTrainingSession(tx, sessionID, &studentID, false)
		if err != nil {
			return err
		}
		if session.Status != models.SessionCompleted {
			return ErrInvalidTrainingSessionState
		}
		var existing models.TrainerReview
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("training_session_id = ?", session.ID).First(&existing).Error; err != nil {
			return mapPhase5Error(err)
		}
		return mapPhase5Error(tx.Model(&existing).Updates(map[string]any{"rating": review.Rating, "feedback": review.Feedback}).Error)
	})
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return r.StudentReview(ctx, studentID, sessionID)
}
func reviewQuery(db *gorm.DB) *gorm.DB {
	return db.Table("trainer_reviews").
		Select("trainer_reviews.*, sessions.trainer_id, trainer.name AS trainer_name, enrollments.student_id, student.name AS student_name").
		Joins("JOIN training_sessions sessions ON sessions.id = trainer_reviews.training_session_id").
		Joins("JOIN student_enrollments enrollments ON enrollments.id = sessions.enrollment_id").
		Joins("JOIN users trainer ON trainer.id = sessions.trainer_id").
		Joins("JOIN users student ON student.id = enrollments.student_id")
}
func (r *TrainingSessionRepository) StudentReview(ctx context.Context, studentID, sessionID int64) (*TrainerReviewRecord, error) {
	var record TrainerReviewRecord
	query := reviewQuery(r.db.WithContext(ctx)).Where("trainer_reviews.training_session_id = ? AND enrollments.student_id = ?", sessionID, studentID)
	if err := query.Take(&record).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &record, nil
}
func (r *TrainingSessionRepository) TrainerReviews(ctx context.Context, trainerID int64) ([]TrainerReviewRecord, error) {
	result := make([]TrainerReviewRecord, 0)
	err := reviewQuery(r.db.WithContext(ctx)).Where("sessions.trainer_id = ?", trainerID).Order("trainer_reviews.created_at DESC, trainer_reviews.id DESC").Scan(&result).Error
	return result, err
}
func (r *TrainingSessionRepository) TrainerReviewSummary(ctx context.Context, trainerID int64) (*TrainerReviewSummaryRecord, error) {
	var result TrainerReviewSummaryRecord
	err := r.db.WithContext(ctx).Table("trainer_reviews").
		Select("COUNT(trainer_reviews.id) AS total_reviews, COALESCE(ROUND(AVG(trainer_reviews.rating)::numeric, 2), 0) AS average_rating").
		Joins("JOIN training_sessions ON training_sessions.id = trainer_reviews.training_session_id").
		Where("training_sessions.trainer_id = ?", trainerID).Scan(&result).Error
	return &result, err
}
func (r *TrainingSessionRepository) AdminReviews(ctx context.Context) ([]TrainerReviewRecord, error) {
	result := make([]TrainerReviewRecord, 0)
	err := reviewQuery(r.db.WithContext(ctx)).Order("trainer_reviews.created_at DESC, trainer_reviews.id DESC").Scan(&result).Error
	return result, err
}
func (r *TrainingSessionRepository) AdminTrainerReviews(ctx context.Context, trainerID int64) ([]TrainerReviewRecord, error) {
	var trainer models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND role = ?", trainerID, models.RoleTrainer).First(&trainer).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return r.TrainerReviews(ctx, trainerID)
}
func (r *TrainingSessionRepository) StudentCertificates(ctx context.Context, studentID int64) ([]CertificateRecord, error) {
	return r.certificates(ctx, &studentID)
}
func (r *TrainingSessionRepository) AdminCertificates(ctx context.Context) ([]CertificateRecord, error) {
	return r.certificates(ctx, nil)
}
func (r *TrainingSessionRepository) StudentCertificate(ctx context.Context, studentID, id int64) (*CertificateRecord, error) {
	return r.certificate(ctx, id, &studentID)
}
func (r *TrainingSessionRepository) AdminCertificate(ctx context.Context, id int64) (*CertificateRecord, error) {
	return r.certificate(ctx, id, nil)
}
func (r *TrainingSessionRepository) certificates(ctx context.Context, studentID *int64) ([]CertificateRecord, error) {
	query := r.db.WithContext(ctx).Model(&models.Certificate{}).Joins("JOIN student_enrollments ON student_enrollments.id = certificates.enrollment_id")
	if studentID != nil {
		query = query.Where("student_enrollments.student_id = ?", *studentID)
	}
	values := make([]models.Certificate, 0)
	if err := query.Order("certificates.issued_at DESC, certificates.id DESC").Find(&values).Error; err != nil {
		return nil, err
	}
	result := make([]CertificateRecord, 0, len(values))
	for i := range values {
		record, err := r.certificateRecord(ctx, &values[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *record)
	}
	return result, nil
}
func (r *TrainingSessionRepository) certificate(ctx context.Context, id int64, studentID *int64) (*CertificateRecord, error) {
	query := r.db.WithContext(ctx).Model(&models.Certificate{}).Joins("JOIN student_enrollments ON student_enrollments.id = certificates.enrollment_id").Where("certificates.id = ?", id)
	if studentID != nil {
		query = query.Where("student_enrollments.student_id = ?", *studentID)
	}
	var certificate models.Certificate
	if err := query.First(&certificate).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return r.certificateRecord(ctx, &certificate)
}
func (r *TrainingSessionRepository) certificateRecord(ctx context.Context, certificate *models.Certificate) (*CertificateRecord, error) {
	result := &CertificateRecord{Certificate: *certificate}
	if err := r.db.WithContext(ctx).First(&result.Enrollment, certificate.EnrollmentID).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	if err := r.db.WithContext(ctx).First(&result.Student, result.Enrollment.StudentID).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return result, nil
}
