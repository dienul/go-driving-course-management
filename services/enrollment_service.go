package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

type EnrollmentService struct {
	records *repositories.EnrollmentRepository
	now     func() time.Time
}

func NewEnrollmentService(records *repositories.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{records: records, now: time.Now}
}

func (s *EnrollmentService) ListPackages(ctx context.Context) ([]models.CoursePackage, error) {
	results, err := s.records.ListActivePackages(ctx)
	return results, enrollmentError(err)
}

func (s *EnrollmentService) GetPackage(ctx context.Context, id int64) (*models.CoursePackage, error) {
	result, err := s.records.GetActivePackage(ctx, id)
	return result, enrollmentError(err)
}

func (s *EnrollmentService) CreateEnrollment(ctx context.Context, studentID int64, request dto.CreateEnrollmentRequest) (*dto.EnrollmentCheckoutData, error) {
	if studentID <= 0 || request.PackageID <= 0 {
		return nil, ErrInvalidInput
	}
	code, err := generatePaymentCode(s.now())
	if err != nil {
		return nil, err
	}
	enrollment, payment, err := s.records.CreateEnrollment(ctx, studentID, request.PackageID, code)
	if err != nil {
		return nil, enrollmentError(err)
	}
	return &dto.EnrollmentCheckoutData{Enrollment: *enrollment, Payment: *payment}, nil
}

func (s *EnrollmentService) ListStudentEnrollments(ctx context.Context, studentID int64) ([]models.StudentEnrollment, error) {
	results, err := s.records.ListStudentEnrollments(ctx, studentID)
	return results, enrollmentError(err)
}

func (s *EnrollmentService) GetStudentEnrollment(ctx context.Context, studentID, id int64) (*models.StudentEnrollment, error) {
	result, err := s.records.GetStudentEnrollment(ctx, studentID, id)
	return result, enrollmentError(err)
}

func (s *EnrollmentService) GetEnrollmentPayment(ctx context.Context, studentID, id int64) (*models.Payment, error) {
	result, err := s.records.GetEnrollmentPayment(ctx, studentID, id)
	return result, enrollmentError(err)
}

func (s *EnrollmentService) Pay(ctx context.Context, studentID, paymentID int64, request dto.PayRequest) (*dto.PaymentCheckoutData, error) {
	if !validPaymentMethod(request.PaymentMethod) {
		return nil, ErrInvalidInput
	}
	enrollment, payment, invoice, err := s.records.Pay(ctx, studentID, paymentID, request.PaymentMethod, s.now().UTC())
	if err != nil {
		return nil, enrollmentError(err)
	}
	record, err := s.records.GetStudentInvoice(ctx, studentID, invoice.ID)
	if err != nil {
		return nil, enrollmentError(err)
	}
	return &dto.PaymentCheckoutData{
		Enrollment: *enrollment, Payment: *payment, Invoice: invoiceResponse(record),
	}, nil
}

func (s *EnrollmentService) ListStudentInvoices(ctx context.Context, studentID int64) ([]dto.InvoiceData, error) {
	records, err := s.records.ListStudentInvoices(ctx, studentID)
	if err != nil {
		return nil, enrollmentError(err)
	}
	return invoiceResponses(records), nil
}

func (s *EnrollmentService) GetStudentInvoice(ctx context.Context, studentID, id int64) (*dto.InvoiceData, error) {
	record, err := s.records.GetStudentInvoice(ctx, studentID, id)
	if err != nil {
		return nil, enrollmentError(err)
	}
	result := invoiceResponse(record)
	return &result, nil
}

func (s *EnrollmentService) ListEnrollments(ctx context.Context) ([]models.StudentEnrollment, error) {
	results, err := s.records.ListEnrollments(ctx)
	return results, enrollmentError(err)
}

func (s *EnrollmentService) GetEnrollment(ctx context.Context, id int64) (*models.StudentEnrollment, error) {
	result, err := s.records.GetEnrollment(ctx, id)
	return result, enrollmentError(err)
}

func (s *EnrollmentService) ListPayments(ctx context.Context) ([]models.Payment, error) {
	results, err := s.records.ListPayments(ctx)
	return results, enrollmentError(err)
}

func (s *EnrollmentService) GetPayment(ctx context.Context, id int64) (*models.Payment, error) {
	result, err := s.records.GetPayment(ctx, id)
	return result, enrollmentError(err)
}

func (s *EnrollmentService) ListInvoices(ctx context.Context) ([]dto.InvoiceData, error) {
	records, err := s.records.ListInvoices(ctx)
	if err != nil {
		return nil, enrollmentError(err)
	}
	return invoiceResponses(records), nil
}

func (s *EnrollmentService) GetInvoice(ctx context.Context, id int64) (*dto.InvoiceData, error) {
	record, err := s.records.GetInvoice(ctx, id)
	if err != nil {
		return nil, enrollmentError(err)
	}
	result := invoiceResponse(record)
	return &result, nil
}

func generatePaymentCode(now time.Time) (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return fmt.Sprintf("PAY-%s-%s", now.UTC().Format("20060102"), strings.ToUpper(hex.EncodeToString(random))), nil
}

func validPaymentMethod(method models.PaymentMethod) bool {
	return method == models.PaymentMethodBankTransfer || method == models.PaymentMethodCash
}

func enrollmentError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrRecordNotFound):
		return ErrResourceNotFound
	case errors.Is(err, repositories.ErrDuplicateRecord):
		return ErrResourceConflict
	case errors.Is(err, repositories.ErrActiveEnrollmentExists):
		return ErrActiveEnrollment
	case errors.Is(err, repositories.ErrPaymentAlreadyProcessed):
		return ErrPaymentProcessed
	case errors.Is(err, repositories.ErrInvalidFinancialState):
		return ErrInvalidState
	default:
		return err
	}
}

func invoiceResponses(records []repositories.InvoiceRecord) []dto.InvoiceData {
	results := make([]dto.InvoiceData, 0, len(records))
	for i := range records {
		results = append(results, invoiceResponse(&records[i]))
	}
	return results
}

func invoiceResponse(record *repositories.InvoiceRecord) dto.InvoiceData {
	invoice := record.Invoice
	return dto.InvoiceData{
		ID: invoice.ID, PaymentID: invoice.PaymentID, InvoiceNumber: invoice.InvoiceNumber,
		Student: userResponse(&record.Student), Enrollment: record.Enrollment, Payment: record.Payment,
		IssuedAt: invoice.IssuedAt, CreatedAt: invoice.CreatedAt, UpdatedAt: invoice.UpdatedAt,
	}
}
