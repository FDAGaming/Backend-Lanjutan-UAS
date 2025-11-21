package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- MONGODB MODELS (SRS 3.2.1) ---

type AchievementContent struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID       string                 `bson:"studentId" json:"studentId"` // UUID dari Postgres Student
	AchievementType string                 `bson:"achievementType" json:"achievementType"` // academic, competition, etc
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	Details         map[string]interface{} `bson:"details" json:"details"` // Field Dinamis
	Attachments     []Attachment           `bson:"attachments" json:"attachments"`
	Tags            []string               `bson:"tags" json:"tags"`
	Points          int                    `bson:"points" json:"points"`
	CreatedAt       time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updatedAt"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileURL    string    `bson:"fileUrl" json:"fileUrl"`
	FileType   string    `bson:"fileType" json:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

// --- POSTGRESQL MODELS (SRS 3.1.7) ---

type AchievementReference struct {
	ID                 string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StudentID          string     `gorm:"type:uuid;not null" json:"studentId"`
	MongoAchievementID string     `gorm:"type:varchar(24);not null" json:"mongoAchievementId"`
	Status             string     `gorm:"type:varchar(20);default:'draft'" json:"status"` // draft, submitted, verified, rejected
	SubmittedAt        *time.Time `json:"submittedAt"`
	VerifiedAt         *time.Time `json:"verifiedAt"`
	VerifiedBy         *string    `gorm:"type:uuid" json:"verifiedBy"`
	RejectionNote      string     `gorm:"type:text" json:"rejectionNote"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// --- DTOs (Request/Response) ---

type CreateAchievementRequest struct {
	Title           string                 `json:"title" validate:"required"`
	AchievementType string                 `json:"achievementType" validate:"required"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
}

type VerifyAchievementRequest struct {
	Status        string `json:"status" validate:"required,oneof=verified rejected"`
	RejectionNote string `json:"rejectionNote"`
}