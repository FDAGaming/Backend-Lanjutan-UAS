package model

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoDB Struct (Detail Prestasi) [cite: 107]
type AchievementDetail struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    StudentID       string             `bson:"studentId" json:"studentId"`
    AchievementType string             `bson:"achievementType" json:"achievementType"` // academic, competition, etc
    Title           string             `bson:"title" json:"title"`
    Description     string             `bson:"description" json:"description"`
    Details         map[string]interface{} `bson:"details" json:"details"` // Dynamic Fields [cite: 114]
    CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

// PostgreSQL Struct (Referensi Status) [cite: 92]
type AchievementReference struct {
    ID                 string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    StudentID          string    `gorm:"type:uuid;not null" json:"studentId"`
    MongoAchievementID string    `gorm:"type:varchar(24);not null" json:"mongoAchievementId"`
    Status             string    `gorm:"type:varchar(20);default:'draft'" json:"status"` // draft, submitted [cite: 96]
    CreatedAt          time.Time `json:"createdAt"`
}