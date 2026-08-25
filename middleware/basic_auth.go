package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/dienulhaq/go-driving-course-management/utils"
	"github.com/gin-gonic/gin"
)

func BasicAuth(expectedUsername, expectedPassword string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedUsername == "" || expectedPassword == "" {
			utils.InternalServerError(c)
			return
		}

		username, password, ok := c.Request.BasicAuth()
		usernameMatches := subtle.ConstantTimeCompare(
			[]byte(username),
			[]byte(expectedUsername),
		) == 1
		passwordMatches := subtle.ConstantTimeCompare(
			[]byte(password),
			[]byte(expectedPassword),
		) == 1
		if !ok || !usernameMatches || !passwordMatches {
			c.Header("WWW-Authenticate", `Basic realm="internal"`)
			utils.Error(c, http.StatusUnauthorized, "invalid basic auth credentials")
			return
		}
		c.Next()
	}
}
