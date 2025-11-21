package model

import "time"

// Tabel students
type Student struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	
	UserID       string    `gorm:"type:uuid;column:user_id" json:"userId"`
	User         User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	
	// Kolom student_id di Database (NIM)
	StudentID    string    `gorm:"unique;not null;type:varchar(20);column:student_id" json:"studentId"` 
	
	ProgramStudy string    `gorm:"type:varchar(100);column:program_study" json:"programStudy"`
	AcademicYear string    `gorm:"type:varchar(10);column:academic_year" json:"academicYear"`
	
	AdvisorID    *string   `gorm:"type:uuid;column:advisor_id" json:"advisorId"`
	Advisor      *Lecturer `gorm:"foreignKey:AdvisorID;references:ID" json:"advisor,omitempty"`
	
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"createdAt"`
}