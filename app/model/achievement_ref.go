package model

import "time"

// Tabel achievement_references
type AchievementReference struct {
	ID                 string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	
	// Foreign Key merujuk ke students.id (UUID)
	StudentID          string     `gorm:"type:uuid;column:student_id" json:"studentId"`
	Student            Student    `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
	
	MongoAchievementID string     `gorm:"not null;type:varchar(24);column:mongo_achievement_id" json:"mongoAchievementId"`
	
	// Field Title (Tambahan Modul 6 Search/Sort) - Tetap di Postgres agar query cepat
	Title              string     `gorm:"type:varchar(255);not null" json:"title"`
	
	// Enum (draft, submitted, verified, rejected)
	Status             string     `gorm:"type:varchar(20);default:'draft'" json:"status"`
	
	SubmittedAt        *time.Time `gorm:"column:submitted_at" json:"submittedAt"`
	VerifiedAt         *time.Time `gorm:"column:verified_at" json:"verifiedAt"`
	
	VerifiedBy         *string    `gorm:"type:uuid;column:verified_by" json:"verifiedBy"`
	Verifier           *User      `gorm:"foreignKey:VerifiedBy;references:ID" json:"verifier,omitempty"`
	
	RejectionNote      string     `gorm:"type:text;column:rejection_note" json:"rejectionNote"`
	CreatedAt          time.Time  `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"default:CURRENT_TIMESTAMP;column:updated_at" json:"updatedAt"`
}