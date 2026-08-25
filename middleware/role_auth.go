package middleware

import (
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/models"
	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

func RoleAuth(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		user, ok := AuthenticatedUser(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "authentication required")
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			utils.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		c.Next()
	}
}
