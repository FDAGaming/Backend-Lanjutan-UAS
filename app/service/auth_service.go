package service

import (
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// =================================================================
// 5.1 AUTHENTICATION
// =================================================================

// POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid request body"})
	}

	// 1. Cari User
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// 2. Cek Password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// 3. Cek Status Aktif
	if !user.IsActive {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "User account is inactive"})
	}

	// 4. Ambil Permissions dari Role
	permsData, err := s.roleRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to load permissions"})
	}

	var permissions []string
	for _, p := range permsData {
		permissions = append(permissions, p.Name)
	}

	// 5. Generate Token
	token, err := utils.GenerateToken(user.ID, user.Role.Name, permissions)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate token"})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Login successful",
		Data: fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        user.Role.Name,
				"permissions": permissions,
			},
		},
	})
}

// POST /api/v1/auth/refresh
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Refresh Token Not Implemented"})
}

// POST /api/v1/auth/logout
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// Karena menggunakan JWT (Stateless), logout cukup dilakukan di client side (hapus token).
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Logged out successfully"})
}

// GET /api/v1/auth/profile
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "User not found"})
	}

	user.PasswordHash = "" // Hide sensitive data

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Profile retrieved",
		Data:    user,
	})
}

// =================================================================
// 5.2 USERS MANAGEMENT (ADMIN)
// =================================================================

// GET /api/v1/users
func (s *AuthService) GetAllUsers(c *fiber.Ctx) error {
	// Panggil Repository FindAll (Native SQL)
	users, err := s.userRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "All users retrieved successfully",
		Data:    users,
	})
}

// GET /api/v1/users/:id
func (s *AuthService) GetUserDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Panggil Repository FindByID
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "User not found"})
	}

	user.PasswordHash = "" // Hide sensitive data
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "User detail retrieved",
		Data:    user,
	})
}

// POST /api/v1/users
func (s *AuthService) CreateUser(c *fiber.Ctx) error {
	// Input DTO khusus untuk create user agar lebih rapi
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
		RoleID   string `json:"roleId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input data"})
	}

	// Hash Password
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to hash password"})
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true, // Default active
	}

	// Panggil Repository Create
	if err := s.userRepo.Create(&user); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	user.PasswordHash = "" // Jangan kembalikan hash ke response
	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "User created successfully",
		Data:    user,
	})
}

// PUT /api/v1/users/:id
func (s *AuthService) UpdateUser(c *fiber.Ctx) error {
	// TODO: Tambahkan method Update di UserRepository untuk mengimplementasikan ini
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update User Not Implemented (Repo missing Update method)"})
}

// DELETE /api/v1/users/:id
func (s *AuthService) DeleteUser(c *fiber.Ctx) error {
	// TODO: Tambahkan method Delete di UserRepository untuk mengimplementasikan ini
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Delete User Not Implemented (Repo missing Delete method)"})
}

// PUT /api/v1/users/:id/role
func (s *AuthService) UpdateUserRole(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Role Not Implemented"})
}

// =================================================================
// 5.5 STUDENTS & LECTURERS (Placeholders)
// =================================================================

func (s *AuthService) GetAllStudents(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Get All Students Not Implemented"})
}

func (s *AuthService) GetStudentDetail(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Get Student Detail Not Implemented"})
}

func (s *AuthService) UpdateStudentAdvisor(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Advisor Not Implemented"})
}

func (s *AuthService) GetAllLecturers(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Get All Lecturers Not Implemented"})
}