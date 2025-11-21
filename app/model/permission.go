package model

// Tabel permissions
type Permission struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string `gorm:"unique;not null;type:varchar(100)" json:"name"`
	Resource    string `gorm:"not null;type:varchar(50)" json:"resource"`
	Action      string `gorm:"not null;type:varchar(50)" json:"action"`
	Description string `gorm:"type:text" json:"description"`
}

// Tabel role_permissions
type RolePermission struct {
	RoleID       string     `gorm:"primaryKey;type:uuid;column:role_id" json:"roleId"`
	PermissionID string     `gorm:"primaryKey;type:uuid;column:permission_id" json:"permissionId"`
	
	// Relasi (Optional, agar GORM tahu link-nya)
	Role         Role       `gorm:"foreignKey:RoleID;references:ID" json:"-"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID" json:"-"`
}

// Override nama tabel agar tidak di-pluralize otomatis (menjadi role_permissions)
func (RolePermission) TableName() string {
	return "role_permissions"
}