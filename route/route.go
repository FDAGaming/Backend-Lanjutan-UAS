package route

import (
	"uas/middleware"
	"uas/app/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authService *service.AuthService, achService *service.AchievementService) {
	api := app.Group("/api/v1")

	// --- AUTH ROUTES (Public) ---
	api.Post("/auth/login", authService.Login)

	// --- ACHIEVEMENT ROUTES (Protected) ---
	// Group middleware check token
	ach := api.Group("/achievements", middleware.AuthRequired())

	// GET: Admin, Dosen, Mahasiswa bisa akses (dengan filter logic di service)
	ach.Get("/", achService.GetAll)

	// POST: Hanya Mahasiswa
	ach.Post("/", middleware.RolesAllowed("Mahasiswa"), achService.Submit)

	// PUT Verify: Hanya Dosen Wali
	ach.Post("/:id/verify", middleware.RolesAllowed("Dosen Wali"), achService.Verify)
}