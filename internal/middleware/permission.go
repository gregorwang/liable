package middleware

import (
	"comment-review-platform/internal/services"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	permissionService     *services.PermissionService
	permissionServiceOnce sync.Once
)

// getPermissionService returns a lazily initialized permission service
func getPermissionService() *services.PermissionService {
	permissionServiceOnce.Do(func() {
		permissionService = services.NewPermissionService()
	})
	return permissionService
}

// RequirePermission checks if user has the specified permission
func RequirePermission(permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetCheckedPermission(c, permissionKey)
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		hasPermission, err := getPermissionService().HasPermission(userID, permissionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "Insufficient permissions",
				"required_permission": permissionKey,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks if user has any of the specified permissions
func RequireAnyPermission(permissionKeys ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetCheckedPermission(c, strings.Join(permissionKeys, ","))
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		for _, key := range permissionKeys {
			hasPermission, err := getPermissionService().HasPermission(userID, key)
			if err == nil && hasPermission {
				c.Next()
				return
			}
		}

		// User doesn't have any of the required permissions
		c.JSON(http.StatusForbidden, gin.H{
			"error":                "Insufficient permissions",
			"required_permissions": permissionKeys,
		})
		c.Abort()
	}
}
