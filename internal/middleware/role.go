package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetRole(c)
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			// Debug: log the role mismatch
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"your_role":      userRole,
				"required_roles": roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a shortcut for requiring admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireReviewer is a shortcut for requiring reviewer role
func RequireReviewer() gin.HandlerFunc {
	return RequireRole("reviewer", "admin") // Admin can also act as reviewer
}

// RequireAdminOrReviewer is a shortcut for requiring admin or reviewer role
func RequireAdminOrReviewer() gin.HandlerFunc {
	return RequireRole("admin", "reviewer")
}
