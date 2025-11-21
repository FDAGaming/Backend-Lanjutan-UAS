package repository

import (
	"uas/app/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	// Preload Role untuk kebutuhan cek permission
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

// Helper untuk mendapatkan Student ID dari User ID
func (r *UserRepository) FindStudentByUserID(userID string) (*model.Student, error) {
	var student model.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	return &student, err
}

// Helper untuk mendapatkan Lecturer ID dari User ID
func (r *UserRepository) FindLecturerByUserID(userID string) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Where("user_id = ?", userID).First(&lecturer).Error
	return &lecturer, err
}