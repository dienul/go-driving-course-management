package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrActiveEnrollmentExists  = errors.New("active enrollment already exists")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
	ErrInvalidFinancialState   = errors.New("invalid financial state")
)

type InvoiceRecord struct {
	Invoice    models.Invoice
	Payment    models.Payment
	Enrollment models.StudentEnrollment
	Student    models.User
}

type EnrollmentRepository struct{ db *gorm.DB }

func NewEnrollmentRepository(db *gorm.DB) *EnrollmentRepository { return &EnrollmentRepository{db: db} }

func (r *EnrollmentRepository) ListActivePackages(ctx context.Context) ([]models.CoursePackage, error) {
	results := make([]models.CoursePackage, 0)
	err := r.db.WithContext(ctx).Where("status = ?", models.StatusActive).Order("id ASC").Find(&results).Error
	return results, err
}

func (r *EnrollmentRepository) GetActivePackage(ctx context.Context, id int64) (*models.CoursePackage, error) {
	var result models.CoursePackage
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, models.StatusActive).First(&result).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &result, nil
}

func (r *EnrollmentRepository) CreateEnrollment(ctx context.Context, studentID, packageID int64, paymentCode string) (*models.StudentEnrollment, *models.Payment, error) {
	var enrollment models.StudentEnrollment
	var payment models.Payment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var student models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND role = ? AND status = ?", studentID, models.RoleStudent, models.StatusActive).
			First(&student).Error; err != nil {
			return mapPhase5Error(err)
		}

		var activeCount int64
		if err := tx.Model(&models.StudentEnrollment{}).
			Where("student_id = ? AND status = ?", studentID, models.EnrollmentActive).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrActiveEnrollmentExists
		}

		var pkg models.CoursePackage
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", packageID, models.StatusActive).
			First(&pkg).Error; err != nil {
			return mapPhase5Error(err)
		}

		enrollment = models.StudentEnrollment{
			StudentID: studentID, PackageID: pkg.ID, PackageName: pkg.Name,
			PackagePrice: pkg.Price, TotalHours: pkg.TotalHours, Status: models.EnrollmentPendingPayment,
		}
		if err := tx.Create(&enrollment).Error; err != nil {
			return mapPhase5Error(err)
		}
		payment = models.Payment{
			EnrollmentID: enrollment.ID, PaymentCode: paymentCode,
			Amount: enrollment.PackagePrice, Status: models.PaymentUnpaid,
		}
		return mapPhase5Error(tx.Create(&payment).Error)
	})
	if err != nil {
		return nil, nil, mapPhase5Error(err)
	}
	return &enrollment, &payment, nil
}

func (r *EnrollmentRepository) ListStudentEnrollments(ctx context.Context, studentID int64) ([]models.StudentEnrollment, error) {
	results := make([]models.StudentEnrollment, 0)
	err := r.db.WithContext(ctx).Where("student_id = ?", studentID).Order("id DESC").Find(&results).Error
	return results, err
}

func (r *EnrollmentRepository) GetStudentEnrollment(ctx context.Context, studentID, id int64) (*models.StudentEnrollment, error) {
	var result models.StudentEnrollment
	err := r.db.WithContext(ctx).Where("id = ? AND student_id = ?", id, studentID).First(&result).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &result, nil
}

func (r *EnrollmentRepository) ListEnrollments(ctx context.Context) ([]models.StudentEnrollment, error) {
	results := make([]models.StudentEnrollment, 0)
	err := r.db.WithContext(ctx).Order("id DESC").Find(&results).Error
	return results, err
}

func (r *EnrollmentRepository) GetEnrollment(ctx context.Context, id int64) (*models.StudentEnrollment, error) {
	var result models.StudentEnrollment
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &result, nil
}

func (r *EnrollmentRepository) GetEnrollmentPayment(ctx context.Context, studentID, enrollmentID int64) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Joins("JOIN student_enrollments ON student_enrollments.id = payments.enrollment_id").
		Where("payments.enrollment_id = ? AND student_enrollments.student_id = ?", enrollmentID, studentID).
		First(&payment).Error
	if err != nil {
		return nil, mapPhase5Error(err)
	}
	return &payment, nil
}

func (r *EnrollmentRepository) Pay(ctx context.Context, studentID, paymentID int64, method models.PaymentMethod, now time.Time) (*models.StudentEnrollment, *models.Payment, *models.Invoice, error) {
	var enrollment models.StudentEnrollment
	var payment models.Payment
	var invoice models.Invoice
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, paymentID).Error; err != nil {
			return mapPhase5Error(err)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND student_id = ?", payment.EnrollmentID, studentID).
			First(&enrollment).Error; err != nil {
			return mapPhase5Error(err)
		}
		if payment.Status != models.PaymentUnpaid {
			return ErrPaymentAlreadyProcessed
		}
		if enrollment.Status != models.EnrollmentPendingPayment || payment.Amount != enrollment.PackagePrice {
			return ErrInvalidFinancialState
		}

		var student models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND role = ? AND status = ?", studentID, models.RoleStudent, models.StatusActive).
			First(&student).Error; err != nil {
			return mapPhase5Error(err)
		}
		var activeCount int64
		if err := tx.Model(&models.StudentEnrollment{}).
			Where("student_id = ? AND status = ? AND id <> ?", studentID, models.EnrollmentActive, enrollment.ID).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrActiveEnrollmentExists
		}

		payment.Status, payment.PaymentMethod, payment.PaidAt = models.PaymentPaid, &method, &now
		if err := tx.Model(&payment).Updates(map[string]any{
			"status": payment.Status, "payment_method": payment.PaymentMethod, "paid_at": payment.PaidAt,
		}).Error; err != nil {
			return mapPhase5Error(err)
		}

		prefix := "INV-" + now.Format("20060102") + "-"
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", prefix).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.Invoice{}).Where("invoice_number LIKE ?", prefix+"%").Count(&count).Error; err != nil {
			return err
		}
		invoice = models.Invoice{
			PaymentID: payment.ID, InvoiceNumber: fmt.Sprintf("%s%04d", prefix, count+1), IssuedAt: now,
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return mapPhase5Error(err)
		}

		enrollment.Status, enrollment.StartedAt = models.EnrollmentActive, &now
		if err := tx.Model(&enrollment).Updates(map[string]any{
			"status": enrollment.Status, "started_at": enrollment.StartedAt,
		}).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrActiveEnrollmentExists
			}
			return mapPhase5Error(err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, mapPhase5Error(err)
	}
	return &enrollment, &payment, &invoice, nil
}

func (r *EnrollmentRepository) ListPayments(ctx context.Context) ([]models.Payment, error) {
	results := make([]models.Payment, 0)
	err := r.db.WithContext(ctx).Order("id DESC").Find(&results).Error
	return results, err
}

func (r *EnrollmentRepository) GetPayment(ctx context.Context, id int64) (*models.Payment, error) {
	var result models.Payment
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return &result, nil
}

func (r *EnrollmentRepository) ListStudentInvoices(ctx context.Context, studentID int64) ([]InvoiceRecord, error) {
	return r.listInvoices(ctx, &studentID)
}

func (r *EnrollmentRepository) GetStudentInvoice(ctx context.Context, studentID, id int64) (*InvoiceRecord, error) {
	return r.invoiceRecord(ctx, id, &studentID)
}

func (r *EnrollmentRepository) ListInvoices(ctx context.Context) ([]InvoiceRecord, error) {
	return r.listInvoices(ctx, nil)
}

func (r *EnrollmentRepository) GetInvoice(ctx context.Context, id int64) (*InvoiceRecord, error) {
	return r.invoiceRecord(ctx, id, nil)
}

func (r *EnrollmentRepository) listInvoices(ctx context.Context, studentID *int64) ([]InvoiceRecord, error) {
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Joins("JOIN payments ON payments.id = invoices.payment_id").
		Joins("JOIN student_enrollments ON student_enrollments.id = payments.enrollment_id")
	if studentID != nil {
		query = query.Where("student_enrollments.student_id = ?", *studentID)
	}
	var invoices []models.Invoice
	if err := query.Order("invoices.id DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}
	results := make([]InvoiceRecord, 0, len(invoices))
	for i := range invoices {
		record, err := r.loadInvoiceRecord(ctx, &invoices[i])
		if err != nil {
			return nil, err
		}
		results = append(results, *record)
	}
	return results, nil
}

func (r *EnrollmentRepository) invoiceRecord(ctx context.Context, id int64, studentID *int64) (*InvoiceRecord, error) {
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Joins("JOIN payments ON payments.id = invoices.payment_id").
		Joins("JOIN student_enrollments ON student_enrollments.id = payments.enrollment_id").
		Where("invoices.id = ?", id)
	if studentID != nil {
		query = query.Where("student_enrollments.student_id = ?", *studentID)
	}
	var invoice models.Invoice
	if err := query.First(&invoice).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return r.loadInvoiceRecord(ctx, &invoice)
}

func (r *EnrollmentRepository) loadInvoiceRecord(ctx context.Context, invoice *models.Invoice) (*InvoiceRecord, error) {
	record := &InvoiceRecord{Invoice: *invoice}
	if err := r.db.WithContext(ctx).First(&record.Payment, invoice.PaymentID).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	if err := r.db.WithContext(ctx).First(&record.Enrollment, record.Payment.EnrollmentID).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	if err := r.db.WithContext(ctx).First(&record.Student, record.Enrollment.StudentID).Error; err != nil {
		return nil, mapPhase5Error(err)
	}
	return record, nil
}

func mapPhase5Error(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateRecord
	}
	return err
}
