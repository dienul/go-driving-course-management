package models

type UserRole string

const (
	RoleAdmin   UserRole = "ADMIN"
	RoleTrainer UserRole = "TRAINER"
	RoleStudent UserRole = "STUDENT"
)

type RecordStatus string

const (
	StatusActive   RecordStatus = "ACTIVE"
	StatusInactive RecordStatus = "INACTIVE"
)

type PackageLevel string

const (
	PackageLevelPemula PackageLevel = "PEMULA"
	PackageLevelDasar  PackageLevel = "DASAR"
)

type EnrollmentStatus string

const (
	EnrollmentPendingPayment EnrollmentStatus = "PENDING_PAYMENT"
	EnrollmentActive         EnrollmentStatus = "ACTIVE"
	EnrollmentCompleted      EnrollmentStatus = "COMPLETED"
	EnrollmentCancelled      EnrollmentStatus = "CANCELLED"
)

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCash         PaymentMethod = "CASH"
)

type PaymentStatus string

const (
	PaymentUnpaid PaymentStatus = "UNPAID"
	PaymentPaid   PaymentStatus = "PAID"
)

type AvailabilityStatus string

const (
	AvailabilityPending   AvailabilityStatus = "PENDING"
	AvailabilityPublished AvailabilityStatus = "PUBLISHED"
	AvailabilityCancelled AvailabilityStatus = "CANCELLED"
)

type TrainingSessionStatus string

const (
	SessionScheduled   TrainingSessionStatus = "SCHEDULED"
	SessionInProgress  TrainingSessionStatus = "IN_PROGRESS"
	SessionCompleted   TrainingSessionStatus = "COMPLETED"
	SessionCancelled   TrainingSessionStatus = "CANCELLED"
	SessionRescheduled TrainingSessionStatus = "RESCHEDULED"
)

type SkillStatus string

const (
	SkillNotStarted SkillStatus = "NOT_STARTED"
	SkillPracticed  SkillStatus = "PRACTICED"
	SkillMastered   SkillStatus = "MASTERED"
)

type EvaluationPredicate string

const (
	PredicateKurang     EvaluationPredicate = "KURANG"
	PredicateCukup      EvaluationPredicate = "CUKUP"
	PredicateBaik       EvaluationPredicate = "BAIK"
	PredicateSangatBaik EvaluationPredicate = "SANGAT_BAIK"
)

type SkillLevel string

const (
	SkillLevelBeginner   SkillLevel = "BEGINNER"
	SkillLevelDeveloping SkillLevel = "DEVELOPING"
	SkillLevelCapable    SkillLevel = "CAPABLE"
	SkillLevelProficient SkillLevel = "PROFICIENT"
)
