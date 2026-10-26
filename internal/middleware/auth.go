// Package middleware holds Gin middlewares.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kthesword/expense-tracker-api/internal/auth"
)

// UserIDKey is the Gin context key under which the authenticated user's ID is stored.
const UserIDKey = "userID"

// RequireAuth rejects requests without a valid "Authorization: Bearer <jwt>" header.
func RequireAuth(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed Authorization header"})
			return
		}
		claims, err := tm.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}
