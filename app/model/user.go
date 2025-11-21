package model

import (
	"time"
	// "gorm.io/gorm"
)

// SRS 3.1.1 & 3.1.2
type Role struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"` // Admin, Mahasiswa, Dosen Wali
	Description string    `json:"description"`
}

type User struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	Email        string    `gorm:"unique;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // Password tidak direturn di JSON
	FullName     string    `gorm:"not null" json:"fullName"`
	RoleID       string    `gorm:"type:uuid;not null" json:"roleId"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SRS 3.1.5
type Student struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"type:uuid;unique;not null" json:"userId"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
StudentID    string    `gorm:"unique;not null;type:varchar(20)" json:"studentId"` // NIM
	ProgramStudy string    `json:"programStudy"`
	AdvisorID    *string   `gorm:"type:uuid" json:"advisorId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// SRS 3.1.6
type Lecturer struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"type:uuid;unique;not null" json:"userId"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LecturerID string    `gorm:"unique;not null" json:"lecturerId"` // NIP
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"createdAt"`
}