package services

import "errors"

var (
	ErrEmailExists             = errors.New("email already registered")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrInactiveUser            = errors.New("user account is inactive")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidToken            = errors.New("invalid token")
	ErrResourceNotFound        = errors.New("resource not found")
	ErrResourceConflict        = errors.New("resource already exists")
	ErrActiveEnrollment        = errors.New("student already has an active enrollment")
	ErrPaymentProcessed        = errors.New("payment has already been processed")
	ErrInvalidState            = errors.New("resource is not in the required state")
	ErrAvailabilityOverlap     = errors.New("availability overlaps an existing trainer schedule")
	ErrNoActiveEnrollment      = errors.New("student does not have an active enrollment")
	ErrTrainerSlotConflict     = errors.New("trainer already has an overlapping training session")
	ErrNoRemainingSessions     = errors.New("no purchased training sessions remain")
	ErrSessionAlreadyOpen      = errors.New("student already has an unfinished training session")
	ErrAvailabilityUnavailable = errors.New("trainer availability is not published or no longer available")
	ErrAssessmentRequired      = errors.New("at least one skill assessment is required before completion")
	ErrEvaluationRequired      = errors.New("a complete session evaluation is required before completion")
)
