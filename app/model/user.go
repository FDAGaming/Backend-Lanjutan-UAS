package model

import "time"

// Tabel users
type User struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"unique;not null;type:varchar(50)" json:"username"`
	Email        string    `gorm:"unique;not null;type:varchar(100)" json:"email"`
	PasswordHash string    `gorm:"not null;type:varchar(255);column:password_hash" json:"-"`
	FullName     string    `gorm:"not null;type:varchar(100);column:full_name" json:"fullName"`
	
	RoleID       string    `gorm:"type:uuid;not null;column:role_id" json:"roleId"`
	Role         Role      `gorm:"foreignKey:RoleID;references:ID" json:"role,omitempty"`
	
	IsActive     bool      `gorm:"default:true;column:is_active" json:"isActive"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:updated_at" json:"updatedAt"`
}