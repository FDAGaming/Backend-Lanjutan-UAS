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

// Create User baru (Register)
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Cari User by Email (Login) - Preload Role agar tahu dia Admin/Mhs
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

// Cari User by ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("id = ?", id).First(&user).Error
	return &user, err
}

// --- Helper untuk Profil ---

// Cari Data Mahasiswa berdasarkan UserID (Untuk validasi saat submit prestasi)
func (r *UserRepository) FindStudentByUserID(userID string) (*model.Student, error) {
	var student model.Student
	// Preload Advisor (Dosen Wali) jika perlu
	err := r.db.Preload("User").Preload("Advisor.User").Where("user_id = ?", userID).First(&student).Error
	return &student, err
}

// Cari Data Dosen berdasarkan UserID (Untuk verifikasi)
func (r *UserRepository) FindLecturerByUserID(userID string) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&lecturer).Error
	return &lecturer, err
}