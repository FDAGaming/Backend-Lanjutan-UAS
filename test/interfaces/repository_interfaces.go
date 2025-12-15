package interfaces

import (
	"uas/app/model"
)

// RoleRepositoryInterface defines the interface for role repository
type RoleRepositoryInterface interface {
	GetPermissionsByRoleID(roleID string) ([]model.Permission, error)
	FindByName(name string) (*model.Role, error)
}
