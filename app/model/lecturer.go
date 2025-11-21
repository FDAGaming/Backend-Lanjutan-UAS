package model

import "time"

// Tabel lecturers
type Lecturer struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	
	UserID     string    `gorm:"type:uuid;column:user_id" json:"userId"`
	User       User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	
	LecturerID string    `gorm:"unique;not null;type:varchar(20);column:lecturer_id" json:"lecturerId"` // NIP
	Department string    `gorm:"type:varchar(100)" json:"department"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt"`
}