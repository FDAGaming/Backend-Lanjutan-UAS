package service

import (
	// "errors"
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository // Perlu repo role untuk ambil permissions
}

func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// FR-001: Login
func (s *AuthService) Login(c *fiber.Ctx) error {
	// 1. User mengirim kredensial
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid request body"})
	}

	// 2. Sistem memvalidasi kredensial (Cari user by Email)
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// Cek Password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// 3. Sistem mengecek status aktif user
	if !user.IsActive {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "User account is inactive"})
	}

	// 4. Sistem generate JWT token dengan role dan permissions
	// Ambil permissions dari database berdasarkan RoleID user
	permsData, err := s.roleRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to load permissions"})
	}

	// Convert struct permission ke slice string (misal: ["achievement:create", "user:read"])
	var permissions []string
	for _, p := range permsData {
		permissions = append(permissions, p.Name)
	}

	// Generate Token
	token, err := utils.GenerateToken(user.ID, user.Role.Name, permissions)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate token"})
	}

	// 5. Return token dan user profile
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

// --- Placeholder Auth ---
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "Refresh Token Not Implemented"})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"message": "Logout Not Implemented"})
}

func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	// Logic ambil profile user dari Repo
	return c.Status(200).JSON(fiber.Map{"message": "Profile Data", "userId": userID})
}

// --- Placeholder User Management (Admin) ---
func (s *AuthService) GetAllUsers(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) GetUserDetail(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) CreateUser(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) UpdateUser(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) DeleteUser(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) UpdateUserRole(c *fiber.Ctx) error { return notImplemented(c) }

// --- Placeholder Students & Lecturers ---
func (s *AuthService) GetAllStudents(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) GetStudentDetail(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) UpdateStudentAdvisor(c *fiber.Ctx) error { return notImplemented(c) }
func (s *AuthService) GetAllLecturers(c *fiber.Ctx) error { return notImplemented(c) }

// Helper Internal
func notImplemented(c *fiber.Ctx) error {
	return c.Status(501).JSON(fiber.Map{"status": "error", "message": "Feature not implemented yet"})
}