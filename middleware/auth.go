package middleware

import (
	"uas/app/model"
	"uas/app/repository"
	"uas/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	roleRepo *repository.RoleRepository
}

// Constructor menerima RoleRepository (Sesuai wiring di main.go)
func NewAuthMiddleware(roleRepo *repository.RoleRepository) *AuthMiddleware {
	return &AuthMiddleware{roleRepo: roleRepo}
}

// ==============================================================
// Middleware 1: AuthRequired (FR-002 Step 1 & 2)
// Memastikan User mengirim token yang valid
// ==============================================================
func (m *AuthMiddleware) AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ekstrak JWT dari header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.WebResponse{
				Code:    401,
				Status:  "error",
				Message: "Missing authorization header",
			})
		}

		// Cek format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(model.WebResponse{
				Code:    401,
				Status:  "error",
				Message: "Invalid token format",
			})
		}

		// 2. Validasi token
		claims, err := utils.ParseToken(tokenParts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(model.WebResponse{
				Code:    401,
				Status:  "error",
				Message: "Invalid or expired token",
			})
		}

		// 3. Simpan data User ke Context (Locals)
		// Agar bisa diakses di Controller/Service (c.Locals("user_id"))
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions) // Permissions dimuat dari Token (Cache Strategy)

		return c.Next()
	}
}

// ==============================================================
// Middleware 2: PermissionRequired (FR-002 Step 4 & 5)
// Memastikan User memiliki Permission spesifik (RBAC)
// ==============================================================
func (m *AuthMiddleware) PermissionRequired(requiredPerm string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil permissions yang sudah disimpan di Locals oleh AuthRequired
		userPermsInterface := c.Locals("permissions")
		if userPermsInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(model.WebResponse{
				Code:    403,
				Status:  "error",
				Message: "No permissions found in context",
			})
		}

		// Casting ke []string
		// utils.JwtClaims mendefinisikan Permissions sebagai []string, jadi aman dicasting
		userPerms, ok := userPermsInterface.([]string)
		if !ok {
			// Fallback jika casting gagal (misal masalah decoding JSON internal)
			return c.Status(fiber.StatusInternalServerError).JSON(model.WebResponse{
				Code:    500,
				Status:  "error",
				Message: "Failed to parse user permissions",
			})
		}

		// 4. Check apakah user memiliki permission yang diperlukan
		hasPermission := false
		for _, p := range userPerms {
			if p == requiredPerm {
				hasPermission = true
				break
			}
		}

		// 5. Allow/deny request
		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(model.WebResponse{
				Code:    403,
				Status:  "error",
				Message: "Access denied. Missing permission: " + requiredPerm,
			})
		}

		return c.Next()
	}
}

// ==============================================================
// Middleware 3: RolesAllowed (Alternatif Simple RBAC Modul 5)
// Jika ingin mengecek Role langsung (misal: hanya "Dosen Wali")
// ==============================================================
func (m *AuthMiddleware) RolesAllowed(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role").(string)

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(model.WebResponse{
			Code:    403,
			Status:  "error",
			Message: "Access denied. Role not authorized.",
		})
	}
}