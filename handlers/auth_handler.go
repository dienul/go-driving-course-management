package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/dto"
	"github.com/dienulhaq/go-driving-course-management/middleware"
	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (*dto.RegisterData, error)
	Login(ctx context.Context, request dto.LoginRequest) (*dto.LoginData, error)
}

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register godoc
// @Summary Register a student
// @Description Creates an ACTIVE STUDENT account and student profile. The role is always determined by the backend.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Student registration"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/users/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	data, err := h.auth.Register(c.Request.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailExists):
			utils.Error(c, http.StatusConflict, err.Error())
		case errors.Is(err, services.ErrInvalidInput):
			utils.Error(c, http.StatusUnprocessableEntity, err.Error())
		default:
			utils.InternalServerError(c)
		}
		return
	}
	utils.Success(c, http.StatusCreated, "student registered successfully", data)
}

// Login godoc
// @Summary Login
// @Description Authenticates any ACTIVE ADMIN, TRAINER, or STUDENT account using email and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/users/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	data, err := h.auth.Login(c.Request.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			utils.Error(c, http.StatusUnauthorized, err.Error())
		case errors.Is(err, services.ErrInactiveUser):
			utils.Error(c, http.StatusForbidden, err.Error())
		default:
			utils.InternalServerError(c)
		}
		return
	}
	utils.Success(c, http.StatusOK, "login successful", data)
}

// Me godoc
// @Summary Get current user
// @Description Returns the currently authenticated ACTIVE user.
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserAPIResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.AuthenticatedUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	data := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	utils.Success(c, http.StatusOK, "success", data)
}
