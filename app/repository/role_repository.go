package repository

import (
	"uas/app/model"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Mencari Role berdasarkan nama (misal: untuk default role saat register)
func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

// Mengambil Permission yang dimiliki oleh sebuah Role (untuk Middleware RBAC)
func (r *RoleRepository) GetPermissionsByRoleID(roleID string) ([]model.Permission, error) {
	var permissions []model.Permission
	
	// Join table role_permissions dan permissions
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
		
	return permissions, err
}