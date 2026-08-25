package dto

type InternalStatsData struct {
	TotalStudents              int64 `json:"total_students" example:"120"`
	TotalTrainers              int64 `json:"total_trainers" example:"12"`
	TotalAdmins                int64 `json:"total_admins" example:"2"`
	TotalEnrollments           int64 `json:"total_enrollments" example:"150"`
	ActiveEnrollments          int64 `json:"active_enrollments" example:"34"`
	TotalTrainingSessions      int64 `json:"total_training_sessions" example:"450"`
	ScheduledTrainingSessions  int64 `json:"scheduled_training_sessions" example:"20"`
	InProgressTrainingSessions int64 `json:"in_progress_training_sessions" example:"3"`
	CompletedTrainingSessions  int64 `json:"completed_training_sessions" example:"410"`
	PaidPayments               int64 `json:"paid_payments" example:"140"`
	TotalCertificates          int64 `json:"total_certificates" example:"88"`
	TotalTrainerReviews        int64 `json:"total_trainer_reviews" example:"305"`
}

type InternalStatsAPIResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"success"`
	Data    InternalStatsData `json:"data"`
}
