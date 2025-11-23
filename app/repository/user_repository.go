package repository

import (
	"database/sql"
	"errors"
	"time"
	"uas/app/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create User baru (Register)
func (r *UserRepository) Create(user *model.User) error {
	// Kita gunakan RETURNING untuk mendapatkan ID dan CreatedAt yang digenerate database
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	// Set default timestamp jika kosong
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	// Eksekusi Query
	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

// Cari User by Email (Login) - JOIN Role agar tahu dia Admin/Mhs
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
			r.id, r.name, r.description
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
		LIMIT 1`

	var user model.User
	user.Role = &model.Role{} // Inisialisasi pointer struct Role

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&user.Role.ID, &user.Role.Name, &user.Role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// Cari User by ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
			r.id, r.name, r.description
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1`

	var user model.User
	user.Role = &model.Role{}

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&user.Role.ID, &user.Role.Name, &user.Role.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// --- Helper untuk Profil ---

// Cari Data Mahasiswa berdasarkan UserID (Untuk validasi saat submit prestasi)
// Join: Students -> Users (Mhs) -> Lecturers (Advisor) -> Users (Dosen)
func (r *UserRepository) FindStudentByUserID(userID string) (*model.Student, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.id, u.username, u.full_name, u.email, -- Info User Mahasiswa
			l.id, l.lecturer_id, l.department, -- Info Advisor (Dosen)
			au.id, au.full_name -- Info User Advisor (Nama Dosen)
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users au ON l.user_id = au.id
		WHERE s.user_id = $1`

	var s model.Student
	s.User = &model.User{}         // Init pointer
	s.Advisor = &model.Lecturer{}  // Init pointer
	s.Advisor.User = &model.User{} // Init pointer user dalam advisor

	// Variable helper untuk scanning field nullable (LEFT JOIN)
	var advisorID sql.NullString
	var advisorLecID sql.NullString
	var advisorDept sql.NullString
	var advisorUserID sql.NullString
	var advisorName sql.NullString

	// Karena Advisor ID di DB disimpan sebagai UUID tapi di scan ke string, kita perlu handling NULL dengan hati-hati
	// Scanning dilakukan berurutan sesuai Query
	err := r.db.QueryRow(query, userID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &advisorID, &s.CreatedAt,
		&s.User.ID, &s.User.Username, &s.User.FullName, &s.User.Email,
		&s.Advisor.ID, &advisorLecID, &advisorDept,
		&advisorUserID, &advisorName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student profile not found")
		}
		return nil, err
	}

	// Mapping kembali NullString ke struct jika ada datanya
	if advisorID.Valid {
		idStr := advisorID.String
		s.AdvisorID = &idStr
		s.Advisor.ID = advisorID.String
		s.Advisor.LecturerID = advisorLecID.String
		s.Advisor.Department = advisorDept.String
		s.Advisor.User.ID = advisorUserID.String
		s.Advisor.User.FullName = advisorName.String
	} else {
		// Jika tidak punya dosen wali, set nil
		s.Advisor = nil
		s.AdvisorID = nil
	}

	return &s, nil
}

// Cari Data Dosen berdasarkan UserID (Untuk verifikasi)
func (r *UserRepository) FindLecturerByUserID(userID string) (*model.Lecturer, error) {
	query := `
		SELECT 
			l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
			u.id, u.username, u.full_name, u.email
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE l.user_id = $1`

	var l model.Lecturer
	l.User = &model.User{}

	err := r.db.QueryRow(query, userID).Scan(
		&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt,
		&l.User.ID, &l.User.Username, &l.User.FullName, &l.User.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("lecturer profile not found")
		}
		return nil, err
	}

	return &l, nil
}