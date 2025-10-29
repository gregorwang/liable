package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"fmt"
)

type PermissionService struct {
	permissionRepo *repository.PermissionRepository
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		permissionRepo: repository.NewPermissionRepository(),
	}
}

// GetUserPermissions retrieves all permissions for a user
func (s *PermissionService) GetUserPermissions(userID int) ([]string, error) {
	return s.permissionRepo.GetUserPermissions(userID)
}

// HasPermission checks if a user has a specific permission
func (s *PermissionService) HasPermission(userID int, permissionKey string) (bool, error) {
	return s.permissionRepo.HasPermission(userID, permissionKey)
}

// GrantPermissions grants multiple permissions to a user
func (s *PermissionService) GrantPermissions(userID int, permissionKeys []string, grantedBy int) error {
	if len(permissionKeys) == 0 {
		return fmt.Errorf("no permissions to grant")
	}

	// Validate that all permission keys exist
	for _, key := range permissionKeys {
		_, err := s.permissionRepo.GetPermissionByKey(key)
		if err != nil {
			return fmt.Errorf("invalid permission key %s: %w", key, err)
		}
	}

	// Grant permissions
	return s.permissionRepo.GrantPermissions(userID, permissionKeys, &grantedBy)
}

// RevokePermissions revokes multiple permissions from a user
func (s *PermissionService) RevokePermissions(userID int, permissionKeys []string) error {
	if len(permissionKeys) == 0 {
		return fmt.Errorf("no permissions to revoke")
	}

	return s.permissionRepo.RevokePermissions(userID, permissionKeys)
}

// GetAllPermissions retrieves all active permissions
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.GetAllPermissions()
}

// ListPermissions retrieves permissions with filtering and pagination
func (s *PermissionService) ListPermissions(resource, category, search string, page, pageSize int) (*models.ListPermissionsResponse, error) {
	permissions, total, err := s.permissionRepo.ListPermissions(resource, category, search, page, pageSize)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.ListPermissionsResponse{
		Data:       permissions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
