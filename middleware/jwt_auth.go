package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/repositories"
	"github.com/dienulhaq/go-driving-course-management/services"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

const authenticatedUserKey = "authenticated_user"

type UserReader interface {
	FindByID(ctx context.Context, id int64) (*models.User, error)
}

func JWTAuth(tokens *services.JWTManager, users UserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "missing or invalid bearer token")
			return
		}

		claims, err := tokens.Parse(rawToken)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		user, err := users.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			if errors.Is(err, repositories.ErrUserNotFound) {
				utils.Error(c, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			utils.InternalServerError(c)
			return
		}
		if user.Status != models.StatusActive {
			utils.Error(c, http.StatusForbidden, "user account is inactive")
			return
		}

		c.Set(authenticatedUserKey, user)
		c.Next()
	}
}

func AuthenticatedUser(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get(authenticatedUserKey)
	if !exists {
		return nil, false
	}
	user, ok := value.(*models.User)
	return user, ok && user != nil
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
