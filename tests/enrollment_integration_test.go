package tests

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/routes"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestEnrollmentPaymentInvoicePostgreSQL(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash fixture password: %v", err)
	}
	fixtures := []models.User{
		{Name: "Phase 5 Admin", Email: "phase5-admin@example.com", PasswordHash: string(hash), Role: models.RoleAdmin, Status: models.StatusActive},
		{Name: "Phase 5 Trainer", Email: "phase5-trainer@example.com", PasswordHash: string(hash), Role: models.RoleTrainer, Status: models.StatusActive},
	}
	for i := range fixtures {
		if err := db.Create(&fixtures[i]).Error; err != nil {
			t.Fatalf("create fixture user: %v", err)
		}
	}
	packages := []models.CoursePackage{
		{Name: "Phase 5 Pemula 6 Jam", Level: models.PackageLevelPemula, TotalHours: 6, Price: 900000, Status: models.StatusActive},
		{Name: "Phase 5 Dasar 10 Jam", Level: models.PackageLevelDasar, TotalHours: 10, Price: 1400000, Status: models.StatusActive},
		{Name: "Phase 5 Hidden Package", Level: models.PackageLevelPemula, TotalHours: 8, Price: 1100000, Status: models.StatusInactive},
	}
	for i := range packages {
		if err := db.Create(&packages[i]).Error; err != nil {
			t.Fatalf("create package fixture: %v", err)
		}
	}
	gin.SetMode(gin.TestMode)
	router, err := routes.New(db, config.Config{JWTSecret: "integration-test-secret-with-at-least-32-bytes", JWTExpiresIn: "1h"})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}
	for _, email := range []string{"phase5-student-one@example.com", "phase5-student-two@example.com"} {
		response := performJSON(t, router, http.MethodPost, "/api/users/register", map[string]any{
			"name": "Phase 5 Student", "email": email, "password": "strong-password",
		}, "")
		requireAdminStatus(t, response, http.StatusCreated)
	}
	login := func(email string) string {
		t.Helper()
		response := performJSON(t, router, http.MethodPost, "/api/users/login", map[string]any{
			"email": email, "password": "strong-password",
		}, "")
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.LoginData `json:"data"`
		}
		decodeResponse(t, response, &result)
		return "Bearer " + result.Data.Token
	}
	adminToken := login(fixtures[0].Email)
	trainerToken := login(fixtures[1].Email)
	studentToken := login("phase5-student-one@example.com")
	otherToken := login("phase5-student-two@example.com")

	t.Run("authorization and active package catalog", func(t *testing.T) {
		path := "/api/v1/student/packages"
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, ""), http.StatusUnauthorized)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, adminToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, path, nil, trainerToken), http.StatusForbidden)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/admin/enrollments", nil, studentToken), http.StatusForbidden)
		response := performJSON(t, router, http.MethodGet, path, nil, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data []models.CoursePackage `json:"data"`
		}
		decodeResponse(t, response, &result)
		for _, item := range result.Data {
			if item.Status != models.StatusActive {
				t.Fatalf("inactive package appeared in student catalog: %+v", item)
			}
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("%s/%d", path, packages[0].ID), nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("%s/%d", path, packages[2].ID), nil, studentToken), http.StatusNotFound)
	})

	var first, second dto.EnrollmentCheckoutData
	t.Run("transactional enrollment and package snapshot", func(t *testing.T) {
		endpoint := "/api/v1/student/enrollments"
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"package_id": 0}, studentToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"package_id": packages[2].ID}, studentToken), http.StatusNotFound)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"package_id": 999999}, studentToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPost, endpoint, map[string]any{
			"package_id": packages[0].ID, "package_price": 1, "amount": 1, "student_id": fixtures[0].ID,
		}, studentToken)
		requireAdminStatus(t, response, http.StatusCreated)
		var result struct {
			Data dto.EnrollmentCheckoutData `json:"data"`
		}
		decodeResponse(t, response, &result)
		first = result.Data
		if first.Enrollment.Status != models.EnrollmentPendingPayment || first.Payment.Status != models.PaymentUnpaid {
			t.Fatalf("checkout state mismatch: %+v", first)
		}
		if first.Enrollment.PackageName != packages[0].Name || first.Enrollment.PackagePrice != 900000 || first.Payment.Amount != 900000 || first.Payment.PaymentMethod != nil {
			t.Fatalf("client changed authoritative package/payment data: %+v", first)
		}
		if first.Enrollment.StudentID == fixtures[0].ID || first.Payment.EnrollmentID != first.Enrollment.ID {
			t.Fatalf("checkout ownership/relationship mismatch: %+v", first)
		}
		if !regexp.MustCompile("^PAY-[0-9]{8}-[A-F0-9]{16}$").MatchString(first.Payment.PaymentCode) {
			t.Fatalf("invalid payment code: %s", first.Payment.PaymentCode)
		}
		if err := db.Model(&models.CoursePackage{}).Where("id = ?", packages[0].ID).Update("price", 1800000).Error; err != nil {
			t.Fatalf("change package master after snapshot: %v", err)
		}
		response = performJSON(t, router, http.MethodPost, endpoint, map[string]any{"package_id": packages[1].ID}, studentToken)
		requireAdminStatus(t, response, http.StatusCreated)
		decodeResponse(t, response, &result)
		second = result.Data
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, endpoint, nil, studentToken), http.StatusOK)
		ownPath := fmt.Sprintf("%s/%d", endpoint, first.Enrollment.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath, nil, otherToken), http.StatusNotFound)
		payment := performJSON(t, router, http.MethodGet, ownPath+"/payment", nil, studentToken)
		requireAdminStatus(t, payment, http.StatusOK)
		var paymentResult struct {
			Data models.Payment `json:"data"`
		}
		decodeResponse(t, payment, &paymentResult)
		if paymentResult.Data.Amount != 900000 {
			t.Fatalf("package master update changed payment snapshot: %+v", paymentResult.Data)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, ownPath+"/payment", nil, otherToken), http.StatusNotFound)
		invoiceList := performJSON(t, router, http.MethodGet, "/api/v1/student/invoices", nil, studentToken)
		requireAdminStatus(t, invoiceList, http.StatusOK)
		var unpaidInvoices struct {
			Data []dto.InvoiceData `json:"data"`
		}
		decodeResponse(t, invoiceList, &unpaidInvoices)
		if len(unpaidInvoices.Data) != 0 {
			t.Fatalf("invoice existed before payment: %+v", unpaidInvoices.Data)
		}
	})

	var issued dto.PaymentCheckoutData
	t.Run("atomic payment invoice and activation", func(t *testing.T) {
		endpoint := fmt.Sprintf("/api/v1/student/payments/%d/pay", first.Payment.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"payment_method": "CARD"}, studentToken), http.StatusBadRequest)
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"payment_method": "CASH"}, otherToken), http.StatusNotFound)
		response := performJSON(t, router, http.MethodPost, endpoint, map[string]any{
			"payment_method": "BANK_TRANSFER", "amount": 1,
		}, studentToken)
		requireAdminStatus(t, response, http.StatusOK)
		var result struct {
			Data dto.PaymentCheckoutData `json:"data"`
		}
		decodeResponse(t, response, &result)
		issued = result.Data
		if issued.Enrollment.Status != models.EnrollmentActive || issued.Enrollment.StartedAt == nil ||
			issued.Payment.Status != models.PaymentPaid || issued.Payment.PaidAt == nil ||
			issued.Payment.PaymentMethod == nil || *issued.Payment.PaymentMethod != models.PaymentMethodBankTransfer {
			t.Fatalf("payment transaction did not complete all state transitions: %+v", issued)
		}
		if issued.Payment.Amount != 900000 || issued.Invoice.Enrollment.PackagePrice != 900000 ||
			issued.Invoice.Enrollment.PackageName != packages[0].Name {
			t.Fatalf("invoice lost historical package snapshot: %+v", issued)
		}
		if !regexp.MustCompile("^INV-[0-9]{8}-[0-9]{4}$").MatchString(issued.Invoice.InvoiceNumber) {
			t.Fatalf("invalid invoice number: %s", issued.Invoice.InvoiceNumber)
		}
		if issued.Invoice.Student.Email != "phase5-student-one@example.com" || issued.Invoice.PaymentID != issued.Payment.ID {
			t.Fatalf("invoice relation data is incorrect: %+v", issued.Invoice)
		}
		requireAdminStatus(t, performJSON(t, router, http.MethodPost, endpoint, map[string]any{"payment_method": "CASH"}, studentToken), http.StatusConflict)
		var invoiceCount int64
		if err := db.Model(&models.Invoice{}).Where("payment_id = ?", first.Payment.ID).Count(&invoiceCount).Error; err != nil || invoiceCount != 1 {
			t.Fatalf("double payment created duplicate invoice: count=%d err=%v", invoiceCount, err)
		}
	})

	t.Run("second active enrollment rolls payment back", func(t *testing.T) {
		endpoint := fmt.Sprintf("/api/v1/student/payments/%d/pay", second.Payment.ID)
		response := performJSON(t, router, http.MethodPost, endpoint, map[string]any{"payment_method": "CASH"}, studentToken)
		requireAdminStatus(t, response, http.StatusConflict)
		var payment models.Payment
		if err := db.First(&payment, second.Payment.ID).Error; err != nil || payment.Status != models.PaymentUnpaid || payment.PaidAt != nil {
			t.Fatalf("failed activation did not roll back payment: payment=%+v err=%v", payment, err)
		}
		var invoiceCount int64
		if err := db.Model(&models.Invoice{}).Where("payment_id = ?", second.Payment.ID).Count(&invoiceCount).Error; err != nil || invoiceCount != 0 {
			t.Fatalf("failed activation generated an invoice: count=%d err=%v", invoiceCount, err)
		}
		response = performJSON(t, router, http.MethodPost, "/api/v1/student/enrollments", map[string]any{"package_id": packages[1].ID}, studentToken)
		requireAdminStatus(t, response, http.StatusConflict)
	})

	t.Run("invoice ownership and administrator monitoring", func(t *testing.T) {
		invoicePath := fmt.Sprintf("/api/v1/student/invoices/%d", issued.Invoice.ID)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, "/api/v1/student/invoices", nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, invoicePath, nil, studentToken), http.StatusOK)
		requireAdminStatus(t, performJSON(t, router, http.MethodGet, invoicePath, nil, otherToken), http.StatusNotFound)
		resources := []struct {
			collection string
			id         int64
		}{
			{"/api/v1/admin/enrollments", first.Enrollment.ID},
			{"/api/v1/admin/payments", first.Payment.ID},
			{"/api/v1/admin/invoices", issued.Invoice.ID},
		}
		for _, resource := range resources {
			requireAdminStatus(t, performJSON(t, router, http.MethodGet, resource.collection, nil, adminToken), http.StatusOK)
			requireAdminStatus(t, performJSON(t, router, http.MethodGet, fmt.Sprintf("%s/%d", resource.collection, resource.id), nil, adminToken), http.StatusOK)
		}
	})

	t.Run("cash payment and sequential invoices", func(t *testing.T) {
		enrollment := performJSON(t, router, http.MethodPost, "/api/v1/student/enrollments", map[string]any{
			"package_id": packages[1].ID,
		}, otherToken)
		requireAdminStatus(t, enrollment, http.StatusCreated)
		var checkout struct {
			Data dto.EnrollmentCheckoutData `json:"data"`
		}
		decodeResponse(t, enrollment, &checkout)
		paid := performJSON(t, router, http.MethodPost, fmt.Sprintf("/api/v1/student/payments/%d/pay", checkout.Data.Payment.ID), map[string]any{
			"payment_method": "CASH",
		}, otherToken)
		requireAdminStatus(t, paid, http.StatusOK)
		var result struct {
			Data dto.PaymentCheckoutData `json:"data"`
		}
		decodeResponse(t, paid, &result)
		if result.Data.Payment.PaymentMethod == nil || *result.Data.Payment.PaymentMethod != models.PaymentMethodCash {
			t.Fatalf("cash payment method was not recorded: %+v", result.Data.Payment)
		}
		if result.Data.Invoice.InvoiceNumber == issued.Invoice.InvoiceNumber {
			t.Fatalf("invoice numbers were reused: %s", issued.Invoice.InvoiceNumber)
		}
	})
}
