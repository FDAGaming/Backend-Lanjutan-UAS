package service

import (
	"uas/app/repository"
	"uas/utils" // Asumsi ada util untuk JWT & Bcrypt
	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// FR-001: Login Handler
func (s *AuthService) Login(c *fiber.Ctx) error {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// 1. Cari User
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 2. Cek Password (Gunakan Utils Bcrypt)
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 3. Generate Token (Gunakan Utils JWT)
	token, err := utils.GenerateToken(user.ID, user.Role.Name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  fiber.Map{"username": user.Username, "role": user.Role.Name},
	})
}