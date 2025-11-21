package middleware

import (
	"uas/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Modul 5: AuthRequired Middleware
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token akses diperlukan",
			})
		}

		// 2. Cek format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Format token tidak valid",
			})
		}

		// 3. Validasi Token
		claims, err := utils.ParseToken(tokenParts[1]) // Fungsi ini ada di utils/jwt.go (jawaban sebelumnya)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token tidak valid atau expired",
			})
		}

        // Casting claims ke Map/Struct yang sesuai
        // Pastikan utils.ParseToken mengembalikan *utils.JwtClaims
        // Disini diasumsikan Anda menyesuaikan utils/jwt.go agar me-return claims struct

		// 4. Simpan data user ke Context (Locals)
		c.Locals("user_id", claims.UserID) // ID dari tabel Users
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// Modul 5: Role Based Access Control (RBAC)
// Menerima variadic parameter (bisa banyak role)
func RolesAllowed(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role").(string)

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Akses ditolak. Anda tidak memiliki izin.",
		})
	}
}