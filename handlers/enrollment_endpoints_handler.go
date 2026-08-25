package handlers

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

// StudentListPackages godoc
// @Summary List active course packages
// @Description List active course packages. Requires an ACTIVE STUDENT account.
// @Tags Student Packages
// @Produce json
// @Success 200 {object} dto.PackageListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/packages [get]
func (h *EnrollmentHandler) StudentListPackages(c *gin.Context) {
	result, err := h.service.ListPackages(c.Request.Context())
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentGetPackage godoc
// @Summary Get active course package
// @Description Get active course package. Requires an ACTIVE STUDENT account.
// @Tags Student Packages
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.PackageAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/packages/{id} [get]
func (h *EnrollmentHandler) StudentGetPackage(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetPackage(c.Request.Context(), id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentCreateEnrollment godoc
// @Summary Create enrollment and unpaid payment
// @Description Create enrollment and unpaid payment. Requires an ACTIVE STUDENT account.
// @Tags Student Enrollments
// @Accept json
// @Produce json
// @Param request body dto.CreateEnrollmentRequest true "Request body"
// @Success 201 {object} dto.EnrollmentCheckoutAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/enrollments [post]
func (h *EnrollmentHandler) StudentCreateEnrollment(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	var request dto.CreateEnrollmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.CreateEnrollment(c.Request.Context(), studentID, request)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusCreated, "enrollment created successfully", result)
}

// StudentListEnrollments godoc
// @Summary List own enrollments
// @Description List own enrollments. Requires an ACTIVE STUDENT account.
// @Tags Student Enrollments
// @Produce json
// @Success 200 {object} dto.EnrollmentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/enrollments [get]
func (h *EnrollmentHandler) StudentListEnrollments(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	result, err := h.service.ListStudentEnrollments(c.Request.Context(), studentID)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentGetEnrollment godoc
// @Summary Get own enrollment
// @Description Get own enrollment. Requires an ACTIVE STUDENT account.
// @Tags Student Enrollments
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.EnrollmentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/enrollments/{id} [get]
func (h *EnrollmentHandler) StudentGetEnrollment(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.GetStudentEnrollment(c.Request.Context(), studentID, id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentGetEnrollmentPayment godoc
// @Summary Get own enrollment payment
// @Description Get own enrollment payment. Requires an ACTIVE STUDENT account.
// @Tags Student Payments
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.PaymentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/enrollments/{id}/payment [get]
func (h *EnrollmentHandler) StudentGetEnrollmentPayment(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.GetEnrollmentPayment(c.Request.Context(), studentID, id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentPayPayment godoc
// @Summary Simulate payment and issue invoice
// @Description Simulate payment and issue invoice. Requires an ACTIVE STUDENT account.
// @Tags Student Payments
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.PayRequest true "Request body"
// @Success 200 {object} dto.PaymentCheckoutAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/payments/{id}/pay [post]
func (h *EnrollmentHandler) StudentPayPayment(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	var request dto.PayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.Pay(c.Request.Context(), studentID, id, request)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "payment completed successfully", result)
}

// StudentListInvoices godoc
// @Summary List own paid invoices
// @Description List own paid invoices. Requires an ACTIVE STUDENT account.
// @Tags Student Invoices
// @Produce json
// @Success 200 {object} dto.InvoiceListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/invoices [get]
func (h *EnrollmentHandler) StudentListInvoices(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	result, err := h.service.ListStudentInvoices(c.Request.Context(), studentID)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// StudentGetInvoice godoc
// @Summary Get own paid invoice
// @Description Get own paid invoice. Requires an ACTIVE STUDENT account.
// @Tags Student Invoices
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.InvoiceAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/student/invoices/{id} [get]
func (h *EnrollmentHandler) StudentGetInvoice(c *gin.Context) {
	studentID, ok := currentStudentID(c)
	if !ok {
		return
	}
	id, valid := adminID(c, "id")
	if !valid {
		return
	}
	result, err := h.service.GetStudentInvoice(c.Request.Context(), studentID, id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminListEnrollments godoc
// @Summary List all student enrollments
// @Description List all student enrollments. Requires an ACTIVE ADMIN account.
// @Tags Admin Enrollments
// @Produce json
// @Success 200 {object} dto.EnrollmentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/enrollments [get]
func (h *EnrollmentHandler) AdminListEnrollments(c *gin.Context) {
	result, err := h.service.ListEnrollments(c.Request.Context())
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminGetEnrollment godoc
// @Summary Get student enrollment
// @Description Get student enrollment. Requires an ACTIVE ADMIN account.
// @Tags Admin Enrollments
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.EnrollmentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/enrollments/{id} [get]
func (h *EnrollmentHandler) AdminGetEnrollment(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetEnrollment(c.Request.Context(), id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminListPayments godoc
// @Summary List all payments
// @Description List all payments. Requires an ACTIVE ADMIN account.
// @Tags Admin Payments
// @Produce json
// @Success 200 {object} dto.PaymentListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments [get]
func (h *EnrollmentHandler) AdminListPayments(c *gin.Context) {
	result, err := h.service.ListPayments(c.Request.Context())
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminGetPayment godoc
// @Summary Get payment
// @Description Get payment. Requires an ACTIVE ADMIN account.
// @Tags Admin Payments
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.PaymentAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments/{id} [get]
func (h *EnrollmentHandler) AdminGetPayment(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetPayment(c.Request.Context(), id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminListInvoices godoc
// @Summary List all issued invoices
// @Description List all issued invoices. Requires an ACTIVE ADMIN account.
// @Tags Admin Invoices
// @Produce json
// @Success 200 {object} dto.InvoiceListAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/invoices [get]
func (h *EnrollmentHandler) AdminListInvoices(c *gin.Context) {
	result, err := h.service.ListInvoices(c.Request.Context())
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}

// AdminGetInvoice godoc
// @Summary Get issued invoice
// @Description Get issued invoice. Requires an ACTIVE ADMIN account.
// @Tags Admin Invoices
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} dto.InvoiceAPIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/invoices/{id} [get]
func (h *EnrollmentHandler) AdminGetInvoice(c *gin.Context) {
	id, ok := adminID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.GetInvoice(c.Request.Context(), id)
	if err != nil {
		enrollmentFailure(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "success", result)
}
