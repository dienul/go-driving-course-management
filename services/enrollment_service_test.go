package services

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
)

func TestPaymentCodeAndMethodValidation(t *testing.T) {
	now := time.Date(2026, time.August, 25, 10, 0, 0, 0, time.UTC)
	first, err := generatePaymentCode(now)
	if err != nil {
		t.Fatalf("generate payment code: %v", err)
	}
	second, err := generatePaymentCode(now)
	if err != nil {
		t.Fatalf("generate second payment code: %v", err)
	}
	pattern := regexp.MustCompile("^PAY-20260825-[A-F0-9]{16}$")
	if !pattern.MatchString(first) || first == second {
		t.Fatalf("unsafe or duplicate payment codes: first=%q second=%q", first, second)
	}
	if !validPaymentMethod(models.PaymentMethodBankTransfer) || !validPaymentMethod(models.PaymentMethodCash) || validPaymentMethod(models.PaymentMethod("CARD")) {
		t.Fatal("payment method validation is incorrect")
	}
}

func TestEnrollmentBusinessErrorMapping(t *testing.T) {
	cases := []struct{ source, expected error }{
		{repositories.ErrRecordNotFound, ErrResourceNotFound},
		{repositories.ErrDuplicateRecord, ErrResourceConflict},
		{repositories.ErrActiveEnrollmentExists, ErrActiveEnrollment},
		{repositories.ErrPaymentAlreadyProcessed, ErrPaymentProcessed},
		{repositories.ErrInvalidFinancialState, ErrInvalidState},
	}
	for _, item := range cases {
		if !errors.Is(enrollmentError(item.source), item.expected) {
			t.Errorf("error %v mapped to %v, want %v", item.source, enrollmentError(item.source), item.expected)
		}
	}
}
